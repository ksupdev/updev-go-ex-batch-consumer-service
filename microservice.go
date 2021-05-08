package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// IMicroservice is interface for centralized service management
type IMicroservice interface {
	Start() error
	Stop()
	Cleanup() error
	Log(message string)

	// Bulk Consumer Services
	ConsumeBatch(servers string, topic string, groupID string, batchSize int, batchTimeout time.Duration, h ServiceHandleFunc) error
}

// ServiceHandleFunc is the handler for each Microservice
type ServiceHandleFunc func(ctx IContext) error

// Microservice struct

// Microservice is the centralized service management
type Microservice struct {
	exitChannel chan bool
}

// NewMicroservice is the constructor function of Microservice
func NewMicroservice() *Microservice {
	return &Microservice{}
}

// Kafka service

// Start start all registered services
func (ms *Microservice) Start() error {
	// There are 2 ways to exit from Microservices
	// 1. The SigTerm can be send from outside program such as from k8s
	// 2. Send true to ms.exitChannel
	osQuit := make(chan os.Signal, 1)
	ms.exitChannel = make(chan bool, 1)
	signal.Notify(osQuit, syscall.SIGTERM, syscall.SIGINT)
	exit := false
	for {
		if exit {
			break
		}
		select {
		case <-osQuit:
			// if exitHTTP != nil {
			// 	exitHTTP <- true
			// }
			exit = true
		case <-ms.exitChannel:
			// if exitHTTP != nil {
			// 	exitHTTP <- true
			// }
			exit = true
		}
	}

	return nil
}

// Stop stop the services
func (ms *Microservice) Stop() {
	if ms.exitChannel == nil {
		return
	}
	ms.exitChannel <- true
}

// Cleanup clean resources up from every registered services before exit
func (ms *Microservice) Cleanup() error {
	return nil
}

// Log log message to console
func (ms *Microservice) Log(tag string, message string) {
	fmt.Println(tag+": ", message)
}

// run with goroutine by ConsumeBatch func
func (ms *Microservice) consumeBatch(
	servers string,
	topic string,
	groupID string,
	readTimeout time.Duration,
	batchSize int,
	batchTimeout time.Duration,
	h ServiceHandleFunc) error {

	// Batch Filler
	fill := func(b *Batch, payload interface{}) error {
		p := payload.(string)
		b.Add(p)
		return nil
	}

	// Batch Executer
	exec := func(b *Batch) error {
		messages := make([]string, 0)
		for {
			item := b.Read()
			if item == nil {
				break
			}
			message := item.(string)
			messages = append(messages, message)
		}

		if len(messages) == 0 {
			return nil
		}

		// Execute Handler
		h(NewBatchConsumerContext(ms, messages))
		return nil
	}

	// Payloads loader
	payload := make(chan interface{})
	quit := make(chan bool, 1)

	go func() {

		c, err := ms.newKafkaConsumer(servers, groupID)
		if err != nil {
			quit <- true
			return
		}

		// This will close kafka consumer before exit function
		defer c.Close()

		c.Subscribe(topic, nil)

		for {

			if readTimeout <= 0 {
				// readtimeout -1 indicates no timeout
				readTimeout = -1
			}

			msg, err := c.ReadMessage(readTimeout)
			if err != nil {
				kafkaErr, ok := err.(kafka.Error)
				if ok {
					if kafkaErr.Code() == kafka.ErrTimedOut {
						if readTimeout == -1 {
							// No timeout just continue to read message again
							continue
						}
					}
				}
				return
			}
			fmt.Println("--Subscribe data from kafka- " + string(msg.Value))
			message := string(msg.Value)
			payload <- message
		}
	}()

	go func() {
		// Gracefull shutdown routine
		osQuit := make(chan os.Signal, 1)
		signal.Notify(osQuit, syscall.SIGTERM, syscall.SIGINT)

		select {
		case <-quit:
			close(payload)
		case <-osQuit:
			close(payload)
		}
	}()

	// Error listener
	errc := make(chan error)
	defer close(errc)
	go func() {
		for err := range errc {
			if err != nil {
				ms.Log("BatchConsumer", err.Error())
			}
		}
	}()

	be := NewBatchEvent(batchSize, batchTimeout, fill, exec, payload, errc)
	be.Start()

	return nil
}

// ConsumeBatch register service endpoint for Batch Consumer service
func (ms *Microservice) ConsumeBatch(
	servers string,
	topic string,
	groupID string,
	readTimeout time.Duration,
	batchSize int,
	batchTimeout time.Duration,
	h ServiceHandleFunc) error {

	go ms.consumeBatch(servers, topic, groupID, readTimeout, batchSize, batchTimeout, h)
	return nil
}

// Kafka consumer

// newKafkaConsumer create new Kafka consumer
func (ms *Microservice) newKafkaConsumer(servers string, groupID string) (*kafka.Consumer, error) {
	// Configurations
	// https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	config := &kafka.ConfigMap{

		// Alias for metadata.broker.list: Initial list of brokers as a CSV list of broker host or host:port.
		// The application may also use rd_kafka_brokers_add() to add brokers during runtime.
		"bootstrap.servers": servers,

		// Client group id string. All clients sharing the same group.id belong to the same group.
		"group.id": groupID,

		// Action to take when there is no initial offset in offset store or the desired offset is out of range:
		// 'smallest','earliest' - automatically reset the offset to the smallest offset,
		// 'largest','latest' - automatically reset the offset to the largest offset,
		// 'error' - trigger an error which is retrieved by consuming messages and checking 'message->err'.
		"auto.offset.reset": "earliest",

		// Protocol used to communicate with brokers.
		// plaintext, ssl, sasl_plaintext, sasl_ssl
		"security.protocol": "plaintext",

		// Automatically and periodically commit offsets in the background.
		// Note: setting this to false does not prevent the consumer from fetching previously committed start offsets.
		// To circumvent this behaviour set specific start offsets per partition in the call to assign().
		"enable.auto.commit": true,

		// The frequency in milliseconds that the consumer offsets are committed (written) to offset storage. (0 = disable).
		// default = 5000ms (5s)
		// 5s is too large, it might cause double process message easily, so we reduce this to 200ms (if we turn on enable.auto.commit)
		"auto.commit.interval.ms": 200,

		// Automatically store offset of last message provided to application.
		// The offset store is an in-memory store of the next offset to (auto-)commit for each partition
		// and cs.Commit() <- offset-less commit
		"enable.auto.offset.store": true,

		// Enable TCP keep-alives (SO_KEEPALIVE) on broker sockets
		"socket.keepalive.enable": true,
	}

	kc, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, err
	}
	return kc, err
}
