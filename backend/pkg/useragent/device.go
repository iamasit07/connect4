package useragent

import (
	"net/http"
	"strings"
)

// ExtractDeviceInfo parses User-Agent header to extract device information
func ExtractDeviceInfo(r *http.Request) string {
	ua := r.Header.Get("User-Agent")
	if ua == "" {
		return "Unknown Device"
	}

	// Parse browser
	browser := "Unknown Browser"
	if strings.Contains(ua, "Chrome/") && !strings.Contains(ua, "Edg") {
		browser = "Chrome"
	} else if strings.Contains(ua, "Safari/") && !strings.Contains(ua, "Chrome") {
		browser = "Safari"
	} else if strings.Contains(ua, "Firefox/") {
		browser = "Firefox"
	} else if strings.Contains(ua, "Edg/") {
		browser = "Edge"
	}

	// Extract version (first number after browser name)
	version := ""
	browserKey := browser + "/"
	if idx := strings.Index(ua, browserKey); idx != -1 {
		versionStart := idx + len(browserKey)
		versionEnd := versionStart
		for versionEnd < len(ua) && (ua[versionEnd] >= '0' && ua[versionEnd] <= '9' || ua[versionEnd] == '.') {
			versionEnd++
		}
		if versionEnd > versionStart {
			version = ua[versionStart:versionEnd]
			// Take only major version
			if dotIdx := strings.Index(version, "."); dotIdx != -1 {
				version = version[:dotIdx]
			}
		}
	}

	// Parse OS
	os := "Unknown OS"
	if strings.Contains(ua, "Windows") {
		if strings.Contains(ua, "Windows NT 10.0") {
			os = "Windows 10/11"
		} else if strings.Contains(ua, "Windows NT 6.3") {
			os = "Windows 8.1"
		} else if strings.Contains(ua, "Windows NT 6.1") {
			os = "Windows 7"
		} else {
			os = "Windows"
		}
	} else if strings.Contains(ua, "Mac OS X") {
		os = "macOS"
	} else if strings.Contains(ua, "Linux") {
		os = "Linux"
	} else if strings.Contains(ua, "Android") {
		os = "Android"
	} else if strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") {
		os = "iOS"
	}

	// Combine browser and OS
	if version != "" {
		return browser + " " + version + " on " + os
	}
	return browser + " on " + os
}

// ExtractIPAddress gets the real IP address from the request
// Handles proxies and load balancers by checking X-Forwarded-For and X-Real-IP headers
func ExtractIPAddress(r *http.Request) string {
	// Try X-Forwarded-For header first (used by most proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Try X-Real-IP header (used by nginx)
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
