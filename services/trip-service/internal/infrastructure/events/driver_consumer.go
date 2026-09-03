package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Sed-Miyuki/OmniRoute/services/trip-service/internal/domain"
	"github.com/Sed-Miyuki/OmniRoute/shared/contracts"
	"github.com/Sed-Miyuki/OmniRoute/shared/messaging"
	pbd "github.com/Sed-Miyuki/OmniRoute/shared/proto/driver"
	"github.com/rabbitmq/amqp091-go"
)

type driverConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service domain.TripService
}

func NewDriverConsumer(rabbitmq *messaging.RabbitMQ,service domain.TripService) *driverConsumer {
	return &driverConsumer{
		rabbitmq: rabbitmq,
		service: service,
	}
}

func (c *driverConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.DriverTripResponseQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		var message contracts.AmqpMessage
		if err:=json.Unmarshal(msg.Body,&message);err!=nil{
			log.Printf("failed to unmarshal message: %v",err)
			return err
		}

		var payload messaging.DriverTripResponseData
		if err:=json.Unmarshal(message.Data,&payload);err!=nil{
			log.Printf("failed to unmarshal message: %v",err)
			return err
		}
		log.Printf("driver response received message: %+v",msg)

		switch msg.RoutingKey{
			case contracts.DriverCmdTripAccept:
				if err:=c.handleTripAccepted(ctx,payload.TripID,payload.Driver);err!=nil{
					log.Printf("failed to handle the trip accept: %v",err)
					return err
				}
				return nil
			case contracts.DriverCmdTripDecline:
				if err:=c.handleTripDeclined(ctx,payload.TripID,payload.RiderID);err!=nil{
					log.Printf("failed to handle the trip decline: %v",err)
					return err
				}
				return nil
		}

		log.Printf("unknown trip event: %+v",payload)
		return nil
	})
}

func (c *driverConsumer) handleTripAccepted(ctx context.Context,tripID string,driver *pbd.Driver) error{
	trip,err:=c.service.GetTripByID(ctx,tripID)
	if err!=nil{
		return err
	}

	if trip==nil{
		return fmt.Errorf("Trip was not found: %s",tripID)
	}
	if err:=c.service.UpdateTrip(ctx,tripID,"accepted",driver);err!=nil{
		log.Printf("Failed to update the trip: %v",err)
		return err
	}

	trip,err=c.service.GetTripByID(ctx,tripID)
	if err!=nil{
		return err
	}

	marshalledTrip,err:=json.Marshal(trip)
	if err!=nil{
		return err
	}

	if err:=c.rabbitmq.PublishMessage(ctx,contracts.TripEventDriverAssigned,contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data: marshalledTrip,
	});err!=nil{
		return err
	}

	marshalledPayload,err:=json.Marshal(messaging.PaymentTripResponseData{
		TripID: tripID,
		UserID: trip.UserID,
		DriverID: driver.Id,
		Amount: trip.RideFare.TotalPriceInCents,
		Currency: "USD",
	})
	if err!=nil{
		return fmt.Errorf("error marshalling data: %v",err)
	}

	if err:=c.rabbitmq.PublishMessage(ctx,contracts.PaymentCmdCreateSession,
		contracts.AmqpMessage{
			OwnerID: trip.UserID,
			Data: marshalledPayload,
		},
	);err!=nil{
		return err
	}
	return nil
}

func (c *driverConsumer) handleTripDeclined(ctx context.Context,tripID string,riderID string) error{
	trip,err:=c.service.GetTripByID(ctx,tripID)
	if err!=nil{
		return err
	}

	newPayload:=messaging.TripEventData{
		Trip: trip.ToProto(),
	}

	marshalledPayload,err:=json.Marshal(newPayload)
	if err!=nil{
		return err
	}

	if err:=c.rabbitmq.PublishMessage(ctx,contracts.TripEventDriverNotInterested,contracts.AmqpMessage{
		OwnerID: riderID,
		Data: marshalledPayload,
	});err!=nil{
		return err
	}
	return nil
}