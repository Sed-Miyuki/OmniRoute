package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Sed-Miyuki/OmniRoute/services/api-gateway/grpc_clients"
	"github.com/Sed-Miyuki/OmniRoute/shared/contracts"
	"github.com/Sed-Miyuki/OmniRoute/shared/messaging"
	"github.com/Sed-Miyuki/OmniRoute/shared/proto/driver"
)

var (
	connManager = messaging.NewConnectionManager()
)
 
func handleRidersWebSocket(w http.ResponseWriter, r *http.Request,rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("Web socket upgrade failed: %v", err)
		return
	}

	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Printf("No User id provided")
		return
	}

	connManager.Add(userID, conn)
	defer connManager.Remove(userID)

	queues:=[]string{
		messaging.DriverCmdTripRequestQueue,
		messaging.NotifyDriverAssignQueue,
	}

	for _,q:=range queues{
		consumer:=messaging.NewQueueConsumer(rb,connManager,q)

		if err:=consumer.Start();err!=nil{
			log.Printf("failed to start consumer for queue: %s: err: %v",q,err)
		}
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading message: %v", err)
			break
		}
		log.Printf("received message: %s", message)
	}
}

func handleDriversWebSocket(w http.ResponseWriter, r *http.Request,rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("Web socket upgrade failed: %v", err)
		return
	}

	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Printf("No User id provided")
		return
	}

	connManager.Add(userID, conn)

	packageSlug := r.URL.Query().Get("packageSlug")
	if packageSlug == "" {
		log.Printf("No package slug provided")
		return
	}

	ctx := r.Context()

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		connManager.Remove(userID)
		driverService.Client.UnRegisterDriver(ctx, &driver.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})

		driverService.Close()

		log.Println("Driver unregistered: ", userID)
	}()

	driverData, err := driverService.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})
	if err != nil {
		log.Printf("Error registering driver: %v", err)
		return
	}

	if err := connManager.SendMessage(userID, contracts.WSMessage{
		Type: contracts.DriverCmdRegister,
		Data: driverData.Driver,
	}); err != nil {
		log.Printf("error sending message: %v", err)
		return
	}

	queues:=[]string{
		messaging.DriverCmdTripRequestQueue,
	}

	for _,q:=range queues{
		consumer:=messaging.NewQueueConsumer(rb,connManager,q)

		if err:=consumer.Start();err!=nil{
			log.Printf("failed to start consumer for queue: %s: err: %v",q,err)
		}
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading message: %v", err)
			break
		}

		type driverMessage struct{
			Type string `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		var driverMSG driverMessage

		if err:=json.Unmarshal(message,&driverMSG);err!=nil{
			log.Printf("error unmarshalling messages: %v",err)
			return
		}

		switch driverMSG.Type{
		case contracts.DriverCmdLocation:
			//maybe in future???
			continue
		case contracts.DriverCmdTripAccept,contracts.DriverCmdTripDecline:
			if err:=rb.PublishMessage(ctx,driverMSG.Type,contracts.AmqpMessage{
				OwnerID: userID,
				Data: driverMSG.Data,
			});err!=nil{
				log.Printf("Error publishing message to RabbitMQ: %v",err)
			}
		default:
			log.Printf("Unknown message Type: %s",driverMSG.Type)
		}
	}
}
