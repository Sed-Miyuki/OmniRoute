package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sed-Miyuki/OmniRoute/services/payment-service/internal/infrastructure/events"
	"github.com/Sed-Miyuki/OmniRoute/services/payment-service/internal/infrastructure/paypal"
	"github.com/Sed-Miyuki/OmniRoute/services/payment-service/internal/service"
	"github.com/Sed-Miyuki/OmniRoute/services/payment-service/pkg/types"
	"github.com/Sed-Miyuki/OmniRoute/shared/env"
	"github.com/Sed-Miyuki/OmniRoute/shared/messaging"
	"github.com/Sed-Miyuki/OmniRoute/shared/tracing"
)

var GrpcAddr = env.GetString("GRPC_ADDR", ":9004")

func main() {

	//initialize tracing
	tracerCfg := tracing.Config{
		ServiceName:    "payment-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JeagerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}

	sh, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("Failed to initialize the tracer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	// Setup graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	appURL := env.GetString("APP_URL", "http://localhost:3000")

	paypalCfg := &types.PaymentConfig{
		PayPalClientID: env.GetString("PAYPAL_CLIENT_ID", ""),
		PayPalSecret:   env.GetString("PAYPAL_SECRET", ""),
		ReturnURL:      env.GetString("PAYPAL_RETURN_URL", appURL+"?payment=success"),
		CancelURL:      env.GetString("PAYPAL_CANCEL_URL", appURL+"?payment=cancel"),
	}

	if paypalCfg.PayPalClientID == "" || paypalCfg.PayPalSecret == "" {
		log.Fatalf("PayPal credentials are not set")
		return
	}

	paymentProcessor := paypal.NewPayPalClient(paypalCfg)
	svc := service.NewPaymentService(paymentProcessor)

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")

	tripConsumer := events.NewTripConsumer(rabbitmq, svc)

	go tripConsumer.Listen()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down payment service...")
}
