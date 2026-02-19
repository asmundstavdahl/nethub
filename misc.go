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

		// ANSI color codes for different units
		getColoredUnit := func(u string) string {
			if len(u) == 1 {
				u = u + " "
			}
			
			switch u {
			case "B ":
				return "\033[32m" + u + "\033[0m" // Green for Bytes
			case "KB":
				return "\033[34m" + u + "\033[0m" // Blue for Kilobytes
			case "MB":
				return "\033[35m" + u + "\033[0m" // Magenta for Megabytes
			case "GB":
				return "\033[33m" + u + "\033[0m" // Yellow for Gigabytes
			case "TB":
				return "\033[31m" + u + "\033[0m" // Red for Terabytes
			default:
				return u
			}
		}
		
		// Format values to never exceed 3 digits using decimal prefixes
		formatSpeed := func(value int) (formattedValue float64, unit string) {
			if value >= BYTES_TERA {
				val := float64(value) / float64(BYTES_TERA)
				if val >= 1000 {
					return val / 1000, "PB" // Petabytes for very large values
				} else if val >= 100 {
					return val / 10, "TB"  // Show as decimal TB (e.g., 123.4 TB)
				} else {
					return val, "TB"
				}
			} else if value >= BYTES_GIGA {
				val := float64(value) / float64(BYTES_GIGA)
				if val >= 1000 {
					return val / 1000, "TB"
				} else if val >= 100 {
					return val / 10, "GB"  // Show as decimal GB (e.g., 123.4 GB)
				} else {
					return val, "GB"
				}
			} else if value >= BYTES_MEGA {
				val := float64(value) / float64(BYTES_MEGA)
				if val >= 1000 {
					return val / 1000, "GB"
				} else if val >= 100 {
					return val / 10, "MB"  // Show as decimal MB (e.g., 123.4 MB)
				} else {
					return val, "MB"
				}
			} else if value >= BYTES_KILO {
				val := float64(value) / float64(BYTES_KILO)
				if val >= 1000 {
					return val / 1000, "MB"
				} else if val >= 100 {
					return val / 10, "KB"  // Show as decimal KB (e.g., 123.4 KB)
				} else {
					return val, "KB"
				}
			} else {
				if value >= 1000 {
					return float64(value) / 1000, "kB" // Note: lowercase k for decimal
				} else {
					return float64(value), "B"
				}
			}
		}
		
		// Format numbers with consistent width (always 3 digits max with 1 decimal place)
		formatNumber := func(n float64) string {
			if n >= 100 {
				return fmt.Sprintf("%3.0f", n) // No decimal for 100+
			} else if n >= 10 {
				return fmt.Sprintf("%2.1f ", n) // 1 decimal for 10-99.9
			} else {
				return fmt.Sprintf("%1.1f  ", n) // 1 decimal for <10
			}
		}

		var statsLine string
		if minSpeed == -1 {
			val, unitStr := formatSpeed(speed)
			statsLine = fmt.Sprintf("%3d clients | %s%s/s (no history)", clients, formatNumber(val), getColoredUnit(unitStr))
		} else {
			currentVal, currentUnit := formatSpeed(speed)
			minVal, minUnit := formatSpeed(minSpeed)
			maxVal, maxUnit := formatSpeed(maxSpeed)
			avgVal, avgUnit := formatSpeed(weightedAvg)
			statsLine = fmt.Sprintf("%3d clients | %s%s/s [min:%s%s/s| max:%s%s/s| avg:%s%s/s]", 
				clients, formatNumber(currentVal), getColoredUnit(currentUnit),
				formatNumber(minVal), getColoredUnit(minUnit),
				formatNumber(maxVal), getColoredUnit(maxUnit),
				formatNumber(avgVal), getColoredUnit(avgUnit))
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
