# Corren

Corren est un moteur de transactions financières programmable, conçu pour les développeurs qui construisent des produits financiers en Afrique et au Moyen-Orient.

Le développement de logiciels financiers est à la fois critique et notoirement difficile. Les mêmes erreurs reviennent systématiquement — états partiels, balances incohérentes, logique métier dispersée dans le code applicatif — créant des risques importants pour les systèmes en production.

Corren répond à ce problème avec un ledger qui offre des transactions multi-postings atomiques, programmables en **FaRl** — un langage dédié aux mouvements de fonds. Il est particulièrement adapté aux applications nécessitant une logique financière complexe :

- Places de marché avec partage de revenus et flux de paiement avancés
- Fintechs et néobanques MENA/Afrique
- Systèmes de monnaies internes (crédits virtuels, loyalty points)
- Passerelles de paiement utilisant des actifs non standards
- Monnaies locales et finance complémentaire
- Systèmes de jeu, inventaires et échange d'actifs

---

## Démarrage rapide

```bash
# Compiler et installer
git clone https://github.com/amezianechayer/corren
cd corren && go build -o corren .
sudo mv corren /usr/local/bin/

# Initialiser et démarrer
corren config init
corren server start
# → Listening on 127.0.0.1:3068

# Émettre des DZD depuis @world et créditer un compte
echo "
transfer [DZD.2 10000] (
  from @world
  to   @banque:reserve
)
" > script.farl
corren exec quickstart script.farl
# 200 {"ok":true}

# Exemple complet — vente en ligne avec partage de revenus
cat > vente.farl << 'EOF'
transfer [DZD.2 5000] (
  from @world
  to   @acheteur:ameziane
)

transfer [DZD.2 5000] (
  from @acheteur:ameziane
  to   @commande:1042:paiement
)

transfer [DZD.2 5000] (
  from @commande:1042:paiement
  to   {
    90/100 to @vendeur:yanis
    10/100 to @plateforme:commission
  }
)
EOF

corren exec quickstart vente.farl
# @vendeur:yanis         → 4500 DZD.2 (90%)
# @plateforme:commission →  500 DZD.2 (10%)

# Consulter le solde d'un compte
curl http://localhost:3068/quickstart/accounts/@vendeur:yanis

# Lister les transactions
curl http://localhost:3068/quickstart/transactions
```

---

## FaRl — Le langage de script

FaRl (Financial accounting Rule Language) est le langage de script intégré à Corren. Il permet d'exprimer des mouvements de fonds complexes de manière lisible, sûre, et validée à la compilation.

### Transfer simple

```farl
transfer [DZD.2 10000] (
  from @world
  to   @banque:reserve
)
```

### Cascade de sources

```farl
# Tire du wallet d'abord, puis du crédit, puis de @world en dernier recours
transfer [DZD.2 3000] (
  from {
    @wallet:ameziane
    @credit:ameziane
    @world
  }
  to @commande:0042
)
```

### Distribution proportionnelle

```farl
transfer [DZD.2 5000] (
  from @client:ameziane
  to   {
    85/100 to @vendeur:yanis
    10/100 to @plateforme:commission
    5/100  to @taxes
  }
)
```

### Avec remaining

```farl
transfer [DZD.2 5000] (
  from @client:ameziane
  to   {
    10%       to @plateforme:commission
    5%        to @taxes
    remaining to @vendeur:yanis
  }
)
```

### Avec variables

```farl
var $montant:    monetary
var $vendeur:    account
var $commission: portion

transfer $montant (
  from @world
  to   {
    $commission to @plateforme
    remaining   to $vendeur
  }
)
```

### Avec métadonnées dynamiques

```farl
var $sale:   account
var $seller: account = meta($sale,   "seller")
var $rate:   portion = meta($seller, "commission_rate")

transfer [DZD.2 10000] (
  from $sale
  to   {
    $rate     to @plateforme:commission
    remaining to $seller
  }
)
```

### Set metadata

```farl
# Attacher des métadonnées à une transaction
set transaction metadata {
  "reference" = "PAY-2026-001"
  "channel"   = "mobile"
}

# Modifier les métadonnées d'un compte
set account metadata of @vendeur:yanis key "commission_rate" = "12%"
```

### Transfer all balance

```farl
# Vider entièrement un compte
transfer [DZD.2 *] (
  from @escrow:commande:0042
  to   @vendeur:yanis
)
```

---

