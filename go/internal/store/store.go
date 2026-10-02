package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/CassiopeiaCode/Keyweave/go/internal/domain"
	_ "modernc.org/sqlite"
)

var ErrVersionConflict = errors.New("version conflict")

type Store struct {
	db *sql.DB
}

func OpenSQLite(ctx context.Context, dsn string, migrationPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	for _, q := range []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	schema, err := os.ReadFile(migrationPath)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.ExecContext(ctx, string(schema)); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) CreateTemplate(ctx context.Context, t domain.CredentialTemplate) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO credential_templates (
	id,name,description,protocols_json,official_defaults_json,custom_fields_json,state_json,js_source,version,created_at,updated_at
) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID,
		t.Name,
		t.Description,
		mustJSON(t.Protocols),
		mustJSON(defaultMap(t.OfficialDefaults)),
		mustJSON(defaultMap(t.CustomFields)),
		mustJSON(defaultMap(t.State)),
		t.JSSource,
		defaultVersion(t.Version),
		ts(t.CreatedAt),
		ts(t.UpdatedAt),
	)
	return err
}

func (s *Store) GetTemplate(ctx context.Context, id string) (domain.CredentialTemplate, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id,name,description,protocols_json,official_defaults_json,custom_fields_json,state_json,js_source,version,created_at,updated_at
FROM credential_templates WHERE id=?`, id)

	var out domain.CredentialTemplate
	var protocols, defaults, custom, state string
	var createdAt, updatedAt string
	if err := row.Scan(
		&out.ID, &out.Name, &out.Description, &protocols, &defaults, &custom, &state, &out.JSSource, &out.Version, &createdAt, &updatedAt,
	); err != nil {
		return domain.CredentialTemplate{}, err
	}
	if err := unmarshal(protocols, &out.Protocols); err != nil {
		return domain.CredentialTemplate{}, err
	}
	if err := unmarshal(defaults, &out.OfficialDefaults); err != nil {
		return domain.CredentialTemplate{}, err
	}
	if err := unmarshal(custom, &out.CustomFields); err != nil {
		return domain.CredentialTemplate{}, err
	}
	if err := unmarshal(state, &out.State); err != nil {
		return domain.CredentialTemplate{}, err
	}
	out.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	out.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return out, nil
}

func (s *Store) UpdateTemplate(ctx context.Context, expectedVersion int64, t domain.CredentialTemplate) error {
	res, err := s.db.ExecContext(ctx, `
UPDATE credential_templates
SET name=?,description=?,protocols_json=?,official_defaults_json=?,custom_fields_json=?,state_json=?,js_source=?,version=version+1,updated_at=?
WHERE id=? AND version=?`,
		t.Name,
		t.Description,
		mustJSON(t.Protocols),
		mustJSON(defaultMap(t.OfficialDefaults)),
		mustJSON(defaultMap(t.CustomFields)),
		mustJSON(defaultMap(t.State)),
		t.JSSource,
		ts(t.UpdatedAt),
		t.ID,
		expectedVersion,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrVersionConflict
	}
	return nil
}

func (s *Store) CreateInstance(ctx context.Context, i domain.CredentialInstance) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO credential_instances (
	id,template_id,name,address,base_url,key_ciphertext,key_nonce,key_kid,models_json,availability,proxy,custom_fields_json,state_json,version,created_at,updated_at
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		i.ID, i.TemplateID, i.Name, i.Address, i.BaseURL,
		nullBlob(i.KeyCiphertext), nullBlob(i.KeyNonce), i.KeyKID,
		mustJSON(i.Models), i.Availability, i.Proxy,
		mustJSON(defaultMap(i.CustomFields)),
		mustJSON(defaultMap(i.State)),
		defaultVersion(i.Version),
		ts(i.CreatedAt), ts(i.UpdatedAt),
	)
	return err
}

func (s *Store) GetInstance(ctx context.Context, id string) (domain.CredentialInstance, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id,template_id,name,address,base_url,key_ciphertext,key_nonce,key_kid,models_json,availability,proxy,custom_fields_json,state_json,version,created_at,updated_at
FROM credential_instances WHERE id=?`, id)
	var out domain.CredentialInstance
	var models, custom, state string
	var createdAt, updatedAt string
	if err := row.Scan(
		&out.ID, &out.TemplateID, &out.Name, &out.Address, &out.BaseURL, &out.KeyCiphertext, &out.KeyNonce, &out.KeyKID,
		&models, &out.Availability, &out.Proxy, &custom, &state, &out.Version, &createdAt, &updatedAt,
	); err != nil {
		return domain.CredentialInstance{}, err
	}
	if err := unmarshal(models, &out.Models); err != nil {
		return domain.CredentialInstance{}, err
	}
	if err := unmarshal(custom, &out.CustomFields); err != nil {
		return domain.CredentialInstance{}, err
	}
	if err := unmarshal(state, &out.State); err != nil {
		return domain.CredentialInstance{}, err
	}
	out.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	out.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return out, nil
}

