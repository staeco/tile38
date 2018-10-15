package endpoint

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

const (
	kafkaExpiresAfter = time.Second * 30
)

// KafkaConn is an endpoint connection
type KafkaConn struct {
	mu   sync.Mutex
	ep   Endpoint
	conn sarama.SyncProducer
	ex   bool
	t    time.Time
}

// Expired returns true if the connection has expired
func (conn *KafkaConn) Expired() bool {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	if !conn.ex {
		if time.Now().Sub(conn.t) > kafkaExpiresAfter {
			if conn.conn != nil {
				conn.close()
			}
			conn.ex = true
		}
	}
	return conn.ex
}

func (conn *KafkaConn) close() {
	if conn.conn != nil {
		conn.conn.Close()
		conn.conn = nil
	}
}

// Send sends a message
func (conn *KafkaConn) Send(msg string) error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.ex {
		return errExpired
	}
	conn.t = time.Now()

	uri := fmt.Sprintf("%s:%d", conn.ep.Kafka.Host, conn.ep.Kafka.Port)
	if conn.conn == nil {
		cfg := sarama.NewConfig()
		cfg.Net.DialTimeout = time.Second
		cfg.Net.ReadTimeout = time.Second * 5
		cfg.Net.WriteTimeout = time.Second * 5
		// Fix #333 : fix backward incompatibility introduced by sarama library
		cfg.Producer.Return.Successes = true

		c, err := sarama.NewSyncProducer([]string{uri}, cfg)
		if err != nil {
			return err
		}

		conn.conn = c
	}

	message := &sarama.ProducerMessage{
		Topic: conn.ep.Kafka.QueueName,
		Value: sarama.StringEncoder(msg),
	}

	_, offset, err := conn.conn.SendMessage(message)
	if err != nil {
		conn.close()
		return err
	}

	if offset < 0 {
		conn.close()
		return errors.New("invalid kafka reply")
	}

	return nil
}

func newKafkaConn(ep Endpoint) *KafkaConn {
	return &KafkaConn{
		ep: ep,
		t:  time.Now(),
	}
}
