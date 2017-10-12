package ntable

import (
	"errors"

	"github.com/nats-io/go-nats-streaming"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	Set(key, val []byte) error
	Get(key []byte) ([]byte, error)
	Del(key []byte) error
}

// HandleFunc takes a store and the message and performs the update to the table.
type HandleFunc func(Store, *stan.Msg)

type Table struct {
	Conn        stan.Conn
	Channel     string
	Store       Store
	Handle      HandleFunc
	DurableName string

	sub stan.Subscription
}

func (t *Table) Open() error {
	if t.sub != nil {
		return errors.New("Table already open")
	}

	if t.Handle == nil {
		return errors.New("Handle required")
	}

	// Default to in memory store.
	if t.Store == nil {
		t.Store = NewMemStore()
	}

	opts := []stan.SubscriptionOption{
		stan.DeliverAllAvailable(),
	}
	if t.DurableName != "" {
		opts = append(opts, stan.DurableName(t.DurableName))
	}

	sub, err := t.Conn.Subscribe(
		t.Channel,
		func(msg *stan.Msg) {
			t.Handle(t.Store, msg)
		},
		opts...,
	)
	if err != nil {
		return err
	}

	t.sub = sub

	return nil
}

func (t *Table) Close() error {
	return t.sub.Close()
}

func (t *Table) Unsubscribe() error {
	return t.sub.Unsubscribe()
}

func (t *Table) Get(key []byte) ([]byte, error) {
	return t.Store.Get(key)
}
