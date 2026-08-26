package main

import (
	// "bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Sed-Miyuki/OmniRoute/services/api-gateway/grpc_clients"
	"github.com/Sed-Miyuki/OmniRoute/shared/contracts"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
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

	tripService,err:=grpc_clients.NewTripServiceClient()
	if err!=nil{
		log.Fatal(err)
	}

	defer tripService.Close()

	tripPreview,err:=tripService.Client.PreviewTrip(r.Context(),reqBody.ToProto())
	if err!=nil{
		log.Printf("failed to preview a trip: %v",err)
		http.Error(w,"failed to preview trip",http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: tripPreview}

	writeJSON(w, http.StatusCreated, response)
}

func handleTripStart(w http.ResponseWriter,r *http.Request){
	var reqBody startTripRequest
	if err:=json.NewDecoder(r.Body).Decode(&reqBody);err!=nil{
		http.Error(w,"failed to parse JSON payload",http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// Why we need to create a new client for each connection:
	// because if a service is down, we don't want to block the whole application
	// so we create a new client for each connection
	tripService,err:=grpc_clients.NewTripServiceClient()
	if err!=nil{
		log.Fatal(err)
	}
	defer tripService.Close()

	trip,err:=tripService.Client.CreateTrip(r.Context(),reqBody.toProto())
	if err!=nil{
		log.Printf("Failed to start a trip: %v", err)
		http.Error(w, "Failed to start trip", http.StatusInternalServerError)
		return
	}

	response:=contracts.APIResponse{Data: trip}
	writeJSON(w,http.StatusCreated,response)
}
