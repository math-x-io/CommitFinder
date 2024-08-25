package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	
)

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func clearDB() {
	db, err := sql.Open("sqlite3", "./cache.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`DELETE FROM generated_strings;`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Cache cleared")
}

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./cache.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS generated_strings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		value TEXT NOT NULL UNIQUE
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func stringExists(db *sql.DB, value string) bool {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM generated_strings WHERE value=? LIMIT 1);`
	err := db.QueryRow(query, value).Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	return exists
}

func insertString(db *sql.DB, value string) {
	_, err := db.Exec(`INSERT INTO generated_strings(value) VALUES(?)`, value)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	db := initDB()
	defer db.Close()

	fmt.Print("Enter the URL of the target repo (https://github.com/torvalds/linux) : ")
	var repoURL string
	fmt.Scan(&repoURL)

	fmt.Print("Do you want to clear the cache? (y/n): ")
	var clearCache string
	fmt.Scan(&clearCache)
	if clearCache == "y" {
		clearDB()
	}

	if !strings.HasSuffix(repoURL, "/") {
		repoURL += "/"
	}

	for {
		randomString := generateRandomString(5)

		if stringExists(db, randomString) {
			continue
		}

		insertString(db, randomString)

		testURL := repoURL + "commit/" + randomString

		resp, err := http.Get(testURL)
		if err != nil {
			log.Println("Erreur lors de la requête:", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("URL found :", testURL)

			file, err := os.OpenFile("found.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			if _, err := file.WriteString(testURL + "\n"); err != nil {
				log.Fatal(err)
			}
			break
		}
	}
}
