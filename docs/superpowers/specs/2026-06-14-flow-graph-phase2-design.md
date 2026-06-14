# Design — Lens Flow Graph (phase 2): interactive d3 graph + clustering + time-scrubber

**Date** : 2026-06-14
**Statut** : validé par le user
**Repo** : Horizon, branche `feature/lens-ui` (frontend pur — **zéro changement backend**).
**Dépend de** : l'endpoint gelé `GET /:ledger/lens/flows` (contrat
`{source, destination, asset, time_bucket, amount, count}`), déjà livré (Lens v1).
**Specs/plans** archivés dans le repo corren `docs/superpowers/` (cerveau projet).

Jamais de commit sur `main`.

## 1. But

Le second rendu du contrat `/flows` (le premier étant le tableau v1) : un graphe
de flux **interactif, force-directed**, avec **clustering automatique par type de
compte** et un **time-scrubber** qui rejoue les flux dans le temps. Thème clair
Lens (Inter, accent `#13e07e`). Construit avec **d3 v7** (nouvelle dépendance ;
on garde les primitives éprouvées — `forceSimulation`, `zoom`, `drag` — et on
écrit une intégration React propre + un style sur mesure ; on n'hérite PAS du
code de la branche `transaction-graph`).

## 2. Données — tout dérive de `/flows` (frontend pur)

`src/lib/buildGraph.js`, fonctions **pures** (inspectables, futures-testables) :

- `buildGraph(flows, asset)` → `{ nodes, links, buckets }` :
  - filtre les lignes `/flows` à l'`asset` choisi ;
  - collapse par `(source, destination)` en sommant `amount` + `count` → **links** ;
  - **nodes** = comptes distincts (union source+destination) ; `volume` du nœud =
    somme des montants incidents ; `kind` = préfixe (`bank|client|supplier|contracts|world|other`) ;
  - **buckets** = liste triée des `time_bucket` distincts (pour le scrubber).
- `filterToBucket(flows, asset, t)` → graphe **cumulatif** : ne garde que les
  lignes dont `time_bucket <= t`, puis `buildGraph`. (Vue cumulative, pas fenêtrée.)
- `kindOf(account)` → le type (mapping préfixe). Pur, testable.

Multi-asset : **un asset à la fois** via un sélecteur (on ne mélange jamais les
assets ; comparer des épaisseurs DZD vs EUR n'a pas de sens). Défaut = asset au
plus gros volume.

## 3. Rendu & interactions

### Layout (d3-force)
`forceLink` (distance ∝ inverse du montant, bornée), `forceManyBody` (charge
négative), `forceCenter`, `forceCollide` (rayon nœud, anti-chevauchement), plus
la **force de cluster** (§4). `d3.zoom` (pan + molette, scaleExtent borné),
`d3.drag` sur les nœuds (fixe la position pendant le drag, relâche après).

### Nœuds
Cercle, rayon ∝ `sqrt(volume)` (bornes min/max), **couleur par `kind`** (palette
sémantique : bank, client, supplier, contracts, world, other — teintes distinctes
cohérentes avec Lens). Label = adresse tronquée, affiché au-delà d'un seuil de
zoom ou au survol.

### Arêtes
Dirigées (marqueur flèche en `defs`), épaisseur ∝ `amount` (échelle bornée),
opacité douce. **Animation de flux** : `stroke-dasharray` animé (ou une particule
`<circle>` glissant le long du chemin) pour évoquer l'argent qui circule
source→destination. Désactivable si perf.

### Interactions
- **Hover arête** → surbrillance + `EdgeDetailPanel` (source→dest, asset, montant,
  count).
- **Hover/clic nœud** → met en avant ses arêtes incidentes + atténue le reste ;
  petit panneau : total in / out / net du nœud.
- **Zoom/pan/drag**, **légende** (couleurs par type), **sélecteur d'asset**,
  **slider top-N** (mappé sur `/flows?limit`).
- **État vide** : « Aucun flux à afficher. »

## 4. Clustering automatique

- **Par type de compte** (déterministe, explicable) : bank / client / supplier /
  contracts / world / other.
- **Force de cluster** : un centroïde par type ; `forceX/forceY` tirent chaque
  nœud vers le centroïde de son type → les groupes s'auto-organisent visuellement.
- **Méta-nœuds repliables** : un cluster peut s'afficher comme **un seul nœud
  agrégé** étiqueté (« Clients (42) », volume = somme du groupe). **Clic =
  déplier/replier.** Les arêtes d'un cluster replié sont agrégées vers/depuis le
  méta-nœud. Défaut : cluster **replié si > seuil** (ex. 12 nœuds), déplié sinon.
- État de repli porté en React (`Set` des types repliés) ; `applyClustering(graph,
  collapsedSet)` (pur) renvoie le graphe effectif (méta-nœuds + arêtes agrégées).

## 5. Time-scrubber

- `src/components/TimeScrubber.jsx` : timeline horizontale sur `buckets` triés.
- **Vue cumulative** : position T → graphe des flux **jusqu'à** T (via
  `filterToBucket`). Curseur par défaut = dernier bucket (tous les flux).
