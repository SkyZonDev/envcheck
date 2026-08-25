package render

import (
	"io"
	"os"
)

// ColorEnabled applique le contrat de la spec (section 2) : pas de couleur
// si --no-color, si NO_COLOR est non vide, ou si w n'est pas un TTY.
func ColorEnabled(w io.Writer, noColor bool) bool {
	if noColor {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
