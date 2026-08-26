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

	// h "github.com/Sed-Miyuki/OmniRoute/services/trip-service/internal/infrastructure/http"
	"github.com/Sed-Miyuki/OmniRoute/services/trip-service/internal/infrastructure/grpc"
	"github.com/Sed-Miyuki/OmniRoute/services/trip-service/internal/infrastructure/repository"
	"github.com/Sed-Miyuki/OmniRoute/services/trip-service/internal/service"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9093"

func main() {
	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func ()  {
		sigCh:=make(chan os.Signal,1)
		signal.Notify(sigCh,os.Interrupt,syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis,err:=net.Listen("tcp",GrpcAddr)
	if err!=nil{
		log.Fatalf("failed to listen: %v",err)
	}

	//starting the gRPC server
	grpcServer:=grpcserver.NewServer()

	grpc.NewGRPCHandler(grpcServer,svc)

	log.Printf("Starting grpc Trip service on port %s",lis.Addr().String())

	go func ()  {
		if err:=grpcServer.Serve(lis);err!=nil{
			log.Printf("failed to serve: %v",err)
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

	// mux.HandleFunc("POST /preview", httphandler.HandleTripPreview)

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
