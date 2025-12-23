# Go-Redis (Lite)

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)

A high-performance, persistent in-memory key-value store built from scratch in Go.

This project mimics the core architecture of **Redis**, featuring a custom TCP protocol, concurrent request handling, automatic key eviction (TTL), and disk persistence (AOF).

## 🚀 Features

- **Custom TCP Protocol:** Handles raw TCP connections and parses custom text commands (`SET`, `GET`, `SETEX`).
- **Concurrency Safe:** Uses `sync.RWMutex` to ensure safe concurrent reads and writes across multiple clients.
- **Persistence (AOF):** Implements an **Append-Only File** strategy to save every write command to disk. Data survives server crashes and restarts.
- **Automatic Eviction (TTL):** Features a background Garbage Collector (Goroutine) that actively scans and deletes expired keys.
- **Lazy Expiration:** Checks key validity on access to ensure no stale data is ever returned.

## 🛠️ Architecture

- **Server:** Listens on port `:6379`. Spawns a dedicated Goroutine for every new connection.
- **Storage:** In-memory `map[string]Item` protected by Read/Write Mutexes.
- **Garbage Collector:** A dedicated background thread that runs every 1 second to clean up expired keys.

## 💻 Tech Stack

- **Language:** Go (Golang)
- **Networking:** `net` (Raw TCP)
- **Concurrency:** Goroutines, Channels, `sync.Mutex`
- **I/O:** `bufio`, `os` (File persistence)

## 🏃‍♂️ Quick Start

1.  **Clone the repository**

    ```bash
    git clone [https://github.com/Khush-Ojha/go-redis.git](https://github.com/Khush-Ojha/go-redis.git)
    cd go-redis
    ```

2.  **Start the Server**

    ```bash
    go run main.go
    ```

    _Server will start on port :6379 and restore any previous data from `database.aof`._

3.  **Run the Client** (In a separate terminal)
    ```bash
    go run client.go
    ```

## 🧪 Commands

| Command   | Usage                     | Description                                      |
| :-------- | :------------------------ | :----------------------------------------------- |
| **SET**   | `SET key value`           | Saves a key permanently.                         |
| **GET**   | `GET key`                 | Retrieves a key. Returns `(nil)` if not found.   |
| **SETEX** | `SETEX key seconds value` | Saves a key that self-destructs after X seconds. |

## 🧠 Key Learnings

- **Memory Management:** Implemented both active (GC) and passive (Lazy) expiration strategies.
- **Durability:** Learned how to bridge the gap between volatile memory (RAM) and non-volatile storage (Disk) using write-ahead logging.
- **Race Conditions:** Solved data corruption issues using granular locking (`RLock` vs `Lock`).
