package messaging

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/teilorbarcelos/backend-go/pkg/config"
)

var RabbitConn *amqp.Connection
var RabbitChannel *amqp.Channel

func ConnectRabbitMQ() {
	var err error
	RabbitConn, err = amqp.Dial(config.AppConfig.RabbitMQUrl)
	if err != nil {
		log.Fatalf("Falha ao conectar no RabbitMQ: %v", err)
	}

	RabbitChannel, err = RabbitConn.Channel()
	if err != nil {
		log.Fatalf("Falha ao abrir canal no RabbitMQ: %v", err)
	}

	log.Println("Conexão com RabbitMQ estabelecida com sucesso.")
}
