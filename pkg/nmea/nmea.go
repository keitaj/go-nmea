// Package nmea provides a parser for NMEA 0183 sentences commonly used in GNSS receivers.
//
// Supported sentence types:
//   - GGA: Global Positioning System Fix Data (position, quality, altitude)
//   - GLL: Geographic Position - Latitude/Longitude
//   - RMC: Recommended Minimum Navigation Data (position, velocity, date)
//   - VTG: Course Over Ground and Ground Speed
//   - GSA: DOP and Active Satellites
//   - GSV: Satellites in View (signal strength, elevation, azimuth)
//   - ZDA: Time and Date
//   - GBS: GNSS Satellite Fault Detection
//   - GST: GNSS Pseudorange Error Statistics
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

// BaseSentence is the base structure for all NMEA sentences.
type BaseSentence struct {
	Talker   TalkerID // e.g., "GP", "GN", "QZ"
	Type     string   // e.g., "GGA", "RMC"
	Fields   []string // raw fields between commas
	Checksum string   // two-character hex checksum
	Raw      string   // original raw sentence
}

// GGA represents a Global Positioning System Fix Data sentence.
type GGA struct {
	BaseSentence
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
	BaseSentence
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
	BaseSentence
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
	BaseSentence
	TotalMsgs  int             // Total number of GSV messages
	MsgNum     int             // Current message number
	TotalSats  int             // Total satellites in view
	Satellites []SatelliteInfo // Satellite data (up to 4 per message)
}

// GLL represents a Geographic Position - Latitude/Longitude sentence.
type GLL struct {
	BaseSentence
	Latitude  float64 // Decimal degrees (north positive)
	Longitude float64 // Decimal degrees (east positive)
	Time      string  // UTC time hhmmss.ss
	Status    string  // A=Valid, V=Invalid
	Mode      string  // A=Autonomous, D=Differential, E=Estimated, N=Not valid
}

// VTG represents a Course Over Ground and Ground Speed sentence.
type VTG struct {
	BaseSentence
	CourseTrue     float64 // Course over ground (true north, degrees)
	CourseMag      float64 // Course over ground (magnetic north, degrees)
	SpeedKnots     float64 // Speed over ground in knots
	SpeedKmh       float64 // Speed over ground in km/h
	Mode           string  // A=Autonomous, D=Differential, E=Estimated, N=Not valid
}

// ZDA represents a Time and Date sentence.
type ZDA struct {
	BaseSentence
	Time       string // UTC time hhmmss.ss
	Day        int    // Day (01-31)
	Month      int    // Month (01-12)
	Year       int    // Year (4-digit)
	LocalHours int    // Local zone hours (-13 to +13)
	LocalMins  int    // Local zone minutes (00-59)
}

// GBS represents a GNSS Satellite Fault Detection sentence.
type GBS struct {
	BaseSentence
	Time      string  // UTC time hhmmss.ss
	ErrLat    float64 // Expected error in latitude (meters, 1-sigma)
	ErrLon    float64 // Expected error in longitude (meters, 1-sigma)
	ErrAlt    float64 // Expected error in altitude (meters, 1-sigma)
	SVID      int     // Satellite ID of most likely failed satellite
	Prob      float64 // Probability of missed detection
	Bias      float64 // Estimate of bias on most likely failed satellite (meters)
	StdDev    float64 // Standard deviation of bias estimate (meters)
}

// GST represents a GNSS Pseudorange Error Statistics sentence.
type GST struct {
	BaseSentence
	Time    string  // UTC time hhmmss.ss
	RangeRMS float64 // RMS value of standard deviation of range inputs (meters)
	StdMajor float64 // Standard deviation of semi-major axis (meters, 1-sigma)
	StdMinor float64 // Standard deviation of semi-minor axis (meters, 1-sigma)
	Orient   float64 // Orientation of semi-major axis (degrees from true north)
	StdLat   float64 // Standard deviation of latitude error (meters, 1-sigma)
	StdLon   float64 // Standard deviation of longitude error (meters, 1-sigma)
	StdAlt   float64 // Standard deviation of altitude error (meters, 1-sigma)
}

