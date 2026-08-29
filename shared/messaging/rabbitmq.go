package messaging

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct{
	conn *amqp.Connection
}

func NewRabbitMQ(uri string) (*RabbitMQ,error){
	//Rabbitmq
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil,fmt.Errorf("failed to connect to rabbitmq: %v",err)
	}

	rmq:=&RabbitMQ{
		conn: conn,
	} 
	
	return rmq,nil
}

func (r *RabbitMQ) Close(){
	if r.conn!=nil{
		r.conn.Close()
	}
}