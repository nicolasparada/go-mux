# Mux

A golang HTTP request multiplexer.

## Install

```bash
go get github.com/nicolasparada/go-router
```

## Usage

**Simple**: familiar API using `http.Handler` and `http.HandlerFunc` interfaces.
```go
func main() {
  r := mux.NewRouter()
  r.Handle("/test", test)

  http.ListenAndServe(":5000", r)
}
```

**URL param**: capture URL parameters with `{myParam}` notation, and access
 them with `mux.URLParam(ctx, "myParam")`.
```go
func main() {
  r := mux.NewRouter()
  r.HandleFunc("/hello/{name}", hello)

  http.ListenAndServe(":5000", r)
}

func hello(w http.ResponseWriter, r *http.Request) {
    name := mux.URLParam(r.Context(), "name")
    fmt.Fprintf(w, "Hello, %s", name)
}
```

**Wildcard**: match anything with `*`.
```go
func main() {
  r := mux.NewRouter()
  r.Handle("/*", http.FileServer(http.FS(static)))

  http.ListenAndServe(":5000", r)
}
```

**REST**: mux by HTTP method using `mux.MethodHandler`.
 It will respond with `405` `Method Not Allowed` for you if none match.
```go
func main() {
  r := mux.NewRouter()
  r.Handle("/api/todos", mux.MethodHandler{
    http.MethodPost: createTodo,
    http.MethodGet:  todos,
  })
  r.Handle("/api/todos/{todoID}", mux.MethodHandler{
    http.MethodGet:    todo,
    http.MethodPatch:  updateTodo,
    http.MethodDelete: deleteTodo,
  })

  http.ListenAndServe(":5000", r)
}
```
