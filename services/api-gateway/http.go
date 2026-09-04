package main

import (
	// "bytes"

	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/Sed-Miyuki/OmniRoute/services/api-gateway/grpc_clients"
	"github.com/Sed-Miyuki/OmniRoute/shared/contracts"
	"github.com/Sed-Miyuki/OmniRoute/shared/env"
	"github.com/Sed-Miyuki/OmniRoute/shared/messaging"
	"github.com/Sed-Miyuki/OmniRoute/shared/tracing"
)

var tracer = tracing.GetTracer("api-gateway")

func handleTripPreview(w http.ResponseWriter, r *http.Request) {

	ctx,span:=tracer.Start(r.Context(),"handleTripPreview")
	defer span.End()

	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// validation
	if reqBody.UserID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	// jsonBody, _ := json.Marshal(reqBody)
	// reader := bytes.NewReader(jsonBody)

	// // TODO: Call trip service
	// resp, err := http.Post("http://trip-service:8083/preview", "application/json", reader)
	// if err != nil {
	// 	log.Print(err)
	// 	return
	// }

	// defer resp.Body.Close()

	// var respBody any
	// if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
	// 	http.Error(w, "failed to parse JSON data from trip service", http.StatusBadRequest)
	// 	return
	// }

	//response := contracts.APIResponse{Data: respBody}

	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	defer tripService.Close()

	tripPreview, err := tripService.Client.PreviewTrip(ctx, reqBody.ToProto())
	if err != nil {
		log.Printf("failed to preview a trip: %v", err)
		http.Error(w, "failed to preview trip", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: tripPreview}

	writeJSON(w, http.StatusCreated, response)
}

func handleTripStart(w http.ResponseWriter, r *http.Request) {
	ctx,span:=tracer.Start(r.Context(),"handleTripStart")
	defer span.End()

	var reqBody startTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON payload", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// Why we need to create a new client for each connection:
	// because if a service is down, we don't want to block the whole application
	// so we create a new client for each connection
	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer tripService.Close()

	trip, err := tripService.Client.CreateTrip(ctx, reqBody.toProto())
	if err != nil {
		log.Printf("Failed to start a trip: %v", err)
		http.Error(w, "Failed to start trip", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: trip}
	writeJSON(w, http.StatusCreated, response)
}

func handlePayPalWebhook(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	ctx,span:=tracer.Start(r.Context(),"handlePayPalWebhook")
	defer span.End()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	webhookID := env.GetString("PAYPAL_WEBHOOK_ID", "")
	if webhookID == "" {
		log.Printf("PAYPAL_WEBHOOK_ID is required")
		http.Error(w, "Webhook ID missing", http.StatusInternalServerError)
		return
	}

	var event struct {
		EventType string `json:"event_type"`
		Resource  struct {
			ID       string `json:"id"`
			Status   string `json:"status"`
			CustomID string `json:"custom_id"`
		} `json:"resource"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error parsing PayPal webhook JSON: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	log.Printf("Received verified PayPal event: %s", event.EventType)

	switch event.EventType {
	case "PAYMENT.CAPTURE.COMPLETED":
		var metadata struct {
			TripID   string `json:"trip_id"`
			UserID   string `json:"user_id"`
			DriverID string `json:"driver_id"`
		}

		if err := json.Unmarshal([]byte(event.Resource.CustomID), &metadata); err != nil {
			log.Printf("Error parsing custom_id metadata (falling back to raw string): %v", err)
			metadata.TripID = event.Resource.CustomID
		}

		payload := messaging.PaymentStatusUpdateData{
			TripID:   metadata.TripID,
			UserID:   metadata.UserID,
			DriverID: metadata.DriverID,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshalling payload: %v", err)
			http.Error(w, "Failed to marshal payload", http.StatusInternalServerError)
			return
		}

		message := contracts.AmqpMessage{
			OwnerID: metadata.UserID,
			Data:    payloadBytes,
		}

		if err := rb.PublishMessage(
			ctx,
			contracts.PaymentEventSuccess,
			message,
		); err != nil {
			log.Printf("Error publishing payment event: %v", err)
			http.Error(w, "Failed to publish payment event", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
