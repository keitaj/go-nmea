package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/keitaj/go-nmea/pkg/nmea"
)

func main() {
	var (
		inputFile  = flag.String("f", "", "Input NMEA log file (default: stdin)")
		filterType = flag.String("type", "", "Filter by sentence type (GGA, GLL, RMC, VTG, GSA, GSV, ZDA, GBS, GST)")
		verbose    = flag.Bool("v", false, "Verbose output: show all fields")
		showErrors = flag.Bool("errors", false, "Show parse errors")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "gnss-nmea-cli - GNSS NMEA sentence parser\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  cat data.nmea | gnss-nmea-cli\n")
		fmt.Fprintf(os.Stderr, "  gnss-nmea-cli -f data.nmea -type GGA\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	var input *os.File
	if *inputFile != "" {
		f, err := os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		input = f
	} else {
		input = os.Stdin
	}

	reader := nmea.NewStreamReader(input)
	stats := &parseStats{}

	reader.OnParsed(func(sentence nmea.Sentence, raw string) {
		stats.total++

		switch s := sentence.(type) {
		case *nmea.GGA:
			stats.gga++
			if *filterType != "" && *filterType != "GGA" {
				return
			}
			if *verbose {
				printGGAVerbose(s)
			} else {
				fmt.Printf("[%s] GGA | %s | Lat:%.6f Lon:%.6f | Fix:%s | Sats:%d | HDOP:%.2f | Alt:%.1fm\n",
					s.Talker, s.Time, s.Latitude, s.Longitude,
					s.Quality, s.NumSatellites, s.HDOP, s.Altitude)
			}

		case *nmea.GLL:
			stats.gll++
			if *filterType != "" && *filterType != "GLL" {
				return
			}
			fmt.Printf("[%s] GLL | %s | Lat:%.6f Lon:%.6f | Status:%s\n",
				s.Talker, s.Time, s.Latitude, s.Longitude, s.Status)

		case *nmea.RMC:
			stats.rmc++
			if *filterType != "" && *filterType != "RMC" {
				return
			}
			fmt.Printf("[%s] RMC | %s | %s | Lat:%.6f Lon:%.6f | Spd:%.1fkn | Crs:%.1f°\n",
				s.Talker, s.Time, s.Date, s.Latitude, s.Longitude, s.Speed, s.Course)

		case *nmea.VTG:
			stats.vtg++
			if *filterType != "" && *filterType != "VTG" {
				return
			}
			fmt.Printf("[%s] VTG | CrsT:%.1f° CrsM:%.1f° | Spd:%.1fkn %.1fkm/h\n",
				s.Talker, s.CourseTrue, s.CourseMag, s.SpeedKnots, s.SpeedKmh)

		case *nmea.GSA:
			stats.gsa++
			if *filterType != "" && *filterType != "GSA" {
				return
			}
			fixStr := map[int]string{1: "No Fix", 2: "2D", 3: "3D"}[s.FixType]
			fmt.Printf("[%s] GSA | Fix:%s | SVs:%v | PDOP:%.2f HDOP:%.2f VDOP:%.2f\n",
				s.Talker, fixStr, s.SVIDs, s.PDOP, s.HDOP, s.VDOP)

		case *nmea.GSV:
			stats.gsv++
			if *filterType != "" && *filterType != "GSV" {
				return
			}
			fmt.Printf("[%s] GSV | Msg %d/%d | Total Sats:%d | ",
				s.Talker, s.MsgNum, s.TotalMsgs, s.TotalSats)
			for i, sat := range s.Satellites {
				if i > 0 {
					fmt.Print(" ")
				}
				if sat.SNR >= 0 {
					fmt.Printf("SV%02d(El:%d°,Az:%d°,SNR:%ddB)", sat.SVID, sat.Elevation, sat.Azimuth, sat.SNR)
				} else {
					fmt.Printf("SV%02d(El:%d°,Az:%d°,--)", sat.SVID, sat.Elevation, sat.Azimuth)
				}
			}
			fmt.Println()

		case *nmea.ZDA:
			stats.zda++
			if *filterType != "" && *filterType != "ZDA" {
				return
			}
			fmt.Printf("[%s] ZDA | %s | %04d-%02d-%02d | TZ:%+03d:%02d\n",
				s.Talker, s.Time, s.Year, s.Month, s.Day, s.LocalHours, s.LocalMins)

		case *nmea.GBS:
			stats.gbs++
			if *filterType != "" && *filterType != "GBS" {
				return
			}
			fmt.Printf("[%s] GBS | %s | ErrLat:%.2fm ErrLon:%.2fm ErrAlt:%.2fm | FailSV:%d Bias:%.2fm\n",
				s.Talker, s.Time, s.ErrLat, s.ErrLon, s.ErrAlt, s.SVID, s.Bias)

		case *nmea.GST:
			stats.gst++
			if *filterType != "" && *filterType != "GST" {
				return
			}
			fmt.Printf("[%s] GST | %s | RMS:%.2fm | σLat:%.2fm σLon:%.2fm σAlt:%.2fm\n",
				s.Talker, s.Time, s.RangeRMS, s.StdLat, s.StdLon, s.StdAlt)

		default:
			stats.other++
			if *filterType == "" {
				fmt.Printf("[%s] %s | (unsupported type)\n", sentence.GetTalker(), sentence.GetType())
			}
		}
	})

	if *showErrors {
		reader.OnError(func(raw string, err error) {
			stats.errors++
			var pe *nmea.ParseError
			if errors.As(err, &pe) {
				switch pe.Kind {
				case nmea.ErrChecksum:
					fmt.Fprintf(os.Stderr, "CHECKSUM: %v\n  -> %s\n", err, raw)
				case nmea.ErrFieldCount:
					fmt.Fprintf(os.Stderr, "FIELDS: %v\n  -> %s\n", err, raw)
				default:
					fmt.Fprintf(os.Stderr, "ERROR: %v\n  -> %s\n", err, raw)
				}
			} else {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n  -> %s\n", err, raw)
			}
		})
	} else {
		reader.OnError(func(_ string, _ error) {
			stats.errors++
		})
	}

	_, err := reader.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "\n--- Summary ---\n")
	fmt.Fprintf(os.Stderr, "Total: %d | GGA: %d | GLL: %d | RMC: %d | VTG: %d | GSA: %d | GSV: %d | ZDA: %d | GBS: %d | GST: %d | Other: %d | Errors: %d\n",
		stats.total, stats.gga, stats.gll, stats.rmc, stats.vtg, stats.gsa, stats.gsv, stats.zda, stats.gbs, stats.gst, stats.other, stats.errors)
}

type parseStats struct {
	total, gga, gll, rmc, vtg, gsa, gsv, zda, gbs, gst, other, errors int
}

func printGGAVerbose(g *nmea.GGA) {
	fmt.Println("━━━ GGA Fix Data ━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Talker:      %s\n", g.Talker)
	fmt.Printf("  Time (UTC):  %s\n", g.Time)
	fmt.Printf("  Position:    %.8f, %.8f\n", g.Latitude, g.Longitude)
	fmt.Printf("  Fix Quality: %s\n", g.Quality)
	fmt.Printf("  Satellites:  %d\n", g.NumSatellites)
	fmt.Printf("  HDOP:        %.2f\n", g.HDOP)
	fmt.Printf("  Altitude:    %.2f m\n", g.Altitude)
	fmt.Printf("  Geoid Sep:   %.2f m\n", g.GeoidSep)
	if g.DGPSAge > 0 {
		fmt.Printf("  DGPS Age:    %.1f s\n", g.DGPSAge)
		fmt.Printf("  DGPS Stn:    %s\n", g.DGPSStationID)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
