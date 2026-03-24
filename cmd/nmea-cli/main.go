package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/keitaj/go-nmea/pkg/nmea"
)

func main() {
	var (
		inputFile  = flag.String("f", "", "Input NMEA log file (default: stdin)")
		filterType = flag.String("type", "", "Filter by sentence type (GGA, RMC, GSA, GSV)")
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

	reader.OnParsed(func(sentence interface{}, raw string) {
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

		case *nmea.RMC:
			stats.rmc++
			if *filterType != "" && *filterType != "RMC" {
				return
			}
			fmt.Printf("[%s] RMC | %s | %s | Lat:%.6f Lon:%.6f | Spd:%.1fkn | Crs:%.1f°\n",
				s.Talker, s.Time, s.Date, s.Latitude, s.Longitude, s.Speed, s.Course)

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

		default:
			stats.other++
			if *filterType == "" {
				if base, ok := sentence.(nmea.Sentence); ok {
					fmt.Printf("[%s] %s | (unsupported type)\n", base.Talker, base.Type)
				}
			}
		}
	})

	if *showErrors {
		reader.OnError(func(raw string, err error) {
			stats.errors++
			fmt.Fprintf(os.Stderr, "ERROR: %v\n  -> %s\n", err, raw)
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
	fmt.Fprintf(os.Stderr, "Total: %d | GGA: %d | RMC: %d | GSA: %d | GSV: %d | Other: %d | Errors: %d\n",
		stats.total, stats.gga, stats.rmc, stats.gsa, stats.gsv, stats.other, stats.errors)
}

type parseStats struct {
	total, gga, rmc, gsa, gsv, other, errors int
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
