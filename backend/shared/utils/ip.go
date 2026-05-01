package utils

import (
	"net/http"
	"strings"
)

// GetIPAddress extracts the client IP address from the request
// Checks X-Forwarded-For and X-Real-IP headers before falling back to RemoteAddr
func GetIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header (proxy/load balancer)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (nginx proxy)
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	// RemoteAddr format is "IP:port", we only want the IP
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}

	return addr
}
