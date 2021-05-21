package main

import (
	"encoding/json"
	"fmt"
)

const (
	txCreated         = "transaction.created"
	blockCreated      = "block.created"
	eventVersion      = "1.0"
	moneroNATSChannel = "monero"
)

type Event struct {
	Type    string      `json:"type"`
	Version string      `json:"version"`
	Data    interface{} `json:"data"`
}

func NewTXCreatedEvent(tx Tx) Event {
	return Event{
		Type:    txCreated,
		Version: eventVersion,
		Data:    tx,
	}
}

func NewBlockCreatedEvent(b Block) Event {
	return Event{
		Type:    blockCreated,
		Version: eventVersion,
		Data:    b,
	}
}

type Publisher interface {
	Publish([]byte, string) error
	IsConnected() bool
}

type EventPublishing struct {
	Publisher Publisher
}

func (ep *EventPublishing) IsConnected() bool {
	return ep.Publisher.IsConnected()
}

func (ep *EventPublishing) PushEvent(ev interface{}) error {
	fmt.Println(fmt.Sprintf("Event Payload: %+v", ev))
	jsonPayload, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	if err := ep.Publisher.Publish(jsonPayload, moneroNATSChannel); err != nil {
		// TODO: return retriable/non-retriable error
		return err
	}

	return nil
}

func (ep *EventPublishing) PushTxEvent(tx Tx) error {
	eventPayload := NewTXCreatedEvent(tx)
	return ep.PushEvent(eventPayload)
}

func (ep *EventPublishing) PushBlockEvent(b Block) error {
	ev := NewBlockCreatedEvent(b)
	return ep.PushEvent(ev)
}

func NewNatsPublishingClient(natsHost string) *EventPublishing {
	return &EventPublishing{
		Publisher: NewNATSClient(natsHost),
	}
}
