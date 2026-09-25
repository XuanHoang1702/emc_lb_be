package service

import (
	"fmt"
	"net/url"
	"strings"
)

func ValidateCallbackURL(rawURL string, allowedHosts []string, env string) error {
	if rawURL == "" {
		return nil // Optional URLs are allowed
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	if u.User != nil {
		return fmt.Errorf("URL userinfo/credentials not allowed")
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}

	inputHost := u.Hostname()

	if u.Scheme == "http" {
		if env == "production" {
			return fmt.Errorf("http scheme not allowed in production")
		}
		// In non-production, http is only allowed for localhost
		if inputHost != "localhost" && inputHost != "127.0.0.1" {
			return fmt.Errorf("http scheme only allowed for localhost in development")
		}
	}

	// Bypass whitelist for localhost in dev
	isLocalDev := env != "production" && (inputHost == "localhost" || inputHost == "127.0.0.1")
	if !isLocalDev {

		// Host whitelist validation
		if len(allowedHosts) > 0 {
			allowed := false
			for _, allowedHost := range allowedHosts {
				if strings.EqualFold(inputHost, allowedHost) {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("host %s is not in the allowed callback hosts", inputHost)
			}
		}
	}

	return nil
}
