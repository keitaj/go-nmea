package nmea

import (
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