- **Play/pause** : avance T automatiquement (vitesse simple, ex. 1 bucket/600ms),
  les arêtes apparaissent/grossissent → effet temporel. Pause/reset.
- Au changement de T, le graphe se met à jour **sans recréer la simulation à
  zéro** quand c'est possible (mise à jour du data-join + `simulation.alpha`
  relancé doucement) pour une transition fluide.

## 6. Architecture (composants frontend)

| Fichier | Rôle |
|---|---|
| `src/lib/buildGraph.js` (new) | transfo pure : `kindOf`, `buildGraph`, `filterToBucket`, `applyClustering` |
| `src/components/FlowGraph.jsx` (new) | la simu d3 dans un `<svg>` géré par ref ; data-join ; zoom/drag/hover ; cleanup |
| `src/components/TimeScrubber.jsx` (new) | timeline + play/pause |
| `src/components/EdgeDetailPanel.jsx` (new) | détail d'arête au survol/clic |
| `src/components/GraphLegend.jsx` (new) | légende couleurs + sélecteur asset + slider top-N |
| `src/pages/Lens.jsx` (modify) | panneau « Flow graph » (en plus du tableau v1) |
| `package.json` (modify) | ajoute `d3` v7 |

## 7. Robustesse (« sans bug » — gravé dans le plan)

- **Cleanup de la simulation** : `useEffect` retourne `() => simulation.stop()` ;
  au changement d'asset/T/limit, on stoppe l'ancienne simu et on
  `svg.selectAll('*').remove()` avant de reconstruire. (Le bug d3+React n°1 :
  ticker après démontage → éliminé.)
- **Pas de setState dans le tick d3** : d3 mute les positions et patche le DOM
  directement ; React ne re-render pas à chaque tick.
- **Cap** : nœuds/arêtes bornés via `/flows?limit` + un cap de sécurité ; le
  collapse de clusters protège les gros graphes.
- **Transfo pure** isolée dans `buildGraph.js` → testable/inspectable même sans
  infra de test JS dans Horizon (décision Lens : pas de Jest en v1).
- **QA navigateur** (browse) avec captures à chaque phase.

## 8. Build en 4 phases (chaque phase QA + commit avant la suivante)

1. **Base** : `buildGraph` + `FlowGraph` force de base (nœuds/arêtes/zoom/drag/
   hover) sur `/flows`, sélecteur d'asset, intégré dans Lens. d3 ajouté.
2. **Habillage** : coloration par type, animation de flux, `EdgeDetailPanel`,
   `GraphLegend` + slider top-N, panneau nœud (in/out/net).
3. **Clustering** : force de cluster + méta-nœuds repliables (`applyClustering`).
4. **Time-scrubber** : `TimeScrubber` + vue cumulative + play/pause + transitions.

## 9. Hors périmètre

Détection de communautés par connectivité (on cluster par type), vue fenêtrée
(on fait cumulatif), export PNG, persistance de la disposition. → ultérieur.

## 10. Risques

- d3 + React mal géré = fuites/jank : mitigé par le pattern cleanup strict (§7).
- Beaucoup de features dans un composant : mitigé par le découpage 4 phases + la
  transfo pure isolée + QA par phase.
- Perf sur gros ledger : cap + clustering replié + animation désactivable.
