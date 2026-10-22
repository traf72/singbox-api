package config

import (
	"fmt"
	"slices"

	"github.com/traf72/singbox-api/internal/apperr"
)

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

func removeOutbound(tag string, conf *Conf) apperr.Err {
	index := slices.IndexFunc(conf.Outbounds, func(item *outbound) bool {
		return item != nil && item.Tag == tag
	})
	if index == -1 {
		return errOutboundNotFound(tag)
	}

	conf.Outbounds = slices.Delete(conf.Outbounds, index, index+1)
	return nil
}
