package server

import (
	"crypto/tls"
	"context"
	"fmt"
	"net/http"
	"os"

	"ordm-main/pkg/config"
)

type HTTPSServer struct {
	server *http.Server
}

func SetupHTTPS(handler http.Handler, port string) (*HTTPSServer, error) {
	if !config.IsProduction() {
		return &HTTPSServer{
			server: &http.Server{
				Addr:    ":" + port,
				Handler: handler,
			},
		}, nil
	}

	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")

	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("certificados SSL não configurados")
	}

	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}

	server := &http.Server{
		Addr:      ":" + port,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	return &HTTPSServer{server: server}, nil
}

func (h *HTTPSServer) Start() error {
	if config.IsProduction() {
		return h.server.ListenAndServeTLS("", "")
	}
	return h.server.ListenAndServe()
}

func (h *HTTPSServer) Shutdown() error {
	return h.server.Shutdown(context.Background())
}

func ForceHTTPS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if config.IsProduction() {
			if r.Header.Get("X-Forwarded-Proto") != "https" && r.TLS == nil {
				httpsURL := "https://" + r.Host + r.RequestURI
				http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
				return
			}
		}
		next(w, r)
	}
}

func SecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		
		if config.IsProduction() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		next(w, r)
	}
}
