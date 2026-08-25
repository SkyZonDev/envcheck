package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SkyZonDev/envcheck/internal/core"
)

// globalFlags regroupe les options communes à plusieurs commandes.
// Elles sont déclarées ici (et non dans check.go) pour rester visibles
// même quand on tape juste `envcheck --help`.
type globalFlags struct {
	config  string
	format  string
	quiet   bool
	noColor bool
	version string
}

// Execute construit l'arbre de commandes, l'exécute, et renvoie le code
// de sortie du processus (voir internal/core/exitcode.go). main.go se
// contente d'appeler os.Exit(app.Execute(version)).
func Execute(version string) int {
	flags := &globalFlags{version: version}

	rootCmd := &cobra.Command{
		Use:           "envcheck",
		Short:         "Vérifie que votre poste satisfait les prérequis d'un projet.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `  envcheck
  envcheck --format json
  envcheck init
  envcheck init --dry-run`,
		// Sans sous-commande, `envcheck` se comporte comme `envcheck check`.
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, flags)
		},
	}

	rootCmd.PersistentFlags().StringVar(&flags.config, "config", "", "chemin explicite vers le fichier de configuration")
	rootCmd.PersistentFlags().StringVar(&flags.format, "format", "text", "format de sortie : text ou json (yaml pour init --dry-run)")
	rootCmd.PersistentFlags().BoolVar(&flags.quiet, "quiet", false, "n'afficher que les échecs")
	rootCmd.PersistentFlags().BoolVar(&flags.noColor, "no-color", false, "désactiver la couleur dans la sortie texte")

	rootCmd.AddCommand(newCheckCmd(flags))
	rootCmd.AddCommand(newInitCmd(flags))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "erreur:", err)
		return core.ExitInternalError
	}

	return lastExitCode
}

// lastExitCode porte le code métier des commandes (0/1/2/3). RunE de Cobra
// ne connaît que error/nil ; on le renseigne ici puis main.go le relit.
var lastExitCode = core.ExitOK
