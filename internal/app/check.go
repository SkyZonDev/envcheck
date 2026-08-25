package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SkyZonDev/envcheck/internal/check"
	"github.com/SkyZonDev/envcheck/internal/config"
	"github.com/SkyZonDev/envcheck/internal/core"
	"github.com/SkyZonDev/envcheck/internal/platform"
	"github.com/SkyZonDev/envcheck/internal/render"
)

func newCheckCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Charge, valide et exécute les règles d'envcheck.yml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, flags)
		},
	}
}

// runCheck est le point de jonction unique entre le CLI et le moteur.
// Flux : découverte/chargement de la config -> exécution des règles
// (internal/core.Run) -> rendu texte ou JSON -> code de sortie.
func runCheck(cmd *cobra.Command, flags *globalFlags) error {
	path, err := resolveConfigPath(flags)
	if err != nil {
		return reportConfigError(cmd, err)
	}

	file, err := config.Load(path)
	if err != nil {
		return reportConfigError(cmd, err)
	}

	switch flags.format {
	case "text", "json":
	default:
		fmt.Fprintf(cmd.ErrOrStderr(), "format %q invalide (attendu : text ou json)\n", flags.format)
		lastExitCode = core.ExitConfigError
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		lastExitCode = core.ExitInternalError
		return fmt.Errorf("impossible de déterminer le répertoire courant : %w", err)
	}

	rt := check.Runtime{
		CWD:  cwd,
		Env:  envMap(),
		Exec: platform.NewExecRunner(),
		FS:   platform.OSFS{},
	}

	report := core.Run(context.Background(), file, path, flags.version, check.DefaultRegistry(), rt)

	if flags.format == "json" {
		// Contrat de la spec (section 2) : en mode JSON, stdout ne contient
		// QUE l'objet JSON. Tout diagnostic humain reste en texte.
		if err := render.JSON(cmd.OutOrStdout(), report); err != nil {
			lastExitCode = core.ExitInternalError
			return fmt.Errorf("écriture du rapport JSON : %w", err)
		}
	} else {
		render.Text(cmd.OutOrStdout(), report, flags.quiet, render.ColorEnabled(cmd.OutOrStdout(), flags.noColor))
	}

	if report.Passed {
		lastExitCode = core.ExitOK
	} else {
		lastExitCode = core.ExitFail
	}
	return nil
}

// resolveConfigPath applique --config si fourni, sinon découvre le fichier
// dans le répertoire courant (jamais dans les parents — section 2 spec).
func resolveConfigPath(flags *globalFlags) (string, error) {
	if flags.config != "" {
		return flags.config, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("impossible de déterminer le répertoire courant : %w", err)
	}
	return config.Discover(cwd)
}

// reportConfigError distingue une erreur que l'utilisateur peut corriger
// (fichier absent, YAML invalide, schéma invalide -> code 2, ErrNotFound ou
// *config.Error) d'une erreur interne inattendue (code 3).
func reportConfigError(cmd *cobra.Command, err error) error {
	var cfgErr *config.Error
	if errors.Is(err, config.ErrNotFound) || errors.As(err, &cfgErr) {
		fmt.Fprintln(cmd.ErrOrStderr(), "erreur de configuration:", err)
		lastExitCode = core.ExitConfigError
		return nil
	}
	lastExitCode = core.ExitInternalError
	return err
}

// envMap copie os.Environ() sous forme de map — la seule interaction du
// contrôle env avec le système réel, sans jamais faire fuiter les valeurs
// ailleurs que dans cette map interne au processus.
func envMap() map[string]string {
	env := os.Environ()
	m := make(map[string]string, len(env))
	for _, kv := range env {
		if k, v, ok := strings.Cut(kv, "="); ok {
			m[k] = v
		}
	}
	return m
}
