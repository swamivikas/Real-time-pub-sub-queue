# Real-Time Pub/Sub Queue in Go (Terminal-Only)

A lightweight publish-subscribe messaging server implemented exclusively with Go’s standard library. Clients connect over plain TCP and issue newline-delimited commands to subscribe, publish, and unsubscribe from topics.

⚡ **Highlights**

* Pure Go 1.22 – no external deps
* Concurrent fan-out with goroutines & channels
* Safe shared state with `sync.RWMutex`
* Testable entirely from the terminal using `ncat`, `telnet`, or any raw TCP client

---

## Quick Start

```bash
# clone and enter the project
$ git clone https://github.com/<you>/realtime-pubsub-go.git
$ cd realtime-pubsub-go

# run (default port 9000)
$ go run .

# or build binary
$ go build -o pubsub
$ ./pubsub            # listens on :9000
$ ./pubsub 8080       # custom port
```

Open two (or more) terminal windows and connect with `ncat` / `telnet`:

```bash
# Terminal A – subscriber
$ ncat localhost 9000
SUBSCRIBE sports

# Terminal B – publisher
$ ncat localhost 9000
PUBLISH sports Ronaldo scored!
```

Terminal A instantly prints:
```
Ronaldo scored!
```

---

## Command Reference

Each command is a single line terminated by `\n`.

| Command | Description |
|---------|-------------|
| `SUBSCRIBE <topic>`        | Start receiving messages for a topic |
| `UNSUBSCRIBE <topic>`      | Stop receiving messages for a topic |
| `PUBLISH <topic> <message>`| Broadcast `<message>` to all subscribers of `<topic>` |
| `EXIT`                     | Gracefully disconnect |

Example session:
```
SUBSCRIBE news
PUBLISH news "Breaking: Go 1.22 released!"
UNSUBSCRIBE news
EXIT
```

---

## Architecture Overview

```mermaid
flowchart TD
    A[Client] --commands--> S(Server)
    subgraph Server
      B[Command parser]
      C[Broker\nmap[topic][]chan]
      D[Forwarder goroutine]
    end

    B --> C
    C --fan-out--> D
    D --messages--> A
```

* **Broker** – central hub maintaining `map[string]map[chan string]struct{}`.
* **Concurrency** – each client connection has two goroutines:
  * reader: parses commands, talks to Broker
  * writer: forwards broker messages to the socket
* **Back-pressure** – per-subscriber buffered channels (size 16) drop messages if a client is too slow.

---

## Testing Tips

### Windows
* Install Ncat via **winget**: `winget install --id Insecure.Nmap -e`  
  (Then restart PowerShell so `ncat` is on PATH.)
* Or enable Telnet: `dism /online /Enable-Feature /FeatureName:TelnetClient`

### Linux / macOS
* Debian/Ubuntu: `sudo apt install netcat-openbsd`
* macOS (Homebrew): `brew install nmap` (includes ncat)

---

## License

MIT – do whatever you want, just preserve copyright.
