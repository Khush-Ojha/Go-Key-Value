package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, _ := net.Dial("tcp", "localhost:6379")
	for {
		// Read from your keyboard
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')

		// Send to server
		fmt.Fprintf(conn, text+"\n")

		// Listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Server: " + message)
	}
}
