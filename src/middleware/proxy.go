package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/casapps/casreg/src/config"
	"github.com/sirupsen/logrus"
)

// Default trusted IP ranges (private networks and loopback)
var defaultTrustedRanges = []string{
	"127.0.0.1/32",     // IPv4 loopback
	"::1/128",          // IPv6 loopback
	"10.0.0.0/8",       // Private IPv4 class A
	"172.16.0.0/12",    // Private IPv4 class B
	"192.168.0.0/16",   // Private IPv4 class C
	"fc00::/7",         // Unique local IPv6 addresses
	"fe80::/10",        // Link-local IPv6 addresses
}

// ProxyHeaders middleware handles reverse proxy headers and extracts real client IP
func ProxyHeaders(cfg *config.Config) func(http.Handler) http.Handler {
	// Parse trusted IP ranges
	trustedNets := parseTrustedIPs(cfg.Security.TrustedIPs)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the immediate remote address
			remoteAddr := r.RemoteAddr
			if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
				remoteAddr = host
			}

			// Check if the request is from a trusted proxy
			isTrusted := isIPTrusted(remoteAddr, trustedNets)

			var realIP string

			if isTrusted {
				// Process proxy headers in order of preference
				realIP = extractRealIP(r)

				// Log proxy header processing for debugging
				logrus.WithFields(logrus.Fields{
					"remote_addr":         remoteAddr,
					"real_ip":             realIP,
					"x_forwarded_for":     r.Header.Get("X-Forwarded-For"),
					"x_real_ip":           r.Header.Get("X-Real-IP"),
					"cf_connecting_ip":    r.Header.Get("CF-Connecting-IP"),
					"true_client_ip":      r.Header.Get("True-Client-IP"),
				}).Debug("Processing proxy headers")
			} else {
				// Not from trusted proxy, use direct connection address
				realIP = remoteAddr

				// Log untrusted proxy attempt if headers are present
				if r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Real-IP") != "" {
					logrus.WithFields(logrus.Fields{
						"remote_addr":     remoteAddr,
						"x_forwarded_for": r.Header.Get("X-Forwarded-For"),
						"x_real_ip":       r.Header.Get("X-Real-IP"),
					}).Warn("Proxy headers from untrusted source ignored")
				}
			}

			// Set real IP in context
			ctx := context.WithValue(r.Context(), "real_ip", realIP)

			// Update RemoteAddr for compatibility with existing code
			r.RemoteAddr = realIP

			// Process other forwarded headers if trusted
			if isTrusted {
				// X-Forwarded-Proto (http/https)
				if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
					r.URL.Scheme = proto
				}

				// X-Forwarded-Host
				if host := r.Header.Get("X-Forwarded-Host"); host != "" {
					r.Host = host
				}

				// X-Forwarded-Port
				if port := r.Header.Get("X-Forwarded-Port"); port != "" {
					r.URL.Host = net.JoinHostPort(r.Host, port)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractRealIP extracts the real client IP from proxy headers
func extractRealIP(r *http.Request) string {
	// Priority order for different proxy headers

	// 1. CF-Connecting-IP (Cloudflare)
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		if parsedIP := parseIP(ip); parsedIP != "" {
			return parsedIP
		}
	}

	// 2. True-Client-IP (Akamai, Cloudflare)
	if ip := r.Header.Get("True-Client-IP"); ip != "" {
		if parsedIP := parseIP(ip); parsedIP != "" {
			return parsedIP
		}
	}

	// 3. X-Real-IP (Nginx)
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		if parsedIP := parseIP(ip); parsedIP != "" {
			return parsedIP
		}
	}

	// 4. X-Forwarded-For (Standard, can contain multiple IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// We want the first (leftmost) IP which is the original client
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if parsedIP := parseIP(ip); parsedIP != "" {
				return parsedIP
			}
		}
	}

	// 5. Forwarded header (RFC 7239)
	if forwarded := r.Header.Get("Forwarded"); forwarded != "" {
		if ip := parseForwardedHeader(forwarded); ip != "" {
			return ip
		}
	}

	// 6. X-Original-Forwarded-For
	if ip := r.Header.Get("X-Original-Forwarded-For"); ip != "" {
		ips := strings.Split(ip, ",")
		for _, ipStr := range ips {
			ipStr = strings.TrimSpace(ipStr)
			if parsedIP := parseIP(ipStr); parsedIP != "" {
				return parsedIP
			}
		}
	}

	// Fallback to remote address
	remoteAddr := r.RemoteAddr
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}

// parseIP validates and returns a clean IP address
func parseIP(ipStr string) string {
	ipStr = strings.TrimSpace(ipStr)
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}
	return ip.String()
}

// parseForwardedHeader parses RFC 7239 Forwarded header
func parseForwardedHeader(forwarded string) string {
	// Forwarded header format: for=192.0.2.60;proto=http;by=203.0.113.43
	parts := strings.Split(forwarded, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "for=") {
			ip := strings.TrimPrefix(part, "for=")
			// Remove quotes if present
			ip = strings.Trim(ip, "\"")
			// Handle IPv6 in brackets [2001:db8::1]
			ip = strings.Trim(ip, "[]")
			if parsedIP := parseIP(ip); parsedIP != "" {
				return parsedIP
			}
		}
	}
	return ""
}

// parseTrustedIPs parses trusted IP ranges from config and defaults
func parseTrustedIPs(configTrusted []string) []*net.IPNet {
	var nets []*net.IPNet

	// Parse default trusted ranges
	for _, cidr := range defaultTrustedRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to parse default trusted range: %s", cidr)
			continue
		}
		nets = append(nets, ipNet)
	}

	// Parse config trusted IPs
	for _, cidr := range configTrusted {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}

		// If no CIDR notation, assume /32 for IPv4 or /128 for IPv6
		if !strings.Contains(cidr, "/") {
			ip := net.ParseIP(cidr)
			if ip == nil {
				logrus.Warnf("Invalid IP address in trusted IPs: %s", cidr)
				continue
			}
			if ip.To4() != nil {
				cidr = cidr + "/32"
			} else {
				cidr = cidr + "/128"
			}
		}

		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to parse trusted IP range: %s", cidr)
			continue
		}
		nets = append(nets, ipNet)
	}

	logrus.Infof("Loaded %d trusted IP ranges", len(nets))
	return nets
}

// isIPTrusted checks if an IP is in the trusted ranges
func isIPTrusted(ipStr string, trustedNets []*net.IPNet) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, ipNet := range trustedNets {
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}
