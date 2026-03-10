package authn

import (
	"log/slog"
	"net/http"
	"strings"
)

const headerAuthorization = "Authorization"

type Middleware struct {
	Config Config
	Parser ClaimsParser
	Logger *slog.Logger
}

func NewMiddleware(config Config, parser ClaimsParser, logger *slog.Logger) Middleware {
	if parser == nil {
		parser = NewParser()
	}
	if logger == nil {
		logger = slog.Default()
	}
	return Middleware{Config: config, Parser: parser, Logger: logger}
}

func (m Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Config.LocalDevBypass {
			m.Logger.Warn("authentication bypass enabled for local development; do not use in production")
			claims := Claims{Subject: m.Config.BypassSubject, Roles: append([]string(nil), m.Config.BypassRoles...)}
			next.ServeHTTP(w, r.WithContext(ContextWithClaims(r.Context(), claims)))
			return
		}

		if m.Config.Mode == ModeDisabled {
			next.ServeHTTP(w, r)
			return
		}

		authz := strings.TrimSpace(r.Header.Get(headerAuthorization))
		if !strings.HasPrefix(authz, "Bearer ") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authz, "Bearer ")
		claims, err := m.Parser.Parse(token)
		if err != nil {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(ContextWithClaims(r.Context(), claims)))
	})
}
