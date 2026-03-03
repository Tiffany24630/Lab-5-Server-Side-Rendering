package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Series struct {
	ID      int
	Name    string
	Current int
	Total   int
}

func ConnectDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./series.db")
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(1)
	return db
}

func GetAllSeries(db *sql.DB) ([]Series, error) {
	rows, err := db.Query("SELECT id,name,current_episode,total_episodes FROM series")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Series

	for rows.Next() {
		var s Series
		rows.Scan(&s.ID, &s.Name, &s.Current, &s.Total)
		list = append(list, s)
	}

	return list, nil
}
