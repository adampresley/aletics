package services

import (
	"net"
	"net/http"
	"strings"
)

func GetIP(r *http.Request) string {
	// Check X-Forwarded-For (comma-separated list of IPs)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP (client's real IP, assuming proxies are trusted)
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fallback to X-Real-Ip
	realIP := r.Header.Get("X-Real-Ip")
	if realIP != "" {
		return realIP
	}

	// Final fallback: use RemoteAddr (e.g., "192.168.1.1:50000")
	// Extract just the IP part
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // Return full if parsing fails
	}
	return ip
}
