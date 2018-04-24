package main

func main() {
	parseFlags()

	verbose("Hubbing on port", port)

	broadcast := make(chan *Packet)

	end := make(chan bool)

	if argMonitor {
		go trafficicityMonitor()
	}

	go acceptClients(end, broadcast)

	// Wait for end by error
	<-end
}
