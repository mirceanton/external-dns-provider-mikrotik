package mikrotik

// config details for authentication with the MikroTik RouterOS API
type Config struct {
	Host          string `env:"MIKROTIK_HOST,notEmpty"`
	Port          string `env:"MIKROTIK_PORT,notEmpty" envDefault:"443"`
	Username      string `env:"MIKROTIK_USERNAME,notEmpty"`
	Password      string `env:"MIKROTIK_PASSWORD,notEmpty"`
	SkipTLSVerify bool   `env:"MIKROTIK_SKIP_TLS_VERIFY" envDefault:"false"`
}

// https://help.mikrotik.com/docs/display/ROS/DNS#DNS-DNSStatic
// TODO: Add all fields here
type DNSRecord struct {
	ID      string `json:".id,omitempty"`
	Name    string `json:"name"`                       // Domain name.
	Address string `json:"address"`                    // The address that will be used for "A" or "AAAA" type records.
	Type    string `default:"A" json:"type,omitempty"` // Type of the DNS record. (A | AAAA | CNAME | FWD | MX | NS | NXDOMAIN | SRV | TXT ; Default: A)
	// TTL     endpoint.TTL `json:"ttl,omitempty"`              // Maximum time-to-live for cached records. (default 24h)  //FIXME
	Comment string `json:"comment,omitempty"` // Comment about the domain name record.
}

// https://help.mikrotik.com/docs/display/ROS/Resource
type SystemInfo struct {
	ArchitectureName     string `json:"architecture-name"`
	BadBlocks            string `json:"bad-blocks"`
	BoardName            string `json:"board-name"`
	BuildTime            string `json:"build-time"`
	CPU                  string `json:"cpu"`
	CPUCount             string `json:"cpu-count"`
	CPUFrequency         string `json:"cpu-frequency"`
	CPULoad              string `json:"cpu-load"`
	FactorySoftware      string `json:"factory-software"`
	FreeHDDSpace         string `json:"free-hdd-space"`
	FreeMemory           string `json:"free-memory"`
	Platform             string `json:"platform"`
	TotalHDDSpace        string `json:"total-hdd-space"`
	TotalMemory          string `json:"total-memory"`
	Uptime               string `json:"uptime"`
	Version              string `json:"version"`
	WriteSectSinceReboot string `json:"write-sect-since-reboot"`
	WriteSectTotal       string `json:"write-sect-total"`
}
