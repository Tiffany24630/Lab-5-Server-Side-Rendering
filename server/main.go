package main

import (
	"database/sql"
	"log"
	"net"
	"strings"
)

func main() {
	db := ConnectDB()
	defer db.Close()

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("http://localhost:8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, db)
	}
}

func handleConnection(conn net.Conn, db *sql.DB) {
	defer conn.Close()

	buf := make([]byte, 8192)

	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return
	}

	req := string(buf[:n])

	lines := strings.Split(req, "\r\n")
	if len(lines) == 0 {
		return
	}

	firstLine := lines[0]
	parts := strings.Split(firstLine, " ")

	if len(parts) < 2 {
		return
	}

	method := parts[0]
	path := parts[1]

	switch {
	case method == "GET" && path == "/":
		Home(conn, db)

	case method == "GET" && path == "/create":
		ShowCreate(conn)

	case method == "POST" && path == "/create":
		CreateSeries(conn, req, db)

	case method == "POST" && strings.HasPrefix(path, "/update"):
		UpdateEpisode(conn, path, db)

	case strings.HasPrefix(path, "/static/"):
		ServeStatic(conn, path)

	default:
		NotFound(conn)
	}
}
