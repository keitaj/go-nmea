# go-nmea

[![CI](https://github.com/keitaj/go-nmea/actions/workflows/ci.yml/badge.svg)](https://github.com/keitaj/go-nmea/actions/workflows/ci.yml)

A lightweight, zero-dependency NMEA 0183 parser written in Go. Supports multi-constellation GNSS receivers (GPS, GLONASS, Galileo, BeiDou, QZSS).

## Features

- Parse 9 sentence types: GGA, GLL, RMC, VTG, GSA, GSV, ZDA, GBS, GST
- Type-safe `Sentence` interface with functional interfaces (`HasPosition`, `HasTimestamp`, `HasSpeed`)
- Structured errors with `errors.Is` / `errors.As` support
- Multi-constellation support (GP, GL, GA, GB, QZ, GN talker IDs)
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
| GGA | Fix Data | Position, fix quality, satellites, HDOP, altitude | HasPosition, HasTimestamp |
| GLL | Geographic Position | Position, status | HasPosition, HasTimestamp |
| RMC | Recommended Minimum | Position, speed, course, date/time | HasPosition, HasTimestamp, HasSpeed |
| VTG | Course & Speed | True/magnetic course, speed (kn/kmh) | HasSpeed |
| GSA | Active Satellites | Fix mode (2D/3D), satellite IDs, PDOP/HDOP/VDOP | — |
| GSV | Satellites in View | Satellite elevation, azimuth, SNR | — |
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

## License

MIT
