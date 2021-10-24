[![Go Reference](https://pkg.go.dev/badge/github.com/nicolasparada/go-mux.svg)](https://pkg.go.dev/github.com/nicolasparada/go-mux)

# Mux

A golang HTTP request multiplexer.

## Install

```bash
go get github.com/nicolasparada/go-mux
```

## Usage

 - **Simple**: familiar API using `http.Handler` and `http.HandlerFunc` interfaces.
```go
func main() {
  m := mux.New()
  m.Handle("/test", testHandler)

  http.ListenAndServe(":5000", m)
}
```

 -  **URL param**: capture URL parameters with `{myParam}` notation, and access
 them with `mux.URLParam(ctx, "myParam")`.
```go
func main() {
  m := mux.New()
  m.HandleFunc("/hello/{name}", helloHandler)

  http.ListenAndServe(":5000", m)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    name := mux.URLParam(ctx, "name")
    fmt.Fprintf(w, "Hello, %s", name)
}
```

 - **Wildcard**: match anything with `*`.
```go
func main() {
  m := mux.New()
  m.Handle("/*", http.FileServer(http.Dir("static")))

  http.ListenAndServe(":5000", m)
}
```

 - **REST**: mux by HTTP method using `mux.MethodHandler`.
 It will respond with `405` `Method Not Allowed` for you if none match.
```go
func main() {
  m := mux.New()
  m.Handle("/api/todos", mux.MethodHandler{
    http.MethodPost: createTodoHandler,
    http.MethodGet:  todosHandler,
  })
  m.Handle("/api/todos/{todoID}", mux.MethodHandler{
    http.MethodGet:    todoHandler,
    http.MethodPatch:  updateTodoHandler,
    http.MethodDelete: deleteTodoHandler,
  })

  http.ListenAndServe(":5000", m)
}
```
