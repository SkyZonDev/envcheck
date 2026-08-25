package render_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/SkyZonDev/envcheck/internal/check"
	"github.com/SkyZonDev/envcheck/internal/core"
	"github.com/SkyZonDev/envcheck/internal/render"
)

func sampleReport(passed bool) core.RunReport {
	status := check.Pass
	if !passed {
		status = check.Fail
	}
	return core.RunReport{
		SchemaVersion: 1,
		ToolVersion:   "0.1.0",
		ConfigPath:    "envcheck.yml",
		Passed:        passed,
		Summary:       core.Summary{Pass: 1},
		Checks: []check.Result{
			{ID: "database-url", Type: "env", Required: true, Status: status, Summary: "test", Hint: "corrige"},
		},
	}
}

func TestJSON_IsOnlyContentAndValid(t *testing.T) {
	var buf bytes.Buffer
	if err := render.JSON(&buf, sampleReport(true)); err != nil {
		t.Fatalf("JSON() erreur: %v", err)
	}

	var decoded core.RunReport
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("la sortie n'est pas un JSON valide: %v\nsortie: %s", err, buf.String())
	}
	if decoded.ConfigPath != "envcheck.yml" {
		t.Fatalf("ConfigPath = %q, attendu envcheck.yml", decoded.ConfigPath)
	}
}

func TestText_ShowsHintOnFailure(t *testing.T) {
	var buf bytes.Buffer
	render.Text(&buf, sampleReport(false), false, false)

	out := buf.String()
	if !strings.Contains(out, "corrige") {
		t.Fatalf("le hint devrait apparaître pour un contrôle en échec, got:\n%s", out)
	}
	if !strings.Contains(out, "environnement non prêt") {
		t.Fatalf("le résumé devrait indiquer 'environnement non prêt', got:\n%s", out)
	}
}

func TestText_QuietHidesPasses(t *testing.T) {
	var buf bytes.Buffer
	render.Text(&buf, sampleReport(true), true, false)

	out := buf.String()
	if strings.Contains(out, "database-url") {
		t.Fatalf("en mode quiet, une règle passée ne devrait pas apparaître, got:\n%s", out)
	}
}

func TestText_ColorWrapsMarks(t *testing.T) {
	var buf bytes.Buffer
	render.Text(&buf, sampleReport(false), false, true)

	out := buf.String()
	if !strings.Contains(out, "\x1b[31m") {
		t.Fatalf("un échec coloré devrait contenir du rouge ANSI, got:\n%s", out)
	}
	if !strings.Contains(out, "corrige") {
		t.Fatalf("le hint doit rester lisible avec la couleur, got:\n%s", out)
	}
}

func TestColorEnabled_DisabledByFlagAndBuffer(t *testing.T) {
	var buf bytes.Buffer
	if render.ColorEnabled(&buf, false) {
		t.Fatal("un Buffer n'est pas un TTY")
	}
	if render.ColorEnabled(os.Stdout, true) {
		t.Fatal("--no-color doit gagner")
	}
}

func TestColorEnabled_NO_COLOR(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if render.ColorEnabled(os.Stdout, false) {
		t.Fatal("NO_COLOR non vide doit désactiver la couleur")
	}
}