func (s *Store) UpdateInstance(ctx context.Context, expectedVersion int64, i domain.CredentialInstance) error {
	res, err := s.db.ExecContext(ctx, `
UPDATE credential_instances
SET name=?,address=?,base_url=?,key_ciphertext=?,key_nonce=?,key_kid=?,models_json=?,availability=?,proxy=?,custom_fields_json=?,state_json=?,version=version+1,updated_at=?
WHERE id=? AND version=?`,
		i.Name, i.Address, i.BaseURL, nullBlob(i.KeyCiphertext), nullBlob(i.KeyNonce), i.KeyKID,
		mustJSON(i.Models), i.Availability, i.Proxy, mustJSON(defaultMap(i.CustomFields)), mustJSON(defaultMap(i.State)),
		ts(i.UpdatedAt), i.ID, expectedVersion,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrVersionConflict
	}
	return nil
}

func (s *Store) CreateGroup(ctx context.Context, g domain.CredentialGroup) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO credential_groups (
	id,name,template_ids_json,scheduler_type,scheduler_code,config_json,state_json,version,created_at,updated_at
) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		g.ID, g.Name, mustJSON(g.TemplateIDs), g.SchedulerType, g.SchedulerCode,
		mustJSON(defaultMap(g.Config)), mustJSON(defaultMap(g.State)), defaultVersion(g.Version), ts(g.CreatedAt), ts(g.UpdatedAt),
	)
	return err
}

func (s *Store) GetGroup(ctx context.Context, id string) (domain.CredentialGroup, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id,name,template_ids_json,scheduler_type,scheduler_code,config_json,state_json,version,created_at,updated_at
FROM credential_groups WHERE id=?`, id)
	var out domain.CredentialGroup
	var tids, cfg, state string
	var createdAt, updatedAt string
	if err := row.Scan(&out.ID, &out.Name, &tids, &out.SchedulerType, &out.SchedulerCode, &cfg, &state, &out.Version, &createdAt, &updatedAt); err != nil {
		return domain.CredentialGroup{}, err
	}
	if err := unmarshal(tids, &out.TemplateIDs); err != nil {
		return domain.CredentialGroup{}, err
	}
	if err := unmarshal(cfg, &out.Config); err != nil {
		return domain.CredentialGroup{}, err
	}
	if err := unmarshal(state, &out.State); err != nil {
		return domain.CredentialGroup{}, err
	}
	out.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	out.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return out, nil
}

func (s *Store) UpdateGroup(ctx context.Context, expectedVersion int64, g domain.CredentialGroup) error {
	res, err := s.db.ExecContext(ctx, `
UPDATE credential_groups
SET name=?,template_ids_json=?,scheduler_type=?,scheduler_code=?,config_json=?,state_json=?,version=version+1,updated_at=?
WHERE id=? AND version=?`,
		g.Name, mustJSON(g.TemplateIDs), g.SchedulerType, g.SchedulerCode, mustJSON(defaultMap(g.Config)), mustJSON(defaultMap(g.State)), ts(g.UpdatedAt), g.ID, expectedVersion,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrVersionConflict
	}
	return nil
}

func (s *Store) CreateAccessKey(ctx context.Context, a domain.AccessKey) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO access_keys (id,name,secret_hash,secret_prefix,group_id,created_at,revoked_at)
VALUES (?,?,?,?,?,?,?)`,
		a.ID, a.Name, a.SecretHash, a.SecretPrefix, a.GroupID, ts(a.CreatedAt), nullableTS(a.RevokedAt),
	)
	return err
}

