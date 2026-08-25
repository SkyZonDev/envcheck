# Changelog

Toutes les évolutions notables de ce projet sont documentées ici.

Le format s'inspire de [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/),
et le versionnement suit [SemVer](https://semver.org/lang/fr/).

## [Unreleased]

## [0.1.0] — 2026-08-26

### Ajouté

- Documentation utilisateur : README, `docs/fonctionnement.md` (pipeline et
  architecture), `docs/configuration.md` (schéma YAML), `docs/cli.md`
  (commandes, rapport JSON, CI).
- Contrôle `docker` : distingue client introuvable, daemon inaccessible,
  permissions insuffisantes et timeout (3 s). Aucun test n'exige un démon.
- `envcheck init`, `--dry-run` (stdout YAML uniquement) et `--force`.
- Coloration ANSI du rendu texte, coupée hors TTY, avec `--no-color` et `NO_COLOR`.
- Contrôles `command` (SemVer, `os/exec` sans shell) et `path`.
- Contrôle `env` : présence d'une variable, jamais sa valeur.
- Chargement YAML strict (`KnownFields`), découverte limitée au répertoire courant.
- GoReleaser : archives Linux/macOS (`amd64`, `arm64`), Windows (`amd64`),
  `checksums.txt`, version injectée au build. Un job CI installe l'artefact
  Linux sur un runner sans toolchain Go.

### Corrigé

- Comparaison du modèle `envcheck init` et de `examples/envcheck.yml` sous
  Windows (CRLF au checkout).

[Unreleased]: https://github.com/SkyZonDev/envcheck/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/SkyZonDev/envcheck/releases/tag/v0.1.0

