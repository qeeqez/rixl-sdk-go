// Package exenv has small helpers shared across the example programs.
package exenv

import (
	"log"
	"os"
)

func MustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing %s", k)
	}
	return v
}

func EnvOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func Deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func Int32Or(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}
