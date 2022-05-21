package main

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"os"
	"os/signal"
	"sync"
)

type MyConsumerGroupHandler struct {
	name  string
	count int64
}

// Setup 执行在 获得新 session 后 的第一步, 在 ConsumeClaim() 之前
func (MyConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup 执行在 session 结束前, 当所有 ConsumeClaim goroutines 都退出时
func (MyConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

//consumeclaim具体消费逻辑
func (m MyConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("[consumer] name:%s topic:%q partition:%d offset:%d\n", m.name, msg.Topic, msg.Partition, msg.Offset)
		sess.MarkMessage(msg, "")
		m.count++
		if m.count%10000 == 0 {
			fmt.Printf("name:%s 消费数:%v\n", m.name, m.count)
		}
	}
	return nil

}

func main() {
	hostinfo := "52.131.238.242:9092"
	topic := "newyearne"
	group := "kafka-group"

	go ConsumerGroupNew(hostinfo, group, "C1", topic)
	go ConsumerGroupNew(hostinfo, group, "C2", topic)
	ConsumerGroupNew(hostinfo, group, "C3", topic)

}

func ConsumerGroupNew(hostinfo, group, name, topic string) {
	var consumerhost []string
	consumerhost = append(consumerhost, hostinfo)

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cg, err := sarama.NewConsumerGroup(consumerhost, group, config)
	if err != nil {
		log.Fatal("NewConsumerGroup err: ", err)
	}
	defer cg.Close()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		handler := MyConsumerGroupHandler{name: name}
		for {
			fmt.Println("running: ", name)
			err := cg.Consume(ctx, []string{topic}, handler)
			if err != nil {
				log.Println("Consume err: ", err)
			}
			// 如果 context 被 cancel 了，那么退出
			if ctx.Err() != nil {
				return
			}
		}
	}()
	wg.Wait()

}

type consumerGroupHandler struct {
	name string
}

func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("%s Message topic:%q partition:%d offset:%d  value:%s\n", h.name, msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
		// 手动确认消息
		sess.MarkMessage(msg, "")
	}
	return nil
}

func handleErrors(group *sarama.ConsumerGroup, wg *sync.WaitGroup) {
	defer wg.Done()
	for err := range (*group).Errors() {
		fmt.Println("ERROR", err)
	}
}

func ConsumerGroup(hostinfo, group, topic string) {
	var consumerhost []string
	consumerhost = append(consumerhost, hostinfo)

	var wg sync.WaitGroup

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = false
	config.Version = sarama.V0_10_2_0

	client, err := sarama.NewClient(consumerhost, config)
	if err != nil {
		log.Fatal("连接错误 ", err)
	}
	defer client.Close()

	group1, err := sarama.NewConsumerGroupFromClient(group, client)
	if err != nil {
		panic(err)
	}

	defer group1.Close()

	wg.Add(2)
	go consume(&group1, &wg, group)

	wg.Wait()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	select {
	case <-signals:
	}

}

func consume(group *sarama.ConsumerGroup, wg *sync.WaitGroup, name string) {
	fmt.Println(name + "start")
	wg.Done()
	ctx := context.Background()
	for {
		topics := []string{"newyearne"}
		handler := consumerGroupHandler{name: name}
		err := (*group).Consume(ctx, topics, handler)
		if err != nil {
			panic(err)
		}
	}
}

func main1() {
	hostinfo := "52.131.238.242:9092"
	topic := "newyearne"

	go ConsumerGroup(hostinfo, topic, "c1")
	go ConsumerGroup(hostinfo, topic, "c2")

	for {

	}

}
