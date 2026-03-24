// Package nmea provides a parser for NMEA 0183 sentences commonly used in GNSS receivers.
//
// Supported sentence types:
//   - GGA: Global Positioning System Fix Data (position, quality, altitude)
//   - RMC: Recommended Minimum Navigation Data (position, velocity, date)
//   - GSA: DOP and Active Satellites
//   - GSV: Satellites in View (signal strength, elevation, azimuth)
package nmea

import (
	"fmt"
	"strconv"
	"strings"
)

// TalkerID represents the GNSS constellation identifier.
type TalkerID string

const (
	TalkerGP TalkerID = "GP" // GPS
	TalkerGL TalkerID = "GL" // GLONASS
	TalkerGA TalkerID = "GA" // Galileo
	TalkerGB TalkerID = "GB" // BeiDou
	TalkerGN TalkerID = "GN" // Multi-constellation
	TalkerQZ TalkerID = "QZ" // QZSS (Michibiki)
)

// FixQuality represents the GPS fix quality indicator in GGA sentences.
type FixQuality int

const (
	FixInvalid    FixQuality = 0
	FixGPS        FixQuality = 1
	FixDGPS       FixQuality = 2
	FixPPS        FixQuality = 3
	FixRTKFixed   FixQuality = 4
	FixRTKFloat   FixQuality = 5
	FixEstimated  FixQuality = 6
	FixManual     FixQuality = 7
	FixSimulation FixQuality = 8
)

func (f FixQuality) String() string {
	names := map[FixQuality]string{
		FixInvalid:    "Invalid",
		FixGPS:        "GPS",
		FixDGPS:       "DGPS",
		FixPPS:        "PPS",
		FixRTKFixed:   "RTK Fixed",
		FixRTKFloat:   "RTK Float",
		FixEstimated:  "Estimated",
		FixManual:     "Manual",
		FixSimulation: "Simulation",
	}
	if name, ok := names[f]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", int(f))
}

// Sentence is the base structure for all NMEA sentences.
type Sentence struct {
	Talker   TalkerID // e.g., "GP", "GN", "QZ"
	Type     string   // e.g., "GGA", "RMC"
	Fields   []string // raw fields between commas
	Checksum string   // two-character hex checksum
	Raw      string   // original raw sentence
}

// GGA represents a Global Positioning System Fix Data sentence.
type GGA struct {
	Sentence
	Time          string     // UTC time hhmmss.ss
	Latitude      float64    // Decimal degrees (north positive)
	Longitude     float64    // Decimal degrees (east positive)
	Quality       FixQuality // Fix quality indicator
	NumSatellites int        // Number of satellites in use
	HDOP          float64    // Horizontal dilution of precision
	Altitude      float64    // Altitude above mean sea level (meters)
	GeoidSep      float64    // Geoidal separation (meters)
	DGPSAge       float64    // Age of DGPS data (seconds)
	DGPSStationID string     // DGPS reference station ID
}

// RMC represents a Recommended Minimum Navigation Data sentence.
type RMC struct {
	Sentence
	Time      string  // UTC time hhmmss.ss
	Status    string  // A=Active, V=Void
	Latitude  float64 // Decimal degrees
	Longitude float64 // Decimal degrees
	Speed     float64 // Speed over ground in knots
	Course    float64 // Course over ground in degrees
	Date      string  // Date ddmmyy
	MagVar    float64 // Magnetic variation in degrees
	MagVarDir string  // E or W
	Mode      string  // A=Autonomous, D=Differential, E=Estimated, N=Not valid
}

// GSA represents a DOP and Active Satellites sentence.
type GSA struct {
	Sentence
	Mode     string  // M=Manual, A=Automatic
	FixType  int     // 1=No fix, 2=2D, 3=3D
	SVIDs    []int   // Satellite vehicle IDs (up to 12)
	PDOP     float64 // Position dilution of precision
	HDOP     float64 // Horizontal dilution of precision
	VDOP     float64 // Vertical dilution of precision
	SystemID int     // GNSS system ID (NMEA 4.10+)
}

