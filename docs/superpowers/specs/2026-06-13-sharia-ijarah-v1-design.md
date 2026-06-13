# Design — Ijarah opérationnelle v1 + moteur sharia multi-contrats

**Date** : 2026-06-13
**Statut** : validé par le user (feu vert, 2 conditions intégrées)
**Base** : branche `feature/sharia-ijarah-v1`, empilée sur `feature/pilot-hardening`
(hérite des fixes : assets pointés, réparation tables legacy, auth, postgres).
**Objectif** : ajouter le contrat Ijarah opérationnelle (location pure) comme
2ᵉ type de contrat, en généralisant d'abord le moteur sharia derrière une
interface `ContractKind`. Prouve que le moteur n'est pas mono-Murabaha.

Jamais de commit sur `main`.

---

## Partie 1 — Refacto moteur multi-contrats (behavior-preserving)

### Avant

`Contract.Params` est un `MurabahaParams` concret ; `Engine.Transition()` est
un `switch` codé en dur avec la logique Murabaha (check qabd sur `sell`, rejet
de tout type ≠ murabaha). Mono-contrat.

### Après

```go
// Params est un marqueur ; chaque kind connaît son type concret.
type Params interface{}

type Event struct {
    Name  string          // "acquire", "sell", "pay_installment", "pay_rent"…
    Input TransitionInput // seq, rebate, penalty… (déjà existant)
}

type ContractKind interface {
    Type() string
    DecodeParams(json.RawMessage) (Params, error)
    Validate(Params) error
    ShariaGate(led LedgerPort, c Contract, p Params, ev Event) error // pré-FSM, nil par défaut
    AllowedTransitions(from string) []string                         // FSM
    Preconditions(led LedgerPort, c Contract, p Params, ev Event) error
    BuildPlan(led LedgerPort, c Contract, p Params, ev Event) (TransitionPlan, error)
}
```

**Ordre des contrôles (préservation du comportement Murabaha).** Aujourd'hui,
vendre sans possession renvoie `ERR_SHARIA_VIOLATION` (SS-8) **même depuis
PROMISE** — la priorité Sharia passe avant l'erreur de séquencement FSM. Pour
préserver ça exactement, le moteur appelle `kind.ShariaGate()` **avant** le
check FSM ; il renvoie une erreur typée prioritaire ou nil. Murabaha
l'implémente pour `sell` (qabd, SS-8) ; Ijarah pour `lease` (propriété, SS-9) ;
défaut = nil pour toutes les autres transitions. Ensuite seulement : FSM →
`Preconditions` → `BuildPlan`.

```go
// TransitionPlan : ce que le moteur commitera de façon générique.
type TransitionPlan struct {
    Postings  []core.Posting
    Reference string
    NewState  string
    StandardRef string
    Event     string // event d'audit (transition|penalty|settled|completed…)
    Payload   string
    // hooks post-commit déclaratifs, exécutés par le moteur :
    MarkInstallment *InstallmentMark // nil si aucun
    ExtraAudit      []AuditEvent     // events additionnels (ex. asset returned)
}
```

- `Contract.Params` devient `json.RawMessage` (le store sérialise déjà en JSON
  → **zéro migration**).
- `registry = map[string]ContractKind`, une impl par contrat :
  `murabahaKind{}`, `ijarahKind{}`.
- `Engine.Transition()` devient générique : lock → load → `kind := registry[c.Type]`
  → décoder params → **`ShariaGate` (pré-FSM)** → FSM (`AllowedTransitions`) →
  `Preconditions` → `BuildPlan` → commit atomique → hooks post-commit déclarés
  dans le plan → audit. Le moteur ne connaît plus aucun contrat en particulier.
- `Engine.Create()` route aussi par type : `kind.DecodeParams` + `kind.Validate`
  + construction d'échéancier déléguée au kind.

### Contrainte stricte (filet de sécurité)

Le refacto Murabaha est **behavior-preserving** : on EXTRAIT la logique du
switch vers `murabahaKind` SANS changer de comportement. Les scénarios A–G et
tous les tests sharia/engine/store/api existants doivent rester **verts et
inchangés** pendant tout le refacto. Aucun test modifié ou supprimé. Si un test
A–G casse, c'est que le refacto n'est pas behavior-preserving → corriger le
refacto, pas le test. Ijarah et ses tests n'arrivent qu'une fois tout vert.

### Après le refacto

Ajouter un contrat = écrire une `ContractKind` + l'enregistrer. Jamais toucher
au moteur. (Mudarabah, Ijarah Muntahia Bittamleek… en v2/v3.)

