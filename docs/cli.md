# Interface en ligne de commande

Le binaire n’a pas d’effet de bord, sauf `envcheck init` qui écrit un fichier.

La commande par défaut est `check` : taper `envcheck` revient à taper `envcheck check`.

## `envcheck` / `envcheck check`

Charge le YAML, valide le schéma, exécute toutes les règles, imprime le rapport.

```bash
envcheck
envcheck check
envcheck --config path/to/envcheck.yml
envcheck --format json
envcheck --quiet --no-color
```

| Option | Défaut | Description |
| --- | --- | --- |
| `--config CHEMIN` | (découverte) | Fichier YAML explicite. Pas de recherche `envcheck.yml` / `.envcheck.yml`. |
| `--format text\|json` | `text` | Format du rapport. Toute autre valeur (y compris `yaml`) est une erreur d’usage, code `2`. |
| `--quiet` | `false` | En texte : n’afficher que les contrôles qui ne sont pas `pass`. Sans effet sur le JSON, qui reste complet. |
| `--no-color` | `false` | Désactive les couleurs ANSI du rendu texte. |

Couleurs : désactivées si `--no-color`, si la variable `NO_COLOR` est non vide, ou si `stdout` n’est pas un TTY (pipe, redirection, CI).

### Sortie texte

```text
✓ git            Git 2.45.1 satisfait >=2.40.0
✗ docker-daemon  Le client Docker est disponible, mais le démon est inaccessible.
  ↳ Démarrez Docker Desktop ou le service Docker.

1/2 contrôles requis réussis — environnement non prêt.
```

| Statut | Marqueur |
| --- | --- |
| `pass` | `✓` |
| `fail` | `✗` |
| `warn` | `!` |
| `error` | `‼` |

Le `hint` n’est imprimé que lorsque le statut n’est pas `pass`.

Le bilan compte uniquement les contrôles **requis**. Un avertissement optionnel n’entre pas dans le `N/M`.

### Sortie JSON

`--format json` écrit **exclusivement** un objet JSON sur `stdout`. Les diagnostics (fichier absent, YAML invalide, option incorrecte) vont sur `stderr`.

```json
{
  "schemaVersion": 1,
  "toolVersion": "v0.1.0",
  "configPath": "envcheck.yml",
  "startedAt": "2026-08-26T01:07:00Z",
  "durationMs": 42,
  "passed": false,
  "summary": {
    "pass": 1,
    "fail": 1,
    "warn": 0,
    "error": 0
  },
  "checks": [
    {
      "id": "git",
      "type": "command",
      "required": true,
      "status": "pass",
      "summary": "Git 2.45.1 satisfait >=2.40.0",
      "hint": "Installez Git depuis https://git-scm.com/downloads.",
      "expected": ">=2.40.0",
      "actual": "2.45.1",
      "durationMs": 12
    }
  ]
}
```

| Champ du rapport | Signification |
| --- | --- |
| `schemaVersion` | Contrat JSON (`1`), distinct de la version du binaire et du `version:` YAML. |
| `toolVersion` | Version compilée (`dev` en local, tag Git en release). |
| `configPath` | Chemin du YAML utilisé. |
| `startedAt` | Horodatage RFC 3339 du début du run. |
| `durationMs` | Durée totale du run. |
| `passed` | `true` s’il n’y a aucun `fail` ni `error`. |
| `summary` | Compteurs par statut, pour éviter de recompter `checks`. |
| `checks[]` | Un objet par règle, dans l’ordre du YAML. |

Champs d’un contrôle :

| Champ | Présence | Contenu |
| --- | --- | --- |
| `id`, `type`, `required`, `status`, `summary`, `durationMs` | toujours | Identité et verdict. |
| `hint` | si déclaré dans le YAML | Conseil humain, même en cas de succès (présent dans le JSON ; masqué en texte si `pass`). |
| `expected` | selon le type | Contrainte de version (`command`) ou `kind` (`path` : `file` / `directory` / `any`). |
| `actual` | selon le type | Version extraite (`command`) ou `file` / `directory` (`path`). **Toujours absent pour `env`.** |
| `detail` | rarement | Réservé ; le MVP le laisse en général vide. |

