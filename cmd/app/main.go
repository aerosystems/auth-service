package main

import (
	"context"
	"golang.org/x/sync/errgroup"
)

// @title Auth Service
// @version 1.0.8
// @description A mandatory part of any microservice infrastructure of a modern WEB application, which is responsible for user authentication and authorization.

// @contact.name Artem Kostenko
// @contact.url https://github.com/aerosystems

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Should contain Access JWT Token, with the Bearer started

// @host gw.verifire.dev/auth
// @schemes https
// @BasePath /
func main() {
	app := InitApp()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return app.httpServer.Run()
	})

	group.Go(func() error {
		return app.handleSignals(ctx, cancel)
	})

	if err := group.Wait(); err != nil {
		app.log.Errorf("error occurred: %v", err)
	}
}
