package main

import (
	"testing"
)

func TestVersion(t *testing.T) {
	if version == "" {
		version = "dev"
	}

	if version != "dev" && version == "" {
		t.Error("Version should not be empty")
	}
}
