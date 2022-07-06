# storagefile

Fast structured storage and retrieval file, this allows you to save the json/file data for user input.

- `StorageFile` has 1 writer and `MAXREADERS=5` concurrent readers
- Ability to `Save()` `type` string and `data` bytes
- Ability to `Get(int64)` the above
- Ability to `GetHeader(int64)` for the item saved
  - this will retrieve `id, date, datalength` for the saved item

# example

```go
storagefile.MAXREADERS=10 // set 10 readers (default = 5)
sf, _ := storagefile.Open("docs.dat")
defer  sf.Close()

// doc save chi middleware
app.Use(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// skip non json
		skip := true
		if r.Header.Get("Content-Type") == "application/json" {
			skip = false
		}
		if skip == false && (r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE") {
			b, _ := io.ReadAll(r.Body)
			sf.Save(r.Method+"|"+r.URL.Path, b)
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewReader(b))
		}
		next.ServeHTTP(w, r)
	}
})

// get the 10th record in the StroageFile
ty, by, err := sf.Get(int64(10))
if err != nil {
    log.Error(err)
}
fmt.Println("type=", ty)
fmt.Println("bytes=", by)

```

# Performance

- running on AMD Ryzen 5500U "performance mode" ~ 640,000 json string saves/sec
