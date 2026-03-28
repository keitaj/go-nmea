package nmea

import (
	"errors"
	"math"
	"testing"
)

const tolerance = 0.0001

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestParseGGA(t *testing.T) {
	raw := "$GPGGA,092725.00,3539.3010,N,13941.2820,E,1,08,1.03,22.5,M,39.5,M,,*6F"

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gga, ok := result.(*GGA)
	if !ok {
		t.Fatalf("expected *GGA, got %T", result)
	}

	if gga.Talker != TalkerGP {
		t.Errorf("talker: got %q, want %q", gga.Talker, TalkerGP)
	}
	if gga.Time != "092725.00" {
		t.Errorf("time: got %q, want %q", gga.Time, "092725.00")
	}
	// 3539.3010 N = 35 + 39.3010/60 = 35.65502
	if !almostEqual(gga.Latitude, 35.6550166) {
		t.Errorf("latitude: got %f, want ~35.6550", gga.Latitude)
	}
	// 13941.2820 E = 139 + 41.2820/60 = 139.68803
	if !almostEqual(gga.Longitude, 139.68803) {
		t.Errorf("longitude: got %f, want ~139.6880", gga.Longitude)
	}
	if gga.Quality != FixGPS {
		t.Errorf("quality: got %v, want %v", gga.Quality, FixGPS)
	}
	if gga.NumSatellites != 8 {
		t.Errorf("satellites: got %d, want 8", gga.NumSatellites)
	}
	if !almostEqual(gga.HDOP, 1.03) {
		t.Errorf("HDOP: got %f, want 1.03", gga.HDOP)
	}
	if !almostEqual(gga.Altitude, 22.5) {
		t.Errorf("altitude: got %f, want 22.5", gga.Altitude)
	}
}

func TestParseGGA_RTKFixed(t *testing.T) {
	raw := "$GNGGA,081234.00,3526.1234,N,13945.6789,E,4,12,0.65,35.2,M,39.5,M,1.2,0001*50"

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gga, ok := result.(*GGA)
	if !ok {
		t.Fatalf("expected *GGA, got %T", result)
	}

	if gga.Talker != TalkerGN {
		t.Errorf("talker: got %q, want %q", gga.Talker, TalkerGN)
	}
	if gga.Quality != FixRTKFixed {
		t.Errorf("quality: got %v, want RTK Fixed", gga.Quality)
	}
	if gga.NumSatellites != 12 {
		t.Errorf("satellites: got %d, want 12", gga.NumSatellites)
	}
}

func TestParseRMC(t *testing.T) {
	raw := "$GPRMC,092725.00,A,3539.3010,N,13941.2820,E,0.05,220.5,240326,,,A*6C"

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rmc, ok := result.(*RMC)
	if !ok {
		t.Fatalf("expected *RMC, got %T", result)
	}

	if rmc.Status != "A" {
		t.Errorf("status: got %q, want %q", rmc.Status, "A")
	}
	if rmc.Date != "240326" {
		t.Errorf("date: got %q, want %q", rmc.Date, "240326")
	}
	if !almostEqual(rmc.Speed, 0.05) {
		t.Errorf("speed: got %f, want 0.05", rmc.Speed)
	}
	if !almostEqual(rmc.Course, 220.5) {
		t.Errorf("course: got %f, want 220.5", rmc.Course)
	}
	if rmc.Mode != "A" {
		t.Errorf("mode: got %q, want %q", rmc.Mode, "A")
	}
}

func TestParseGSA(t *testing.T) {
	raw := "$GPGSA,A,3,01,03,06,09,12,17,19,22,25,28,,,1.50,0.90,1.20*01"

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gsa, ok := result.(*GSA)
	if !ok {
		t.Fatalf("expected *GSA, got %T", result)
	}

	if gsa.Mode != "A" {
		t.Errorf("mode: got %q, want %q", gsa.Mode, "A")
	}
	if gsa.FixType != 3 {
		t.Errorf("fix type: got %d, want 3", gsa.FixType)
	}
	if len(gsa.SVIDs) != 10 {
		t.Errorf("SVIDs count: got %d, want 10", len(gsa.SVIDs))
	}
	if !almostEqual(gsa.PDOP, 1.50) {
		t.Errorf("PDOP: got %f, want 1.50", gsa.PDOP)
	}
}

