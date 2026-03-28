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

// StreamReader reads NMEA sentences from an io.Reader and dispatches them.
type StreamReader struct {
	reader   *bufio.Reader
	onParsed Handler
	onError  ErrorHandler
}

// NewStreamReader creates a new NMEA stream reader.
func NewStreamReader(r io.Reader) *StreamReader {
	return &StreamReader{
		reader: bufio.NewReader(r),
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
	for {
		line, err := sr.reader.ReadString('\n')
		if len(line) > 0 {
			// Only process lines that look like NMEA sentences
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
		if err == io.EOF {
			return count, nil
		}
		if err != nil {
			return count, fmt.Errorf("read error: %w", err)
		}
	}
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
