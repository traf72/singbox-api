package config

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/utils"
)

type Conf struct {
	Log       *logging    `json:"log"`
	DNS       dns         `json:"dns"`
	Inbounds  []*inbound  `json:"inbounds"`
	Outbounds []*outbound `json:"outbounds"`
	Route     route       `json:"route"`
}

type logging struct {
	Disabled  bool   `json:"disabled"`
	Level     string `json:"level"`
	Output    string `json:"output"`
	Timestamp bool   `json:"timestamp"`
}

type dns struct {
	IndependentCache bool        `json:"independent_cache,omitempty"`
	CacheCapacity    int         `json:"cache_capacity,omitempty"`
	ReverseMapping   bool        `json:"reverse_mapping,omitempty"`
	Final            string      `json:"final"`
	Rules            []DNSRule   `json:"rules"`
	Servers          []dnsServer `json:"servers"`
}

type Rule struct {
	Domain        []string `json:"domain,omitempty"`
	DomainKeyword []string `json:"domain_keyword,omitempty"`
	DomainRegex   []string `json:"domain_regex,omitempty"`
	DomainSuffix  []string `json:"domain_suffix,omitempty"`
}

type DNSRule struct {
	Rule
	Server string `json:"server"`
}

type dnsServer struct {
	Address         string `json:"address"`
	AddressResolver string `json:"address_resolver,omitempty"`
	Detour          string `json:"detour,omitempty"`
	Tag             string `json:"tag"`
}

type inbound struct {
	Listen                 string   `json:"listen,omitempty"`
	ListenPort             int      `json:"listen_port,omitempty"`
	AutoRoute              bool     `json:"auto_route,omitempty"`
	AutoRedirect           bool     `json:"auto_redirect,omitempty"`
	EndpointIndependentNAT *bool    `json:"endpoint_independent_nat,omitempty"`
	Address                []string `json:"address,omitempty"`
	InterfaceName          string   `json:"interface_name,omitempty"`
	MTU                    int      `json:"mtu,omitempty"`
	Stack                  string   `json:"stack,omitempty"`
	StrictRoute            *bool    `json:"strict_route,omitempty"`
	Tag                    string   `json:"tag"`
	Type                   string   `json:"type"`
}

type outbound struct {
	Flow           string `json:"flow,omitempty"`
	PacketEncoding string `json:"packet_encoding,omitempty"`
	Server         string `json:"server,omitempty"`
	ServerPort     int    `json:"server_port,omitempty"`
	Tag            string `json:"tag"`
	TLS            *tls   `json:"tls,omitempty"`
	Type           string `json:"type"`
	UUID           string `json:"uuid,omitempty"`
}

type tls struct {
	ALPN       []string `json:"alpn"`
	Enabled    bool     `json:"enabled"`
	Reality    *reality `json:"reality,omitempty"`
	ServerName string   `json:"server_name"`
	UTLS       *utls    `json:"utls,omitempty"`
}

type reality struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"public_key"`
	ShortID   string `json:"short_id"`
}

type utls struct {
	Enabled     bool   `json:"enabled"`
	Fingerprint string `json:"fingerprint"`
}

type route struct {
	AutoDetectInterface bool        `json:"auto_detect_interface"`
	Final               string      `json:"final"`
	Rules               []RouteRule `json:"rules"`
}

type RouteRule struct {
	Rule
	IP_CIDR  []string `json:"ip_cidr,omitempty"`
	Inbound  []string `json:"inbound,omitempty"`
	Outbound string   `json:"outbound,omitempty"`
	Protocol string   `json:"protocol,omitempty"`
	Action   string   `json:"action,omitempty"`
	Strategy string   `json:"strategy,omitempty"`
	Timeout  string   `json:"timeout,omitempty"`
}

type Config struct {
	Conf         *Conf
	lastModified time.Time
}

var serializeOptions = &utils.JSONOptions{Indent: "    ", EscapeHTML: false}

var errEmptyPath = apperr.NewFatalErr("Config_EmptyPath", "path to the configuration file is not specified")

func errStatReading(err string) apperr.Err {
	return apperr.NewFatalErr("Config_StatReadError", err)
}

func Load() (*Config, apperr.Err) {
	path, appErr := getConfPath()
	if appErr != nil {
		return nil, appErr
	}

	stat, err := os.Stat(path)
	if err != nil {
		return nil, errStatReading(err.Error())
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, apperr.NewFatalErr("Config_OpenError", err.Error())
	}
	defer file.Close()

	c := new(Conf)
	if err := utils.FromJSON(file, c); err != nil {
		return nil, apperr.NewFatalErr("Config_JsonDecodeError", err.Error())
	}

	return &Config{Conf: c, lastModified: stat.ModTime()}, nil
}

var saveMutex sync.Mutex

func Save(c *Config) apperr.Err {
	path, appErr := getConfPath()
	if appErr != nil {
		return appErr
	}

	stat, err := os.Stat(path)
	if err != nil {
		return errStatReading(err.Error())
	}

	if stat.ModTime() != c.lastModified {
		return apperr.NewConflictErr("Config_Conflict", "the configuration has been modified by another request")
	}

	saveMutex.Lock()
	defer saveMutex.Unlock()

	tempPath := path + ".tmp"
	tmpFile, err := os.Create(tempPath)
	if err != nil {
		return apperr.NewFatalErr("Config_TmpFileCreateError", err.Error())
	}

	removeTmpFile := func() {
		if err != nil {
			if removeErr := os.Remove(tempPath); removeErr != nil {
				log.Println("failed to remove temp config file:", removeErr)
			}
		}
	}

	defer removeTmpFile()
	defer tmpFile.Close()

	if err := utils.ToJSON(tmpFile, c.Conf, serializeOptions); err != nil {
		return apperr.NewFatalErr("Config_JsonEncodeError", err.Error())
	}

	tmpFile.Close()
	if err := os.Rename(tempPath, path); err != nil {
		return apperr.NewFatalErr("Config_TmpFileRenameError", err.Error())
	}

	return nil
}

func getConfPath() (string, apperr.Err) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		return "", errEmptyPath
	}

	return path, nil
}
