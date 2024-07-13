package main

import (
	"github.com/google/uuid"
	"log"
	"time"
)

type FileRecord struct {
	Id              uuid.UUID
	Name            string
	Size            int64
	ContentType     string
	UploadTimestamp int64
	SecondsToLive   int64
}

func (c *AppContext) saveFileRecord(name string, size int64, contentType string) (uuid.UUID, error) {
	id := uuid.New()
	uploadTimestamp := time.Now().UTC().Unix()

	sqlFileInsertStmt := "INSERT INTO file_record(id, name, size, content_type, upload_timestamp, seconds_to_live) VALUES (?, ?, ?, ?, ?, ?)"
	stmt, err := c.Db.Prepare(sqlFileInsertStmt)
	if err != nil {
		log.Printf("Failed to create file record insert statement: %s\n", err)
		return id, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, name, size, contentType, uploadTimestamp, 86400 /* 1 day */)
	if err != nil {
		log.Printf("Failed to insert file record: %s\n", err)
		return id, err
	}

	return id, nil
}

func (c *AppContext) removeFileRecord(id uuid.UUID) error {
	_, err := c.Db.Exec("delete from file_record where id = ?", id)
	if err != nil {
		return err
	}

	return nil
}

func (c *AppContext) getFileRecordById(id uuid.UUID) (*FileRecord, error) {
	var result FileRecord

	row := c.Db.QueryRow("select * from file_record where id = ?", id)
	err := row.Scan(&result.Id, &result.Name, &result.Size, &result.ContentType, &result.UploadTimestamp, &result.SecondsToLive)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *AppContext) getFilesToRemove() ([]FileRecord, error) {
	var records []FileRecord

	timeNow := time.Now().UTC().Unix()
	rows, err := c.Db.Query("select * from file_record where upload_timestamp + seconds_to_live <= ? and seconds_to_live > 0;", timeNow)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var result FileRecord
		err = rows.Scan(&result.Id, &result.Name, &result.Size, &result.ContentType, &result.UploadTimestamp, &result.SecondsToLive)
		if err != nil {
			return nil, err
		}

		if result.SecondsToLive != -1 {
			records = append(records, result)
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return records, nil
}
