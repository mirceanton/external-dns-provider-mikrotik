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
EEEEE  X     X TTTTTTT EEEEE  RRRRR   N   N   A   L          DDDD    N   N    SSS
E       X   X     T    E      R    R  NN  N  A A  L          D   D   NN  N  S
EEE       X       T    EEE    RRRRR   N N N AAAAA L     ===  D    D  N N N   SSS
E       X   X     T    E      R   R   N  NN A   A L          D   D   N  NN      S
EEEEE  X     X    T    EEEEE  R    R  N   N A   A LLLLL      DDDD    N   N  SSSS
==================================================================================
     MMM      MMM       KKK                          TTTTTTTTTTT      KKK
     MMMM    MMMM       KKK                          TTTTTTTTTTT      KKK
     MMM MMMM MMM  III  KKK  KKK  RRRRRR     OOOOOO      TTT     III  KKK  KKK
     MMM  MM  MMM  III  KKKKK     RRR  RRR  OOO  OOO     TTT     III  KKKKK
     MMM      MMM  III  KKK KKK   RRRRRR    OOO  OOO     TTT     III  KKK KKK
     MMM      MMM  III  KKK  KKK  RRR  RRR   OOOOOO      TTT     III  KKK  KKK

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
