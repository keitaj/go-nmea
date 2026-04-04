# go-nmea

[![CI](https://github.com/keitaj/go-nmea/actions/workflows/ci.yml/badge.svg)](https://github.com/keitaj/go-nmea/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/keitaj/go-nmea.svg)](https://pkg.go.dev/github.com/keitaj/go-nmea)
[![Go Report Card](https://goreportcard.com/badge/github.com/keitaj/go-nmea)](https://goreportcard.com/report/github.com/keitaj/go-nmea)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight, zero-dependency NMEA 0183 parser written in Go. Supports multi-constellation GNSS receivers (GPS, GLONASS, Galileo, BeiDou, QZSS) with NMEA 4.10/4.11 extensions.

## Features

- Parse 13 sentence types: DTM, GFA, GGA, GLL, GNS, GRS, RMC, VTG, GSA, GSV, ZDA, GBS, GST
- Type-safe `Sentence` interface with functional interfaces (`HasPosition`, `HasTimestamp`, `HasSpeed`)
- Structured errors with `errors.Is` / `errors.As` support
- Multi-constellation support (GP, GL, GA, GB, QZ, GN talker IDs)
- NMEA 4.10/4.11 extensions (GSV SignalID, GSA SystemID, SBAS)
- Custom parser registration for proprietary sentence types
- Checksum validation
- Stream reader for continuous data (serial ports, log files)
- CLI tool for NMEA log analysis

## Install

```bash
go get github.com/keitaj/go-nmea
```

## Usage as Library

```go
package main

import (
    "errors"
    "fmt"
    "os"

    "github.com/keitaj/go-nmea/pkg/nmea"
)

func main() {
    // Parse a single sentence — returns nmea.Sentence interface
    s, err := nmea.Parse("$GNGGA,023042.00,3527.2700,N,13937.8900,E,1,10,0.92,8.2,M,36.7,M,,*4F")
    if err != nil {
        // Structured error handling
        if errors.Is(err, nmea.ErrChecksumMismatch) {
            fmt.Println("bad checksum")
        }
        panic(err)
    }

    // Use the Sentence interface for common fields
    fmt.Printf("[%s] %s\n", s.GetTalker(), s.GetType())

    // Use functional interfaces — no type switch needed
    if pos, ok := s.(nmea.HasPosition); ok {
        lat, lon := pos.GetPosition()
        fmt.Printf("Position: %.6f, %.6f\n", lat, lon)
    }

    // Type assertion for sentence-specific fields
    if gga, ok := s.(*nmea.GGA); ok {
        fmt.Printf("Fix: %s, Sats: %d, HDOP: %.2f\n", gga.Quality, gga.NumSatellites, gga.HDOP)
    }

    // Stream from file
    f, _ := os.Open("data.nmea")
    defer f.Close()

    reader := nmea.NewStreamReader(f)
    reader.OnParsed(func(sentence nmea.Sentence, raw string) {
        if pos, ok := sentence.(nmea.HasPosition); ok {
            lat, lon := pos.GetPosition()
            fmt.Printf("[%s] %.6f, %.6f\n", sentence.GetType(), lat, lon)
        }
    })
    reader.ReadAll()
}
```

## Custom Sentence Types

Register your own parsers for proprietary or unsupported sentence types:

```go
// Define your custom type
type HDT struct {
    nmea.BaseSentence
    Heading float64
    True    string
}

// Register it
nmea.RegisterParser("HDT", func(s nmea.BaseSentence) (nmea.Sentence, error) {
    if len(s.Fields) < 2 {
        return nil, fmt.Errorf("HDT requires 2 fields, got %d", len(s.Fields))
    }
    return &HDT{
        BaseSentence: s,
        Heading:      nmea.ParseFloat(s.Fields[0]),
        True:         s.Fields[1],
    }, nil
})

// Now Parse() handles HDT sentences automatically
s, _ := nmea.Parse("$GPHDT,274.07,T*03")
hdt := s.(*HDT)
fmt.Printf("Heading: %.2f°\n", hdt.Heading)
```

Helper functions `ParseFloat`, `ParseInt`, and `ParseLatLon` are exported for use in custom parsers.

## CLI Tool

```bash
make build

# Parse NMEA log file
./bin/nmea-cli -f data.nmea

# Filter by sentence type
./bin/nmea-cli -f data.nmea -type GGA

# Verbose output (GGA details)
./bin/nmea-cli -f data.nmea -type GGA -v

# Show parse errors
./bin/nmea-cli -f data.nmea -errors

# Read from stdin
cat data.nmea | ./bin/nmea-cli
```

## Supported Sentence Types

| Type | Description | Key Fields | Interfaces |
|------|-------------|------------|------------|
| DTM | Datum Reference | Local datum, lat/lon/alt offsets, reference datum | — |
| GFA | Fix Accuracy & Integrity | Protection levels (HPL/VPL), position errors, integrity status (4.11) | HasTimestamp |
| GGA | Fix Data | Position, fix quality, satellites, HDOP, altitude | HasPosition, HasTimestamp |
| GLL | Geographic Position | Position, status | HasPosition, HasTimestamp |
| GNS | GNSS Fix Data | Position, per-system mode indicators, NavStatus (4.10+) | HasPosition, HasTimestamp |
| GRS | Range Residuals | Per-satellite range residuals, SystemID/SignalID (4.10+) | HasTimestamp |
| RMC | Recommended Minimum | Position, speed, course, date/time | HasPosition, HasTimestamp, HasSpeed |
| VTG | Course & Speed | True/magnetic course, speed (kn/kmh) | HasSpeed |
| GSA | Active Satellites | Fix mode (2D/3D), satellite IDs, PDOP/HDOP/VDOP, SystemID (4.10+) | — |
| GSV | Satellites in View | Satellite elevation, azimuth, SNR, SignalID (4.10+) | — |
| ZDA | Time and Date | UTC time, date, timezone offset | HasTimestamp |
| GBS | Fault Detection | Position errors, failed satellite ID, bias | HasTimestamp |
| GST | Error Statistics | RMS, error ellipse, lat/lon/alt std dev | HasTimestamp |

## Supported Constellations

| Talker ID | Constellation |
|-----------|---------------|
| GP | GPS (USA) |
| GL | GLONASS (Russia) |
| GA | Galileo (EU) |
| GB | BeiDou (China) |
| QZ | QZSS / Michibiki (Japan) |
| GN | Multi-constellation |

## NMEA 4.10/4.11 Extensions

This library supports modern NMEA protocol extensions:

- **GFA Integrity** — Protection levels, alert limits, and integrity status for safety-of-life applications
- **GNS NavStatus** — Navigational status indicator (Safe, Caution, Unsafe, Not valid)
- **GRS SystemID/SignalID** — Per-constellation and per-signal range residuals
- **GSA SystemID** — Identifies the GNSS constellation per sentence, enabling per-system DOP tracking
- **GSV SignalID** — Identifies signal types for multi-frequency receivers (e.g., L1 C/A, L2C, L5)
- **SystemID constants** — GPS (1), GLONASS (2), Galileo (3), BeiDou (4), QZSS (5), NavIC (6), SBAS (7)

These fields are backward compatible — they default to zero when not present in standard NMEA sentences.

## License

MIT