// SatelliteInfo holds data for a single satellite from a GSV sentence.
type SatelliteInfo struct {
	SVID      int // Satellite vehicle ID
	Elevation int // Elevation in degrees (0-90)
	Azimuth   int // Azimuth in degrees (0-359)
	SNR       int // Signal-to-noise ratio in dB-Hz (0-99), -1 if not tracked
}

// GSV represents a Satellites in View sentence.
type GSV struct {
	Sentence
	TotalMsgs  int             // Total number of GSV messages
	MsgNum     int             // Current message number
	TotalSats  int             // Total satellites in view
	Satellites []SatelliteInfo // Satellite data (up to 4 per message)
}

// Parse parses a raw NMEA sentence string and returns the appropriate typed struct.
func Parse(raw string) (interface{}, error) {
	s, err := parseSentence(raw)
	if err != nil {
		return nil, err
	}

	switch s.Type {
	case "GGA":
		return parseGGA(s)
	case "RMC":
		return parseRMC(s)
	case "GSA":
		return parseGSA(s)
	case "GSV":
		return parseGSV(s)
	default:
		// Return the base sentence for unsupported types
		return s, nil
	}
}

// parseSentence extracts the base sentence structure from a raw NMEA string.
func parseSentence(raw string) (Sentence, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) == 0 {
		return Sentence{}, fmt.Errorf("empty sentence")
	}
	if raw[0] != '$' && raw[0] != '!' {
		return Sentence{}, fmt.Errorf("sentence must start with '$' or '!': %q", raw)
	}

	// Split checksum
	var body, checksum string
	if idx := strings.IndexByte(raw, '*'); idx >= 0 {
		body = raw[1:idx]
		checksum = raw[idx+1:]
	} else {
		return Sentence{}, fmt.Errorf("no checksum found in: %q", raw)
	}

	// Validate checksum
	if err := validateChecksum(body, checksum); err != nil {
		return Sentence{}, err
	}

	fields := strings.Split(body, ",")
	if len(fields) < 1 || len(fields[0]) < 3 {
		return Sentence{}, fmt.Errorf("invalid sentence header: %q", raw)
	}

	header := fields[0]
	talker := TalkerID(header[:2])
	sentType := header[2:]

	return Sentence{
		Talker:   talker,
		Type:     sentType,
		Fields:   fields[1:],
		Checksum: checksum,
		Raw:      raw,
	}, nil
}

// validateChecksum verifies the XOR checksum of the NMEA sentence body.
func validateChecksum(body, checksum string) error {
	checksum = strings.TrimSpace(checksum)
	if len(checksum) < 2 {
		return fmt.Errorf("invalid checksum: %q", checksum)
	}

	expected, err := strconv.ParseUint(checksum[:2], 16, 8)
	if err != nil {
		return fmt.Errorf("invalid checksum hex: %q", checksum)
	}

	var computed uint8
	for i := 0; i < len(body); i++ {
		computed ^= body[i]
	}

	if uint8(expected) != computed {
		return fmt.Errorf("checksum mismatch: expected 0x%02X, got 0x%02X", expected, computed)
	}
	return nil
}

// parseGGA parses a GGA sentence.
func parseGGA(s Sentence) (*GGA, error) {
	if len(s.Fields) < 14 {
		return nil, fmt.Errorf("GGA requires at least 14 fields, got %d", len(s.Fields))
	}

	gga := &GGA{Sentence: s}
	gga.Time = s.Fields[0]
	gga.Latitude = parseLatLon(s.Fields[1], s.Fields[2])
	gga.Longitude = parseLatLon(s.Fields[3], s.Fields[4])
	gga.Quality = FixQuality(parseInt(s.Fields[5]))
	gga.NumSatellites = parseInt(s.Fields[6])
	gga.HDOP = parseFloat(s.Fields[7])
	gga.Altitude = parseFloat(s.Fields[8])
	// Fields[9] = altitude units (M)
	gga.GeoidSep = parseFloat(s.Fields[10])
	// Fields[11] = geoidal sep units (M)
	gga.DGPSAge = parseFloat(s.Fields[12])
	gga.DGPSStationID = s.Fields[13]

	return gga, nil
}

