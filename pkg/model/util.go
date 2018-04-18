package model

import (
	"os"
)

func GetEnv() string {
	env := os.Getenv("KUBEENV")
	if "" == env {
		return "unknown"
	}
	return env
}
