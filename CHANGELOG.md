# Changelog

Toutes les évolutions notables de ce projet sont documentées ici.

Le format s'inspire de [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/),
et le versionnement suit [SemVer](https://semver.org/lang/fr/).

## [Unreleased]

## [0.1.0] — 2026-08-26

Première version publique. Un binaire, un `envcheck.yml` dans le dépôt, un
verdict : le poste — ou le runner CI — a-t-il vraiment les prérequis ?

### Ajouté

- Contrôles `command` : le binaire est dans le PATH, avec une plage SemVer
  optionnelle. Exécution via `os/exec` sans shell, délai de 3 s.
- Contrôles `docker` : messages distincts si le client manque, si le démon
  ne répond pas, ou si les permissions sont insuffisantes.
- Contrôles `env` : la variable est définie (ou vide si `allowEmpty`). Sa
  valeur n'apparaît jamais dans le rapport.
- Contrôles `path` : un fichier ou un dossier existe, relativement au dépôt.
- `envcheck init` génère un modèle (`--dry-run` pour l'afficher, `--force`
  pour écraser).
- Rapport texte (couleurs, `--quiet`, `--no-color`) ou JSON pour la CI.
  Codes de sortie `0` (prêt), `1` (échec), `2` (config), `3` (interne).
- Archives Linux et macOS (`amd64`, `arm64`), Windows (`amd64`), et
  `checksums.txt`.

### Corrigé

- Comparaison du modèle `envcheck init` et de `examples/envcheck.yml` sous
  Windows (CRLF au checkout).

[Unreleased]: https://github.com/SkyZonDev/envcheck/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/SkyZonDev/envcheck/releases/tag/v0.1.0
