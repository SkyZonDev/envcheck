package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/SkyZonDev/envcheck/internal/config"
	"github.com/SkyZonDev/envcheck/internal/core"
	"github.com/SkyZonDev/envcheck/internal/template"
)

func newInitCmd(flags *globalFlags) *cobra.Command {
	var dryRun bool
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Crée un fichier envcheck.yml à partir d'un modèle",
		Long:  "Écrit envcheck.yml dans le répertoire courant. Sans --force, refuse d'écraser un fichier existant (envcheck.yml ou .envcheck.yml).",
		Example: `  envcheck init
  envcheck init --dry-run
  envcheck init --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				lastExitCode = core.ExitInternalError
				return fmt.Errorf("impossible de déterminer le répertoire courant : %w", err)
			}
			return runInit(cmd, initOptions{
				dir:    cwd,
				dryRun: dryRun,
				force:  force,
				format: flags.format,
			})
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "afficher le modèle sans écrire de fichier")
	cmd.Flags().BoolVar(&force, "force", false, "écraser un fichier existant")

	return cmd
}

type initOptions struct {
	dir    string
	dryRun bool
	force  bool
	format string
}

func runInit(cmd *cobra.Command, opts initOptions) error {
	if opts.dryRun {
		if opts.format == "json" {
			fmt.Fprintln(cmd.ErrOrStderr(), "init --dry-run n'accepte que le format yaml")
			lastExitCode = core.ExitConfigError
			return nil
		}
		fmt.Fprint(cmd.OutOrStdout(), template.DefaultConfig)
		lastExitCode = core.ExitOK
		return nil
	}

	existing, err := config.Discover(opts.dir)
	if err == nil && !opts.force {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s existe déjà ; passez --force pour l'écraser\n", existing)
		lastExitCode = core.ExitConfigError
		return nil
	}
	if err != nil && !errors.Is(err, config.ErrNotFound) {
		lastExitCode = core.ExitInternalError
		return err
	}

	path := filepath.Join(opts.dir, "envcheck.yml")
	if err := os.WriteFile(path, []byte(template.DefaultConfig), 0o644); err != nil {
		lastExitCode = core.ExitInternalError
		return fmt.Errorf("écriture de %s : %w", path, err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "écrit", path)
	lastExitCode = core.ExitOK
	return nil
}
