# Aurex ledger 

Aurex est un moteur de transactions financières programmable conçu pour rendre le développement d'applications financières plus sûr, plus simple et plus fiable.

Le développement de logiciels financiers est à la fois critique et notoirement difficile. Les mêmes erreurs reviennent systématiquement, créant des risques importants pour les systèmes en production.

Aurex répond à ce problème avec un ledger qui offre des transactions multi-postings atomiques, programmables en **FaRl** — un langage dédié aux mouvements de fonds. Il est particulièrement adapté aux applications nécessitant une logique financière complexe, comme :

- Les places de marché avec des flux de paiement avancés et du partage de revenus
- Les systèmes de monnaies internes, par exemple des crédits virtuels
- Les monnaies en jeu, inventaires et systèmes d'échange
- Les passerelles de paiement utilisant des actifs non standards
- Les monnaies locales et la finance complémentaire

## Démarrage rapide

```bash
# Émettre des DZD depuis le compte world et créditer un utilisateur
curl -X POST \
  -H 'Content-Type: application/json' \
  -d '{
    "postings": [
      {
        "source": "@world",
        "destination": "@central-bank",
        "asset": "DZD.2",
        "amount": 10000
      },
      {
        "source": "@central-bank",
        "destination": "@users:001",
        "asset": "DZD.2",
        "amount": 10000
      }
    ]
  }' http://localhost:3068/transactions

# Consulter le solde de users:001
curl -X GET http://localhost:3068/accounts/@users:001

# Lister les transactions
curl -X GET http://localhost:3068/transactions
```