package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	rabbit_mq_host := os.Getenv("RABBIT_MQ_HOST")
	rabbit_mq_port := os.Getenv("RABBIT_MQ_PORT")
	rabbit_mq_auth_username := os.Getenv("RABBIT_MQ_AUTH_USERNAME")
	rabbit_mq_auth_password := os.Getenv("RABBIT_MQ_AUTH_PASSWORD")

	rabbitMqDsnUrl := url.URL{
		Scheme: "amqp",
		Host:   rabbit_mq_host + ":" + rabbit_mq_port,
		User:   url.UserPassword(rabbit_mq_auth_username, rabbit_mq_auth_password),
	}
	color.Cyan("Dialing rabbit mq connection to : %s", rabbitMqDsnUrl.String())

	rabiitConnection, err := amqp.Dial(rabbitMqDsnUrl.String())
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	color.Cyan("rabbit mq connected ! \n\n\n")
	defer rabiitConnection.Close()

	rabbitChannel, err := rabiitConnection.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel %s", err)
	}
	defer rabbitChannel.Close()

	q, err := getQueue(rabbitChannel)

	if err != nil {
		log.Fatalf("Failed to declare a queue : %s", err)
	}

	timeOurPublish := 1 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeOurPublish)
	defer cancel()

	msg, _ := getMessage(q)

	err = rabbitChannel.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		msg,
	)

	if err != nil {
		log.Fatalf("Failed to publish a message : %s", err)
	}
	log.Printf(" [x] Sent %s\n", string(msg.Body))

}

func getMessage(queue amqp.Queue) (amqp.Publishing, error) {
	color.Cyan("\n\n==========================================")
	color.Cyan("Message Declaration")
	color.Cyan("==========================================\n")
	color.Green("Whats message you want to sent into %s ? ", queue.Name)
	var body string

	_, err := fmt.Scanf("%s", &body)
	if err != nil {
		fmt.Println("Error:", err)
		return amqp.Publishing{}, err
	}

	fmt.Println(body)

	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(body),
	}

	return msg, nil
}
func getQueue(channel *amqp.Channel) (amqp.Queue, error) {
	color.Cyan("==========================================")
	color.Cyan("Queue declaration")
	color.Cyan("==========================================")
	color.Green("Please input the queue name : ")

	var queueName string
	fmt.Scan(&queueName)
	var queueDurable, queueAutoDelete, queueExclusive, queueNoWait bool
	queueDurable = askQueueDurable()
	queueAutoDelete = askQueueAutoDelete()
	queueExclusive = askQueueExclusive()

	q, err := channel.QueueDeclare(
		queueName,       // name
		queueDurable,    // durable
		queueAutoDelete, // delete when unused
		queueExclusive,  // exclusive
		queueNoWait,     // no-wait
		nil,             // arguments
	)

	color.Cyan("\nThis is your Queue summary : ")
	color.Cyan("\nName : %s", q.Name)
	color.Cyan("\nDurable : %v", queueDurable)
	color.Cyan("\nDurable : %v", queueDurable)
	color.Cyan("\nAutoDelete : %v", queueAutoDelete)
	color.Cyan("\nExclusive : %v\n\n", queueExclusive)

	var confirmationQuestion string
	color.Green("Sounds good ? y / n (retry) [n] : ")
	fmt.Scan(&confirmationQuestion)
	confirmationQuestion = strings.ToUpper(strings.TrimSpace(confirmationQuestion))
	if confirmationQuestion == "Y" {
		return q, err
	}
	return getQueue(channel)
}

func askQueueDurable() bool {
	var confirmationQuestion string
	color.Green("Is the queue are durable ? y / n / h (help)   [n] : ")
	fmt.Scan(&confirmationQuestion)
	confirmationQuestion = strings.ToUpper(strings.TrimSpace(confirmationQuestion))
	if confirmationQuestion == "H" {
		color.Yellow("A durable queue is a type of queue that is stored persistently on disk \nand will persist even after the RabbitMQ server is shut down or restarted.\n\nMeanwhile, non-durable queues are only stored in memory\nand will be deleted when the RabbitMQ server is shut down or restarted.")
		color.Yellow("\nSo with that condition\n\n")
		return askQueueDurable()
	} else if confirmationQuestion == "Y" {
		return true
	}
	return false
}

func askQueueAutoDelete() bool {
	var confirmationQuestion string
	color.Green("Is it an Auto-delete queue ? y / n / h (help)   [n] : ")
	fmt.Scan(&confirmationQuestion)
	confirmationQuestion = strings.ToUpper(strings.TrimSpace(confirmationQuestion))
	if confirmationQuestion == "H" {
		color.Yellow("Auto-delete queue is a type of queue that will be deleted automatically when there are no consumers (consumers) connected to the queue.")
		color.Yellow("\nSo with that condition\n\n")
		return askQueueAutoDelete()
	} else if confirmationQuestion == "Y" {
		return true
	}
	return false
}

func askQueueExclusive() bool {
	var confirmationQuestion string
	color.Green("Is it an exclusive queue ? y / n / h (help)   [n] : ")
	fmt.Scan(&confirmationQuestion)
	confirmationQuestion = strings.ToUpper(strings.TrimSpace(confirmationQuestion))
	if confirmationQuestion == "H" {
		color.Yellow("Exclusive queue is a type of queue that can only be accessed by one connection at a time. Exclusive queues are usually used in situations where we want to ensure that only one consumer can consume messages from that queue.")
		color.Yellow("\nSo with that condition\n\n")
		return askQueueExclusive()
	} else if confirmationQuestion == "Y" {
		return true
	}
	return false
}
