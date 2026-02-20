# nethub

TCP packet hub that broadcasts data to all connected clients except the sender.

## Installation

```sh
go install github.com/asmundstavdahl/nethub
```

## Usage

```sh
nethub [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | 9001 | Port to listen on |
| `-maxread` | 2000 | Max bytes to read at once |
| `-verbose` | false | Enable logging output |
| `-monitor` | false | Enable traffic monitor |
| `-moninterval` | 500 | Monitor update interval (ms) |
| `-quiet` | false | Suppress all output except errors |

## Architecture

Two main components:
- **Accept loop**: Accepts connections, spawns handlers
- **Connection handler**: One per client, reads data and broadcasts

### Broadcasting

When a client sends data:
1. Connection handler reads into buffer
2. Creates a packet with sender's channel and data
3. Broadcasts to all connected channels except sender's
4. Updates traffic statistics

Each client has a dedicated channel. A write goroutine continuously reads from this channel and writes to the connection.

### Monitor

When enabled (`-monitor`), displays real-time stats:
- Current/min/max/avg speed (color-coded by unit)
- Client count
- 1-minute rolling history
