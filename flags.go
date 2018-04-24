package main

import (
	"flag"
)

var port int
var maxReadBytes int
var argVerbose bool
var argMonitor bool
var argMonitorInterval int
var argQuiet bool

func parseFlags() {
	flag.IntVar(&port, "port", 9001, "port for hub to listen to")
	flag.IntVar(&maxReadBytes, "maxread", 2000, "maximum number of bytes to read at once")
	flag.BoolVar(&argVerbose, "verbose", false, "turn on logging output")
	flag.BoolVar(&argMonitor, "monitor", false, "turn on traffic monitor")
	flag.IntVar(&argMonitorInterval, "moninterval", 200, "milliseconds delay for each monitor update")
	flag.BoolVar(&argQuiet, "quiet", false, "surpress all output except errors")

	flag.Parse()

	if argQuiet {
		argMonitor = false
		argVerbose = false
	}
}
