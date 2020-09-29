package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type Todo struct {
	Title string
	Done  bool
}

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

func queryRecords(db *sql.DB) (string, string) {
	loc, _ := time.LoadLocation("Europe/Berlin")
	timestamps := ""
	downloadpoints := ""
	row, err := db.Query("SELECT * FROM speedrecords ORDER BY id")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var id int
		var timestamp int64
		var download float64
		var upload float64
		row.Scan(&id, &timestamp, &download, &upload)
		log.Println("Record", id, timestamp, download, upload)
		timestamps += time.Unix(timestamp, 0).In(loc).Format("'2006-01-02 15:04:05'") + ", "
		downloadpoints += strconv.FormatFloat(download, 'f', 2, 64) + ", "
		log.Println(timestamps)
		log.Println(downloadpoints)
	}
	return timestamps, downloadpoints
}

func main() {

	db, _ := sql.Open("sqlite3", "./sqlite.db")
	defer db.Close()

	queryRecords(db)

	loc, _ := time.LoadLocation("Europe/Berlin")

	speed, err := exec.Command("speedtest-cli", "--csv").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("speedtest done")
	fmt.Println(string(speed))

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

	tmpl := template.Must(template.ParseFiles("stats.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		timestamps, downloadpoints := queryRecords(db)
		data := TodoPageData{
			Timestamps:     template.JS(timestamps),
			Downloadpoints: template.JS(downloadpoints),
		}
		tmpl.Execute(w, data)
	})
	go log.Fatal(http.ListenAndServe(":8080", nil))
}
