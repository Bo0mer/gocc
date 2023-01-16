package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

var (
	addr      string
	connLimit int
)

func init() {
	flag.StringVar(&addr, "address", "127.0.0.1:2345", "Address to listen.")
	flag.IntVar(&connLimit, "connection-limit", 2, "Max number of connections.")
}

type ConnectionRegistry struct {
	mu          sync.RWMutex
	connections map[string]net.Conn
}

func (r *ConnectionRegistry) AddConnection(id string, conn net.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.connections[id] = conn
}

func (r *ConnectionRegistry) RemoveConnection(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.connections, id)
}

func (r *ConnectionRegistry) ActiveConnections() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.connections)
}

func (r *ConnectionRegistry) ForEach(action func(string, net.Conn)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, conn := range r.connections {
		action(id, conn)
	}
}

var registry ConnectionRegistry

func main() {
	flag.Parse()

	fmt.Println("address provided:", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("cannot start listener: %v", err)
	}

	registry = ConnectionRegistry{
		connections: make(map[string]net.Conn),
	}

	go printConnectionStats()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
		}

		activeConnections := registry.ActiveConnections()
		if activeConnections >= connLimit {
			fmt.Fprintf(conn, "sorry, max connection limit is reached, please try again later\n")
			conn.Close()
		} else {
			go handleConnection(conn)
		}
	}
}

func printConnectionStats() {
	t := time.Tick(2 * time.Second)
	for range t {
		connCount := registry.ActiveConnections()
		log.Printf("Number of active connections: %d", connCount)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var id = conn.RemoteAddr().String()
	// Step 1: Add conn to the connection registry.
	registry.AddConnection(id, conn)
	defer registry.RemoveConnection(id)

	reader := bufio.NewReader(conn)
	for {
		// Step 2: Read each new message from the client.
		message, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			log.Printf("error reading message from client: %v", err)
			return
		}

		registry.ForEach(func(cid string, cconn net.Conn) {
			if id == cid {
				// Do nothing.
				return
			}

			// Send the newly read message to all other clients.
			fmt.Fprintf(cconn, "<%s> %s", id, message)
		})
	}
}
