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

Two goroutine types:
- **acceptClients** (`handleclient.go:48`): Accepts connections, spawns handlers
- **handleConnection** (`handleclient.go:12`): One per client, reads data and broadcasts

### Broadcasting

When a client sends data:
1. `handleConnection` reads into buffer
2. Creates `Packet` with sender's channel and data
3. `Packet.Broadcast()` iterates `clientChannels` list
4. Sends data to all channels except sender's
5. Updates global `trafficicity` counter

Each client has a dedicated channel. A write goroutine continuously reads from this channel and writes to the connection.

### Monitor

When enabled (`-monitor`), displays real-time stats:
- Current/min/max/avg speed (color-coded by unit)
- Client count
- 1-minute rolling history