## API REST

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| `GET` | `/:ledger/stats` | Statistiques du ledger |
| `GET` | `/:ledger/transactions` | Lister les transactions |
| `POST` | `/:ledger/transactions` | Créer une transaction manuelle |
| `POST` | `/:ledger/script` | Exécuter un script FaRl |
| `GET` | `/:ledger/accounts` | Lister les comptes |
| `GET` | `/:ledger/accounts/:address` | Solde et infos d'un compte |

### Exécuter un script via l'API

```bash
curl -X POST http://localhost:3068/quickstart/script \
  -H "Content-Type: application/json" \
  -d '{
    "plain": "transfer [DZD.2 1000] (\n  from @world\n  to @alice\n)",
    "vars": {}
  }'
```

### Injecter des variables

```bash
curl -X POST http://localhost:3068/quickstart/script \
  -H "Content-Type: application/json" \
  -d '{
    "plain": "var $montant: monetary\nvar $dest: account\ntransfer $montant (\n  from @world\n  to $dest\n)",
    "vars": {
      "montant": { "asset": "DZD.2", "amount": 5000 },
      "dest": "@vendeur:yanis"
    }
  }'
```

---

## Devises

Exemples de devises utilisées avec Corren :

| Devise | Asset FaRl | Précision |
|--------|-----------|-----------|
| Dinar algérien | `DZD.2` | 2 décimales |
| Dirham marocain | `MAD.2` | 2 décimales |
| Dinar tunisien | `TND.3` | 3 décimales |
| Livre égyptienne | `EGP.2` | 2 décimales |
| Franc CFA | `XOF.0` | 0 décimale |
| Naira nigérian | `NGN.2` | 2 décimales |
| Shilling kényan | `KES.2` | 2 décimales |
| Riyal saoudien | `SAR.2` | 2 décimales |
| Dirham UAE | `AED.2` | 2 décimales |
| Euro | `EUR.2` | 2 décimales |
| Dollar US | `USD.2` | 2 décimales |

---

## CLI

```bash
corren config init                    # Initialiser la configuration
corren server start                   # Démarrer le serveur HTTP
corren exec <ledger> <script.farl>    # Exécuter un script FaRl
corren storage init                   # Initialiser le schéma de base de données
```

**Prérequis :** Go 1.16+

---

## Roadmap

### Phase 1 — Fondations `en cours`

- [x] Ledger core avec transactions atomiques multi-postings
- [x] Langage FaRl avec compilateur et VM
- [x] API REST — script, transactions, comptes
- [x] Storage SQLite et PostgreSQL
- [x] CLI — run, server, config
- [x] Docker Compose — déploiement en une commande
- [x] Tests d'intégration complets
- [x] Documentation API OpenAPI/Swagger

### Phase 2 — Connecteurs PSP `priorité`

Connecteurs de paiement pour le marché MENA et africain :

- [ ] **Chargily** — passerelle de paiement algérienne
- [ ] **SlickPay** — paiement en ligne Algérie
- [ ] **BaridiMob Pay** — paiement mobile La Poste Algérie
- [ ] **SATIM** — réseau interbancaire algérien, paiement par carte CIB
- [ ] **Wave** — paiement mobile Afrique de l'Ouest (Sénégal, Côte d'Ivoire)
- [ ] **Flouci** — paiement mobile Tunisie
- [ ] **Paymob** — MENA (Égypte, Maroc, UAE)
- [ ] **PayTabs** — Arabie Saoudite, UAE, MENA
- [ ] **Flutterwave** — Afrique anglophone
- [ ] **Paystack** — Nigeria, Ghana, Kenya
- [ ] **Stripe** — global
- [ ] **Adyen** — global

Chaque connecteur expose une interface unifiée — une seule API pour tous les PSP, les différences sont abstraites dans le connecteur.

### Phase 3 — Developer Experience

- [ ] **SDK Go** — client officiel
- [ ] **SDK Python** — client officiel
- [ ] **SDK JavaScript / TypeScript** — client officiel
- [ ] **Playground FaRl** — éditeur web interactif pour tester des scripts sans installation
- [ ] **Dashboard web** — interface de visualisation des balances, transactions et comptes

### Phase 4 — Intelligence financière

- [ ] **FaRl Guard** — monitoring de transactions en temps réel, règles déclaratives intégrées dans FaRl pour détecter la fraude et enforcer les limites
- [ ] **FaRl Lens** — observabilité financière, visualisation des flux de fonds, analytics par compte et par ledger
- [ ] **Reconciliation** — rapprochement automatique du ledger avec les relevés bancaires et rapports PSP externes

---

*Corren est en développement actif. Les contributions sont les bienvenues.*