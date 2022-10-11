package config

import (
	"fmt"
	"time"
)

// Environment variable constants.
const (
	EnvServerPort                  = "server.port"
	EnvServerHost                  = "server.host"
	EnvServerFQDN                  = "server.fqdn"
	EnvServerSwaggerEnabled        = "server.swagger.enabled"
	EnvServerSwaggerHost           = "server.swagger.host"
	EnvEnvironment                 = "env.environment"
	EnvRegion                      = "env.region"
	EnvVersion                     = "env.version"
	EnvCommit                      = "env.commit"
	EnvBuildDate                   = "env.builddate"
	EnvLogLevel                    = "log.level"
	EnvPaydNoop                    = "payd.noop"
	EnvSocketChannelTimeoutSeconds = "socket.channel.timeoutseconds"
	EnvSocketMaxMessageBytes       = "socket.maxmessage.bytes"
	EnvTransportMode               = "transport.mode"

	LogDebug = "debug"
	LogInfo  = "info"
	LogError = "error"
	LogWarn  = "warn"

	TransportModeHybrid = "hybrid"
	TransportModeSocket = "socket"
)

// Config returns strongly typed config values.
type Config struct {
	Logging    *Logging
	Server     *Server
	Deployment *Deployment
	PayD       *PayD
	Sockets    *Socket
	Transports *Transports
}

// Deployment contains information relating to the current
// deployed instance.
type Deployment struct {
	Environment string
	AppName     string
	Region      string
	Version     string
	Commit      string
	BuildDate   time.Time
}

// IsDev determines if this app is running on a dev environment.
func (d *Deployment) IsDev() bool {
	return d.Environment == "dev"
}

func (d *Deployment) String() string {
	return fmt.Sprintf("Environment: %s \n AppName: %s\n Region: %s\n Version: %s\n Commit:%s\n BuildDate: %s\n",
		d.Environment, d.AppName, d.Region, d.Version, d.Commit, d.BuildDate)
}

// Logging contains log configuration.
type Logging struct {
	Level string
}

// Server contains all settings required to run a web server.
type Server struct {
	Port     string
	Hostname string
	// FQDN - fully qualified domain name, used to form the PaymentTerms
	// payment URL as this may be different from the hostname + port.
	FQDN string
	// SwaggerEnabled if true we will include an endpoint to serve swagger documents.
	SwaggerEnabled bool
	SwaggerHost    string
}

// PayD is for mock testing.
type PayD struct {
	Noop bool
}

// Socket contains config items for a socket server.
type Socket struct {
	MaxMessageBytes int
	ChannelTimeout  time.Duration
}

// Transports enables or disables dpp transports.
type Transports struct {
	Mode string
}

// ConfigurationLoader will load configuration items
// into a struct that contains a configuration.
type ConfigurationLoader interface {
	WithServer() ConfigurationLoader
	WithDeployment(app string) ConfigurationLoader
	WithLog() ConfigurationLoader
	WithPayD() ConfigurationLoader
	WithSockets() ConfigurationLoader
	WithTransports() ConfigurationLoader
	Load() *Config
}
