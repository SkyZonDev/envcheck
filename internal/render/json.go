package render

import (
	"encoding/json"
	"io"

	"github.com/SkyZonDev/envcheck/internal/core"
)

// JSON écrit report comme SEUL contenu de w. C'est le contrat d'intégration
// CI de la spec (section 2) : stdout ne doit jamais contenir autre chose
// qu'un objet JSON valide en mode --format json.
func JSON(w io.Writer, report core.RunReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
