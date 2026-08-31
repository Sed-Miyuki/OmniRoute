package main

import (
	"context"
	"log"
	"net"

	// "net/http"
	"os"
	"os/signal"
	"syscall"

	"time"

	// h "github.com/Sed-Miyuki/OmniRoute/services/Driver-service/internal/infrastructure/http"
	"github.com/Sed-Miyuki/OmniRoute/shared/env"
	"github.com/Sed-Miyuki/OmniRoute/shared/messaging"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9092"

func main() {

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	service := NewService()

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")

	//starting the gRPC server
	grpcServer := grpcserver.NewServer()
	NewGrpcHandler(grpcServer, service)

	consumer:=NewTripConsumer(rabbitmq,service)
	go func ()  {
		if err:=consumer.Listen();err!=nil{
			log.Fatalf("Failed to listen to the message: %v",err)
		}
	}()

	log.Printf("Starting grpc Driver service on port %s", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down the server...")
	grpcServer.GracefulStop()

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("Server gracefully stopped.")
	case <-time.After(10 * time.Second):
		log.Println("Graceful shutdown timed out. Forcing stop...")
		grpcServer.Stop() // Force close all active connections
	}

	// mux := http.NewServeMux()
	// httphandler := h.HttpHandler{Service: svc}

	// mux.HandleFunc("POST /preview", httphandler.HandleDriverPreview)

	// server := &http.Server{
	// 	Addr:    ":8083",
	// 	Handler: mux,
	// }

	// serverErrors := make(chan error, 1)

	// go func() {
	// 	log.Printf("Server listening on %s", server.Addr)
	// 	serverErrors <- server.ListenAndServe()
	// }()

	// shutdown := make(chan os.Signal, 1)
	// signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// select {
	// case err := <-serverErrors:
	// 	log.Printf("Error starting server: %v", err)

	// case sig := <-shutdown:
	// 	log.Printf("Server is shutting down due to %v signal", sig)

	// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 	defer cancel()

	// 	if err := server.Shutdown(ctx); err != nil {
	// 		log.Printf("Could not stop server gracefully: %v", err)
	// 		server.Close()
	// 	}
	// }
}
