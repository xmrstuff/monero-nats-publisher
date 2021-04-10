package main

import (
	"encoding/json"
	"fmt"
)

const (
	txCreated         = "transaction.created"
	eventVersion      = "1.0"
	moneroNATSChannel = "monero"
)

type Event struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Data    Tx     `json:"data"`
}

func NewTXCreatedEvent(tx Tx) Event {
	return Event{
		Type:    txCreated,
		Version: eventVersion,
		Data:    tx,
	}
}

type Publisher interface {
	Publish([]byte, string) error
}

type EventPublishing struct {
	Publisher Publisher
}

func (ep *EventPublishing) PushTxEvent(tx Tx) error {
	eventPayload := NewTXCreatedEvent(tx)
	fmt.Printf("Event Payload: %+v", eventPayload)
	jsonPayload, err := json.Marshal(eventPayload)
	if err != nil {
		return err
	}

	if err := ep.Publisher.Publish(jsonPayload, moneroNATSChannel); err != nil {
		// TODO: return retriable/non-retriable error
		return err
	}

	return nil
}

func NewNatsPublishingClient(natsHost string) *EventPublishing {
	return &EventPublishing{
		Publisher: NewNATSClient(natsHost),
	}
}
