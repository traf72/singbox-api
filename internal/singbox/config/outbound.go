package config

import (
	"fmt"
	"slices"

	"github.com/traf72/singbox-api/internal/apperr"
)

const proxySelectorTag = "proxy"

type VLESSOutboundOptions struct {
	Tag         string
	Server      string
	ServerPort  int
	UUID        string
	Flow        string
	ServerName  string
	Fingerprint string
	PublicKey   string
	ShortID     string
}

func errOutboundTagExists(tag string) apperr.Err {
	return apperr.NewConflictErr("Outbound_TagExists", fmt.Sprintf("outbound tag '%s' already exists", tag))
}

func errOutboundNotFound(tag string) apperr.Err {
	return apperr.NewNotFoundErr("Outbound_NotFound", fmt.Sprintf("outbound tag '%s' was not found", tag))
}

func errProxySelectorNotFound() apperr.Err {
	return apperr.NewNotFoundErr("Outbound_ProxySelectorNotFound", fmt.Sprintf("selector outbound tag '%s' was not found", proxySelectorTag))
}

func errSelectorMemberNotFound(tag string) apperr.Err {
	return apperr.NewNotFoundErr("Outbound_SelectorMemberNotFound", fmt.Sprintf("outbound tag '%s' is not a member of selector '%s'", tag, proxySelectorTag))
}

func AddVLESSOutbound(options *VLESSOutboundOptions) apperr.Err {
	c, err := Load()
	if err != nil {
		return err
	}

	if err := addVLESSOutbound(options, c.Conf); err != nil {
		return err
	}

	return Save(c)
}

func addVLESSOutbound(options *VLESSOutboundOptions, conf *Conf) apperr.Err {
	if slices.ContainsFunc(conf.Outbounds, func(item *outbound) bool {
		return item != nil && item.Tag == options.Tag
	}) {
		return errOutboundTagExists(options.Tag)
	}

	selector := findProxySelector(conf)
	if selector == nil {
		return errProxySelectorNotFound()
	}

	conf.Outbounds = append(conf.Outbounds, &outbound{
		Type:       "vless",
		Tag:        options.Tag,
		Server:     options.Server,
		ServerPort: options.ServerPort,
		UUID:       options.UUID,
		Flow:       options.Flow,
		TLS: &tls{
			Enabled:    true,
			ServerName: options.ServerName,
			UTLS: &utls{
				Enabled:     true,
				Fingerprint: options.Fingerprint,
			},
			Reality: &reality{
				Enabled:   true,
				PublicKey: options.PublicKey,
				ShortID:   options.ShortID,
			},
		},
	})
	if !slices.Contains(selector.Outbounds, options.Tag) {
		selector.Outbounds = append(selector.Outbounds, options.Tag)
	}

	return nil
}

func RemoveOutbound(tag string) apperr.Err {
	c, err := Load()
	if err != nil {
		return err
	}

	if err := removeOutbound(tag, c.Conf); err != nil {
		return err
	}

	return Save(c)
}

func SetProxySelectorDefault(tag string) apperr.Err {
	c, err := Load()
	if err != nil {
		return err
	}

	if err := setProxySelectorDefault(tag, c.Conf); err != nil {
		return err
	}

	return Save(c)
}

func setProxySelectorDefault(tag string, conf *Conf) apperr.Err {
	selector := findProxySelector(conf)
	if selector == nil {
		return errProxySelectorNotFound()
	}
	if !slices.Contains(selector.Outbounds, tag) {
		return errSelectorMemberNotFound(tag)
	}

	selector.Default = tag
	return nil
}

func removeOutbound(tag string, conf *Conf) apperr.Err {
	index := slices.IndexFunc(conf.Outbounds, func(item *outbound) bool {
		return item != nil && item.Tag == tag
	})
	if index == -1 {
		return errOutboundNotFound(tag)
	}

	conf.Outbounds = slices.Delete(conf.Outbounds, index, index+1)

	selector := findProxySelector(conf)
	if selector != nil {
		selector.Outbounds = slices.DeleteFunc(selector.Outbounds, func(item string) bool {
			return item == tag
		})
		if selector.Default == tag {
			selector.Default = ""
			if len(selector.Outbounds) > 0 {
				selector.Default = selector.Outbounds[0]
			}
		}
	}

	return nil
}

func findProxySelector(conf *Conf) *outbound {
	index := slices.IndexFunc(conf.Outbounds, func(item *outbound) bool {
		return item != nil && item.Type == "selector" && item.Tag == proxySelectorTag
	})
	if index == -1 {
		return nil
	}

	return conf.Outbounds[index]
}