// parseRMC parses an RMC sentence.
func parseRMC(s Sentence) (*RMC, error) {
	if len(s.Fields) < 11 {
		return nil, fmt.Errorf("RMC requires at least 11 fields, got %d", len(s.Fields))
	}

	rmc := &RMC{Sentence: s}
	rmc.Time = s.Fields[0]
	rmc.Status = s.Fields[1]
	rmc.Latitude = parseLatLon(s.Fields[2], s.Fields[3])
	rmc.Longitude = parseLatLon(s.Fields[4], s.Fields[5])
	rmc.Speed = parseFloat(s.Fields[6])
	rmc.Course = parseFloat(s.Fields[7])
	rmc.Date = s.Fields[8]
	rmc.MagVar = parseFloat(s.Fields[9])
	rmc.MagVarDir = s.Fields[10]
	if len(s.Fields) > 11 {
		rmc.Mode = s.Fields[11]
	}

	return rmc, nil
}

// parseGSA parses a GSA sentence.
// Standard GSA has 17 fields (mode, fix, 12 SVIDs, PDOP, HDOP, VDOP),
// but some receivers (e.g., QZSS) output fewer satellite ID slots.
// PDOP/HDOP/VDOP are always the last 3 fields (before optional SystemID).
func parseGSA(s Sentence) (*GSA, error) {
	if len(s.Fields) < 5 {
		return nil, fmt.Errorf("GSA requires at least 5 fields, got %d", len(s.Fields))
	}

	gsa := &GSA{Sentence: s}
	gsa.Mode = s.Fields[0]
	gsa.FixType = parseInt(s.Fields[1])

	// Determine DOP field positions.
	// Standard: fields 14-16. With fewer satellite slots, DOP shifts left.
	n := len(s.Fields)
	dopStart := n - 3
	if n > 17 {
		// NMEA 4.10+: SystemID is the last field, DOP is 3 fields before it
		gsa.SystemID = parseInt(s.Fields[n-1])
		dopStart = n - 4
	}

	// Satellite IDs: fields 2 through dopStart-1
	for i := 2; i < dopStart; i++ {
		if id := parseInt(s.Fields[i]); id > 0 {
			gsa.SVIDs = append(gsa.SVIDs, id)
		}
	}

	gsa.PDOP = parseFloat(s.Fields[dopStart])
	gsa.HDOP = parseFloat(s.Fields[dopStart+1])
	gsa.VDOP = parseFloat(s.Fields[dopStart+2])

	return gsa, nil
}

// parseGSV parses a GSV sentence.
func parseGSV(s Sentence) (*GSV, error) {
	if len(s.Fields) < 3 {
		return nil, fmt.Errorf("GSV requires at least 3 fields, got %d", len(s.Fields))
	}

	gsv := &GSV{Sentence: s}
	gsv.TotalMsgs = parseInt(s.Fields[0])
	gsv.MsgNum = parseInt(s.Fields[1])
	gsv.TotalSats = parseInt(s.Fields[2])

	// Each satellite takes 4 fields: SVID, elevation, azimuth, SNR
	for i := 3; i+3 < len(s.Fields); i += 4 {
		sat := SatelliteInfo{
			SVID:      parseInt(s.Fields[i]),
			Elevation: parseInt(s.Fields[i+1]),
			Azimuth:   parseInt(s.Fields[i+2]),
			SNR:       -1, // default: not tracked
		}
		if s.Fields[i+3] != "" {
			sat.SNR = parseInt(s.Fields[i+3])
		}
		gsv.Satellites = append(gsv.Satellites, sat)
	}

	return gsv, nil
}

// --- Helper functions ---

// parseLatLon converts NMEA lat/lon format (ddmm.mmmm) to decimal degrees.
func parseLatLon(value, direction string) float64 {
	if value == "" || direction == "" {
		return 0
	}

	// Find the decimal point to split degrees and minutes
	dotIdx := strings.IndexByte(value, '.')
	if dotIdx < 2 {
		return 0
	}

	degStr := value[:dotIdx-2]
	minStr := value[dotIdx-2:]

	deg, _ := strconv.ParseFloat(degStr, 64)
	min, _ := strconv.ParseFloat(minStr, 64)

	result := deg + min/60.0

	if direction == "S" || direction == "W" {
		result = -result
	}
	return result
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}
