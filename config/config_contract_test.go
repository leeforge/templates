package config

import "testing"

func TestConfigMustContainBackendParitySections(t *testing.T) {
	cfg := Default()
	if cfg.Database == nil || cfg.Security == nil || cfg.AccessControl == nil || cfg.Plugins == nil {
		t.Fatalf("examples config missing backend parity sections")
	}
}
