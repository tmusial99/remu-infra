package main

type DomainConfig struct {
	Host       string `json:"host"`
	StaticPath string `json:"static_path"`
	APIEnabled bool   `json:"api_enabled"`
}

type ServerConfig struct {
	Port                string         `json:"port"`
	LocalhostTestDomain string         `json:"localhost_test_domain"`
	NoCacheExtensions   []string       `json:"exts_to_cache"`
	Domains             []DomainConfig `json:"domains"`
}

func getConfig() *ServerConfig {
	return &ServerConfig{
		Port:                ":3000",
		LocalhostTestDomain: "novi-tech.net",
		NoCacheExtensions:   []string{".html"},
		Domains: []DomainConfig{
			{
				Host:       "tmdev.pl",
				StaticPath: "./public/tmdev",
				APIEnabled: true,
			},
			{
				Host:       "novi-tech.net",
				StaticPath: "./public/novi-tech",
			},
		},
	}
}
