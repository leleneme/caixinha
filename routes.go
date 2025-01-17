package main

import (
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func (c *AppContext) uploadFile(w http.ResponseWriter, r *http.Request) {
	// 1 megabyte = 1 million bytes, i guess
	maxBytesSize := *c.MaxFileSize * 1_000_000

	// ParseMultipartForm only limits how much of the server's memory is used for parsing.
	// Anything exceeding maxBytesSize is saved to disk, so we need truncate the request
	// body first to maxBytesSize and a plus of 500 bytes (just to be safe) ;)
	r.Body = http.MaxBytesReader(w, r.Body, maxBytesSize+500)
	err := r.ParseMultipartForm(maxBytesSize)
	if err != nil {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		w.Write([]byte(fmt.Sprintf("maximum file size of %d bytes exceeded", maxBytesSize)))
		return
	}

	uploadedFile, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error retriving 'file': %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to retrive file from request (is it missing?)"))
		return
	}
	defer uploadedFile.Close()

	filename, size := handler.Filename, handler.Size
	contentType := atIndexOr(0, "application/octet-stream", handler.Header["Content-Type"])

	id, err := c.saveFileRecord(filename, size, contentType)
	if err != nil {
		log.Printf("Failed to save file record: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer func() {
		if err != nil {
			// a error ocurred after the record was inserted!
			// so we need to remove it
			c.removeFileRecord(id)
		}
	}()

	bytes, err := io.ReadAll(uploadedFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Failed to read uploaded file: %s\n", err)
		return
	}

	filePath := fmt.Sprintf("%s/%s", *c.StoragePath, id)
	err = os.WriteFile(filePath, bytes, 0644)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Failed to save file to server: %s\n", err)
		return
	}

	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}

	shortId := toBase62(id)
	fmt.Fprintf(w, "%s://%s/f/%s", proto, r.Host, shortId)
}

func (c *AppContext) getFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseBase62(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	record, err := c.getFileRecordById(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if record.UploadTimestamp+record.SecondsToLive <= time.Now().UTC().Unix() {
		log.Printf("File %s expired on request, wow!\n", id)
		w.WriteHeader(http.StatusGone)
		return
	}

	savedPath := fmt.Sprintf("%s/%s", *c.StoragePath, record.Id)

	w.Header().Set("Content-Type", record.ContentType)
	http.ServeFile(w, r, savedPath)
}
