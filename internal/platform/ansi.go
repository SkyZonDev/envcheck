package platform

import "regexp"

// ansiEscape couvre les séquences CSI les plus courantes (couleurs, styles)
// que certains binaires injectent dans `--version`. On les retire avant
// d'extraire un numéro de version, jamais pour les mettre dans un rapport.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

// StripANSI retire les séquences d'échappement ANSI de s.
func StripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}
