package main

import (
	"fmt"
	"net"
	"time"
)

func handleConnection(conn net.Conn) {
	defer func() {
		fmt.Printf("Client disconnected: %s at %s\n", conn.RemoteAddr(), time.Now().Format(time.RFC1123))
		conn.Close()
	}()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from client:", err)
			return
		}

		_, err = conn.Write(buf[:n])
		if err != nil {
			fmt.Println("Error writing to client:", err)
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

		fmt.Printf("Client connected: %s at %s\n", conn.RemoteAddr(), time.Now().Format(time.RFC1123))
		go handleConnection(conn)
	}
}
