package main

import (
	"database/sql"
		"html/template"
	"log"

	"net/http"
	"os"

	"strconv"

	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type TodoPageData struct {
	Timestamps     template.JS
	Downloadpoints template.JS
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
	sqlitestring := os.Args[len(os.Args)-1] + "?_query_only=true"
	db, _ := sql.Open("sqlite3", sqlitestring)
	defer db.Close()


	tmpl := template.Must(template.ParseFiles("stats.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		timestamps, downloadpoints := queryRecords(db)
		data := TodoPageData{
			Timestamps:     template.JS(timestamps),
			Downloadpoints: template.JS(downloadpoints),
		}
		tmpl.Execute(w, data)
	})
	go log.Fatal(http.ListenAndServe(":1337", nil))
}
