package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

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
	conn, err := amqp.Dial(rabbitMqDsnUrl.String())
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel : %s", err)
	}
	defer ch.Close()
	queueList, err := getList()
	if err != nil {
		log.Fatalf("Failed to get the queue lists : %s", err)
	}

	if len(queueList) == 0 {
		log.Fatalf("Sorry, there's no available queue, on your rabbit mq server")
	}

	color.Cyan("Which queue you want to listen ? 1-%v : ", len(queueList))

	for i, queue := range queueList {
		color.Cyan("%v. %s", i+1, queue.Name)
	}

	var confirmationQuestion int
	fmt.Scan(&confirmationQuestion)
	if confirmationQuestion == 0 || confirmationQuestion > len(queueList) {
		log.Fatalf("You are inputting wrong number, please input between 1-%v", len(queueList))
	}
	queueChoosen := queueList[confirmationQuestion-1]

	color.Cyan("Trying to listen queue %s ", queueChoosen.Name)

	msgs, err := ch.Consume(
		queueChoosen.Name, // queue
		"",                // consumer
		true,              // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer : %s", err)
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

func getList() (queues []amqp.Queue, err error) {

	rabbit_mq_host := os.Getenv("RABBIT_MQ_HOST")
	rabbit_mq_api_port := os.Getenv("RABBIT_MQ_API_PORT")
	rabbit_mq_auth_username := os.Getenv("RABBIT_MQ_AUTH_USERNAME")
	rabbit_mq_auth_password := os.Getenv("RABBIT_MQ_AUTH_PASSWORD")
	dsn := url.URL{
		Scheme: "http",
		Host:   rabbit_mq_host + ":" + rabbit_mq_api_port,
		User:   url.UserPassword(rabbit_mq_auth_username, rabbit_mq_auth_password),
		Path:   "api/queues",
	}

	// Set up the HTTP client
	client := &http.Client{}
	// Set up the HTTP request to get the list of queues
	req, err := http.NewRequest("GET", dsn.String(), nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	// Set the authentication credentials for the request
	// password, _ := dsn.User.Password()
	// req.SetBasicAuth(dsn.User.Username(), password)

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return nil, err
	}

	// Unmarshal the response body into a slice of Queue structs
	err = json.Unmarshal(body, &queues)
	if err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return nil, err
	}
	return queues, nil
}
