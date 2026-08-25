package render

import (
	"fmt"
	"io"

	"github.com/SkyZonDev/envcheck/internal/check"
	"github.com/SkyZonDev/envcheck/internal/core"
)

const (
	passMark = "✓"
	failMark = "✗"
	warnMark = "!"
	errMark  = "‼"

	ansiReset   = "\x1b[0m"
	ansiGreen   = "\x1b[32m"
	ansiRed     = "\x1b[31m"
	ansiYellow  = "\x1b[33m"
	ansiMagenta = "\x1b[35m"
	ansiDim     = "\x1b[2m"
)

// Text rend report au format humain montré en section 2 de la spec.
// quiet n'affiche que les lignes qui ne sont pas "pass". color wrappe les
// marqueurs (et le hint) en ANSI ; l'appelant décide via ColorEnabled.
func Text(w io.Writer, report core.RunReport, quiet, color bool) {
	for _, r := range report.Checks {
		if quiet && r.Status == check.Pass {
			continue
		}
		fmt.Fprintf(w, "%s %-15s %s\n", paint(color, statusColor(r.Status), markFor(r.Status)), r.ID, r.Summary)
		if r.Status != check.Pass && r.Hint != "" {
			fmt.Fprintf(w, "%s\n", paint(color, ansiDim, "  ↳ "+r.Hint))
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%d/%d contrôles requis réussis — %s.\n",
		countPassedRequired(report), countRequired(report),
		paint(color, overallColor(report.Passed), overallLabel(report.Passed)))
}

func statusColor(s check.Status) string {
	switch s {
	case check.Pass:
		return ansiGreen
	case check.Warn:
		return ansiYellow
	case check.Error:
		return ansiMagenta
	default:
		return ansiRed
	}
}

func overallColor(passed bool) string {
	if passed {
		return ansiGreen
	}
	return ansiRed
}

func paint(enabled bool, code, s string) string {
	if !enabled || s == "" {
		return s
	}
	return code + s + ansiReset
}

func markFor(s check.Status) string {
	switch s {
	case check.Pass:
		return passMark
	case check.Warn:
		return warnMark
	case check.Error:
		return errMark
	default:
		return failMark
	}
}

func countRequired(report core.RunReport) int {
	n := 0
	for _, r := range report.Checks {
		if r.Required {
			n++
		}
	}
	return n
}

func countPassedRequired(report core.RunReport) int {
	n := 0
	for _, r := range report.Checks {
		if r.Required && r.Status == check.Pass {
			n++
		}
	}
	return n
}

func overallLabel(passed bool) string {
	if passed {
		return "environnement prêt"
	}
	return "environnement non prêt"
}
