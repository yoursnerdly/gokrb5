package client

import (
	"encoding/json"
	"fmt"
	"log"
)

// Settings holds optional client settings.
type Settings struct {
	autoPAFXFast            bool
	disablePAFXFast         bool
	assumePreAuthentication bool
	preAuthEType            int32
	logger                  *log.Logger
}

// jsonSettings is used when marshaling the Settings details to JSON format.
type jsonSettings struct {
	AutoPAFXFast            bool
	DisablePAFXFast         bool
	AssumePreAuthentication bool
}

// NewSettings creates a new client settings struct.
func NewSettings(settings ...func(*Settings)) *Settings {
	s := &Settings{
		autoPAFXFast: true,
	}
	for _, set := range settings {
		set(s)
	}
	return s
}

// AutoPAFXFAST used to configure the client to automatically try PA_FX_FAST and retry without it.
//
// s := NewSettings(AutoPAFXFAST(true))
func AutoPAFXFAST(b bool) func(*Settings) {
	return func(s *Settings) {
		s.autoPAFXFast = b
	}
}

// AutoPAFXFAST indicates if the client should automatically try PA_FX_FAST and fallback without it.
func (s *Settings) AutoPAFXFAST() bool {
	return s.autoPAFXFast
}

// DisablePAFXFAST used to configure the client to not use PA_FX_FAST.
//
// s := NewSettings(DisablePAFXFAST(true))
func DisablePAFXFAST(b bool) func(*Settings) {
	return func(s *Settings) {
		s.disablePAFXFast = b
	}
}

// DisablePAFXFAST indicates is the client should disable the use of PA_FX_FAST.
func (s *Settings) DisablePAFXFAST() bool {
	return s.disablePAFXFast
}

// AssumePreAuthentication used to configure the client to assume pre-authentication is required.
//
// s := NewSettings(AssumePreAuthentication(true))
func AssumePreAuthentication(b bool) func(*Settings) {
	return func(s *Settings) {
		s.assumePreAuthentication = b
	}
}

// AssumePreAuthentication indicates if the client should proactively assume using pre-authentication.
func (s *Settings) AssumePreAuthentication() bool {
	return s.assumePreAuthentication
}

// Logger used to configure client with a logger.
//
// s := NewSettings(kt, Logger(l))
func Logger(l *log.Logger) func(*Settings) {
	return func(s *Settings) {
		s.logger = l
	}
}

// Logger returns the client logger instance.
func (s *Settings) Logger() *log.Logger {
	return s.logger
}

// Log will write to the service's logger if it is configured.
func (cl *Client) Log(format string, v ...interface{}) {
	if cl.settings.Logger() != nil {
		cl.settings.Logger().Output(2, fmt.Sprintf(format, v...))
	}
}

// JSON returns a JSON representation of the settings.
func (s *Settings) JSON() (string, error) {
	js := jsonSettings{
		AutoPAFXFast:            s.autoPAFXFast,
		DisablePAFXFast:         s.disablePAFXFast,
		AssumePreAuthentication: s.assumePreAuthentication,
	}
	b, err := json.MarshalIndent(js, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil

}
