package config

import (
	"encoding/json"
	"os"

	"github.com/traf72/singbox-api/internal/apperr"
)

type config struct {
	Log       logConfig   `json:"log"`
	DNS       dnsConfig   `json:"dns"`
	Inbounds  []inbound   `json:"inbounds"`
	Outbounds []outbound  `json:"outbounds"`
	Route     routeConfig `json:"route"`
}

type logConfig struct {
	Disabled  bool   `json:"disabled"`
	Level     string `json:"level"`
	Output    string `json:"output"`
	Timestamp bool   `json:"timestamp"`
}

type dnsConfig struct {
	IndependentCache bool        `json:"independent_cache"`
	Final            string      `json:"final"`
	Rules            []dnsRule   `json:"rules"`
	Servers          []dnsServer `json:"servers"`
}

type dnsRule struct {
	Domain        []string `json:"domain"`
	DomainKeyword []string `json:"domain_keyword"`
	DomainRegex   []string `json:"domain_regex"`
	DomainSuffix  []string `json:"domain_suffix"`
	Geosite       []string `json:"geosite"`
	Server        string   `json:"server"`
}

type dnsServer struct {
	Address         string `json:"address"`
	AddressResolver string `json:"address_resolver"`
	Detour          string `json:"detour"`
	Tag             string `json:"tag"`
}

type inbound struct {
	Listen                   string `json:"listen"`
	ListenPort               int    `json:"listen_port"`
	Sniff                    bool   `json:"sniff"`
	SniffOverrideDestination bool   `json:"sniff_override_destination"`
	Tag                      string `json:"tag"`
	Type                     string `json:"type"`
}

type outbound struct {
	Flow           string    `json:"flow"`
	PacketEncoding string    `json:"packet_encoding"`
	Server         string    `json:"server"`
	ServerPort     int       `json:"server_port"`
	Tag            string    `json:"tag"`
	TLS            tlsConfig `json:"tls"`
	Type           string    `json:"type"`
	UUID           string    `json:"uuid"`
}

type tlsConfig struct {
	ALPN       []string      `json:"alpn"`
	Enabled    bool          `json:"enabled"`
	Reality    realityConfig `json:"reality"`
	ServerName string        `json:"server_name"`
	UTLS       utlsConfig    `json:"utls"`
}

type realityConfig struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"public_key"`
	ShortID   string `json:"short_id"`
}

type utlsConfig struct {
	Enabled     bool   `json:"enabled"`
	Fingerprint string `json:"fingerprint"`
}

type routeConfig struct {
	AutoDetectInterface bool        `json:"auto_detect_interface"`
	Final               string      `json:"final"`
	Rules               []routeRule `json:"rules"`
}

type routeRule struct {
	Domain        []string `json:"domain"`
	DomainKeyword []string `json:"domain_keyword"`
	DomainRegex   []string `json:"domain_regex"`
	DomainSuffix  []string `json:"domain_suffix"`
	Geosite       []string `json:"geosite"`
	Outbound      string   `json:"outbound"`
}

var errEmptyPath = apperr.NewFatalErr("EmptyConfigPath", "Path to the configuration file is not specified")

func load() (*config, *apperr.Err) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		return nil, errEmptyPath
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, apperr.NewFatalErr("ConfigOpenError", err.Error())
	}

	d := json.NewDecoder(file)
	config := new(config)
	if err := d.Decode(config); err != nil {
		return nil, apperr.NewFatalErr("ConfigJsonDecodeError", err.Error())
	}

	return config, nil
}
