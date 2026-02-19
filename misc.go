package main

import (
	"fmt"
	"log"
	"time"
)

const (
	BYTES_KILO = 1000
	BYTES_MEGA = 1000 * 1000
	BYTES_GIGA = 1000 * 1000 * 1000
	BYTES_TERA = 1000 * 1000 * 1000 * 1000
)

var trafficicity int

func verbose(a ...interface{}) {
	if argVerbose {
		log.Println(a...)
	}
}

func trafficicityMonitor() {
	tickCh := time.Tick(time.Duration(argMonitorInterval) * time.Millisecond)
	var lastTraffic int
	var clients int
	var speed int
	var unit string

	// Dynamic history size: maintain 1 minute of samples regardless of interval
	// Calculate: 60000ms / monitor_interval = number of samples for 1 minute
	historySize := 60000 / argMonitorInterval
	if historySize < 1 {
		historySize = 1 // Minimum of 1 sample
	}
	if historySize > 1000 {
		historySize = 1000 // Maximum of 1000 samples to prevent excessive memory usage
	}

	// Create dynamic history array
	speedHistory := make([]int, historySize)
	var historyIndex = 0
	var historyCount = 0

	for {
		<-tickCh
		currentTraffic := trafficicity

		// Calculate current speed: bytes per second
		if lastTraffic > 0 && argMonitorInterval > 0 {
			timeFactor := 1000.0 / float64(argMonitorInterval) // Convert to per-second
			trafficDiff := currentTraffic - lastTraffic
			if trafficDiff > 0 {
				speed = int(float64(trafficDiff) * timeFactor)
			} else {
				speed = 0
			}
		} else {
			speed = 0
		}
		lastTraffic = currentTraffic

		// Store speed in history
		speedHistory[historyIndex] = speed
		historyIndex = (historyIndex + 1) % historySize
		if historyCount < historySize {
			historyCount++
		}

		// Calculate statistics from history
		minSpeed := -1
		maxSpeed := 0
		totalSpeed := 0
		validSamples := 0

		for i := 0; i < historyCount; i++ {
			s := speedHistory[i]
			// Count all speeds including 0 to properly track lulls
			if minSpeed == -1 || s < minSpeed {
				minSpeed = s
			}
			if s > maxSpeed {
				maxSpeed = s
			}
			totalSpeed += s
			validSamples++
		}

		// Calculate weighted average
		weightedAvg := 0
		if validSamples > 0 {
			weightedAvg = totalSpeed / validSamples
		}

		// Gentle decay for display (10% reduction per interval)
		decayAmount := int(float64(trafficicity) * 0.1)
		if decayAmount > 0 {
			trafficicity -= decayAmount
		}

		// Clear line
		fmt.Print("\r\x1B[K\r")

		clients = clientChannels.Len()
		unit = getUnit(speed)

		// Display format: clients | current/min/max/avg unit/s
		// Pad units to 2 characters and numbers to 5 characters to prevent line shifting
		padUnit := func(u string) string {
			if len(u) == 1 {
				return u + " "
			}
			return u
		}
		padNum := func(n int) string {
			return fmt.Sprintf("%5d", n)
		}

		var statsLine string
		if minSpeed == -1 {
			statsLine = fmt.Sprintf("%3d clients | %s%s/s (no history)", clients, padNum(speed), padUnit(unit))
		} else {
			statsLine = fmt.Sprintf("%3d clients | %s%s/s [min:%s%s/s| max:%s%s/s| avg:%s%s/s]", 
				clients, padNum(speed), padUnit(unit),
				padNum(minSpeed), padUnit(getUnit(minSpeed)),
				padNum(maxSpeed), padUnit(getUnit(maxSpeed)),
				padNum(weightedAvg), padUnit(getUnit(weightedAvg)))
		}

		fmt.Print(statsLine, "\r")
	}
}

func getUnit(traf int) (unit string) {
	if traf < BYTES_KILO {
		unit = "B"
	} else if traf < BYTES_MEGA {
		unit = "KB"
	} else if traf < BYTES_GIGA {
		unit = "MB"
	} else if traf < BYTES_TERA {
		unit = "GB"
	} else if traf >= BYTES_TERA {
		unit = "TB"
	}
	return
}
