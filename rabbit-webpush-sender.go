package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/streadway/amqp"
)

// 
const (
	vapidPublicKey  = "BASCPGpcH6qCEcGlGjww_RWkz016t2LVbrRT4gA3h5KrlwjfEpGEECJWdJleZdnG9ngEQ-Qwn0vKNjuln7USv18"
	vapidPrivateKey = "BLPerEX8aCQec1_Zzj5jp-EGsstW6MNTgL_bcv-fDh8"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
  // connect to rabbitmq 
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

  // prefeach jobs
	err = ch.Qos(
		10,   // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		"webpush", // queue
		"",     // consumer
		false,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {

			start := time.Now()

			// Get webpush subscriber from queue
			s := &webpush.Subscription{}
			json.Unmarshal([]byte(d.Body), s)

			// Send Notification
			_, err := webpush.SendNotification([]byte("Payload!"), s, &webpush.Options{
				Subscriber:      "fabian.lenczewski@gmailc.om", // Do not include "mailto:"
				VAPIDPublicKey:  vapidPublicKey,
				VAPIDPrivateKey: vapidPrivateKey,
				TTL:             10,
			})

			if err != nil {
				fmt.Println(err)
			}

			d.Ack(false)
			log.Printf("Received a message and time took %s", time.Since(start))
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