`schemaVersion` du JSON n’est pas le `version:` du YAML. Le YAML dit « ce fichier est du schéma 1 » ; le JSON dit « ce rapport est du contrat 1 ». Les deux valent `1` aujourd’hui, mais ils peuvent diverger plus tard.

## `envcheck init`

Écrit le modèle YAML dans `./envcheck.yml`.

```bash
envcheck init
envcheck init --dry-run
envcheck init --force
```

| Option | Effet |
| --- | --- |
| `--dry-run` | Affiche le modèle sur `stdout`, n’écrit rien. Format accepté : YAML uniquement. `--format json` avec `--dry-run` → code `2`. |
| `--force` | Écrase `envcheck.yml` (ou refuse de le faire s’il existe, sans ce flag). |

Sans `--force`, si `envcheck.yml` **ou** `.envcheck.yml` existe déjà dans le CWD, la commande refuse (code `2`) et rappelle le chemin trouvé.

Le fichier écrit a toujours le nom `envcheck.yml`, même si c’est `.envcheck.yml` qui existait.

Le modèle est statique : Git, Docker, une variable `DATABASE_URL`, un dossier `migrations`. Adaptez-le ; ce n’est pas une détection automatique du projet.

## `envcheck --version` / `envcheck --help`

`--version` affiche la version injectée au build (`dev` sans ldflags). `--help` liste les commandes, les flags persistants et les exemples.

## Codes de sortie

| Code | Constante interne | Quand |
| --- | --- | --- |
| `0` | `ExitOK` | YAML valide, tous les contrôles **obligatoires** en `pass`. Les `warn` sont autorisés. |
| `1` | `ExitFail` | YAML valide, au moins un `fail` ou `error` sur une règle obligatoire. |
| `2` | `ExitConfigError` | Fichier introuvable, YAML illisible, schéma invalide, `--format` inconnu, `init` sans `--force` sur un fichier existant, `init --dry-run` avec `--format json`. |
| `3` | `ExitInternalError` | Erreur inattendue (CWD illisible, écriture `init` impossible, etc.). |

En shell :

```bash
envcheck
case $? in
  0) echo "prêt" ;;
  1) echo "prérequis manquants" ;;
  2) echo "configuration à corriger" ;;
  *) echo "erreur interne" ;;
esac
```

Le code `1` est un **échec métier attendu** : la CI doit le traiter comme « environnement non conforme », pas comme un crash de l’outil. Le code `2` signale que le contrat YAML lui-même est cassé.

## Intégration CI

Exemple GitHub Actions : le job échoue si un contrôle obligatoire échoue.

```yaml
- name: Vérifier l'environnement
  run: envcheck --format json --no-color
```

`--no-color` est redondant en CI (stdout n’est en général pas un TTY) mais rend le log prévisible.

Pour archiver le rapport :

```yaml
- name: envcheck
  run: envcheck --format json > envcheck-report.json
- uses: actions/upload-artifact@v4
  if: always()
  with:
    name: envcheck-report
    path: envcheck-report.json
```

Le JSON est le contrat stable à parser. Ne parsez pas la sortie texte : les marqueurs, les couleurs et les libellés français peuvent évoluer ; les champs JSON (`status`, `passed`, `id`) sont le point d’extension prévu.

## Découverte du fichier en CI

Comme en local : `envcheck` ne cherche que dans le répertoire de travail du step. Placez `envcheck.yml` à la racine du dépôt, ou :

```yaml
- run: envcheck --config ./deploy/envcheck.yml
  working-directory: .
```

Si vous changez `working-directory`, les contrôles `path` relatifs suivent ce nouveau CWD.
