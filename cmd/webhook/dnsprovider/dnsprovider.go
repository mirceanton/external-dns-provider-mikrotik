package dnsprovider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/mirceanton/external-dns-provider-mikrotik/cmd/webhook/init/configuration"
	"github.com/mirceanton/external-dns-provider-mikrotik/internal/mikrotik"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"

	log "github.com/sirupsen/logrus"
)

type MikrotikProviderFactory func(baseProvider *provider.BaseProvider, mikrotikConfig *mikrotik.Config) provider.Provider

func Init(config configuration.Config) (provider.Provider, error) {
	var domainFilter endpoint.DomainFilter
	createMsg := "creating mikrotik provider with "

	if config.RegexDomainFilter != "" {
		createMsg += fmt.Sprintf("regexp domain filter: '%s', ", config.RegexDomainFilter)
		if config.RegexDomainExclusion != "" {
			createMsg += fmt.Sprintf("with exclusion: '%s', ", config.RegexDomainExclusion)
		}
		domainFilter = endpoint.NewRegexDomainFilter(
			regexp.MustCompile(config.RegexDomainFilter),
			regexp.MustCompile(config.RegexDomainExclusion),
		)
	} else {
		if config.DomainFilter != nil && len(config.DomainFilter) > 0 {
			createMsg += fmt.Sprintf("domain filter: '%s', ", strings.Join(config.DomainFilter, ","))
		}
		if config.ExcludeDomains != nil && len(config.ExcludeDomains) > 0 {
			createMsg += fmt.Sprintf("exclude domain filter: '%s', ", strings.Join(config.ExcludeDomains, ","))
		}
		domainFilter = endpoint.NewDomainFilterWithExclusions(config.DomainFilter, config.ExcludeDomains)
	}

	createMsg = strings.TrimSuffix(createMsg, ", ")
	if strings.HasSuffix(createMsg, "with ") {
		createMsg += "no kind of domain filters"
	}
	log.Info(createMsg)

	mikrotikConfig := mikrotik.Config{}
	if err := env.Parse(&mikrotikConfig); err != nil {
		return nil, fmt.Errorf("reading mikrotik configuration failed: %v", err)
	}

	return mikrotik.NewMikrotikProvider(domainFilter, &mikrotikConfig)
}