// Parse parses a raw NMEA sentence string and returns the appropriate typed struct.
// All returned types implement the Sentence interface.
func Parse(raw string) (Sentence, error) {
	s, err := parseSentence(raw)
	if err != nil {
		return nil, err
	}

	switch s.Type {
	case "GGA":
		return parseGGA(s)
	case "GLL":
		return parseGLL(s)
	case "RMC":
		return parseRMC(s)
	case "VTG":
		return parseVTG(s)
	case "GSA":
		return parseGSA(s)
	case "GSV":
		return parseGSV(s)
	case "ZDA":
		return parseZDA(s)
	case "GBS":
		return parseGBS(s)
	case "GST":
		return parseGST(s)
	default:
		// Return the base sentence for unsupported types
		return &s, nil
	}
}

// parseSentence extracts the base sentence structure from a raw NMEA string.
func parseSentence(raw string) (BaseSentence, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) == 0 {
		return BaseSentence{}, newParseError(ErrInvalidFormat, raw, "empty sentence")
	}
	if raw[0] != '$' && raw[0] != '!' {
		return BaseSentence{}, newParseError(ErrInvalidFormat, raw, "sentence must start with '$' or '!': %q", raw)
	}

	// Split checksum
	var body, checksum string
	if idx := strings.IndexByte(raw, '*'); idx >= 0 {
		body = raw[1:idx]
		checksum = raw[idx+1:]
	} else {
		return BaseSentence{}, newParseError(ErrInvalidFormat, raw, "no checksum found in: %q", raw)
	}

	// Validate checksum
	if err := validateChecksum(body, checksum, raw); err != nil {
		return BaseSentence{}, err
	}

	fields := strings.Split(body, ",")
	if len(fields) < 1 || len(fields[0]) < 3 {
		return BaseSentence{}, newParseError(ErrInvalidFormat, raw, "invalid sentence header: %q", raw)
	}

	header := fields[0]
	talker := TalkerID(header[:2])
	sentType := header[2:]

	return BaseSentence{
		Talker:   talker,
		Type:     sentType,
		Fields:   fields[1:],
		Checksum: checksum,
		Raw:      raw,
	}, nil
}

// validateChecksum verifies the XOR checksum of the NMEA sentence body.
func validateChecksum(body, checksum, raw string) error {
	checksum = strings.TrimSpace(checksum)
	if len(checksum) < 2 {
		return newParseError(ErrChecksum, raw, "invalid checksum: %q", checksum)
	}

	expected, err := strconv.ParseUint(checksum[:2], 16, 8)
	if err != nil {
		return newParseError(ErrChecksum, raw, "invalid checksum hex: %q", checksum)
	}

	var computed uint8
	for i := 0; i < len(body); i++ {
		computed ^= body[i]
	}

	if uint8(expected) != computed {
		return newParseError(ErrChecksum, raw, "checksum mismatch: expected 0x%02X, got 0x%02X", expected, computed)
	}
	return nil
}

