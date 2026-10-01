package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/CassiopeiaCode/Keyweave/go/internal/domain"
	"github.com/CassiopeiaCode/Keyweave/go/internal/secrets"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	migrationPath := filepath.Clean(filepath.Join("..", "..", "..", "migrations", "001_init.sql"))
	s, err := OpenSQLite(ctx, dbPath, migrationPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestTemplateCRUDAndVersionConflict(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	tpl := domain.CredentialTemplate{
		ID:               "tpl-1",
		Name:             "generic",
		Protocols:        []domain.ProtocolID{"openai.chat"},
		OfficialDefaults: map[string]interface{}{"availability": 100},
		CustomFields:     map[string]interface{}{"a": "b"},
		State:            map[string]interface{}{},
		JSSource:         "export default {};",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := s.CreateTemplate(ctx, tpl); err != nil {
		t.Fatalf("create template: %v", err)
	}
	got, err := s.GetTemplate(ctx, tpl.ID)
	if err != nil {
		t.Fatalf("get template: %v", err)
	}
	if got.Name != tpl.Name {
		t.Fatalf("unexpected template name: %s", got.Name)
	}

	got.Name = "generic-2"
	got.UpdatedAt = time.Now().UTC()
	if err := s.UpdateTemplate(ctx, got.Version, got); err != nil {
		t.Fatalf("update template: %v", err)
	}

	if err := s.UpdateTemplate(ctx, 1, got); err != ErrVersionConflict {
		t.Fatalf("expected version conflict, got: %v", err)
	}
}

func TestSecretStorageNoPlaintext(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	now := time.Now().UTC()

	g := domain.CredentialGroup{
		ID:            "g-1",
		Name:          "default",
		TemplateIDs:   []string{"tpl-1"},
		SchedulerType: "availability-round-robin",
		Config:        map[string]interface{}{},
		State:         map[string]interface{}{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.CreateGroup(ctx, g); err != nil {
		t.Fatalf("create group: %v", err)
	}

	tpl := domain.CredentialTemplate{
		ID:               "tpl-1",
		Name:             "template",
		Protocols:        []domain.ProtocolID{"openai.chat"},
		OfficialDefaults: map[string]interface{}{},
		CustomFields:     map[string]interface{}{},
		State:            map[string]interface{}{},
		JSSource:         "x",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.CreateTemplate(ctx, tpl); err != nil {
		t.Fatalf("create template: %v", err)
	}

	plaintext := []byte("upstream-secret-token")
	masterKey := []byte("0123456789abcdef0123456789abcdef")
	sealed, err := secrets.Seal(masterKey, "k1", plaintext)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}

	inst := domain.CredentialInstance{
		ID:         "i-1",
		TemplateID: tpl.ID,
		CredentialOfficialFields: domain.CredentialOfficialFields{
			Availability: 100,
			Models:       []domain.ModelDefinition{{ID: "gpt-4.1"}},
		},
		KeyCiphertext: sealed.Ciphertext,
		KeyNonce:      sealed.Nonce,
		KeyKID:        &sealed.KeyID,
		CustomFields:  map[string]interface{}{},
		State:         map[string]interface{}{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.CreateInstance(ctx, inst); err != nil {
		t.Fatalf("create instance: %v", err)
	}

	item := domain.SecretItem{
		ID:         "sec-1",
		OwnerType:  "instance",
		OwnerID:    inst.ID,
		Name:       "apiKey",
		Ciphertext: sealed.Ciphertext,
		Nonce:      sealed.Nonce,
		KeyID:      sealed.KeyID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.PutSecretItem(ctx, item); err != nil {
		t.Fatalf("put secret item: %v", err)
	}

	got, err := s.GetSecretItem(ctx, "instance", inst.ID, "apiKey")
	if err != nil {
		t.Fatalf("get secret item: %v", err)
	}
	open, err := secrets.Open(masterKey, got.Ciphertext, got.Nonce)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if string(open) != string(plaintext) {
		t.Fatalf("unexpected plaintext")
	}

	var keyCipher []byte
	if err := s.DB().QueryRowContext(ctx, `SELECT key_ciphertext FROM credential_instances WHERE id=?`, inst.ID).Scan(&keyCipher); err != nil {
		t.Fatalf("read key_ciphertext: %v", err)
	}
	if string(keyCipher) == string(plaintext) {
		t.Fatalf("plaintext was persisted in key_ciphertext")
	}

	var leakCount int
	if err := s.DB().QueryRowContext(ctx, `
SELECT count(*) FROM credential_instances
WHERE id=? AND (
  custom_fields_json LIKE '%' || ? || '%' OR
  state_json LIKE '%' || ? || '%'
)`, inst.ID, string(plaintext), string(plaintext)).Scan(&leakCount); err != nil {
		t.Fatalf("query plaintext leak: %v", err)
	}
	if leakCount != 0 {
		t.Fatalf("plaintext leaked in json fields")
	}
}

func TestAccessKeyLookupByHash(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	now := time.Now().UTC()

	g := domain.CredentialGroup{
		ID:            "g-1",
		Name:          "default",
		TemplateIDs:   []string{},
		SchedulerType: "availability-round-robin",
		Config:        map[string]interface{}{},
		State:         map[string]interface{}{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.CreateGroup(ctx, g); err != nil {
		t.Fatalf("create group: %v", err)
	}

	hash := []byte("hash-value")
	in := domain.AccessKey{
		ID:        "ak-1",
		Name:      "k",
		SecretHash: hash,
		GroupID:   g.ID,
		CreatedAt: now,
	}
	if err := s.CreateAccessKey(ctx, in); err != nil {
		t.Fatalf("create access key: %v", err)
	}

	out, err := s.GetAccessKeyByHash(ctx, hash)
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if out.ID != in.ID {
		t.Fatalf("unexpected id: %s", out.ID)
	}
}

func TestPRAGMAsApplied(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	var mode string
	if err := s.DB().QueryRowContext(ctx, "PRAGMA journal_mode;").Scan(&mode); err != nil {
		t.Fatalf("read journal_mode: %v", err)
	}
	if mode == "" {
		t.Fatalf("journal mode empty")
	}

	var fk int
	if err := s.DB().QueryRowContext(ctx, "PRAGMA foreign_keys;").Scan(&fk); err != nil {
		t.Fatalf("read foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys disabled")
	}
}

func TestGetMissingReturnsNoRows(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	_, err := s.GetSource(ctx, "missing")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errorsIsNoRows(err) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func errorsIsNoRows(err error) bool {
	return err == sql.ErrNoRows
}
