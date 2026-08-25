# Sécurité

`envcheck` ne collecte aucune télémétrie et n'envoie rien à un serveur. Il
lit un YAML local, interroge le poste, et imprime un rapport.

## Ce qui ne doit jamais fuiter

Les valeurs des variables d'environnement n'apparaissent ni dans le rapport
JSON, ni dans la sortie texte, ni dans les messages d'erreur. Un contrôle
`env` dit seulement si la variable est absente, vide, ou définie.

Si vous ouvrez une issue ou une PR, retirez secrets, tokens et extraits
`.env` des logs et des fichiers YAML collés.

## Signaler une vulnérabilité

Écrivez à l'équipe de maintenance via GitHub Security Advisories sur le
dépôt, plutôt que d'ouvrir une issue publique. Décrivez l'impact sans y
joindre de secrets réels.
