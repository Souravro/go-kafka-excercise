package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

type User struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

var (
	brokers             = "0.0.0.0:8097"
	topic               = "user_details_2"
	recordsNumber int64 = 10

	// following will be used to randomly generate json objects
	names  = []string{"Alice", "Bob", "Chris", "Danny", "Eric", "Freud", "Gabriel", "Hardik"}
	phones = []string{"1323231", "123123123", "312545242", "315125524", "312312", "123132"}
	emails = []string{"s@d.com", "r@o.in", "f@rr.con", "g@foo.com", "sd@ss.uk", "koo@voo.in"}
)

func createConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	return config
}

func main() {
	log.Println("Starting a new Sarama producer...")
	ctx, cancel := context.WithCancel(context.Background())

	config := createConfig()
	producer, err := sarama.NewSyncProducer(strings.Split(brokers, ","), config)
	if err != nil {
		log.Printf("Producer. Error in creating producer. Error: [%v]", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		default:
			produceRecord(producer)
		}
	}()

	wg.Wait()
	cancel()
}

func getEncodedMessage() []byte {
	// Randomly create a json object of type User
	rand.Seed(time.Now().Unix())
	return encodeMessage(User{
		Name:  names[rand.Intn(len(names))],
		Phone: phones[rand.Intn(len(phones))],
		Email: emails[rand.Intn(len(emails))],
	})
}

func encodeMessage(msg interface{}) []byte {
	val, er := json.Marshal(msg)
	if er != nil {
		log.Printf("Error in converting struct to json. Error: [%v]", val)
	}

	return val
}

func produceRecord(producer sarama.SyncProducer) {
	// Produce records
	for i := int64(0); i < recordsNumber; i++ {
		msgBytes := getEncodedMessage()
		producerMsg := &sarama.ProducerMessage{Topic: topic, Key: nil, Value: sarama.StringEncoder(msgBytes)}
		partition, offset, er := producer.SendMessage(producerMsg)
		if er != nil {
			log.Printf("Producer. Unable to Send Message to topic. Error: [%v]", er)
			return
		}
		log.Printf("Producer: produced message- [%v]. Partition: [%v]. Offset: [%v]", string(msgBytes), partition, offset)
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("Producer: message successfully published.")
}