---

## Partie 2 — Ijarah opérationnelle (location pure)

### 2.1 Périmètre v1 (et ce qui est explicitement reporté)

**Dans la v1 :**
- Location pure : la banque achète l'actif, le **maintient à son bilan**
  (compte d'actif loué dédié), perçoit des loyers, **récupère l'actif à la fin**.
- Propriété **conservée par le bailleur** tout du long — c'est ce qui rend le
  loyer licite (usufruit ≠ propriété).
- Amortissement **linéaire** de l'actif sur la durée du bail (valeur résiduelle 0).
- Loyer reconnu en revenu **à chaque paiement de période**.

**Reporté en v2 (NE PAS coder maintenant) :**
- Ijarah Muntahia Bittamleek (transfert de propriété en fin de bail), `transfer_price`.
- `early_terminate` (retour anticipé). Raison : actif non fini d'amortir →
  valeur résiduelle à sortir vers un inventaire à valeur restante = comptabilité
  partielle, précisément la complexité qu'on écarte. v1 = cycle nominal, point.
- Valeur résiduelle, amortissement dégressif.
- Comptabilité d'engagement (accruals). En v1, le loyer devient revenu au
  moment du posting de paiement de chaque période — naturellement « sur la
  durée », sans moteur d'accruals.

**Posture conformité :** l'audit labellise « FAS 32 — amortissement simplifié
(v1) », **jamais** « conforme FAS 32 ». La base d'amortissement est un paramètre
marqué *à valider par un advisor Sharia/comptable* avant production. Référence
AAOIFI au niveau document : **SS-9** (Ijarah) pour les invariants Sharia.

### 2.2 Machine à états

```
PROMISE ──acquire──▶ ACQUIRED ──lease──▶ LEASED ──pay_rent (×N)──▶ COMPLETED
   │                    │                   │
   └──cancel──▶ CANCELLED◀──cancel          └──(à la dernière échéance :
                                               actif rendu au bailleur,
                                               état → COMPLETED)
```

États : `PROMISE`, `ACQUIRED`, `LEASED`, `COMPLETED`, `CANCELLED`.
`cancel` interdit depuis `LEASED`. Transitions hors FSM → `ERR_INVALID_TRANSITION`
journalisé (= I-5 hérité).

### 2.3 Paramètres (IjarahParams)

```go
type IjarahParams struct {
    AssetCode    string   `json:"asset_code"`    // ^[A-Z][A-Z0-9]{0,7}$
    Cost         Monetary `json:"cost"`          // prix d'achat fournisseur
    Rent         Monetary `json:"rent"`          // loyer FIXE par période, même asset que Cost
    Client       string   `json:"client"`        // "@client:xxx"
    Supplier     string   `json:"supplier"`      // "@supplier:xxx"
    BankTreasury string   `json:"bank_treasury"` // défaut "@bank:treasury"
    Periods      int      `json:"periods"`       // >= 1 (nb de loyers)
    FirstDue     string   `json:"first_due"`     // RFC3339
    PeriodDays   int      `json:"period_days"`   // défaut 30
}
```

Validation (sinon `ERR_INVALID_PARAMS`, détail par champ) : `Cost.Amount > 0` ;
`Rent.Amount > 0` ; `Cost.Asset == Rent.Asset` et asset monétaire valide
(`^[A-Z][A-Z0-9]{0,7}(\.[0-9]{1,2})?$`, convention pointée acceptée) ;
`AssetCode` valide et ≠ asset monétaire ; `Periods >= 1` ; `FirstDue` RFC3339 ;
comptes client/supplier matchent le lexer ACCOUNT. Le loyer étant fixe par
période, `total_rent = Rent.Amount × Periods` doit être `> Cost.Amount` pour que
la banque dégage une marge (sinon `ERR_INVALID_PARAMS` : « rent over the term
must exceed cost »).

### 2.4 Comptes ledger (par contrat `<id>`)

| Compte | Rôle | Négatif |
|---|---|---|
| `@contracts:<id>:asset` | actif loué : unité physique `[ASSET 1]` + valeur comptable `[money cost]` qui s'amortit. Propriété du bailleur. | non |
| `@client:<x>:in_use` | possession de l'unité pendant le bail (la banque garde la valeur) | non |
| `@bank:treasury` | trésorerie réelle | non |
| `@bank:income:ijarah` | loyers reconnus en revenu | non |
| `@bank:expense:depreciation` | amortissement cumulé | non |
| `@bank:inventory:returned` | où l'unité revient en fin de bail | non |
| `@bank:inventory:unsold` | où l'unité va si annulation depuis ACQUIRED | non |
| `@charity:pool` | destination exclusive des pénalités | non |
| `@world`, `@supplier:<x>` | sources/contreparties externes | (`@world` exempté) |

**L'Ijarah n'utilise PAS `:counterpart`.** Le counterpart était le réservoir
Murabaha pour la créance/profit différé créés d'avance à la vente. L'Ijarah n'a
pas de créance d'avance (IJ-3) ; la valeur comptable et le revenu sont
capitalisés/reconnus depuis `@world`, exactement comme l'unité physique entre
au ledger. Démonstration que chaque kind n'utilise que les comptes nécessaires.

### 2.5 Plans de postings (figés)

**acquire** — PROMISE → ACQUIRED. Précondition : `treasury.Balances[Cost.Asset] >= Cost.Amount`.
Reference `<id>:acquire`.
```
P1  @bank:treasury → @supplier:<x>            [money  Cost]      (cash sort, achat)
P2  @world         → @contracts:<id>:asset    [ASSET  1]         (unité physique, qabd)
P3  @world         → @contracts:<id>:asset    [money  Cost]      (valeur comptable capitalisée)
```

**lease** — ACQUIRED → LEASED. Précondition **IJ-1** : `@contracts:<id>:asset`
détient `[AssetCode 1]` (la banque possède l'actif). Échec → `ERR_SHARIA_VIOLATION`
/ SS-9 / reason "lease of non-owned asset", journalisé. Reference `<id>:lease`.
```
P1  @contracts:<id>:asset → @client:<x>:in_use  [ASSET  1]       (possession/usufruit, PAS la propriété)
```
Aucune créance de loyer créée (IJ-3). La valeur reste dans `:asset`.

**pay_rent seq n** — LEASED → LEASED (ou → COMPLETED à la dernière).
Précondition : échéance n existante et non payée ; `client.Balances[asset] >= Rent`.
Reference `<id>:rent:<n>`.
```
P1  @client:<x>           → @bank:treasury              [money  Rent]            (cash réel)
P2  @world                → @bank:income:ijarah         [money  Rent]            (revenu reconnu, FAS 32 simplifié)
P3  @contracts:<id>:asset → @bank:expense:depreciation  [money  depreciation_n]  (amortissement de la période)
```
À la **dernière** échéance, dans la MÊME transaction atomique, posting
supplémentaire de retour de l'actif au bailleur :
```
P4  @client:<x>:in_use → @bank:inventory:returned       [ASSET  1]
```
puis état → `COMPLETED`. **Limitation v1 documentée** : « actif rendu » n'a pas
son propre event d'audit séparé (il partage la transition du dernier loyer) →
non requêtable seul. À éclater en v2 via `ExtraAudit`.

**late_penalty** — LEASED → LEASED. Règle dure **IJ-5** (= I-3) : destination
doit commencer par `@charity:` sinon `ERR_SHARIA_VIOLATION` / SS-3 journalisé.
Reference `<id>:late_penalty[:n]`.
```
P1  @client:<x> → @charity:pool   [money  amount]
```

**cancel** — PROMISE → CANCELLED (aucun posting) ; ACQUIRED → CANCELLED.
Depuis ACQUIRED la banque garde l'actif (son risque, comme Murabaha) :
Reference `<id>:cancel`.
```
P1  @contracts:<id>:asset → @bank:inventory:unsold  [ASSET  1]
P2  @contracts:<id>:asset → @world                  [money  Cost]   (sort la valeur comptable)
```

### 2.6 Arithmétique de l'échéancier (exactitude entière)

- Loyer : **fixe** par période = `Rent.Amount`. Chaque ligne porte ce montant
  (pas de SplitEven : un bail a un loyer mensuel constant).
- Amortissement : `depreciation = SplitEven(Cost.Amount, Periods)` — linéaire,
  reste sur la dernière période (réutilise la fonction Murabaha).
- Une ligne d'échéancier Ijarah : `{seq, due_date, rent_amount, depreciation_part, status}`.
- Dates : `due[i] = FirstDue + (i-1)·PeriodDays` (UTC).

**Exemple de référence (à coder en test table-driven) :**
cost = 10 000 000, rent = 500 000/période, periods = 24.
- Loyer chaque période : **500 000** (24 × 500 000 = 12 000 000 = total_rent).
- Amortissement périodes 1–23 : **416 666** ; période 24 : **416 682**
  (Σ = 10 000 000).
- Contrôles fin de bail : treasury = −10 000 000 + 12 000 000 = **+2 000 000**
  (marge en cash) ; `:asset` money = 0 ; `:asset` ASSET = 0 (rendu) ;
  income:ijarah = **12 000 000** ; expense:depreciation = **10 000 000** ;
  counterpart = **0** (jamais touché) ; P&L = 12M − 10M = **2M** = marge.

### 2.7 Invariants (→ standard → test)

| Invariant | Règle | Standard | Test |
|---|---|---|---|
| IJ-1 | Propriété avant location : `lease` exige l'actif dans `:asset` | AAOIFI-SS-9 | scénario "lease sans acquire → 422" |
| IJ-2 | Bailleur reste propriétaire : l'unité ne passe jamais en propriété au locataire ; la valeur reste au bilan | AAOIFI-SS-9 | cycle nominal (asset money sur `:asset` jusqu'à amortissement) |
| IJ-3 | Loyer reconnu période par période, jamais de créance totale d'avance | FAS-32 simplifié | absence de posting de créance à `lease` ; income croît par période |
| IJ-4 | Amortissement linéaire, labellisé « FAS 32 simplifié (v1) », base à valider advisor | FAS-32 simplifié | échéancier de référence ; label dans le payload audit |
| IJ-5 | Pénalités → charité uniquement | AAOIFI-SS-3 | pénalité vers income → 422 SS-3 |
| hérités | FSM strict, exactitude int64, atomicité (1 transition = 1 tx) | — | scénarios FSM / idempotence / comptes protégés |

### 2.8 API

Aucune nouvelle route. Les endpoints contrats existants servent l'Ijarah :
`POST /:ledger/contracts` avec `"type": "ijarah"`, transitions
`acquire|lease|pay_rent|late_penalty|cancel`, `GET …/schedule`, `…/audit`.
Le contrôleur ne change pas (il délègue au moteur, qui route par type).

---

## Partie 3 — Tests & critères d'acceptation

**Refacto (Partie 1) :** `go test ./...` vert, A–G inchangés, tout au vert
AVANT d'écrire la moindre ligne d'Ijarah.

**Ijarah (Partie 2), en TDD :**
- `ijarah_schedule_test.go` : exemple de référence §2.6 + cas limites (periods=1,
  cost non divisible, total_rent ≤ cost rejeté).
- FSM : transitions autorisées/refusées, refus journalisés.
- Plans de postings : acquire / lease / pay_rent (3 postings + P4 au dernier) /
  late_penalty (charité only) / cancel.
- **Scénario nominal** (sqlite, ledger de test) : provisionner treasury ;
  create (24 périodes, valeurs §2.6) ; acquire (supplier payé, `:asset` = 1 unité
  + cost) ; lease sans acquire AILLEURS → 422 SS-9 ; lease (unité en `:in_use`) ;
  payer 24 loyers (vérifier après chacun : income += rent, depreciation += part,
  `:asset` money décroît) ; à la fin : treasury +2M, `:asset` money 0, unité
  rendue, income 12M, depreciation 10M, counterpart 0, état COMPLETED.
- **IJ-5** : late_penalty vers `@bank:income:ijarah` → 422 SS-3, aucun posting.
- **Idempotence** : rejouer `pay_rent seq=1` → 409 ERR_DUPLICATE.
- **Comptes protégés** : un script FaRl ne peut pas rendre `:asset` négatif.
- **Multi-contrats** : un ledger porte un Murabaha ET un Ijarah, les deux
  audits vérifient `chain_valid`.

## Partie 4 — README

Étendre `sharia/README.md` : section Ijarah (FSM, comptes, mapping IJ-x →
SS-9/FAS-32 → test), et noter que le moteur est désormais multi-contrats via
`ContractKind` (comment ajouter un 3ᵉ contrat).

## Hors périmètre (rappel)

IMB / transfert de propriété, early_terminate, valeur résiduelle, amortissement
dégressif, accruals, autres contrats. Voir Partie 2.1.

## Risques connus

- Le refacto `ContractKind` touche le cœur du moteur : risque de régression
  Murabaha. Mitigation : behavior-preserving strict + A–G en filet, refacto
  committé et vert AVANT Ijarah.
- `@world → income:ijarah` : convention de reconnaissance de revenu (pas un
  mouvement de cash). Documentée ; un comptable validera la présentation.
- Amortissement « FAS 32 simplifié » : explicitement non certifié, à valider.
