package mikrotik

import "sigs.k8s.io/external-dns/endpoint"

// config details for authentication with the MikroTik RouterOS API
type Config struct {
	Host          string `env:"MIKROTIK_HOST,notEmpty"`
	Port          string `env:"MIKROTIK_PORT,notEmpty" envDefault:"443"`
	Username      string `env:"MIKROTIK_USERNAME,notEmpty"`
	Password      string `env:"MIKROTIK_PASSWORD,notEmpty"`
	SkipTLSVerify bool   `env:"MIKROTIK_SKIP_TLS_VERIFY" envDefault:"false"`
}

// https://help.mikrotik.com/docs/display/ROS/DNS#DNS-DNSStatic
type DNSRecord struct {
	ID      string       `json:".id"`
	Name    string       `json:"name"`                       // Domain name.
	Address string       `json:"address"`                    // The address that will be used for "A" or "AAAA" type records.
	Type    string       `default:"A" json:"type,omitempty"` // Type of the DNS record. (A | AAAA | CNAME | FWD | MX | NS | NXDOMAIN | SRV | TXT ; Default: A)
	TTL     endpoint.TTL `json:"ttl,omitempty"`              // Maximum time-to-live for cached records. (default 24h)
	Comment string       `json:"comment,omitempty"`          // Comment about the domain name record.
}
