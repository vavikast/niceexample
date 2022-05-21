package main

import (
	"github.com/Shopify/sarama"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	hostinfo := "52.131.238.242:9092"
	topic := "newyearne"

	Partitions(hostinfo, topic)
}

func Partitions(hostinfo, topic string) {
	var consumerhost []string
	consumerhost = append(consumerhost, hostinfo)
	config := sarama.NewConfig()

	config.Producer.Return.Successes = true

	client, err := sarama.NewClient(consumerhost, config)
	if err != nil {
		log.Fatal("client 连接错误")
	}
	defer client.Close()

	producer, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		log.Fatal("NewAsyncProducer 错误")
	}
	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var (
		wg                          sync.WaitGroup
		enqueued, successes, errors int
	)

	go func() {
		for {
			select {
			case <-producer.Successes():
				successes++
			case <-producer.Errors():
				errors++
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range producer.Successes() {
			successes++
		}
	}()

	wg.Add(1)
	// start a groutines to count error num
	go func() {
		defer wg.Done()
		for err := range producer.Errors() {
			log.Println(err)
			errors++
		}
	}()
	//go func(p sarama.AsyncProducer) {
	//	for{
	//		select {
	//		case  <-producer.Successes():
	//			//fmt.Println("offset: ", suc.Offset, "timestamp: ", suc.Timestamp.String(), "partitions: ", suc.Partition)
	//		case fail := <-producer.Errors():
	//			fmt.Println("err: ", fail.Err)
	//		}
	//	}
	//}

ProducerLoop:
	for {
		time11 := time.Now()
		value := "this is a message 0606 " + time11.Format("15:04:05")
		message := &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(value)}
		select {
		case producer.Input() <- message:
			enqueued++
			time.Sleep(time.Second)
		case <-signals:
			producer.AsyncClose() // Trigger a shutdown of the producer.
			break ProducerLoop
		}
	}

	wg.Wait()
	log.Printf("Successfully produced: %d; errors: %d\n", successes, errors)
}
