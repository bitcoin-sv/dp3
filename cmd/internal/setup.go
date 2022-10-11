package internal

import (
	"fmt"
	"net/http"

	"github.com/bitcoin-sv/dpp-proxy/docs"
	"github.com/bitcoin-sv/dpp-proxy/log"
	dppHandlers "github.com/bitcoin-sv/dpp-proxy/transports/http"
	dppMiddleware "github.com/bitcoin-sv/dpp-proxy/transports/http/middleware"
	dppSoc "github.com/bitcoin-sv/dpp-proxy/transports/sockets"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoProm "github.com/labstack/echo-contrib/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/viper"
	echoSwagger "github.com/swaggo/echo-swagger"
	smw "github.com/theflyingcodr/sockets/middleware"
	"github.com/theflyingcodr/sockets/server"

	"github.com/bitcoin-sv/dpp-proxy/config"
	"github.com/bitcoin-sv/dpp-proxy/data/noop"
	socData "github.com/bitcoin-sv/dpp-proxy/data/sockets"
	"github.com/bitcoin-sv/dpp-proxy/service"
	"github.com/libsv/go-dpp"
)

// Deps holds all the dependencies.
type Deps struct {
	PaymentService      dpp.PaymentService
	PaymentTermsService dpp.PaymentTermsService
	ProofsService       dpp.ProofsService
}

// SetupEcho will set up and return an echo server.
func SetupEcho(cfg *config.Config, l log.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return cfg.Logging.Level != config.LogDebug
		},
	}))
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	p := echoProm.NewPrometheus("dpp", nil)
	p.Use(e)
	e.HTTPErrorHandler = dppMiddleware.ErrorHandler(l)
	return e
}

// SetupSwagger will enable the swagger endpoints.
func SetupSwagger(cfg config.Server, e *echo.Echo) {
	docs.SwaggerInfo.Host = cfg.SwaggerHost
	e.GET("/swagger/*", echoSwagger.WrapHandler)
}

// SetupSockets will setup handlers and socket server.
func SetupSockets(cfg config.Socket, e *echo.Echo) *server.SocketServer {
	g := e.Group("/")
	// create socket server
	s := server.New(
		server.WithMaxMessageSize(int64(cfg.MaxMessageBytes)),
		server.WithChannelTimeout(cfg.ChannelTimeout))

	// add middleware, with panic going first
	s.WithMiddleware(smw.PanicHandler, smw.Timeout(smw.NewTimeoutConfig()), smw.Metrics())

	dppSoc.NewPaymentTerms().Register(s)
	dppSoc.NewPayment().Register(s)
	dppHandlers.NewProofs(service.NewProof(socData.NewPaymentStore(s))).RegisterRoutes(g)

	// this is our websocket endpoint, clients will hit this with the channelID they wish to connect to
	e.GET("/ws/:channelID", wsHandler(s))
	return s
}

// SetupHybrid will setup handlers for http=>socket communication.
func SetupHybrid(cfg config.Config, l log.Logger, e *echo.Echo) *server.SocketServer {
	g := e.Group("/")
	s := server.New(
		server.WithMaxMessageSize(int64(cfg.Sockets.MaxMessageBytes)),
		server.WithChannelTimeout(cfg.Sockets.ChannelTimeout))
	// add middleware, with panic going first
	s.WithMiddleware(smw.PanicHandler, smw.Timeout(smw.NewTimeoutConfig()), smw.Metrics())

	paymentStore := socData.NewPaymentStore(s)
	paymentSvc := service.NewPayment(l, paymentStore)
	if cfg.PayD.Noop {
		noopStore := noop.NewNoOp(log.Noop{})
		paymentSvc = service.NewPayment(log.Noop{}, noopStore)
	}
	paymentReqSvc := service.NewPaymentTermsProxy(paymentStore, cfg.Transports, cfg.Server)
	proofsSvc := service.NewProof(paymentStore)

	dppHandlers.NewPaymentHandler(paymentSvc).RegisterRoutes(g)
	dppHandlers.NewPaymentTermsHandler(paymentReqSvc).RegisterRoutes(g)
	dppHandlers.NewProofs(proofsSvc).RegisterRoutes(g)
	dppSoc.NewHealthHandler().Register(s)

	e.GET("/ws/:channelID", wsHandler(s))
	return s
}

// wsHandler will upgrade connections to a websocket and then wait for messages.
func wsHandler(svr *server.SocketServer) echo.HandlerFunc {
	upgrader := websocket.Upgrader{}
	return func(c echo.Context) error {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		defer func() {
			_ = ws.Close()
		}()

		chID := c.Param("channelID")

		if c.QueryParam("internal") != "true" {
			if !svr.HasChannel(chID) {
				return c.JSON(http.StatusNotFound, fmt.Sprintf("Connection for invoice '%s' not found", chID))
			}
		}

		return svr.Listen(ws, chID)
	}
}

// SetupSocketMetrics will setup the socket server metrics.
func SetupSocketMetrics(s *server.SocketServer) {
	// simple metrics
	gCo := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "sockets",
		Subsystem: "server",
		Name:      "gauge_total_connections",
	})

	s.OnClientJoin(func(clientID, channelID string) {
		gCo.Inc()
	})

	s.OnClientLeave(func(clientID, channelID string) {
		gCo.Dec()
	})

	gCh := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "sockets",
		Subsystem: "server",
		Name:      "gauge_total_channels",
	})

	s.OnChannelCreate(func(channelID string) {
		gCh.Inc()
	})

	s.OnChannelClose(func(channelID string) {
		gCh.Dec()
	})
}

// PrintDev outputs some useful dev information such as http routes
// and current settings being used.
func PrintDev(e *echo.Echo) {
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
