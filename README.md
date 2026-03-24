# go-nmea

A lightweight, zero-dependency NMEA 0183 parser written in Go. Supports multi-constellation GNSS receivers (GPS, GLONASS, Galileo, BeiDou, QZSS).

## Features

- Parse GGA, RMC, GSA, GSV sentence types
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
    "fmt"
    "os"

    "github.com/keitaj/go-nmea/pkg/nmea"
)

func main() {
    // Parse a single sentence
    result, err := nmea.Parse("$GNGGA,023042.00,3527.2700,N,13937.8900,E,1,10,0.92,8.2,M,36.7,M,,*4F")
    if err != nil {
        panic(err)
    }

    gga := result.(*nmea.GGA)
    fmt.Printf("Position: %.6f, %.6f (Fix: %s)\n", gga.Latitude, gga.Longitude, gga.Quality)

    // Stream from file
    f, _ := os.Open("data.nmea")
    defer f.Close()

    reader := nmea.NewStreamReader(f)
    reader.OnParsed(func(sentence interface{}, raw string) {
        if g, ok := sentence.(*nmea.GGA); ok {
            fmt.Printf("%.6f, %.6f | Sats: %d | HDOP: %.2f\n",
                g.Latitude, g.Longitude, g.NumSatellites, g.HDOP)
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

| Type | Description | Key Fields |
|------|-------------|------------|
| GGA | Fix Data | Position, fix quality, satellites, HDOP, altitude |
| RMC | Recommended Minimum | Position, speed, course, date/time |
| GSA | Active Satellites | Fix mode (2D/3D), satellite IDs, PDOP/HDOP/VDOP |
| GSV | Satellites in View | Satellite elevation, azimuth, SNR |

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
