package config

import "os"

func GetCorsAllowedOrigins() []string {
	if IsLocal() {
		return []string{"http://localhost:3000"}
	}
	return []string{os.Getenv("FRONT_END_ORIGIN")}
}
