package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type TodoPageData struct {
	Timestamps     template.JS
	Downloadpoints template.JS
}

func insertRecord(db *sql.DB, timestamp int64, download float64, upload float64) {
	log.Println("Inserting record...")
	insertStudentSQL := `INSERT INTO speedrecords(timestamp, download, upload) VALUES (?, ?, ?)`
	statement, err := db.Prepare(insertStudentSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(timestamp, download, upload)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func main() {

	sqlitestring := os.Args[len(os.Args)-1]
	db, _ := sql.Open("sqlite3", sqlitestring)
	defer db.Close()

	loc, _ := time.LoadLocation("Europe/Berlin")

	speed, err := exec.Command("speedtest-cli", "--csv").Output()
	if err != nil {
		log.Fatal(err)
	}

	speedSlice := strings.Split(string(speed), ",")

	timestamp, _ := time.Parse("2006-01-02T15:04:05.999999Z", speedSlice[3])
	download, _ := strconv.ParseFloat(speedSlice[6], 64)
	download /= 1000000
	download = math.Round(download*100) / 100
	upload, _ := strconv.ParseFloat(speedSlice[7], 64)
	upload /= 1000000
	upload = math.Round(upload*100) / 100

	fmt.Println("Timestamp:", timestamp.In(loc).Format("2006-01-02 15:04:05"), "Download:", download, "Upload: ", upload)

	insertRecord(db, timestamp.Unix(), download, upload)
}
