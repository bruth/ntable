# NTable

[![GoDoc](https://godoc.org/github.com/bruth/ntable?status.svg)](http://godoc.org/github.com/bruth/ntable)

A table data structure for NATS Streaming.

The primary use case for maintaining a lookup table of data derived from a stream. It does this by maintaining an internal key-value store. The default implementation uses an in-memory store, but stores are easy to implement since the interface is a basic key-value interface.

```go
type Store interface {
  Get(key []byte) ([]byte, error)
  Set(key, val []byte) error
  Del(key []byte) error
}
```

## Status

Design phase. Suggestions welcome.

## Example

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bruth/ntable"
	"github.com/nats-io/go-nats-streaming"
)

// Record with an ID and some data.
type Record struct {
	ID   string
	Data string
}

func main() {
	conn, _ := stan.Connect("test-cluster", "test-client")

	t := &ntable.Table{
		Conn:    conn,
		Channel: "users",
		// Takes a message, derives the key and value and updates the store.
		Handle: func(s ntable.Store, m *stan.Msg) {
			var r Record
			json.Unmarshal(m.Data, &r)
			s.Set([]byte(r.ID), []byte(r.Data))
		},
	}

	// Open connection (subscription) to channel.
	t.Open()
	defer t.Close()

	// Publish a record about "pam" (likely happening from another process or thread).
	b, _ := json.Marshal(&Record{
		ID:   "pam",
		Data: "color=blue city=Philadelphia food=sushi",
	})
	conn.Publish("users", b)

	// A few moments pass..
	time.Sleep(5 * time.Millisecond)

	// We can get the data about pam using the table.
	val, _ := t.Get([]byte("pam"))

	fmt.Println(string(val))
}
```

## License

MIT.
