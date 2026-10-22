package app

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox"
	"github.com/traf72/singbox-api/internal/singbox/config"
)

var (
	errOutboundEmptyTag = apperr.NewValidationErr("Outbound_EmptyTag", "outbound tag is empty")
	errOutboundEmptyURL = apperr.NewValidationErr("Outbound_EmptyURL", "outbound URL is empty")
)

func errOutboundInvalidURL(reason string) apperr.Err {
	return apperr.NewValidationErr("Outbound_InvalidURL", fmt.Sprintf("invalid VLESS URL: %s", reason))
}

type OutboundRequest struct {
	Tag string `json:"tag"`
	URL string `json:"url"`
}

type SelectorDefaultRequest struct {
	Tag string `json:"tag"`
}

var uuidRegex = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func (r *OutboundRequest) toVLESSOptions() (*config.VLESSOutboundOptions, apperr.Err) {
	tag := strings.TrimSpace(r.Tag)
	if tag == "" {
		return nil, errOutboundEmptyTag
	}

	rawURL := strings.TrimSpace(r.URL)
	if rawURL == "" {
		return nil, errOutboundEmptyURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, errOutboundInvalidURL(err.Error())
	}
	if !strings.EqualFold(parsed.Scheme, "vless") {
		return nil, errOutboundInvalidURL("scheme must be 'vless'")
	}

	uuid := ""
	if parsed.User != nil {
		uuid = parsed.User.Username()
	}
	if !uuidRegex.MatchString(uuid) {
		return nil, errOutboundInvalidURL("UUID is missing or invalid")
	}

	server := parsed.Hostname()
	if net.ParseIP(server) == nil {
		return nil, errOutboundInvalidURL("server must be a valid IP address")
	}

	port, err := strconv.Atoi(parsed.Port())
	if err != nil || port < 1 || port > 65535 {
		return nil, errOutboundInvalidURL("port must be between 1 and 65535")
	}

	query := parsed.Query()
	if query.Get("type") != "tcp" {
		return nil, errOutboundInvalidURL("type must be 'tcp'")
	}
	if query.Get("security") != "reality" {
		return nil, errOutboundInvalidURL("security must be 'reality'")
	}

	requiredQuery := []struct {
		key  string
		name string
	}{
		{key: "pbk", name: "public key"},
		{key: "fp", name: "fingerprint"},
		{key: "sni", name: "server name"},
		{key: "sid", name: "short ID"},
		{key: "flow", name: "flow"},
	}
	for _, parameter := range requiredQuery {
		if query.Get(parameter.key) == "" {
			return nil, errOutboundInvalidURL(fmt.Sprintf("%s query parameter '%s' is required", parameter.name, parameter.key))
		}
	}

	return &config.VLESSOutboundOptions{
		Tag:         tag,
		Server:      server,
		ServerPort:  port,
		UUID:        uuid,
		Flow:        query.Get("flow"),
		ServerName:  query.Get("sni"),
		Fingerprint: query.Get("fp"),
		PublicKey:   query.Get("pbk"),
		ShortID:     query.Get("sid"),
	}, nil
}

func AddOutbound(request *OutboundRequest, restart bool) apperr.Err {
	options, err := request.toVLESSOptions()
	if err != nil {
		return err
	}

	if err := config.AddVLESSOutbound(options); err != nil {
		return err
	}

	if restart {
		return singbox.Restart()
	}

	return nil
}

func RemoveOutbound(tag string, restart bool) apperr.Err {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return errOutboundEmptyTag
	}

	if err := config.RemoveOutbound(tag); err != nil {
		return err
	}

	if restart {
		return singbox.Restart()
	}

	return nil
}

func SetSelectorDefault(request *SelectorDefaultRequest, restart bool) apperr.Err {
	tag := strings.TrimSpace(request.Tag)
	if tag == "" {
		return errOutboundEmptyTag
	}

	if err := config.SetProxySelectorDefault(tag); err != nil {
		return err
	}

	if restart {
		return singbox.Restart()
	}

	return nil
}