func TestParseGSV(t *testing.T) {
	raw := "$GPGSV,3,1,12,01,40,083,42,03,55,220,47,06,25,315,35,09,72,010,44*7D"

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gsv, ok := result.(*GSV)
	if !ok {
		t.Fatalf("expected *GSV, got %T", result)
	}

	if gsv.TotalMsgs != 3 {
		t.Errorf("total msgs: got %d, want 3", gsv.TotalMsgs)
	}
	if gsv.TotalSats != 12 {
		t.Errorf("total sats: got %d, want 12", gsv.TotalSats)
	}
	if len(gsv.Satellites) != 4 {
		t.Fatalf("satellites: got %d, want 4", len(gsv.Satellites))
	}

	sat := gsv.Satellites[0]
	if sat.SVID != 1 {
		t.Errorf("sat[0] SVID: got %d, want 1", sat.SVID)
	}
	if sat.Elevation != 40 {
		t.Errorf("sat[0] elevation: got %d, want 40", sat.Elevation)
	}
	if sat.SNR != 42 {
		t.Errorf("sat[0] SNR: got %d, want 42", sat.SNR)
	}
}

func TestParseQZSS(t *testing.T) {
	// QZSS (Michibiki) GGA sentence
	raw := "$QZGGA,092725.00,3539.3010,N,13941.2820,E,1,04,1.50,22.5,M,39.5,M,,*79"

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gga, ok := result.(*GGA)
	if !ok {
		t.Fatalf("expected *GGA, got %T", result)
	}

	if gga.Talker != TalkerQZ {
		t.Errorf("talker: got %q, want %q", gga.Talker, TalkerQZ)
	}
}

func TestParseGLL(t *testing.T) {
	raw := "$GPGLL,3539.3010,N,13941.2820,E,092725.00,A,A*6A"

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gll, ok := result.(*GLL)
	if !ok {
		t.Fatalf("expected *GLL, got %T", result)
	}

	if gll.Talker != TalkerGP {
		t.Errorf("talker: got %q, want %q", gll.Talker, TalkerGP)
	}
	if gll.Time != "092725.00" {
		t.Errorf("time: got %q, want %q", gll.Time, "092725.00")
	}
	if !almostEqual(gll.Latitude, 35.6550166) {
		t.Errorf("latitude: got %f, want ~35.6550", gll.Latitude)
	}
	if !almostEqual(gll.Longitude, 139.68803) {
		t.Errorf("longitude: got %f, want ~139.6880", gll.Longitude)
	}
	if gll.Status != "A" {
		t.Errorf("status: got %q, want %q", gll.Status, "A")
	}
	if gll.Mode != "A" {
		t.Errorf("mode: got %q, want %q", gll.Mode, "A")
	}
}

func TestParseVTG(t *testing.T) {
	raw := "$GPVTG,220.5,T,218.3,M,0.05,N,0.09,K,A*22"

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	vtg, ok := result.(*VTG)
	if !ok {
		t.Fatalf("expected *VTG, got %T", result)
	}

	if !almostEqual(vtg.CourseTrue, 220.5) {
		t.Errorf("course true: got %f, want 220.5", vtg.CourseTrue)
	}
	if !almostEqual(vtg.CourseMag, 218.3) {
		t.Errorf("course mag: got %f, want 218.3", vtg.CourseMag)
	}
	if !almostEqual(vtg.SpeedKnots, 0.05) {
		t.Errorf("speed knots: got %f, want 0.05", vtg.SpeedKnots)
	}
	if !almostEqual(vtg.SpeedKmh, 0.09) {
		t.Errorf("speed km/h: got %f, want 0.09", vtg.SpeedKmh)
	}
	if vtg.Mode != "A" {
		t.Errorf("mode: got %q, want %q", vtg.Mode, "A")
	}
}

