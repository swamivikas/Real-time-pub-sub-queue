# Real-Time Pub/Sub Queue in Go (Terminal-Only)

A lightweight publish-subscribe messaging server implemented exclusively with Go’s standard library. Clients connect over plain TCP and issue newline-delimited commands to subscribe, publish, and unsubscribe from topics.

 **Highlights**

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

<img width="580" height="1140" alt="image" src="https://github.com/user-attachments/assets/441b4954-4d1b-475c-86b9-31d86a48a485" />


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



## here are ss attchaed with examples 



<img width="959" height="187" alt="Screenshot 2025-07-30 225055" src="https://github.com/user-attachments/assets/5234e496-8ce4-4dd9-b12a-9964f3f6fe1b" />



<img width="1026" height="247" alt="Screenshot 2025-07-30 225122" src="https://github.com/user-attachments/assets/c94eeeea-98ff-4369-b508-aba41dc6c511" />



<img width="1035" height="225" alt="Screenshot 2025-07-30 225140" src="https://github.com/user-attachments/assets/0fda3c5d-826a-4ffd-a21c-27e75ad6d3b9" />


<img width="1028" height="290" alt="Screenshot 2025-07-30 225145" src="https://github.com/user-attachments/assets/47f8f0b2-cf85-46fa-bcbc-217d3737dcdb" />






