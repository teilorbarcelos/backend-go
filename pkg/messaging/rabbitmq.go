package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/teilorbarcelos/backend-go/pkg/config"
)

var RabbitConn *amqp.Connection
var RabbitChannel *amqp.Channel

func ConnectRabbitMQ() {
...
	log.Println("Conexão com RabbitMQ estabelecida com sucesso.")
}

func Publish(queueName string, body interface{}) error {
	if RabbitChannel == nil {
		return nil // Ou retorne um erro se for mandatório
	}

	q, err := RabbitChannel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return RabbitChannel.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		})
}
