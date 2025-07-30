package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

// Broker maintains topic subscribers and handles publishing.
type Broker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan string]struct{}
}

// NewBroker creates a new Broker instance.
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string]map[chan string]struct{}),
	}
}

// Subscribe registers a new subscriber channel for a topic.
// It returns the channel that will receive published messages and
// a function to call to unsubscribe.
func (b *Broker) Subscribe(topic string) (chan string, func()) {
	ch := make(chan string, 16) // buffered to prevent blocking publisher
	b.mu.Lock()
	if _, ok := b.subscribers[topic]; !ok {
		b.subscribers[topic] = make(map[chan string]struct{})
	}
	b.subscribers[topic][ch] = struct{}{}
	b.mu.Unlock()

	// Unsubscribe closure
	unsub := func() {
		b.mu.Lock()
		if subs, ok := b.subscribers[topic]; ok {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(b.subscribers, topic)
			}
		}
		b.mu.Unlock()
		close(ch)
	}

	return ch, unsub
}

// Publish sends a message to all subscribers of the given topic.
func (b *Broker) Publish(topic, msg string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if subs, ok := b.subscribers[topic]; ok {
		for ch := range subs {
			select {
			case ch <- msg:
			default:
				// if subscriber is slow, drop the message to keep system responsive
			}
		}
	}
}

func main() {
	port := "9000"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	broker := NewBroker()

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Pub/Sub server listening on %s", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		log.Printf("client connected: %s", conn.RemoteAddr())
		go handleConnection(conn, broker)
	}
}

func handleConnection(conn net.Conn, broker *Broker) {
	defer conn.Close()

	// Track this client's subscriptions to allow clean-up.
	type sub struct {
		topic string
		ch    chan string
		unsub func()
	}
	var subs []sub

	// Goroutine to forward messages from broker to client.
	var wg sync.WaitGroup

	cleanup := func() {
		for _, s := range subs {
			s.unsub()
		}
		wg.Wait() // ensure all forwarders finished before closing
		log.Printf("client disconnected: %s", conn.RemoteAddr())
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		cmd := strings.ToUpper(parts[0])

		switch cmd {
		case "SUBSCRIBE":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR missing topic for SUBSCRIBE")
				continue
			}
			topic := parts[1]
			ch, unsub := broker.Subscribe(topic)
			s := sub{topic: topic, ch: ch, unsub: unsub}
			subs = append(subs, s)

			// Start forwarder
			wg.Add(1)
			go func(c chan string) {
				defer wg.Done()
				for msg := range c {
					fmt.Fprintf(conn, "%s\n", msg)
				}
			}(ch)
			fmt.Fprintf(conn, "OK subscribed to %s\n", topic)

		case "UNSUBSCRIBE":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR missing topic for UNSUBSCRIBE")
				continue
			}
			topic := parts[1]
			found := false
			for i, s := range subs {
				if s.topic == topic {
					s.unsub()
					subs = append(subs[:i], subs[i+1:]...)
					found = true
					break
				}
			}
			if found {
				fmt.Fprintf(conn, "OK unsubscribed from %s\n", topic)
			} else {
				fmt.Fprintf(conn, "ERR not subscribed to %s\n", topic)
			}

		case "PUBLISH":
			if len(parts) < 3 {
				fmt.Fprintln(conn, "ERR usage: PUBLISH <topic> <message>")
				continue
			}
			topic := parts[1]
			msg := parts[2]
			broker.Publish(topic, msg)
			fmt.Fprintln(conn, "OK")

		case "EXIT":
			fmt.Fprintln(conn, "bye")
			cleanup()
			return

		default:
			fmt.Fprintf(conn, "ERR unknown command %s\n", cmd)
		}
	}
	cleanup()
}
