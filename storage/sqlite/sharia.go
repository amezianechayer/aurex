package sqlite

import (
	"database/sql"
	"encoding/json"
	"math"
	"time"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/ledger/query"
	"github.com/huandu/go-sqlbuilder"
)

func (s *SQLiteStore) SaveAsset(a core.AssetEntry) error {
	metaJSON, _ := json.Marshal(a.Metadata)
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("assets")
	ib.Cols("id", "name", "precision", "category", "aaoifi_class", "metadata", "created_at")
	ib.Values(a.ID, a.Name, a.Precision, string(a.Category), a.AaoifiClass, string(metaJSON), a.CreatedAt)
	q, args := ib.BuildWithFlavor(sqlbuilder.SQLite)
	_, err := s.db.Exec(q, args...)
	return err
}

func (s *SQLiteStore) FindAssets() ([]core.AssetEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, name, precision, category, aaoifi_class, COALESCE(metadata,'{}'), created_at
		 FROM assets ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var assets []core.AssetEntry
	for rows.Next() {
		var a core.AssetEntry
		var category, metaStr string
		if err := rows.Scan(&a.ID, &a.Name, &a.Precision, &category, &a.AaoifiClass, &metaStr, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Category = core.AssetCategory(category)
		json.Unmarshal([]byte(metaStr), &a.Metadata)
		assets = append(assets, a)
	}
	return assets, nil
}

func (s *SQLiteStore) FindAsset(id string) (*core.AssetEntry, error) {
	row := s.db.QueryRow(
		`SELECT id, name, precision, category, aaoifi_class, COALESCE(metadata,'{}'), created_at
		 FROM assets WHERE id = ?`, id,
	)
	var a core.AssetEntry
	var category, metaStr string
	err := row.Scan(&a.ID, &a.Name, &a.Precision, &category, &a.AaoifiClass, &metaStr, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.Category = core.AssetCategory(category)
	json.Unmarshal([]byte(metaStr), &a.Metadata)
	return &a, nil
}

func (s *SQLiteStore) SaveContract(c core.ShariaContract) error {
	partiesJSON, _ := json.Marshal(c.Parties)
	termsJSON, _ := json.Marshal(c.Terms)
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("sharia_contracts")
	ib.Cols("id", "type", "status", "ledger", "parties", "terms", "aaoifi_fas", "created_at", "updated_at")
	ib.Values(
		c.ID, string(c.Type), string(c.Status), c.Ledger,
		string(partiesJSON), string(termsJSON), c.AaoifiFAS,
		c.CreatedAt, c.UpdatedAt,
	)
	q, args := ib.BuildWithFlavor(sqlbuilder.SQLite)
	_, err := s.db.Exec(q, args...)
	return err
}

func (s *SQLiteStore) UpdateContractStatus(id string, status core.ContractStatus) error {
	_, err := s.db.Exec(
		`UPDATE sharia_contracts SET status = ?, updated_at = ? WHERE id = ?`,
		string(status), time.Now().Format(time.RFC3339), id,
	)
	return err
}

func (s *SQLiteStore) FindContracts(q query.Query) (query.Cursor, error) {
	q.Limit = int(math.Max(-1, math.Min(float64(q.Limit), 100))) + 1
	c := query.Cursor{}
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "type", "status", "ledger", "parties", "terms", "aaoifi_fas", "created_at", "updated_at")
	sb.From("sharia_contracts")
	sb.OrderBy("created_at desc")
	sb.Limit(q.Limit)
	if q.After != "" {
		sb.Where(sb.LessThan("id", q.After))
	}
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.SQLite)
	rows, err := s.db.Query(sqlq, args...)
	if err != nil {
		return c, err
	}
	defer rows.Close()
	var results []core.ShariaContract
	for rows.Next() {
		var contract core.ShariaContract
		var ctype, status, partiesStr, termsStr string
		if err := rows.Scan(
			&contract.ID, &ctype, &status, &contract.Ledger,
			&partiesStr, &termsStr, &contract.AaoifiFAS,
			&contract.CreatedAt, &contract.UpdatedAt,
		); err != nil {
			return c, err
		}
		contract.Type = core.ContractType(ctype)
		contract.Status = core.ContractStatus(status)
		json.Unmarshal([]byte(partiesStr), &contract.Parties)
		json.Unmarshal([]byte(termsStr), &contract.Terms)
		results = append(results, contract)
	}
	c.PageSize = q.Limit - 1
	c.HasMore = len(results) == q.Limit
	if c.HasMore {
		results = results[:len(results)-1]
	}
	c.Data = results
	return c, nil
}

func (s *SQLiteStore) FindContract(id string) (*core.ShariaContract, error) {
	row := s.db.QueryRow(
		`SELECT id, type, status, ledger, parties, terms, aaoifi_fas, created_at, updated_at
		 FROM sharia_contracts WHERE id = ?`, id,
	)
	var contract core.ShariaContract
	var ctype, status, partiesStr, termsStr string
	err := row.Scan(
		&contract.ID, &ctype, &status, &contract.Ledger,
		&partiesStr, &termsStr, &contract.AaoifiFAS,
		&contract.CreatedAt, &contract.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	contract.Type = core.ContractType(ctype)
	contract.Status = core.ContractStatus(status)
	json.Unmarshal([]byte(partiesStr), &contract.Parties)
	json.Unmarshal([]byte(termsStr), &contract.Terms)
	return &contract, nil
}

func (s *SQLiteStore) SaveCertificate(cert core.ShariaCertificate) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("sharia_certificates")
	ib.Cols("id", "contract_id", "txid", "constraint", "result", "issued_at", "authority")
	ib.Values(cert.ID, cert.ContractID, cert.TxID, cert.Constraint, cert.Result, cert.IssuedAt, cert.Authority)
	q, args := ib.BuildWithFlavor(sqlbuilder.SQLite)
	_, err := s.db.Exec(q, args...)
	return err
}

func (s *SQLiteStore) FindCertificates(contractID string) ([]core.ShariaCertificate, error) {
	rows, err := s.db.Query(
		`SELECT id, COALESCE(contract_id,''), COALESCE(txid,0), "constraint", result, issued_at, authority
		 FROM sharia_certificates WHERE contract_id = ? ORDER BY issued_at DESC`, contractID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var certs []core.ShariaCertificate
	for rows.Next() {
		var cert core.ShariaCertificate
		if err := rows.Scan(&cert.ID, &cert.ContractID, &cert.TxID, &cert.Constraint, &cert.Result, &cert.IssuedAt, &cert.Authority); err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	return certs, nil
}

func (s *SQLiteStore) SaveAPIKey(k core.APIKey) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("api_keys")
	ib.Cols("key_hash", "name", "role", "created_at", "expires_at")
	ib.Values(k.KeyHash, k.Name, k.Role, k.CreatedAt, k.ExpiresAt)
	q, args := ib.BuildWithFlavor(sqlbuilder.SQLite)
	_, err := s.db.Exec(q, args...)
	return err
}

func (s *SQLiteStore) FindAPIKey(keyHash string) (*core.APIKey, error) {
	row := s.db.QueryRow(
		`SELECT key_hash, name, role, created_at, COALESCE(expires_at,'')
		 FROM api_keys WHERE key_hash = ?`, keyHash,
	)
	var k core.APIKey
	err := row.Scan(&k.KeyHash, &k.Name, &k.Role, &k.CreatedAt, &k.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (s *SQLiteStore) DeleteAPIKey(keyHash string) error {
	_, err := s.db.Exec(`DELETE FROM api_keys WHERE key_hash = ?`, keyHash)
	return err
}
