package ntable

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/nats-io/go-nats-streaming"
)

const hexChars = "abcdefgh012345678"

func randChannel(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = hexChars[rand.Intn(len(hexChars))]
	}
	return string(b)
}

func marshalData(t *testing.T, v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func unmarshalData(t *testing.T, b []byte, v interface{}) {
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}

func publishMsg(t *testing.T, conn stan.Conn, ch string, v interface{}) {
	if err := conn.Publish(ch, marshalData(t, v)); err != nil {
		t.Fatal(err)
	}
}

type kvPair struct {
	Key   string
	Value string
}

func expectValue(t *testing.T, tb *Table, key, exp string) {
	act, err := tb.Get([]byte(key))
	if err != nil {
		t.Fatal(err)
	}
	if string(act) != exp {
		t.Errorf("expected %s, got %s", exp, string(act))
	}
}

func expectNothing(t *testing.T, tb *Table, key string) {
	act, err := tb.Get([]byte(key))
	if err == ErrNotFound {
		return
	} else if err != nil {
		t.Fatal(err)
	}
	t.Errorf("expected nothing, got %s", string(act))
}

func TestTable(t *testing.T) {
	conn, err := stan.Connect(
		"test-cluster",
		"test-client",
		stan.NatsURL("nats://localhost:4222"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	channel := randChannel(8)

	tb := &Table{
		Conn:    conn,
		Channel: channel,
		Handle: func(s Store, m *stan.Msg) {
			var kv kvPair
			unmarshalData(t, m.Data, &kv)
			if err := s.Set([]byte(kv.Key), []byte(kv.Value)); err != nil {
				t.Fatal(err)
			}
		},
	}

	if err := tb.Open(); err != nil {
		t.Error(err)
		return
	}
	defer tb.Close()

	publishMsg(t, conn, channel, &kvPair{
		Key:   "bob",
		Value: "blue",
	})

	publishMsg(t, conn, channel, &kvPair{
		Key:   "pam",
		Value: "red",
	})

	time.Sleep(5 * time.Millisecond)
	expectValue(t, tb, "bob", "blue")
	expectValue(t, tb, "pam", "red")
	expectNothing(t, tb, "joe")

	publishMsg(t, conn, channel, &kvPair{
		Key:   "bob",
		Value: "green",
	})

	time.Sleep(5 * time.Millisecond)
	expectValue(t, tb, "bob", "green")

	publishMsg(t, conn, channel, &kvPair{
		Key:   "joe",
		Value: "purple",
	})

	time.Sleep(5 * time.Millisecond)
	expectValue(t, tb, "joe", "purple")
}
