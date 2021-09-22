package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/libsv/pptcl/data"
	"github.com/libsv/pptcl/data/noop"
	_ "github.com/libsv/pptcl/docs"

	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"

	"github.com/libsv/pptcl/config"
	"github.com/libsv/pptcl/data/payd"
	"github.com/libsv/pptcl/service"
	pptclHandlers "github.com/libsv/pptcl/transports/http"

	paydMiddleware "github.com/libsv/pptcl/transports/http/middleware"
)

const appname = "payment-protocol-rest-server"
const banner = `
====================================================================
         _         _       _            _            _     
        /\ \      /\ \    /\ \        /\ \          _\ \   
       /  \ \    /  \ \   \_\ \      /  \ \        /\__ \  
      / /\ \ \  / /\ \ \  /\__ \    / /\ \ \      / /_ \_\ 
     / / /\ \_\/ / /\ \_\/ /_ \ \  / / /\ \ \    / / /\/_/ 
    / / /_/ / / / /_/ / / / /\ \ \/ / /  \ \_\  / / /      
   / / /__\/ / / /__\/ / / /  \/_/ / /    \/_/ / / /       
  / / /_____/ / /_____/ / /     / / /         / / / ____   
 / / /     / / /     / / /     / / /________ / /_/_/ ___/\ 
/ / /     / / /     /_/ /     / / /_________/_______/\__\/ 
\/_/      \/_/      \_\/      \/____________\_______\/     

====================================================================
`

// main is the entry point of the application.
// @title Payment Protocol Server
// @version 0.0.1
// @description Payment Protocol Server is an implementation of a Bip-270 payment flow.
// @termsOfService https://github.com/libsv/go-payment_protocol/blob/master/CODE_STANDARDS.md
// @license.name ISC
// @license.url https://github.com/libsv/go-payment_protocol/blob/master/LICENSE
// @host localhost:8445
// @schemes:
//	- http
//	- https
func main() {
	println("\033[32m" + banner + "\033[0m")
	config.SetupDefaults()
	cfg := config.NewViperConfig(appname).
		WithServer().
		WithDeployment(appname).
		WithLog().
		WithPayD().
		Load()
	// validate the config, fail if it fails.
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	config.SetupLog(cfg.Logging)
	log.Infof("\n------Environment: %s -----\n", cfg.Server)

	e := echo.New()
	e.HideBanner = true
	g := e.Group("/")
	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.HTTPErrorHandler = paydMiddleware.ErrorHandler
	if cfg.Server.SwaggerEnabled {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	if cfg.Deployment.IsDev() {
		printDev(e)
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}
	if !cfg.PayD.Secure { // for testing, don't validate server cert
		// #nosec
		httpClient.Transport = &http.Transport{
			// #nosec
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	// stores
	paydStore := payd.NewPayD(cfg.PayD, data.NewClient(httpClient))
	noopStore := noop.NewNoOp()

	// services
	paymentSvc := service.NewPayment(paydStore)
	paymentReqSvc := service.NewPaymentRequest(cfg.Server, paydStore, paydStore)
	if cfg.PayD.Noop {
		paymentSvc = service.NewPayment(noopStore)
		paymentReqSvc = service.NewPaymentRequest(cfg.Server, noopStore, noopStore)
	}
	// handlers
	pptclHandlers.NewPaymentHandler(paymentSvc).RegisterRoutes(g)
	pptclHandlers.NewPaymentRequestHandler(paymentReqSvc).RegisterRoutes(g)

	e.Logger.Fatal(e.Start(cfg.Server.Port))
}

// printDev outputs some useful dev information such as http routes
// and current settings being used.
func printDev(e *echo.Echo) {
	fmt.Println("==================================")
	fmt.Println("DEV mode, printing http routes:")
	for _, r := range e.Routes() {
		fmt.Printf("%s: %s\n", r.Method, r.Path)
	}
	fmt.Println("==================================")
	fmt.Println("DEV mode, printing settings:")
	for _, v := range viper.AllKeys() {
		fmt.Printf("%s: %v\n", v, viper.Get(v))
	}
	fmt.Println("==================================")
}