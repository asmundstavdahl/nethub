package main

import (
	"fmt"
	"log"
	"math"
	"strings"
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
	//lastRest := 0
	var clients int
	var speed int
	var unit string

	for {
		<-tickCh
		trafDiff := int(math.Ceil(float64(trafficicity) / (1000.0 / 500.0)))
		//lastRest = trafficicity - trafDiff

		// Clear line
		fmt.Print("\r\x1B[K\r")

		// Speed-o-meter
		bar, relativeTraf := makeBar(trafficicity)

		clients = clientChannels.Len()
		speed = relativeTraf //((trafficicity - lastRest) * 1000 / argMonitorInterval) / 1000
		unit = getUnit(trafficicity)

		// Clients and speed
		linePart1 := fmt.Sprintf("%3d clients | %3d%s/s ", clients, speed, unit)

		fmt.Print(linePart1, bar, "\r")

		trafficicity -= trafDiff
	}
}

func barLengthify(traf int) int {
	barLength := int(math.Pow(float64(traf), 0.5) - 0.1)
	if barLength < 0 {
		barLength = 0
	}
	return barLength
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

func makeBar(traf int) (bar string, relativeTraf int) {
	maxBarLength := barLengthify(999)

	var barLength int
	var barChar string
	var barCharLesser string

	if traf < BYTES_KILO {
		// Byte
		barChar = "·"
		barCharLesser = " "
		relativeTraf = int(traf / 1)
		barLength = barLengthify(relativeTraf)
	} else if traf >= BYTES_KILO && traf < BYTES_MEGA {
		// Kilo
		barChar = "–"
		barCharLesser = "·"
		relativeTraf = int(traf / BYTES_KILO)
		barLength = barLengthify(relativeTraf)
	} else if traf < BYTES_GIGA {
		// Mega
		barChar = "→"
		barCharLesser = "–"
		relativeTraf = int(traf / BYTES_MEGA)
		barLength = barLengthify(relativeTraf)
	} else if traf < BYTES_TERA {
		// Giga
		barChar = "»"
		barCharLesser = "→"
		relativeTraf = int(traf / BYTES_GIGA)
		barLength = barLengthify(relativeTraf)
	} else if traf >= BYTES_TERA {
		// Terra
		barChar = "÷"
		barCharLesser = "»"
		relativeTraf = int(traf / BYTES_TERA)
		barLength = barLengthify(relativeTraf)
	}
	bar = strings.Repeat(barChar, barLength) + strings.Repeat(barCharLesser, maxBarLength-barLength)
	return
}
