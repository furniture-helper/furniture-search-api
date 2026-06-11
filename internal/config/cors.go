package config

import (
	"fmt"
	"os"
)

func GetCorsAllowedOrigins() []string {
	if IsLocal() {
		return []string{"http://localhost:3000"}
	}

	return []string{
		fmt.Sprintf("https://%s", os.Getenv("FRONTEND_ORIGIN")),
		fmt.Sprintf("https://www.%s", os.Getenv("FRONTEND_ORIGIN")),
	}
}
