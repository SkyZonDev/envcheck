# Documentation envcheck

`envcheck` lit un YAML local, vérifie les prérequis du poste, et produit un rapport. Il n’installe rien et n’envoie rien sur le réseau.

| Document | Pour qui | Contenu |
| --- | --- | --- |
| [../README.md](../README.md) | Tout le monde | Présentation, installation, démarrage |
| [fonctionnement.md](fonctionnement.md) | Utilisateurs et contributeurs | Pipeline, verdicts, contrôles, architecture |
| [configuration.md](configuration.md) | Mainteneurs d’un dépôt | Schéma YAML, champs, exemples |
| [cli.md](cli.md) | Utilisateurs et CI | Commandes, JSON, codes de sortie |
| [../SECURITY.md](../SECURITY.md) | Tout le monde | Secrets, signalement |
| [../CONTRIBUTING.md](../CONTRIBUTING.md) | Contributeurs | Build, tests, release |

Parcours conseillé : README → ce dossier → [configuration.md](configuration.md) pour écrire votre premier `envcheck.yml` → [cli.md](cli.md) pour brancher la CI.
