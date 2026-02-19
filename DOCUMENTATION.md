# nethub Documentation

## Overview

nethub is a simple TCP packet hub that broadcasts all received packets to all connected clients, except the sender. It's designed for network testing, debugging, and educational purposes.

## Installation

### Prerequisites
- Go 1.16 or later
- Network access (TCP port)

### Important Constraints
**This project uses ONLY the Go standard library.** No third-party modules or external dependencies are allowed. All functionality must be implemented using pure Go standard library packages.

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

The monitor displays comprehensive traffic statistics:
- Number of connected clients
- Current traffic speed (B/s, KB/s, etc.)
- Minimum speed over last minute
- Maximum speed over last minute
- Weighted average speed over last minute

Traffic units:
- B: Bytes (0-999)
- KB: Kilobytes (1000-999,999)
- MB: Megabytes (1,000,000-999,999,999)
- GB: Gigabytes (1,000,000,000-999,999,999,999)
- TB: Terabytes (1,000,000,000,000+)

The monitoring system maintains a dynamic 1-minute history that automatically adjusts based on the monitor interval setting. The history size is calculated as: `60000ms / monitor_interval = samples`. This ensures statistics always represent the last 60 seconds of activity regardless of the monitoring frequency.

Examples:
- 250ms interval: 240 samples (60000/250)
- 500ms interval: 120 samples (60000/500)
- 1000ms interval: 60 samples (60000/1000)

The monitor updates every interval regardless of traffic activity, ensuring accurate statistics even during periods of no traffic. Units are space-padded to prevent display shifting when values change.

Example display:
```
  2 clients |    15KB/s [min:    5KB/s| max:   42KB/s| avg:   18KB/s]
```

The display uses color coding for different units:
- **B** (Bytes): Green
- **KB** (Kilobytes): Blue  
- **MB** (Megabytes): Magenta
- **GB** (Gigabytes): Yellow
- **TB** (Terabytes): Red

### Smart Display Features

**3-Digit Limit**: All numbers are formatted to never exceed 3 digits using decimal prefixes:
- `1234 B/s` → `1.2 kB/s` (note lowercase k for decimal kilo)
- `12345 KB/s` → `12.3 MB/s`
- `123456 MB/s` → `123 GB/s`
- `1234567 GB/s` → `1.2 TB/s`

**Verified Behavior**: The formatting logic is thoroughly tested in `misc_test.go` with comprehensive test cases covering:
- Zero and small values
- Unit conversion thresholds
- Decimal precision handling
- Color coding verification
- Edge cases and boundary conditions

**Consistent Formatting**: Numbers display with appropriate decimal precision:
- `100+` → `123` (no decimals)
- `10-99.9` → `12.3` (1 decimal place)
- `<10` → `1.23` (2 decimal places)

**Auto-Scaling**: Values automatically convert to the most appropriate unit to keep numbers readable while maintaining the 3-digit limit. The system uses both binary (KB, MB, GB, TB) and decimal (kB) prefixes as needed.

This ensures the display remains clean and readable even with extremely high traffic volumes.

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

### Dependency Policy
This project maintains a strict **no third-party dependencies** policy. All features must be implemented using only the Go standard library. This ensures:
- Maximum compatibility
- Minimal security surface
- Easy deployment
- No dependency management overhead

The ANSI color codes used in the monitoring display are implemented with pure ANSI escape sequences, not external libraries.

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
