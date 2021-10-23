A golang HTTP request multiplexer.

## Install

```bash
go get github.com/nicolasparada/go-mux
```

## Usage

```go
func main() {
  m := mux.New()
  m.HandleFunc("/hello/{name}", helloWorld)
  m.Handle("/*", http.FileServer(http.Dir("static")))

  http.ListenAndServe(":5000", m)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    name := mux.URLParam(ctx, "name")
    fmt.Fprintf(w, "Hello, %s", name)
}
```

Capture URL parameters with `{myParam}` notation. Wildcard `*` are also supported in URL and method too.<br>
And that's it.
