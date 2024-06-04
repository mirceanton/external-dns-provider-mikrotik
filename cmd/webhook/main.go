package main

import (
	"fmt"

	"github.com/mirceanton/external-dns-provider-mikrotik/cmd/webhook/init/configuration"
	"github.com/mirceanton/external-dns-provider-mikrotik/cmd/webhook/init/dnsprovider"
	"github.com/mirceanton/external-dns-provider-mikrotik/cmd/webhook/init/logging"
	"github.com/mirceanton/external-dns-provider-mikrotik/cmd/webhook/init/server"
	"github.com/mirceanton/external-dns-provider-mikrotik/pkg/webhook"
	log "github.com/sirupsen/logrus"
)

const banner = `
external-dns-provider-mikrotik
version: %s (%s)
`

var (
	Version = "local"
	Gitsha  = "?"
)

func main() {
	fmt.Printf(banner, Version, Gitsha)

	logging.Init()

	config := configuration.Init()
	provider, err := dnsprovider.Init(config)
	if err != nil {
		log.Fatalf("failed to initialize provider: %v", err)
	}

	main, health := server.Init(config, webhook.New(provider))
	server.ShutdownGracefully(main, health)
}
