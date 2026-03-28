package nmea

import "fmt"

// ErrorKind classifies the type of parse error.
type ErrorKind int

const (
	ErrChecksum      ErrorKind = iota + 1 // Checksum mismatch
	ErrFieldCount                         // Insufficient fields for sentence type
	ErrInvalidFormat                      // Malformed sentence structure
)

// ParseError is the structured error type returned by Parse.
type ParseError struct {
	Kind     ErrorKind // Error classification
	Sentence string    // The raw sentence that failed
	Message  string    // Human-readable detail
}

func (e *ParseError) Error() string {
	return e.Message
}

// Is supports errors.Is matching by ErrorKind.
func (e *ParseError) Is(target error) bool {
	if t, ok := target.(*ParseError); ok {
		return e.Kind == t.Kind
	}
	return false
}

// Sentinel errors for use with errors.Is.
var (
	ErrChecksumMismatch   = &ParseError{Kind: ErrChecksum}
	ErrInsufficientFields = &ParseError{Kind: ErrFieldCount}
	ErrInvalidSentence    = &ParseError{Kind: ErrInvalidFormat}
)

// newParseError creates a ParseError with formatted message.
func newParseError(kind ErrorKind, raw, format string, args ...interface{}) *ParseError {
	return &ParseError{
		Kind:     kind,
		Sentence: raw,
		Message:  fmt.Sprintf(format, args...),
	}
}
