package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

func getBody(req string) string {
	parts := strings.SplitN(req, "\r\n\r\n", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func sendHTML(conn net.Conn, html string) {
	resp := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nContent-Type: text/html; charset=utf-8\r\nContent-Length: %d\r\n\r\n%s",
		len(html), html,
	)
	conn.Write([]byte(resp))
}

func sendText(conn net.Conn, text string) {
	resp := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
		len(text), text,
	)
	conn.Write([]byte(resp))
}

func redirect(conn net.Conn, loc string) {
	conn.Write([]byte("HTTP/1.1 303 See Other\r\nLocation: " + loc + "\r\n\r\n"))
}

func NotFound(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func ServeStatic(conn net.Conn, path string) {
	filePath := "." + path

	data, err := os.ReadFile(filePath)
	if err != nil {
		NotFound(conn)
		return
	}

	contentType := "text/plain"

	if strings.HasSuffix(path, ".css") {
		contentType = "text/css"
	}
	if strings.HasSuffix(path, ".js") {
		contentType = "application/javascript"
	}

	resp := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n",
		contentType, len(data),
	)

	conn.Write([]byte(resp))
	conn.Write(data)
}

func Home(conn net.Conn, db *sql.DB) {
	series, err := GetAllSeries(db)
	if err != nil {
		sendText(conn, "DB error")
		return
	}

	htmlBytes, _ := os.ReadFile("templates/index.html")
	page := string(htmlBytes)

	rows := ""

	for _, s := range series {

		status := ""
		if s.Current >= s.Total {
			status = " ✔ Completada"
		}

		rows += fmt.Sprintf(`
<tr>
<td>%s%s</td>
<td>%d/%d</td>
<td><progress value="%d" max="%d"></progress></td>
<td><button onclick="nextEpisode(%d)">+1</button></td>
</tr>`,
			s.Name, status,
			s.Current, s.Total,
			s.Current, s.Total,
			s.ID)
	}

	page = strings.Replace(page, "{{rows}}", rows, 1)

	sendHTML(conn, page)
}

func ShowCreate(conn net.Conn) {
	html, _ := os.ReadFile("templates/create.html")
	sendHTML(conn, string(html))
}

func CreateSeries(conn net.Conn, req string, db *sql.DB) {
	body := getBody(req)
	values, _ := url.ParseQuery(body)

	name := values.Get("series_name")
	current := values.Get("current_episode")
	total := values.Get("total_episodes")

	if name == "" || current == "" || total == "" {
		sendText(conn, "Invalid data")
		return
	}

	db.Exec(
		"INSERT INTO series (name,current_episode,total_episodes) VALUES (?,?,?)",
		name, current, total,
	)

	redirect(conn, "/")
}

func UpdateEpisode(conn net.Conn, path string, db *sql.DB) {
	parts := strings.Split(path, "?")
	if len(parts) < 2 {
		sendText(conn, "bad request")
		return
	}

	params, _ := url.ParseQuery(parts[1])
	id := params.Get("id")

	db.Exec(`
	UPDATE series
	SET current_episode = current_episode + 1
	WHERE id=? AND current_episode < total_episodes
	`, id)

	sendText(conn, "ok")
}