// parseGGA parses a GGA sentence.
func parseGGA(s BaseSentence) (*GGA, error) {
	if len(s.Fields) < 14 {
		return nil, newParseError(ErrFieldCount, s.Raw, "GGA requires at least 14 fields, got %d", len(s.Fields))
	}

	gga := &GGA{BaseSentence: s}
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
func parseRMC(s BaseSentence) (*RMC, error) {
	if len(s.Fields) < 11 {
		return nil, newParseError(ErrFieldCount, s.Raw, "RMC requires at least 11 fields, got %d", len(s.Fields))
	}

	rmc := &RMC{BaseSentence: s}
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
func parseGSA(s BaseSentence) (*GSA, error) {
	if len(s.Fields) < 5 {
		return nil, newParseError(ErrFieldCount, s.Raw, "GSA requires at least 5 fields, got %d", len(s.Fields))
	}

	gsa := &GSA{BaseSentence: s}
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
func parseGSV(s BaseSentence) (*GSV, error) {
	if len(s.Fields) < 3 {
		return nil, newParseError(ErrFieldCount, s.Raw, "GSV requires at least 3 fields, got %d", len(s.Fields))
	}

	gsv := &GSV{BaseSentence: s}
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

// parseGLL parses a GLL sentence.
func parseGLL(s BaseSentence) (*GLL, error) {
	if len(s.Fields) < 5 {
		return nil, newParseError(ErrFieldCount, s.Raw, "GLL requires at least 5 fields, got %d", len(s.Fields))
	}

	gll := &GLL{BaseSentence: s}
	gll.Latitude = parseLatLon(s.Fields[0], s.Fields[1])
	gll.Longitude = parseLatLon(s.Fields[2], s.Fields[3])
	gll.Time = s.Fields[4]
	if len(s.Fields) > 5 {
		gll.Status = s.Fields[5]
	}
	if len(s.Fields) > 6 {
		gll.Mode = s.Fields[6]
	}

	return gll, nil
}

// parseVTG parses a VTG sentence.
func parseVTG(s BaseSentence) (*VTG, error) {
	if len(s.Fields) < 8 {
		return nil, newParseError(ErrFieldCount, s.Raw, "VTG requires at least 8 fields, got %d", len(s.Fields))
	}

	vtg := &VTG{BaseSentence: s}
	vtg.CourseTrue = parseFloat(s.Fields[0])
	// Fields[1] = "T" (true)
	vtg.CourseMag = parseFloat(s.Fields[2])
	// Fields[3] = "M" (magnetic)
	vtg.SpeedKnots = parseFloat(s.Fields[4])
	// Fields[5] = "N" (knots)
	vtg.SpeedKmh = parseFloat(s.Fields[6])
	// Fields[7] = "K" (km/h)
	if len(s.Fields) > 8 {
		vtg.Mode = s.Fields[8]
	}

	return vtg, nil
}

// parseZDA parses a ZDA sentence.
func parseZDA(s BaseSentence) (*ZDA, error) {
	if len(s.Fields) < 4 {
		return nil, newParseError(ErrFieldCount, s.Raw, "ZDA requires at least 4 fields, got %d", len(s.Fields))
	}

	zda := &ZDA{BaseSentence: s}
	zda.Time = s.Fields[0]
	zda.Day = parseInt(s.Fields[1])
	zda.Month = parseInt(s.Fields[2])
	zda.Year = parseInt(s.Fields[3])
	if len(s.Fields) > 4 {
		zda.LocalHours = parseInt(s.Fields[4])
	}
	if len(s.Fields) > 5 {
		zda.LocalMins = parseInt(s.Fields[5])
	}

	return zda, nil
}

// parseGBS parses a GBS sentence.
func parseGBS(s BaseSentence) (*GBS, error) {
	if len(s.Fields) < 8 {
		return nil, newParseError(ErrFieldCount, s.Raw, "GBS requires at least 8 fields, got %d", len(s.Fields))
	}

	gbs := &GBS{BaseSentence: s}
	gbs.Time = s.Fields[0]
	gbs.ErrLat = parseFloat(s.Fields[1])
	gbs.ErrLon = parseFloat(s.Fields[2])
	gbs.ErrAlt = parseFloat(s.Fields[3])
	gbs.SVID = parseInt(s.Fields[4])
	gbs.Prob = parseFloat(s.Fields[5])
	gbs.Bias = parseFloat(s.Fields[6])
	gbs.StdDev = parseFloat(s.Fields[7])

	return gbs, nil
}

// parseGST parses a GST sentence.
func parseGST(s BaseSentence) (*GST, error) {
	if len(s.Fields) < 8 {
		return nil, newParseError(ErrFieldCount, s.Raw, "GST requires at least 8 fields, got %d", len(s.Fields))
	}

	gst := &GST{BaseSentence: s}
	gst.Time = s.Fields[0]
	gst.RangeRMS = parseFloat(s.Fields[1])
	gst.StdMajor = parseFloat(s.Fields[2])
	gst.StdMinor = parseFloat(s.Fields[3])
	gst.Orient = parseFloat(s.Fields[4])
	gst.StdLat = parseFloat(s.Fields[5])
	gst.StdLon = parseFloat(s.Fields[6])
	gst.StdAlt = parseFloat(s.Fields[7])

	return gst, nil
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
