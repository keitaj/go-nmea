package nmea

// Sentence is the interface that all parsed NMEA sentences implement.
type Sentence interface {
	GetTalker() TalkerID
	GetType() string
	GetRaw() string
}

// HasPosition is implemented by sentence types that contain geographic coordinates.
type HasPosition interface {
	Sentence
	GetPosition() (lat, lon float64)
}

// HasTimestamp is implemented by sentence types that contain UTC time.
type HasTimestamp interface {
	Sentence
	GetTimestamp() string
}

// HasSpeed is implemented by sentence types that contain speed data.
type HasSpeed interface {
	Sentence
	GetSpeedKnots() float64
}

// BaseSentence interface methods.

func (s BaseSentence) GetTalker() TalkerID { return s.Talker }
func (s BaseSentence) GetType() string     { return s.Type }
func (s BaseSentence) GetRaw() string      { return s.Raw }

// HasPosition implementations.

func (g *GGA) GetPosition() (float64, float64) { return g.Latitude, g.Longitude }
func (g *GLL) GetPosition() (float64, float64) { return g.Latitude, g.Longitude }
func (r *RMC) GetPosition() (float64, float64) { return r.Latitude, r.Longitude }

// HasTimestamp implementations.

func (g *GGA) GetTimestamp() string { return g.Time }
func (g *GLL) GetTimestamp() string { return g.Time }
func (r *RMC) GetTimestamp() string { return r.Time }
func (z *ZDA) GetTimestamp() string { return z.Time }
func (g *GBS) GetTimestamp() string { return g.Time }
func (g *GST) GetTimestamp() string { return g.Time }

// HasSpeed implementations.

func (r *RMC) GetSpeedKnots() float64 { return r.Speed }
func (v *VTG) GetSpeedKnots() float64 { return v.SpeedKnots }
