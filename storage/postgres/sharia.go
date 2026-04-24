package postgres

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/ledger/query"
	"github.com/jackc/pgx/v4"
)

func (s *PGStore) SaveAsset(a core.AssetEntry) error {
	metaJSON, _ := json.Marshal(a.Metadata)
	_, err := s.Conn().Exec(context.Background(),
		`INSERT INTO `+s.table("assets")+
			` (id, name, precision, category, aaoifi_class, metadata, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		a.ID, a.Name, a.Precision, string(a.Category), a.AaoifiClass, string(metaJSON), a.CreatedAt,
	)
	return err
}

func (s *PGStore) FindAssets() ([]core.AssetEntry, error) {
	rows, err := s.Conn().Query(context.Background(),
		`SELECT id, name, precision, category, aaoifi_class, COALESCE(metadata,'{}'), created_at
		 FROM `+s.table("assets")+` ORDER BY created_at DESC`,
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

func (s *PGStore) FindAsset(id string) (*core.AssetEntry, error) {
	row := s.Conn().QueryRow(context.Background(),
		`SELECT id, name, precision, category, aaoifi_class, COALESCE(metadata,'{}'), created_at
		 FROM `+s.table("assets")+` WHERE id = $1`, id,
	)
	var a core.AssetEntry
	var category, metaStr string
	err := row.Scan(&a.ID, &a.Name, &a.Precision, &category, &a.AaoifiClass, &metaStr, &a.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.Category = core.AssetCategory(category)
	json.Unmarshal([]byte(metaStr), &a.Metadata)
	return &a, nil
}

func (s *PGStore) SaveContract(c core.ShariaContract) error {
	partiesJSON, _ := json.Marshal(c.Parties)
	termsJSON, _ := json.Marshal(c.Terms)
	_, err := s.Conn().Exec(context.Background(),
		`INSERT INTO `+s.table("sharia_contracts")+
			` (id, type, status, ledger, parties, terms, aaoifi_fas, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		c.ID, string(c.Type), string(c.Status), c.Ledger,
		string(partiesJSON), string(termsJSON), c.AaoifiFAS,
		c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (s *PGStore) UpdateContractStatus(id string, status core.ContractStatus) error {
	_, err := s.Conn().Exec(context.Background(),
		`UPDATE `+s.table("sharia_contracts")+` SET status=$1, updated_at=$2 WHERE id=$3`,
		string(status), time.Now().Format(time.RFC3339), id,
	)
	return err
}

func (s *PGStore) FindContracts(q query.Query) (query.Cursor, error) {
	q.Limit = int(math.Max(-1, math.Min(float64(q.Limit), 100))) + 1
	c := query.Cursor{}
	var sqlq string
	var args []interface{}
	if q.After != "" {
		sqlq = `SELECT id, type, status, ledger, parties, terms, aaoifi_fas, created_at, updated_at
				FROM ` + s.table("sharia_contracts") + ` WHERE id < $1 ORDER BY created_at DESC LIMIT $2`
		args = []interface{}{q.After, q.Limit}
	} else {
		sqlq = `SELECT id, type, status, ledger, parties, terms, aaoifi_fas, created_at, updated_at
				FROM ` + s.table("sharia_contracts") + ` ORDER BY created_at DESC LIMIT $1`
		args = []interface{}{q.Limit}
	}
	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
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

func (s *PGStore) FindContract(id string) (*core.ShariaContract, error) {
	row := s.Conn().QueryRow(context.Background(),
		`SELECT id, type, status, ledger, parties, terms, aaoifi_fas, created_at, updated_at
		 FROM `+s.table("sharia_contracts")+` WHERE id = $1`, id,
	)
	var contract core.ShariaContract
	var ctype, status, partiesStr, termsStr string
	err := row.Scan(
		&contract.ID, &ctype, &status, &contract.Ledger,
		&partiesStr, &termsStr, &contract.AaoifiFAS,
		&contract.CreatedAt, &contract.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
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

func (s *PGStore) SaveCertificate(cert core.ShariaCertificate) error {
	_, err := s.Conn().Exec(context.Background(),
		`INSERT INTO `+s.table("sharia_certificates")+
			` (id, contract_id, txid, "constraint", result, issued_at, authority)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		cert.ID, cert.ContractID, cert.TxID, cert.Constraint, cert.Result, cert.IssuedAt, cert.Authority,
	)
	return err
}

func (s *PGStore) FindCertificates(contractID string) ([]core.ShariaCertificate, error) {
	rows, err := s.Conn().Query(context.Background(),
		`SELECT id, COALESCE(contract_id,''), COALESCE(txid,0), "constraint", result, issued_at, authority
		 FROM `+s.table("sharia_certificates")+
			` WHERE contract_id = $1 ORDER BY issued_at DESC`, contractID,
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

func (s *PGStore) SaveAPIKey(k core.APIKey) error {
	_, err := s.Conn().Exec(context.Background(),
		`INSERT INTO `+s.table("api_keys")+
			` (key_hash, name, role, tier, created_at, expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		k.KeyHash, k.Name, k.Role, k.Tier, k.CreatedAt, k.ExpiresAt,
	)
	return err
}

func (s *PGStore) FindAPIKey(keyHash string) (*core.APIKey, error) {
	row := s.Conn().QueryRow(context.Background(),
		`SELECT key_hash, name, role, COALESCE(tier,'sandbox'), created_at, COALESCE(expires_at,'')
		 FROM `+s.table("api_keys")+` WHERE key_hash = $1`, keyHash,
	)
	var k core.APIKey
	err := row.Scan(&k.KeyHash, &k.Name, &k.Role, &k.Tier, &k.CreatedAt, &k.ExpiresAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (s *PGStore) ListAPIKeys() ([]core.APIKey, error) {
	rows, err := s.Conn().Query(context.Background(),
		`SELECT key_hash, name, role, COALESCE(tier,'sandbox'), created_at, COALESCE(expires_at,'')
		 FROM `+s.table("api_keys")+` ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []core.APIKey
	for rows.Next() {
		var k core.APIKey
		if err := rows.Scan(&k.KeyHash, &k.Name, &k.Role, &k.Tier, &k.CreatedAt, &k.ExpiresAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *PGStore) DeleteAPIKey(keyHash string) error {
	_, err := s.Conn().Exec(context.Background(),
		`DELETE FROM `+s.table("api_keys")+` WHERE key_hash = $1`, keyHash,
	)
	return err
}
