package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"flag"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

type AppContext struct {
	Db          *sql.DB
	Encoding    *base64.Encoding
	StoragePath *string
	TimeToLive  *int64
	MaxFileSize *int64
}

func renderIndex(daysToLive int64, maxSize int64) []byte {
	type IndexData struct {
		Days    int64
		MaxSize int64
	}

	tmpl, err := template.ParseFiles("./static/index.html")
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	tmpl.Execute(&buf, IndexData{Days: daysToLive, MaxSize: maxSize})
	return buf.Bytes()
}

func serveBytes(r *mux.Router, pattern string, content []byte) {
	if content == nil {
		log.Fatalf("Content for pattern %s is nil\n", pattern)
	}

	r.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	})
}

var (
	bindAddressFlag     = flag.String("bind", ":8080", "Bind Address")
	sqliteDbNameFlag    = flag.String("db-file", "./caixinha.db", "SQLite database file path")
	fileStoragePathFlag = flag.String("storage-path", os.TempDir()+"/caixinha/", "Where the files are saved")
	daysToLiveFlag      = flag.Int64("days-to-live", 30, "How many days should files be kept saved, a ttl smaller than 0 means the files are never deleted")
	maxFileSizeFlag     = flag.Int64("max-size", 100 /* MB */, "Max file size in megabytes")
)

func main() {
	flag.Parse()

	// Ensure the storage path exists
	err := os.MkdirAll(*fileStoragePathFlag, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create the storage folder: %s\n", err)
	}

	db := initDatabase(*sqliteDbNameFlag)
	defer db.Close()

	encoding := base64.RawURLEncoding
	ttlInSeconds := *daysToLiveFlag * int64(24*time.Hour.Seconds())
	app := AppContext{db, encoding, fileStoragePathFlag, &ttlInSeconds, maxFileSizeFlag}

	log.Printf("Starting with flags: days-to-live: %d; max-size: %d\n", *daysToLiveFlag, *maxFileSizeFlag)

	index := renderIndex(*daysToLiveFlag, *maxFileSizeFlag)
	robots, _ := os.ReadFile("./static/robots.txt")

	r := mux.NewRouter()
	r.HandleFunc("/upload", app.uploadFile).Methods("POST")
	r.HandleFunc("/f/{id}", app.getFile).Methods("GET")
	serveBytes(r, "/robots.txt", robots)
	serveBytes(r, "/", index)

	if *app.TimeToLive >= 0 {
		go cleanerService(&app)
	}

	log.Printf("Listening on %s\n", *bindAddressFlag)
	http.ListenAndServe(*bindAddressFlag, r)
}
