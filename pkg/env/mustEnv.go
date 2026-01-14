package env

import (
	"log"
	"os"
)

func MustEnv(name string, omitempty bool) string {
	value := os.Getenv(name)
	if value == "" && omitempty == false {
		log.Fatalf("environment variable %s is required", name)
	}
	return value
}