func TestParseZDA(t *testing.T) {
	raw := "$GPZDA,092725.00,24,03,2026,09,00*67"

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	zda, ok := result.(*ZDA)
	if !ok {
		t.Fatalf("expected *ZDA, got %T", result)
	}

	if zda.Time != "092725.00" {
		t.Errorf("time: got %q, want %q", zda.Time, "092725.00")
	}
	if zda.Day != 24 {
		t.Errorf("day: got %d, want 24", zda.Day)
	}
	if zda.Month != 3 {
		t.Errorf("month: got %d, want 3", zda.Month)
	}
	if zda.Year != 2026 {
		t.Errorf("year: got %d, want 2026", zda.Year)
	}
	if zda.LocalHours != 9 {
		t.Errorf("local hours: got %d, want 9", zda.LocalHours)
	}
	if zda.LocalMins != 0 {
		t.Errorf("local mins: got %d, want 0", zda.LocalMins)
	}
}

func TestParseGBS(t *testing.T) {
	raw := "$GPGBS,092725.00,1.5,2.0,3.5,17,0.03,1.2,0.8*5A"

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gbs, ok := result.(*GBS)
	if !ok {
		t.Fatalf("expected *GBS, got %T", result)
	}

	if gbs.Time != "092725.00" {
		t.Errorf("time: got %q, want %q", gbs.Time, "092725.00")
	}
	if !almostEqual(gbs.ErrLat, 1.5) {
		t.Errorf("ErrLat: got %f, want 1.5", gbs.ErrLat)
	}
	if !almostEqual(gbs.ErrLon, 2.0) {
		t.Errorf("ErrLon: got %f, want 2.0", gbs.ErrLon)
	}
	if !almostEqual(gbs.ErrAlt, 3.5) {
		t.Errorf("ErrAlt: got %f, want 3.5", gbs.ErrAlt)
	}
	if gbs.SVID != 17 {
		t.Errorf("SVID: got %d, want 17", gbs.SVID)
	}
	if !almostEqual(gbs.Prob, 0.03) {
		t.Errorf("Prob: got %f, want 0.03", gbs.Prob)
	}
	if !almostEqual(gbs.Bias, 1.2) {
		t.Errorf("Bias: got %f, want 1.2", gbs.Bias)
	}
	if !almostEqual(gbs.StdDev, 0.8) {
		t.Errorf("StdDev: got %f, want 0.8", gbs.StdDev)
	}
}

func TestParseGST(t *testing.T) {
	raw := "$GPGST,092725.00,1.8,3.2,2.1,45.0,1.5,1.2,2.8*6B"

	result, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gst, ok := result.(*GST)
	if !ok {
		t.Fatalf("expected *GST, got %T", result)
	}

	if gst.Time != "092725.00" {
		t.Errorf("time: got %q, want %q", gst.Time, "092725.00")
	}
	if !almostEqual(gst.RangeRMS, 1.8) {
		t.Errorf("RangeRMS: got %f, want 1.8", gst.RangeRMS)
	}
	if !almostEqual(gst.StdMajor, 3.2) {
		t.Errorf("StdMajor: got %f, want 3.2", gst.StdMajor)
	}
	if !almostEqual(gst.StdMinor, 2.1) {
		t.Errorf("StdMinor: got %f, want 2.1", gst.StdMinor)
	}
	if !almostEqual(gst.Orient, 45.0) {
		t.Errorf("Orient: got %f, want 45.0", gst.Orient)
	}
	if !almostEqual(gst.StdLat, 1.5) {
		t.Errorf("StdLat: got %f, want 1.5", gst.StdLat)
	}
	if !almostEqual(gst.StdLon, 1.2) {
		t.Errorf("StdLon: got %f, want 1.2", gst.StdLon)
	}
	if !almostEqual(gst.StdAlt, 2.8) {
		t.Errorf("StdAlt: got %f, want 2.8", gst.StdAlt)
	}
}

func TestChecksumValidation(t *testing.T) {
	// Corrupt checksum
	raw := "$GPGGA,092725.00,3539.3010,N,13941.2820,E,1,08,1.03,22.5,M,39.5,M,,*FF"

	_, err := Parse(raw)
	if err == nil {
		t.Error("expected checksum error, got nil")
	}
}

func TestInvalidSentences(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"empty", ""},
		{"no prefix", "GPGGA,1,2,3*00"},
		{"no checksum", "$GPGGA,1,2,3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.raw)
			if err == nil {
				t.Errorf("expected error for %q", tt.raw)
			}
		})
	}
}

