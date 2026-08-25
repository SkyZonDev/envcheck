package check_test

import (
	"testing"

	"github.com/SkyZonDev/envcheck/internal/check"
	"github.com/SkyZonDev/envcheck/internal/config"
)

func TestDefaultRegistry_IncludesAllMVPTypes(t *testing.T) {
	r := check.DefaultRegistry()

	if _, ok := r.Get(config.TypeEnv); !ok {
		t.Fatal("env devrait rester enregistré")
	}
	if _, ok := r.Get(config.TypePath); !ok {
		t.Fatal("path devrait être enregistré (M3)")
	}
	if _, ok := r.Get(config.TypeCommand); !ok {
		t.Fatal("command devrait être enregistré (M3)")
	}
	if _, ok := r.Get(config.TypeDocker); !ok {
		t.Fatal("docker devrait être enregistré (M4)")
	}
}
