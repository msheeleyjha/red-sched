package middleware

import (
	"github.com/rs/cors"
)

// NewCORSHandler creates a CORS handler with the given frontend URL
func NewCORSHandler(frontendURL string) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{frontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
}
