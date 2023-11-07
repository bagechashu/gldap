package server

import (
	"fmt"

	"github.com/rs/zerolog"

	"github.com/bagechashu/gldap/internal/monitoring"
	"github.com/bagechashu/gldap/pkg/config"
	"github.com/bagechashu/gldap/pkg/handler"
	"github.com/nmcclain/ldap"
)

type LdapSvc struct {
	c *config.Config
	l *ldap.Server

	monitor monitoring.MonitorInterface
	log     zerolog.Logger
}

func NewServer(opts ...Option) (*LdapSvc, error) {
	options := newOptions(opts...)

	s := LdapSvc{
		log:     options.Logger,
		c:       options.Config,
		monitor: options.Monitor,
	}

	var helper handler.Handler

	loh := handler.NewLDAPOpsHelper()

	// instantiate the helper, if any
	if s.c.Helper.Enabled {
		switch s.c.Helper.Datastore {
		case "config":
			helper = handler.NewConfigHandler(
				handler.Logger(&s.log),
				handler.Config(s.c),
				handler.LDAPHelper(loh),
			)
		case "mysql":
			helper = handler.NewMySQLHandler(
				handler.Logger(&s.log),
				handler.Config(s.c),
				handler.LDAPHelper(loh),
			)
		default:
			return nil, fmt.Errorf("unsupported helper %s", s.c.Helper.Datastore)
		}
		s.log.Info().Str("datastore", s.c.Helper.Datastore).Msg("Using helper")
	}

	backendCounter := -1
	allHandlers := handler.HandlerWrapper{Handlers: make([]handler.Handler, 10), Count: &backendCounter}

	// configure the backends
	s.l = ldap.NewServer()
	s.l.EnforceLDAP = true
	for i, backend := range s.c.Backends {
		var h handler.Handler
		switch backend.Datastore {
		case "ldap":
			h = handler.NewLdapHandler(
				handler.Backend(backend),
				handler.Handlers(allHandlers),
				handler.Logger(&s.log),
				handler.Helper(helper),
				handler.Monitor(s.monitor),
			)
		case "config":
			h = handler.NewConfigHandler(
				handler.Backend(backend),
				handler.Logger(&s.log),
				handler.Config(s.c), // TODO only used to access Users and Groups, move that to dedicated options
				handler.LDAPHelper(loh),
				handler.Monitor(s.monitor),
			)
		case "mysql":
			h = handler.NewMySQLHandler(
				handler.Backend(backend),
				handler.Logger(&s.log),
				handler.Config(s.c),
				handler.LDAPHelper(loh),
				handler.Monitor(s.monitor),
			)
		default:
			return nil, fmt.Errorf("unsupported backend %s", backend.Datastore)
		}
		s.log.Info().Str("datastore", backend.Datastore).Int("position", i).Msg("Loading backend")

		// Only our first backend will answer proper LDAP queries.
		// Note that this could evolve towars something nicer where we would maintain
		// multiple binders in addition to the existing multiple LDAP backends
		if i == 0 {
			s.l.BindFunc("", h)
			s.l.SearchFunc("", h)
			s.l.CloseFunc("", h)
		}
		allHandlers.Handlers[i] = h
		backendCounter++
	}

	monitoring.NewLDAPMonitorWatcher(s.l, s.monitor, &s.log)

	return &s, nil
}

// ListenAndServe listens on the TCP network address s.c.LDAP.Listen
func (s *LdapSvc) ListenAndServe() error {
	s.log.Info().Str("address", s.c.LDAP.Listen).Msg("LDAP server listening")
	return s.l.ListenAndServe(s.c.LDAP.Listen)
}

// ListenAndServeTLS listens on the TCP network address s.c.LDAPS.Listen
func (s *LdapSvc) ListenAndServeTLS() error {
	s.log.Info().Str("address", s.c.LDAPS.Listen).Msg("LDAPS server listening")
	return s.l.ListenAndServeTLS(
		s.c.LDAPS.Listen,
		s.c.LDAPS.Cert,
		s.c.LDAPS.Key,
	)
}

// Shutdown ends listeners by sending true to the ldap serves quit channel
func (s *LdapSvc) Shutdown() {
	s.l.Quit <- true
}
