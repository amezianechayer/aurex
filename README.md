# Aurex Ledger

Aurex est un moteur de transactions financières programmable conçu pour rendre le développement d'applications financières plus sûr, plus simple et plus fiable.

Le développement de logiciels financiers est à la fois critique et notoirement difficile. Les mêmes erreurs reviennent systématiquement, créant des risques importants pour les systèmes en production.

Aurex répond à ce problème avec un ledger qui offre des transactions multi-postings atomiques, programmables en FaRl — un langage dédié aux mouvements de fonds. Il est particulièrement adapté aux applications nécessitant une logique financière complexe, comme :

* Les places de marché avec des flux de paiement avancés et du partage de revenus
* Les systèmes de monnaies internes, par exemple des crédits virtuels
* Les monnaies en jeu, inventaires et systèmes d'échange
* Les passerelles de paiement utilisant des actifs non standards
* Les monnaies locales et la finance complémentaire

## Démarrage rapide

```bash
# Démarrer le serveur
aurex server start

# Émettre des DZD depuis @world et créditer un commerçant
echo "
transfer [DZD.2 10000] from @world to @banque:reserve
" > script.farl
aurex exec quickstart script.farl

# Exemple complet — vente en ligne avec partage de revenus
echo "
transfer [DZD.2 5000] from @world to @acheteur:ameziane

transfer [DZD.2 5000] from @acheteur:ameziane to @commande:1042:paiement

transfer [DZD.2 5000] from @commande:1042:paiement
send 90/100 to @vendeur:yanis
send 10/100 to @plateforme:commission
" > vente.farl
aurex exec quickstart vente.farl

# Résultat
# @vendeur:yanis      → 4500 DZD.2 (90%)
# @plateforme:commission →  500 DZD.2 (10%)

# Consulter le solde d'un compte
curl -X GET http://localhost:3068/quickstart/accounts/@vendeur:yanis

# Lister les transactions
curl -X GET http://localhost:3068/quickstart/transactions
```

## API REST

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/:ledger/stats` | Statistiques du ledger |
| GET | `/:ledger/transactions` | Lister les transactions |
| POST | `/:ledger/transactions` | Créer une transaction manuelle |
| POST | `/:ledger/script` | Exécuter un script FaRl |
| GET | `/:ledger/accounts` | Lister les comptes |
| GET | `/:ledger/accounts/:address` | Solde et infos d'un compte |

## FaRl — Le langage de script

FaRl (Financial aRrangement Language) est le langage de script intégré à Aurex. Il permet d'exprimer des mouvements de fonds complexes de manière lisible et sûre.

```farl
# Transfer simple
transfer [DZD.2 1000] from @world to @banque:reserve

# Transfer avec cascade de sources
transfer [DZD.2 500] from @client:alice then @world to @vendeur:bob

# Transfer avec distribution proportionnelle
transfer [DZD.2 1000] from @client:alice
send 85/100 to @vendeur:bob
send 15/100 to @plateforme:commission

# Transfer avec variable
var $montant: monetary
transfer $montant from @world to @beneficiaire:001
```

## Configuration

```yaml
# ~/.aurex/aurex.yaml
server:
  http:
    bind_address: 127.0.0.1:3068

storage:
  driver: sqlite
  dir: ~/.aurex/storage
  sqlite:
    db_name: ledger
```

## Installation

```bash
# Compiler depuis les sources
git clone https://github.com/amezianechayer/aurex
cd aurex
go build -o aurex .
sudo mv aurex /usr/local/bin/

# Initialiser la configuration
aurex config init

# Démarrer le serveur
aurex server start
```