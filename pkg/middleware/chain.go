package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"ordm-main/pkg/auth"
)

func AuthMiddleware(authManager *auth.UserManager) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Verificar se há usuário ativo
			activeUser := authManager.GetActiveUser()
			if activeUser == nil {
				http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
				return
			}

			// Adicionar dados do usuário ao contexto
			ctx := context.WithValue(r.Context(), "user_id", activeUser.ID)
			ctx = context.WithValue(ctx, "username", activeUser.Username)

			next(w, r.WithContext(ctx))
		}
	}
}

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrapper para capturar status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(ww, r)

		duration := time.Since(start)
		log.Printf("[%s] %s %s - %d - %v", r.Method, r.RemoteAddr, r.URL.Path, ww.statusCode, duration)
	}
}

func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Rate limiting simplificado - implementação futura
		// Por enquanto, apenas passa adiante
		next(w, r)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
