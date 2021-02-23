package main

import (
	"github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

const (
	ClusterID = "test-cluster"
)

type NATSClient struct {
	ClusterID string
	ClientID  string
	NATSHost  string
}

func (c *NATSClient) Publish(payload []byte, channel string) error {
	nc, err := nats.Connect(c.NATSHost)
	if err != nil {
		return err
	}

	sc, err := stan.Connect(c.ClusterID, c.ClientID, stan.NatsConn(nc))
	if err != nil {
		return err
	}
	defer sc.Close()

	if err := sc.Publish(channel, payload); err != nil {
		// TODO: return retriable/non-retriable error
		return err
	}

	return nil
}

func NewNATSClient(host string) *NATSClient {
	return &NATSClient{
		NATSHost:  host,
		ClientID:  "random", // TODO: randomize
		ClusterID: ClusterID,
	}
}
