package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	inactivityTimeout = 30 * time.Second // Timeout duration
	maxMessageLength  = 1024             // Maximum allowed message length
)

func handleConnection(conn net.Conn) {
	defer func() {
		// Log disconnection
		fmt.Printf("Client disconnected: %s at %s\n", conn.RemoteAddr(), time.Now().Format(time.RFC1123))
		conn.Close()
	}()

	clientIP := conn.RemoteAddr().String()
	logFileName := fmt.Sprintf("%s.log", strings.ReplaceAll(clientIP, ":", "_"))
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer logFile.Close()

	fmt.Printf("Client connected: %s at %s\n", conn.RemoteAddr(), time.Now().Format(time.RFC1123))

	reader := bufio.NewReader(conn)

	// Create a channel to handle disconnection after timeout
	inactivityTimer := time.NewTimer(inactivityTimeout)
	defer inactivityTimer.Stop()

	// Goroutine to handle inactivity timeout
	go func() {
		<-inactivityTimer.C
		fmt.Printf("Inactivity timeout reached for client: %s\n", conn.RemoteAddr())
		conn.Close()
	}()

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

		// Reset the inactivity timer whenever we get a message
		inactivityTimer.Reset(inactivityTimeout)

		// Check the length of the message and handle long messages
		if len(message) > maxMessageLength {
			fmt.Printf("Received long message from %s (more than %d bytes). Rejecting or truncating.\n", conn.RemoteAddr(), maxMessageLength)

			// Option 1: Reject the message (send a message back saying it's too long)
			_, err = conn.Write([]byte("Message too long. Maximum allowed length is 1024 bytes.\n"))
			if err != nil {
				fmt.Printf("Error writing to %s: %v\n", conn.RemoteAddr(), err)
			}
			continue // Skip further processing

			// Option 2: Truncate the message (uncomment the next line if you want truncation instead of rejection)
			// message = message[:maxMessageLength] // Truncate the message
		}

		// Trim the message
		trimmed := strings.TrimSpace(message)
		switch strings.ToLower(trimmed) {
		case "hello":
			conn.Write([]byte("Hi there!\n"))
			continue
		case "":
			conn.Write([]byte("Say something...\n"))
			continue
		case "bye":
			conn.Write([]byte("Goodbye!\n"))
			return
		}
		if trimmed == "" {
			fmt.Printf("Received empty/whitespace message from %s\n", conn.RemoteAddr())
			continue
		}

		if strings.HasPrefix(trimmed, "/") {
			fields := strings.Fields(trimmed)
			command := fields[0]

			switch command {
			case "/time":
				now := time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST")
				conn.Write([]byte("Current server time: " + now + "\n"))
			case "/quit":
				conn.Write([]byte("Goodbye!\n"))
				return
			case "/echo":
				if len(fields) > 1 {
					// Reconstruct message without /echo
					echoMsg := strings.Join(fields[1:], " ")
					conn.Write([]byte(echoMsg + "\n"))
				} else {
					conn.Write([]byte("Usage: /echo [message]\n"))
				}
			default:
				conn.Write([]byte("Unknown command.\n"))
			}
			continue
		}
		// Log the message with a timestamp
		timestamp := time.Now().Format(time.RFC1123)
		logMessage := fmt.Sprintf("[%s] %s\n", timestamp, trimmed)
		_, err = logFile.WriteString(logMessage)
		if err != nil {
			fmt.Printf("Error writing to log file for %s: %v\n", clientIP, err)
			return
		}

		// Echo back clean message
		fmt.Printf("Received from %s: %q\n", conn.RemoteAddr(), trimmed)
		_, err = conn.Write([]byte(trimmed + "\n"))
		if err != nil {
			fmt.Printf("Error writing to %s: %v\n", conn.RemoteAddr(), err)
			return
		}
	}
}

func main() {

	port := flag.String("port", "4000", "Port number to listen on")
	flag.Parse()
	// define the target host and port we want to connect to
	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Printf("Server listening on :%s\n", *port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		go handleConnection(conn)
	}
}
