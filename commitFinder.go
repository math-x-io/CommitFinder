package main

import (
	"bufio"
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

const (
	charset           = "abcdefghijklmnopqrstuvwxyz0123456789"
	totalCombinations = 60466176
	barWidth          = 50
)

func generateRandomString(length int) string {
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

	_, err = db.Exec("DELETE FROM generated_strings;")
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
	query := "SELECT EXISTS(SELECT 1 FROM generated_strings WHERE value=? LIMIT 1);"
	err := db.QueryRow(query, value).Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	return exists
}

func insertString(db *sql.DB, value string) {
	_, err := db.Exec("INSERT INTO generated_strings(value) VALUES(?)", value)
	if err != nil {
		log.Fatal(err)
	}
}

func getLastString(db *sql.DB) string {
	var lastValue string
	query := "SELECT value FROM generated_strings ORDER BY id DESC LIMIT 1;"
	err := db.QueryRow(query).Scan(&lastValue)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	}
	return lastValue
}

func getRowCount(db *sql.DB) int {
	var count int
	query := "SELECT COUNT(*) FROM generated_strings;"
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	return count
}

func printProgressBar(count int, w *bufio.Writer) {
	progress := float64(count) / float64(totalCombinations)
	filled := int(progress * float64(barWidth))
	fmt.Fprintf(w, "\rProgress: [%s%s] %.2f%%", strings.Repeat("=", filled), strings.Repeat(" ", barWidth-filled), progress*100)
	w.Flush()
}

func nextString(s string) string {
	b := []byte(s)
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] == '9' {
			b[i] = 'a'
		} else if b[i] == 'z' {
			b[i] = '0'
			break
		} else {
			b[i]++
			break
		}
	}
	return string(b)
}

func enumerateStrings(db *sql.DB, repoURL string) {
	lastValue := getLastString(db)
	if lastValue == "" {
		lastValue = "aaaaa"
	}

	writer := bufio.NewWriter(os.Stdout)

	for {
		if stringExists(db, lastValue) {
			lastValue = nextString(lastValue)
			continue
		}

		insertString(db, lastValue)

		count := getRowCount(db)
		printProgressBar(count, writer)

		testURL := repoURL + "commit/" + lastValue
		resp, err := http.Get(testURL)
		if err != nil {
			log.Println("Erreur lors de la requête:", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("\nURL found:", testURL)

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

		lastValue = nextString(lastValue)
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

	fmt.Print("Do you want to enumerate all possible strings or generate random ones? (e/r): ")
	var choice string
	fmt.Scan(&choice)

	if choice == "e" {
		enumerateStrings(db, repoURL)
	} else {
		writer := bufio.NewWriter(os.Stdout)

		for {
			randomString := generateRandomString(5)

			if stringExists(db, randomString) {
				continue
			}

			insertString(db, randomString)

			count := getRowCount(db)
			printProgressBar(count, writer)

			testURL := repoURL + "commit/" + randomString

			resp, err := http.Get(testURL)
			if err != nil {
				log.Println("Erreur lors de la requête:", err)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Println("\nURL found:", testURL)

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
}
