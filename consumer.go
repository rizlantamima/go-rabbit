package main

import (
	"log"
	"net/url"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {

	rabbitMqDsnUrl := url.URL{
		Scheme: "amqp",
		Host:   "localhost:5672",
		User:   url.UserPassword("guest", "guest"),
	}
	conn, err := amqp.Dial(rabbitMqDsnUrl.String())
	if err != nil {
		log.Panicf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("Failed to open a channel : %s", err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		"baru", // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Panicf("Failed to register a consumer : %s", err)
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}