func TestFixQualityString(t *testing.T) {
	tests := []struct {
		quality FixQuality
		want    string
	}{
		{FixInvalid, "Invalid"},
		{FixRTKFixed, "RTK Fixed"},
		{FixRTKFloat, "RTK Float"},
		{FixQuality(99), "Unknown(99)"},
	}

	for _, tt := range tests {
		if got := tt.quality.String(); got != tt.want {
			t.Errorf("FixQuality(%d).String() = %q, want %q", tt.quality, got, tt.want)
		}
	}
}

// --- Interface tests ---

func TestSentenceInterface(t *testing.T) {
	raw := "$GPGGA,092725.00,3539.3010,N,13941.2820,E,1,08,1.03,22.5,M,39.5,M,,*6F"
	s, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.GetTalker() != TalkerGP {
		t.Errorf("GetTalker: got %q, want %q", s.GetTalker(), TalkerGP)
	}
	if s.GetType() != "GGA" {
		t.Errorf("GetType: got %q, want %q", s.GetType(), "GGA")
	}
	if s.GetRaw() != raw {
		t.Errorf("GetRaw: got %q, want %q", s.GetRaw(), raw)
	}
}

func TestHasPosition(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantLat float64
		wantLon float64
	}{
		{"GGA", "$GPGGA,092725.00,3539.3010,N,13941.2820,E,1,08,1.03,22.5,M,39.5,M,,*6F", 35.6550166, 139.68803},
		{"GLL", "$GPGLL,3539.3010,N,13941.2820,E,092725.00,A,A*6A", 35.6550166, 139.68803},
		{"RMC", "$GPRMC,092725.00,A,3539.3010,N,13941.2820,E,0.05,220.5,240326,,,A*6C", 35.6550166, 139.68803},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Parse(tt.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			pos, ok := s.(HasPosition)
			if !ok {
				t.Fatalf("%s should implement HasPosition", tt.name)
			}
			lat, lon := pos.GetPosition()
			if !almostEqual(lat, tt.wantLat) {
				t.Errorf("lat: got %f, want %f", lat, tt.wantLat)
			}
			if !almostEqual(lon, tt.wantLon) {
				t.Errorf("lon: got %f, want %f", lon, tt.wantLon)
			}
		})
	}
}

func TestHasPositionNotImplemented(t *testing.T) {
	// GSA should NOT implement HasPosition
	raw := "$GPGSA,A,3,01,03,06,09,12,17,19,22,25,28,,,1.50,0.90,1.20*01"
	s, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.(HasPosition); ok {
		t.Error("GSA should not implement HasPosition")
	}
}

func TestHasTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantTime string
	}{
		{"GGA", "$GPGGA,092725.00,3539.3010,N,13941.2820,E,1,08,1.03,22.5,M,39.5,M,,*6F", "092725.00"},
		{"GLL", "$GPGLL,3539.3010,N,13941.2820,E,092725.00,A,A*6A", "092725.00"},
		{"RMC", "$GPRMC,092725.00,A,3539.3010,N,13941.2820,E,0.05,220.5,240326,,,A*6C", "092725.00"},
		{"ZDA", "$GPZDA,092725.00,24,03,2026,09,00*67", "092725.00"},
		{"GBS", "$GPGBS,092725.00,1.5,2.0,3.5,17,0.03,1.2,0.8*5A", "092725.00"},
		{"GST", "$GPGST,092725.00,1.8,3.2,2.1,45.0,1.5,1.2,2.8*6B", "092725.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Parse(tt.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			ts, ok := s.(HasTimestamp)
			if !ok {
				t.Fatalf("%s should implement HasTimestamp", tt.name)
			}
			if ts.GetTimestamp() != tt.wantTime {
				t.Errorf("GetTimestamp: got %q, want %q", ts.GetTimestamp(), tt.wantTime)
			}
		})
	}
}

func TestHasSpeed(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantSpeed float64
	}{
		{"RMC", "$GPRMC,092725.00,A,3539.3010,N,13941.2820,E,0.05,220.5,240326,,,A*6C", 0.05},
		{"VTG", "$GPVTG,220.5,T,218.3,M,0.05,N,0.09,K,A*22", 0.05},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Parse(tt.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			spd, ok := s.(HasSpeed)
			if !ok {
				t.Fatalf("%s should implement HasSpeed", tt.name)
			}
			if !almostEqual(spd.GetSpeedKnots(), tt.wantSpeed) {
				t.Errorf("GetSpeedKnots: got %f, want %f", spd.GetSpeedKnots(), tt.wantSpeed)
			}
		})
	}
}

