## Caixinha - File storage

Fully anonymous temporary (by default) file upload and sharing website. Written in Go.

### Compilation

```
git clone https://github.com/leleneme/caixinha
cd caixinha
go build .
```

### Usage and configuration

#### Running

By default the service is served at :8080.
```
./caixinha
```
After starting visit localhost:8080 on a web browser for a simple web-based UI.

#### Flags
- --bind
  - Bind address, default ":8080"
- --db-file
  - SQLite database file path, default "./caixinha.db"
- --storage-path
  - Where the files are saved, default (Linux) /tmp/caixinha
- --days-to-live
  - How many days should files be kept saved, a ttl smaller than 0 means the files are never deleted, default 30
- --max-size
  - Max file size in megabytes, default 100
