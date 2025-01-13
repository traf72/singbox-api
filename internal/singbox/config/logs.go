package config

import (
	"fmt"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
)

func errInvalidLogLevel(l string) apperr.Err {
	return apperr.NewValidationErr("LogLevel_Invalid", fmt.Sprintf("Invalid log level '%s'", l))
}

type LogLevel string

const (
	Trace LogLevel = "trace"
	Debug LogLevel = "debug"
	Info  LogLevel = "info"
	Warn  LogLevel = "warn"
	Error LogLevel = "error"
	Fatal LogLevel = "fatal"
	Panic LogLevel = "panic"
)

func (l LogLevel) String() string {
	return strings.ToLower(strings.TrimSpace(string(l)))
}

func (l LogLevel) isValid() bool {
	level := LogLevel(l.String())
	switch level {
	case Trace, Debug, Info, Warn, Error, Fatal, Panic:
		return true
	default:
		return false
	}
}

func EnableLog(l LogLevel) apperr.Err {
	if l != "" && !l.isValid() {
		return errInvalidLogLevel(string(l))
	}

	c, err := Load()
	if err != nil {
		return err
	}

	c.Conf.Log.Disabled = false
	if l != "" {
		c.Conf.Log.Level = l.String()
	}

	if err := Save(c); err != nil {
		return err
	}

	return nil
}

func DisableLog() apperr.Err {
	c, err := Load()
	if err != nil {
		return err
	}

	c.Conf.Log.Disabled = true
	if err := Save(c); err != nil {
		return err
	}

	return nil
}

func SetLogLevel(l LogLevel) apperr.Err {
	if !l.isValid() {
		return errInvalidLogLevel(string(l))
	}

	c, err := Load()
	if err != nil {
		return err
	}

	c.Conf.Log.Level = l.String()
	if err := Save(c); err != nil {
		return err
	}

	return nil
}
