#!/usr/bin/env bash
#
# Murabaha end-to-end demo against a running `corren server start`.
#
# Flow (spec §15):
#   1. provision the bank treasury
#   2. create the contract (PROMISE)
#   3. try to SELL BEFORE ACQUIRING  -> 422 ERR_SHARIA_VIOLATION (AAOIFI-SS-8)
#   4. acquire (bank pays supplier, takes possession)
#   5. sell (asset to client, receivable + deferred profit are born)
#   6. pay 3 installments (profit recognized pro rata, FAS 28)
#   7. late penalty -> charity pool ONLY (AAOIFI-SS-3)
#   8. full audit trail with chain verification
#
set -euo pipefail

HOST="${CORREN_HOST:-http://localhost:3068}"
LEDGER="${CORREN_LEDGER:-demo}"
ID="murdemo$(date +%Y%m%d)$RANDOM"

say()  { printf '\n\e[1;36m== %s\e[0m\n' "$*"; }
call() { # call METHOD PATH [JSON_BODY]
  local method="$1" path="$2" body="${3:-}"
  if [ -n "$body" ]; then
    curl -sS -X "$method" "$HOST/$LEDGER$path" \
      -H 'Content-Type: application/json' -d "$body"
  else
    curl -sS -X "$method" "$HOST/$LEDGER$path"
  fi
  echo
}

say "1. Provision treasury: 20,000,000 (200,000.00 SAR) from @world"
call POST /transactions '{
  "postings": [
    {"source": "@world", "destination": "@bank:treasury", "asset": "SAR2", "amount": 20000000}
  ]
}'

say "2. Create Murabaha contract $ID (cost 100,000.00 + fixed markup 10,000.00 SAR, 24 installments)"
call POST /contracts '{
  "type": "murabaha",
  "id": "'"$ID"'",
  "params": {
    "asset_code": "VHCL42A",
    "cost":   {"asset": "SAR2", "amount": 10000000},
    "markup": {"asset": "SAR2", "amount": 1000000},
    "client": "@client:ameziane",
    "supplier": "@supplier:toyota",
    "bank_treasury": "@bank:treasury",
    "installments": 24,
    "first_due": "2026-07-01T00:00:00Z",
    "period_days": 30
  }
}'

say "3. Try to sell BEFORE acquiring -> rejected, referenced (I-1 qabd, AAOIFI-SS-8)"
call POST "/contracts/$ID/transitions/sell" || true

say "4. Acquire: bank pays the supplier and takes possession"
call POST "/contracts/$ID/transitions/acquire"

say "5. Sell: asset to client, receivable born, profit deferred (FAS 28)"
call POST "/contracts/$ID/transitions/sell"

say "6. Fund client and pay 3 installments"
call POST /transactions '{
  "postings": [
    {"source": "@world", "destination": "@client:ameziane", "asset": "SAR2", "amount": 1500000}
  ]
}'
for seq in 1 2 3; do
  call POST "/contracts/$ID/transitions/pay_installment" '{"seq": '"$seq"'}'
done

say "7. Late penalty: 200.00 SAR -> charity pool (never bank income, AAOIFI-SS-3)"
call POST "/contracts/$ID/transitions/late_penalty" '{
  "seq": 4, "amount": 20000, "destination": "@charity:pool"
}'

say "7b. Proof: penalty to bank income is REFUSED"
call POST "/contracts/$ID/transitions/late_penalty" '{
  "seq": 5, "amount": 20000, "destination": "@bank:income:murabaha"
}' || true

say "8. Full audit trail, hash chain re-verified (chain_valid)"
call GET "/contracts/$ID/audit?verify=true"

say "Done. Contract state:"
call GET "/contracts/$ID"
