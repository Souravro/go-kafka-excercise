package main

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type User struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

var (
	brokers             = "0.0.0.0:9092"
	version             = "0.11.0.0"
	topic               = "user_details"
	recordsNumber int64 = 10

	// following will be used to randomly generate json objects
	names  = []string{"Alice", "Bob", "Chris", "Danny", "Eric", "Freud", "Gabriel", "Hardik"}
	phones = []string{"1323231", "123123123", "312545242", "315125524", "312312", "123132"}
	emails = []string{"s@d.com", "r@o.in", "f@rr.con", "g@foo.com", "sd@ss.uk", "koo@voo.in"}
)

func createConfig() *sarama.Config {
	version, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	config := sarama.NewConfig()
	config.Version = version
	config.Producer.Idempotent = true
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.Transaction.Retry.Backoff = 10
	config.Producer.Transaction.ID = "txn_producer"
	config.Net.MaxOpenRequests = 1

	return config
}

func main() {
	log.Println("Starting a new Sarama producer...")
	ctx, cancel := context.WithCancel(context.Background())

	config := createConfig()
	//prodProvider := createProducerProvider(brokers, config)
	producer, err := sarama.NewSyncProducer(strings.Split(brokers, ","), config)
	if err != nil {
		log.Printf("Producer. Error in creating producer. Error: [%v]", producer)
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

func getMessage() User {
	// Randomly create a json object of type User
	rand.Seed(time.Now().Unix())
	return User{
		Name:  names[rand.Intn(len(names))],
		Phone: phones[rand.Intn(len(phones))],
		Email: emails[rand.Intn(len(emails))],
	}
}

func encodeMessage(msg interface{}) []byte {
	val, er := json.Marshal(msg)
	if er != nil {
		log.Printf("Error in converting struct to json. Error: [%v]", val)
	}

	return val

	//buf := bytes.Buffer{}
	//enc := gob.NewEncoder(&buf)
	//err := enc.Encode(msg)
	//if err != nil {
	//	log.Fatalf("Error in converting msg into []byte. Error: [%v]", err)
	//}
	//return buf.Bytes()
}

func produceRecord(producer sarama.SyncProducer) {
	// Start kafka transaction
	err := producer.BeginTxn()
	if err != nil {
		log.Printf("Producer: unable to start txn %s\n", err)
		return
	}

	// Produce records
	for i := int64(0); i < recordsNumber; i++ {
		msgToProduce := getMessage()
		msgBytes := encodeMessage(msgToProduce)
		producerMsg := &sarama.ProducerMessage{Topic: topic, Key: nil, Value: sarama.StringEncoder(msgBytes)}
		partition, offset, er := producer.SendMessage(producerMsg)
		if er != nil {
			log.Printf("Producer. Unable to Send Message to topic. Error: [%v]", er)
			return
		}
		log.Printf("Producer: produced message- [%v]. Partition: [%v]. Offset: [%v]", msgToProduce, partition, offset)
		time.Sleep(500 * time.Millisecond)
	}

	err = producer.CommitTxn()
	if err != nil {
		log.Printf("Producer: unable to commit txn %s\n", err)
		for {
			if producer.TxnStatus()&sarama.ProducerTxnFlagFatalError != 0 {
				// fatal error. need to recreate producer.
				log.Printf("Producer: producer is in a fatal state, need to recreate it")
				break
			}
			// If producer is in abortable state, try to abort current transaction.
			if producer.TxnStatus()&sarama.ProducerTxnFlagAbortableError != 0 {
				err = producer.AbortTxn()
				if err != nil {
					// If an error occurred just retry it.
					log.Printf("Producer: unable to abort transaction: %+v", err)
					continue
				}
				break
			}
			// if not you can retry
			err = producer.CommitTxn()
			if err != nil {
				log.Printf("Producer: unable to commit txn %s\n", err)
				continue
			}
		}
		return
	} else {
		log.Println("Producer: txn successfully committed.")
	}
}
