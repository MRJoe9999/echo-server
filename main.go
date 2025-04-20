package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

func handleConnection(conn net.Conn) {
	defer func() {
		fmt.Printf("Client disconnected: %s at %s\n", conn.RemoteAddr(), time.Now().Format(time.RFC1123))
		conn.Close()
	}()

	fmt.Printf("Client connected: %s at %s\n", conn.RemoteAddr(), time.Now().Format(time.RFC1123))

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Printf("Client closed the connection: %s\n", conn.RemoteAddr())
			} else {
				fmt.Printf("Error reading from %s: %v\n", conn.RemoteAddr(), err)
			}
			return
		}

		// Trim the message
		trimmed := strings.TrimSpace(message)
		if trimmed == "" {
			fmt.Printf("Received empty/whitespace message from %s\n", conn.RemoteAddr())
			continue
		}

		fmt.Printf("Received from %s: %q\n", conn.RemoteAddr(), trimmed)

		// Echo back clean message
		_, err = conn.Write([]byte(trimmed + "\n"))
		if err != nil {
			fmt.Printf("Error writing to %s: %v\n", conn.RemoteAddr(), err)
			return
		}
	}
}

func main() {
	// define the target host and port we want to connect to
	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("Server listening on :4000")
	// Our program runs an infinite loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}

		go handleConnection(conn)
	}
}
