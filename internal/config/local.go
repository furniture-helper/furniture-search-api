package config

import "os"

func IsLocal() bool {
	return os.Getenv("LOCAL") == "1"
}
