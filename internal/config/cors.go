package config

import (
	"fmt"
	"os"
	"strings"
)

func GetCorsAllowedOrigins() []string {
	if IsLocal() {
		return []string{"http://localhost:3000"}
	}

	return parseCorsOrigins(strings.TrimSpace(os.Getenv("FRONTEND_ORIGINS")))
}

func parseCorsOrigins(origins string) []string {
	allowedOrigins := []string{}
	seenOrigins := map[string]struct{}{}

	for _, origin := range strings.Split(origins, ",") {
		for _, expandedOrigin := range expandOrigin(origin) {
			if _, seen := seenOrigins[expandedOrigin]; seen {
				continue
			}

			seenOrigins[expandedOrigin] = struct{}{}
			allowedOrigins = append(allowedOrigins, expandedOrigin)
		}
	}

	return allowedOrigins
}

func expandOrigin(origin string) []string {
	trimmedOrigin := strings.TrimSpace(strings.TrimSuffix(origin, "/"))
	if trimmedOrigin == "" {
		return []string{}
	}

	if strings.HasPrefix(trimmedOrigin, "http://") || strings.HasPrefix(trimmedOrigin, "https://") {
		return []string{trimmedOrigin}
	}

	normalizedDomain := strings.TrimPrefix(trimmedOrigin, "www.")
	return []string{
		fmt.Sprintf("https://%s", normalizedDomain),
		fmt.Sprintf("https://www.%s", normalizedDomain),
	}
}
