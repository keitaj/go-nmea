package nmea

import (
	"bufio"
	"fmt"
	"io"
)

// Handler is called for each parsed NMEA sentence.
type Handler func(sentence Sentence, raw string)

// ErrorHandler is called when a sentence fails to parse.
type ErrorHandler func(raw string, err error)

// MaxLineLength is the maximum number of bytes read per line.
// NMEA 0183 specifies 82 characters max; 1024 provides generous headroom.
const MaxLineLength = 1024

// StreamReader reads NMEA sentences from an io.Reader and dispatches them.
type StreamReader struct {
	scanner  *bufio.Scanner
	onParsed Handler
	onError  ErrorHandler
}

// NewStreamReader creates a new NMEA stream reader.
func NewStreamReader(r io.Reader) *StreamReader {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, MaxLineLength), MaxLineLength)
	return &StreamReader{
		scanner: sc,
	}
}

// OnParsed registers a handler for successfully parsed sentences.
func (sr *StreamReader) OnParsed(h Handler) {
	sr.onParsed = h
}

// OnError registers a handler for parse errors.
func (sr *StreamReader) OnError(h ErrorHandler) {
	sr.onError = h
}

// ReadAll reads all sentences until EOF, dispatching each to handlers.
// Returns the total number of sentences processed and any I/O error.
func (sr *StreamReader) ReadAll() (int, error) {
	count := 0
	for sr.scanner.Scan() {
		line := sr.scanner.Text()
		trimmed := trimLine(line)
		if len(trimmed) > 0 && (trimmed[0] == '$' || trimmed[0] == '!') {
			count++
			parsed, parseErr := Parse(trimmed)
			if parseErr != nil {
				if sr.onError != nil {
					sr.onError(trimmed, parseErr)
				}
			} else if sr.onParsed != nil {
				sr.onParsed(parsed, trimmed)
			}
		}
	}
	if err := sr.scanner.Err(); err != nil {
		return count, fmt.Errorf("read error: %w", err)
	}
	return count, nil
}

func trimLine(s string) string {
	// Trim \r\n and spaces
	end := len(s)
	for end > 0 && (s[end-1] == '\n' || s[end-1] == '\r' || s[end-1] == ' ') {
		end--
	}
	start := 0
	for start < end && s[start] == ' ' {
		start++
	}
	return s[start:end]
}