// --- Structured error tests ---

func TestChecksumParseError(t *testing.T) {
	raw := "$GPGGA,092725.00,3539.3010,N,13941.2820,E,1,08,1.03,22.5,M,39.5,M,,*FF"
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrChecksumMismatch) {
		t.Errorf("expected ErrChecksumMismatch, got %v", err)
	}

	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatal("expected *ParseError")
	}
	if pe.Kind != ErrChecksum {
		t.Errorf("expected ErrChecksum kind, got %d", pe.Kind)
	}
	if pe.Sentence != raw {
		t.Errorf("expected sentence %q, got %q", raw, pe.Sentence)
	}
}

func TestInvalidFormatParseError(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidSentence) {
		t.Errorf("expected ErrInvalidSentence, got %v", err)
	}

	_, err = Parse("GPGGA,1,2,3*00")
	if !errors.Is(err, ErrInvalidSentence) {
		t.Errorf("expected ErrInvalidSentence for no prefix, got %v", err)
	}
}

func TestUnknownSentenceReturnsBaseSentence(t *testing.T) {
	// Unknown sentence type should not error, returns *BaseSentence
	raw := "$GPXYZ,1,2,3*50"
	s, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.GetType() != "XYZ" {
		t.Errorf("expected type XYZ, got %q", s.GetType())
	}
	if _, ok := s.(*BaseSentence); !ok {
		t.Errorf("expected *BaseSentence for unknown type, got %T", s)
	}
}

// --- Custom parser registration tests ---

// HDT is a custom sentence type for testing: Heading True.
type HDT struct {
	BaseSentence
	Heading float64
	True    string
}

func TestRegisterParser(t *testing.T) {
	RegisterParser("HDT", func(s BaseSentence) (Sentence, error) {
		if len(s.Fields) < 2 {
			return nil, newParseError(ErrFieldCount, s.Raw, "HDT requires at least 2 fields, got %d", len(s.Fields))
		}
		return &HDT{
			BaseSentence: s,
			Heading:      ParseFloat(s.Fields[0]),
			True:         s.Fields[1],
		}, nil
	})
	defer UnregisterParser("HDT")

	// $GPHDT,274.07,T*03
	raw := "$GPHDT,274.07,T*03"
	s, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hdt, ok := s.(*HDT)
	if !ok {
		t.Fatalf("expected *HDT, got %T", s)
	}
	if !almostEqual(hdt.Heading, 274.07) {
		t.Errorf("heading: got %f, want 274.07", hdt.Heading)
	}
	if hdt.True != "T" {
		t.Errorf("true: got %q, want %q", hdt.True, "T")
	}
	if hdt.GetTalker() != TalkerGP {
		t.Errorf("talker: got %q, want %q", hdt.GetTalker(), TalkerGP)
	}
}

func TestUnregisterParser(t *testing.T) {
	RegisterParser("FOO", func(s BaseSentence) (Sentence, error) {
		return &s, nil
	})
	UnregisterParser("FOO")

	// After unregister, should fall back to BaseSentence
	raw := "$GPFOO,1,2*52"
	s, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.(*BaseSentence); !ok {
		t.Errorf("expected *BaseSentence after unregister, got %T", s)
	}
}

func TestCustomParserOverridesBuiltIn(t *testing.T) {
	// Override the built-in GGA parser
	called := false
	RegisterParser("GGA", func(s BaseSentence) (Sentence, error) {
		called = true
		return parseGGA(s)
	})
	defer UnregisterParser("GGA")

	raw := "$GPGGA,092725.00,3539.3010,N,13941.2820,E,1,08,1.03,22.5,M,39.5,M,,*6F"
	_, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("custom parser should have been called instead of built-in")
	}
}

func TestCustomParserError(t *testing.T) {
	RegisterParser("ERR", func(s BaseSentence) (Sentence, error) {
		return nil, newParseError(ErrFieldCount, s.Raw, "custom error")
	})
	defer UnregisterParser("ERR")

	raw := "$GPERR,1*4F"
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error from custom parser")
	}
	if !errors.Is(err, ErrInsufficientFields) {
		t.Errorf("expected ErrInsufficientFields, got %v", err)
	}
}
