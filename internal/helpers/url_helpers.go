package helpers

import (
	"net/url"
	"strings"
)

func GetNonVariantUrl(rawURL string) string {
	normalizedURL := strings.TrimSpace(rawURL)
	if normalizedURL == "" {
		return ""
	}

	parsedURL, err := url.Parse(normalizedURL)
	if err != nil {
		return normalizedURL
	}

	queryValues := parsedURL.Query()
	queryValues.Del("variant")
	parsedURL.RawQuery = queryValues.Encode()

	return parsedURL.String()
}
