package core

import (
	"context"
	"fmt"
	"time"

	"github.com/SkyZonDev/envcheck/internal/check"
	"github.com/SkyZonDev/envcheck/internal/config"
)

// Run exécute chaque règle déclarée dans file, DANS L'ORDRE DU FICHIER et
// de façon séquentielle (section 6 de la spec : rapport stable, diagnostics
// faciles à comparer ; la parallélisation pourra être ajoutée plus tard
// sans changer l'ordre d'affichage). Une règle en échec ou en erreur
// n'empêche jamais l'évaluation des suivantes.
func Run(ctx context.Context, file *config.File, configPath, toolVersion string, registry *check.Registry, rt check.Runtime) RunReport {
	started := time.Now()
	if rt.Now != nil {
		started = rt.Now()
	}

	results := make([]check.Result, 0, len(file.Checks))
	var summary Summary

	for _, def := range file.Checks {
		result := evaluate(ctx, def, registry, rt)
		results = append(results, result)

		switch result.Status {
		case check.Pass:
			summary.Pass++
		case check.Fail:
			summary.Fail++
		case check.Warn:
			summary.Warn++
		case check.Error:
			summary.Error++
		}
	}

	return RunReport{
		SchemaVersion: 1,
		ToolVersion:   toolVersion,
		ConfigPath:    configPath,
		StartedAt:     started.Format(time.RFC3339),
		DurationMS:    time.Since(started).Milliseconds(),
		// Un échec optionnel devient "warn" (jamais "fail"), donc il suffit
		// de vérifier fail/error pour savoir si l'environnement est prêt.
		Passed:  summary.Fail == 0 && summary.Error == 0,
		Summary: summary,
		Checks:  results,
	}
}

func evaluate(ctx context.Context, def config.Check, registry *check.Registry, rt check.Runtime) check.Result {
	checker, ok := registry.Get(def.Type)
	if !ok {
		status := check.Error
		if !def.IsRequired() {
			status = check.Warn
		}
		return check.Result{
			ID:       def.ID,
			Type:     def.Type,
			Required: def.IsRequired(),
			Status:   status,
			Summary:  fmt.Sprintf("le type %q n'est pas encore implémenté", def.Type),
			Hint:     def.Hint,
		}
	}
	return checker.Run(ctx, def, rt)
}
