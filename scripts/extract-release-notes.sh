#!/usr/bin/env bash
# Extrait la section CHANGELOG.md correspondant à un tag (v0.1.0 ou 0.1.0).
# Échoue si la section est absente : une release sans notes utilisateur
# n'est pas publiable.
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <tag> [fichier-sortie]" >&2
  exit 2
fi

tag="$1"
version="${tag#v}"
out="${2:-}"
root="$(cd "$(dirname "$0")/.." && pwd)"
changelog="$root/CHANGELOG.md"

if [[ ! -f "$changelog" ]]; then
  echo "CHANGELOG.md introuvable" >&2
  exit 1
fi

notes="$(awk -v ver="$version" '
  $0 ~ "^## \\[" ver "\\]" { found = 1; next }
  found && /^## / { exit }
  found && /^\[[^]]+\]: / { exit }
  found { print }
' "$changelog")"

trimmed="$(printf '%s\n' "$notes" | sed -e '/./,$!d' | sed -e :a -e '/^\n*$/{$d;N;ba' -e '}')"

if [[ -z "$trimmed" ]]; then
  echo "aucune section CHANGELOG.md pour ${version} (ajoutez ## [${version}] avant de taguer)" >&2
  exit 1
fi

if [[ -n "$out" ]]; then
  printf '%s\n' "$trimmed" >"$out"
else
  printf '%s\n' "$trimmed"
fi
