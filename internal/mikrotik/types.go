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
type DNSRecord struct {
	ID             string `json:".id,omitempty"`
	Address        string `json:"address,omitempty"`
	CName          string `json:"cname,omitempty"`
	ForwardTo      string `json:"forward-to,omitempty"`
	MXExchange     string `json:"mx-exchange,omitempty"`
	Name           string `json:"name"`
	SrvPort        int    `json:"srv-port,omitempty"`
	SrvTarget      string `json:"srv-target,omitempty"`
	Text           string `json:"text,omitempty"`
	Type           string `default:"A" json:"type,omitempty"`
	AddressList    string `json:"address-list,omitempty"`
	Comment        string `json:"comment,omitempty"`
	Disabled       bool   `default:"false" json:"disabled,omitempty"`
	MatchSubdomain bool   `json:"match-subdomain,omitempty"`
	MXPreference   int    `json:"mx-preference,omitempty"`
	NS             string `json:"ns,omitempty"`
	Regexp         string `json:"regexp,omitempty"`
	SrvPriority    int    `json:"srv-priority,omitempty"`
	SrvWeight      int    `json:"srv-wright,omitempty"`
	TTL            string `json:"ttl,omitempty"`
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