// Package main is the auth-service HTTP server entrypoint.
//
// @title           Auth Service API
// @version         1.0
// @description     Authentication and authorization service for the marketplace.
// @contact.name    leenwood
// @contact.email   george200135@gmail.com
// @host            localhost:8081
// @BasePath        /
// @schemes         http https
//
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Paste your JWT access token: Bearer <token>
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/leenwood/market-auth-service/internal/app/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := service.RunServer(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
