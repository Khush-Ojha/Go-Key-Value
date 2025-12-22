package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Item struct {
	Value  string
	Expiry int64
}

type Database struct {
	data  map[string]Item
	mutex sync.RWMutex
	file  *os.File
}

var db = Database{
	data: make(map[string]Item),
}

func main() {
	fmt.Println("Starting Persistent Redis-lite on port :6379...")

	// 1. Open (or create) the AOF file for saving data
	var err error
	db.file, err = os.OpenFile("database.aof", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening AOF file:", err)
		return
	}
	defer db.file.Close()

	// 2. Load previous data from disk
	loadAOF()

	// 3. Start GC and Server
	go startGarbageCollector()

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

// Replay the AOF file to restore the database state
func loadAOF() {
	file, err := os.Open("database.aof")
	if err != nil {
		return // File doesn't exist yet, that's fine
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		cmd := parts[0]
		key := parts[1]

		if cmd == "SET" {
			db.data[key] = Item{Value: parts[2], Expiry: 0}
			count++
		} else if cmd == "SETEX" && len(parts) > 3 {
			// We skip the seconds parsing since we are making it permanent on restore
			_ = parts[2]

			db.data[key] = Item{Value: parts[3], Expiry: 0}
			count++
		}
	}
	fmt.Printf("💾 Restored %d keys from disk\n", count)
}

func startGarbageCollector() {
	for {
		time.Sleep(1 * time.Second)
		db.mutex.Lock()
		now := time.Now().Unix()
		deleted := 0
		for key, item := range db.data {
			if item.Expiry > 0 && item.Expiry < now {
				delete(db.data, key)
				deleted++
			}
		}
		db.mutex.Unlock()
		if deleted > 0 {
			fmt.Printf("♻️  GC cleaned up %d keys\n", deleted)
		}
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		rawLine := scanner.Text()
		parts := strings.Fields(rawLine)
		if len(parts) == 0 {
			continue
		}
		cmd := strings.ToUpper(parts[0])

		switch cmd {
		case "SET":
			if len(parts) < 3 {
				fmt.Fprintln(conn, "ERROR: Usage SET key value")
				continue
			}

			db.mutex.Lock()
			db.data[parts[1]] = Item{Value: parts[2], Expiry: 0}
			// Write to Disk
			fmt.Fprintln(db.file, rawLine)
			db.mutex.Unlock()

			fmt.Fprintln(conn, "OK")

		case "SETEX":
			if len(parts) < 4 {
				fmt.Fprintln(conn, "ERROR: Usage SETEX key seconds value")
				continue
			}
			seconds, _ := strconv.Atoi(parts[2])
			expiryTime := time.Now().Unix() + int64(seconds)

			db.mutex.Lock()
			db.data[parts[1]] = Item{Value: parts[3], Expiry: expiryTime}
			// Write to Disk
			fmt.Fprintln(db.file, rawLine)
			db.mutex.Unlock()

			fmt.Fprintln(conn, "OK")

		case "GET":
			if len(parts) < 2 {
				continue
			}
			key := parts[1]

			db.mutex.RLock()
			item, exists := db.data[key]
			db.mutex.RUnlock()

			if !exists {
				fmt.Fprintln(conn, "(nil)")
				continue
			}

			if item.Expiry > 0 && item.Expiry < time.Now().Unix() {
				db.mutex.Lock()
				delete(db.data, key)
				db.mutex.Unlock()
				fmt.Fprintln(conn, "(nil)")
			} else {
				fmt.Fprintln(conn, item.Value)
			}

		default:
			fmt.Fprintln(conn, "UNKNOWN")
		}
	}
}
