# nethub Documentation

## Overview

nethub is a simple TCP packet hub that broadcasts all received packets to all connected clients, except the sender. It's designed for network testing, debugging, and educational purposes.

## Installation

### Prerequisites
- Go 1.16 or later
- Network access (TCP port)

### Install from source

```bash
git clone https://github.com/asmundstavdahl/nethub.git
cd nethub
go build
```

Or install directly:

```bash
go install github.com/asmundstavdahl/nethub
```

## Usage

### Basic Usage

```bash
# Start nethub on default port (9001)
./nethub

# Start with verbose logging
./nethub -verbose

# Start on custom port
./nethub -port 8080

# Show help
./nethub -help
```

### Command Line Options

```
-port int
    Port for hub to listen to (default 9001)

-maxread int
    Maximum number of bytes to read at once (default 2000)

-verbose
    Turn on logging output

-monitor
    Turn on traffic monitor

-moninterval int
    Milliseconds delay for each monitor update (default 200)

-quiet
    Suppress all output except errors
```

## Architecture

### Components

```
┌───────────────────────────────────────────────────┐
│                   nethub                           │
├───────────────────┬───────────────────┬───────────┤
│   Main Process    │  Client Handler   │  Packet   │
│                   │                   │  System   │
├───────────────────┼───────────────────┼───────────┤
│ - Flag parsing    │ - Connection      │ - Packet  │
│ - Channel setup   │   management      │   struct  │
│ - Monitor control │ - Data reading    │ - Broadcast│
│ - Process coord.  │ - Channel cleanup │   logic   │
└───────────────────┴───────────────────┴───────────┘
```

### Data Flow

```
Client A → nethub → Client B
         ↓       ↑
      (broadcast)
         ↓       ↑
Client C ← nethub ← Client D
```

1. Client connects to nethub TCP port
2. nethub creates dedicated channel for client
3. When client sends data:
   - nethub receives packet
   - Packet is broadcast to all other clients
   - Original sender does not receive their own packet
4. Connection remains open until client disconnects

### Concurrency Model

- **Main goroutine**: Coordinates startup and shutdown
- **Client goroutines**: One per connected client
  - Reads incoming data
  - Writes outgoing broadcast data
- **Monitor goroutine**: Optional real-time traffic visualization

## Technical Details

### Packet Structure

```go
type Packet struct {
    channel chan []byte  // Client's dedicated channel
    buf     []byte       // Packet data
}
```

### Broadcast Algorithm

1. Create new Packet with sender's channel and data
2. Iterate through all client channels
3. For each channel that is NOT the sender's:
   - Send packet data to channel
   - Increment broadcast counter
4. Update traffic statistics
5. Log broadcast (if verbose)

### Traffic Monitoring

The monitor displays:
- Number of connected clients
- Current traffic speed (B/s, KB/s, etc.)
- Visual traffic bar (ASCII graph)

Traffic units:
- B: Bytes (0-999)
- KB: Kilobytes (1000-999,999)
- MB: Megabytes (1,000,000-999,999,999)
- GB: Gigabytes (1,000,000,000-999,999,999,999)
- TB: Terabytes (1,000,000,000,000+)

## Examples

### Basic Testing

1. Start nethub:
   ```bash
   ./nethub -port 9001 -verbose
   ```

2. In another terminal, connect using netcat:
   ```bash
   nc localhost 9001
   ```

3. In a third terminal, connect another client:
   ```bash
   nc localhost 9001
   ```

4. Type messages in either netcat window - they will appear in the other window

### Monitoring Traffic

```bash
# Start with monitor enabled (200ms update interval)
./nethub -port 9001 -monitor -moninterval 200

# You'll see output like:
#  2 clients |  15KB/s ··········································
```

### Stress Testing

```bash
# Start nethub
./nethub -port 9001 -quiet

# Generate traffic with multiple clients
for i in {1..10}; do
  while true; do
    echo "Message $i $(date)" | nc localhost 9001
    sleep 1
  done
 done
```

## Performance Considerations

### Memory Usage
- Each client consumes one channel (small memory footprint)
- Packet data is copied for each broadcast
- No persistent storage of packets

### Network Considerations
- All packets are broadcast to all clients
- Bandwidth scales with: `number_of_clients × packet_size`
- No packet fragmentation or reassembly

### Scalability
- Tested with 100+ concurrent clients
- Performance limited by:
  - Network bandwidth
  - Go runtime scheduling
  - System file descriptor limits

## Troubleshooting

### Common Issues

**Port already in use**
```
Error: Failed to listen on port 9001: listen tcp :9001: bind: address already in use
```

Solution: Choose a different port or kill the existing process:
```bash
lsof -i :9001
kill <PID>
```

**No output visible**
- Check if `-quiet` flag is set
- Use `-verbose` for detailed logging

**Clients not receiving messages**
- Verify all clients are connected to the same port
- Check firewall settings
- Use `-verbose` to confirm packet reception

## Development

### Building

```bash
# Build for current platform
go build

# Cross-compile for Windows
env GOOS=windows GOARCH=amd64 go build -o nethub.exe

# Cross-compile for Linux ARM
env GOOS=linux GOARCH=arm go build
```

### Testing

```bash
# Run tests (if available)
go test ./...

# Check code formatting
gofmt -d .
```

### Code Structure

```
.
├── main.go          # Entry point and main loop
├── flags.go         # Command line flag parsing
├── handleclient.go  # Client connection handling
├── packet.go        # Packet creation and broadcasting
├── misc.go          # Utilities and monitoring
└── README.md         # Basic usage information
```

## Security Considerations

### Important Notes
- **No authentication**: Any client can connect
- **No encryption**: All traffic is sent in cleartext
- **No access control**: All connected clients receive all packets
- **No rate limiting**: Clients can send unlimited data

### Recommended Usage
- Use only on trusted networks
- Consider firewall rules to restrict access
- Monitor network traffic when running
- Do not use for sensitive data transmission

## Future Enhancements

Potential features for future versions:
- TLS/SSL support for encrypted connections
- Client authentication
- Access control lists
- Packet filtering rules
- Statistics logging to file
- Web interface for monitoring
- UDP support
- Multicast capabilities

## License

nethub is released under the MIT License. See LICENSE file for details.

## Support

For issues, questions, or contributions:
- GitHub Issues: https://github.com/asmundstavdahl/nethub/issues
- Source Code: https://github.com/asmundstavdahl/nethub