func (s *Store) GetAccessKeyByHash(ctx context.Context, hash []byte) (domain.AccessKey, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id,name,secret_hash,secret_prefix,group_id,created_at,revoked_at
FROM access_keys WHERE secret_hash=?`, hash)
	var out domain.AccessKey
	var createdAt string
	var revokedAt sql.NullString
	if err := row.Scan(&out.ID, &out.Name, &out.SecretHash, &out.SecretPrefix, &out.GroupID, &createdAt, &revokedAt); err != nil {
		return domain.AccessKey{}, err
	}
	out.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	if revokedAt.Valid {
		t, _ := time.Parse(time.RFC3339Nano, revokedAt.String)
		out.RevokedAt = &t
	}
	return out, nil
}

func (s *Store) CreateSource(ctx context.Context, src domain.CredentialSource) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO credential_sources (id,name,type,template_id,config_json,code,state_json,version,created_at,updated_at)
VALUES (?,?,?,?,?,?,?,?,?,?)`,
		src.ID, src.Name, src.Type, src.TemplateID, mustJSON(defaultMap(src.Config)), src.Code, mustJSON(defaultMap(src.State)),
		defaultVersion(src.Version), ts(src.CreatedAt), ts(src.UpdatedAt),
	)
	return err
}

func (s *Store) GetSource(ctx context.Context, id string) (domain.CredentialSource, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id,name,type,template_id,config_json,code,state_json,version,created_at,updated_at
FROM credential_sources WHERE id=?`, id)
	var out domain.CredentialSource
	var cfg, state string
	var createdAt, updatedAt string
	if err := row.Scan(&out.ID, &out.Name, &out.Type, &out.TemplateID, &cfg, &out.Code, &state, &out.Version, &createdAt, &updatedAt); err != nil {
		return domain.CredentialSource{}, err
	}
	if err := unmarshal(cfg, &out.Config); err != nil {
		return domain.CredentialSource{}, err
	}
	if err := unmarshal(state, &out.State); err != nil {
		return domain.CredentialSource{}, err
	}
	out.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	out.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return out, nil
}

func (s *Store) UpdateSource(ctx context.Context, expectedVersion int64, src domain.CredentialSource) error {
	res, err := s.db.ExecContext(ctx, `
UPDATE credential_sources
SET name=?,type=?,template_id=?,config_json=?,code=?,state_json=?,version=version+1,updated_at=?
WHERE id=? AND version=?`,
		src.Name, src.Type, src.TemplateID, mustJSON(defaultMap(src.Config)), src.Code, mustJSON(defaultMap(src.State)), ts(src.UpdatedAt), src.ID, expectedVersion,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrVersionConflict
	}
	return nil
}

func (s *Store) PutSecretItem(ctx context.Context, item domain.SecretItem) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO secret_items (id,owner_type,owner_id,name,ciphertext,nonce,key_id,created_at,updated_at)
VALUES (?,?,?,?,?,?,?,?,?)
ON CONFLICT(owner_type,owner_id,name) DO UPDATE SET
  ciphertext=excluded.ciphertext,
  nonce=excluded.nonce,
  key_id=excluded.key_id,
  updated_at=excluded.updated_at`,
		item.ID, item.OwnerType, item.OwnerID, item.Name, item.Ciphertext, item.Nonce, item.KeyID, ts(item.CreatedAt), ts(item.UpdatedAt),
	)
	return err
}

func (s *Store) GetSecretItem(ctx context.Context, ownerType, ownerID, name string) (domain.SecretItem, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id,owner_type,owner_id,name,ciphertext,nonce,key_id,created_at,updated_at
FROM secret_items WHERE owner_type=? AND owner_id=? AND name=?`, ownerType, ownerID, name)
	var out domain.SecretItem
	var createdAt, updatedAt string
	if err := row.Scan(&out.ID, &out.OwnerType, &out.OwnerID, &out.Name, &out.Ciphertext, &out.Nonce, &out.KeyID, &createdAt, &updatedAt); err != nil {
		return domain.SecretItem{}, err
	}
	out.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	out.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return out, nil
}

func defaultVersion(v int64) int64 {
	if v == 0 {
		return 1
	}
	return v
}

func ts(t time.Time) string {
	if t.IsZero() {
		t = time.Now().UTC()
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func nullableTS(t *time.Time) any {
	if t == nil {
		return nil
	}
	return ts(*t)
}

func mustJSON(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

func unmarshal(raw string, target any) error {
	if raw == "" {
		raw = "{}"
	}
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}

func defaultMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return map[string]interface{}{}
	}
	return m
}

func nullBlob(v []byte) any {
	if len(v) == 0 {
		return nil
	}
	return v
}
