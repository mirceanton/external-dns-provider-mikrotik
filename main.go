package main

import (
	"github.com/mirceanton/external-dns-provider-mikrotik/internal/configuration"
	"github.com/mirceanton/external-dns-provider-mikrotik/internal/dnsprovider"
	"github.com/mirceanton/external-dns-provider-mikrotik/internal/logging"
	"github.com/mirceanton/external-dns-provider-mikrotik/internal/server"
	"github.com/mirceanton/external-dns-provider-mikrotik/pkg/webhook"
	log "github.com/sirupsen/logrus"
)

var (
	Version = "local"
	Gitsha  = "?"
)

func main() {
	logging.Init()

	log.Infof("starting external-dns-provider-mikrotik")
	log.Infof("version: %s (%s)", Version, Gitsha)

	config := configuration.Init()
	provider, err := dnsprovider.Init(config)
	if err != nil {
		log.Fatalf("failed to initialize provider: %v", err)
	}

	main, health := server.Init(config, webhook.New(provider))
	server.ShutdownGracefully(main, health)
}
