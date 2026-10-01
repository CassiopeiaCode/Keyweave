PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  applied_at TEXT NOT NULL
);

CREATE TABLE credential_templates (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  protocols_json TEXT NOT NULL,
  official_defaults_json TEXT NOT NULL DEFAULT '{}',
  custom_fields_json TEXT NOT NULL DEFAULT '{}',
  state_json TEXT NOT NULL DEFAULT '{}',
  js_source TEXT NOT NULL,
  version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE credential_instances (
  id TEXT PRIMARY KEY,
  template_id TEXT NOT NULL REFERENCES credential_templates(id)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  name TEXT,

  address TEXT,
  base_url TEXT,

  key_ciphertext BLOB,
  key_nonce BLOB,
  key_kid TEXT,

  models_json TEXT,
  availability REAL NOT NULL DEFAULT 100,
  proxy TEXT,

  custom_fields_json TEXT NOT NULL DEFAULT '{}',
  state_json TEXT NOT NULL DEFAULT '{}',

  version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,

  CHECK(availability = availability)
);

CREATE INDEX idx_credential_instances_template
  ON credential_instances(template_id);

CREATE INDEX idx_credential_instances_availability
  ON credential_instances(availability DESC);

CREATE TABLE credential_groups (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  template_ids_json TEXT NOT NULL,
  scheduler_type TEXT NOT NULL,
  scheduler_code TEXT,
  config_json TEXT NOT NULL DEFAULT '{}',
  state_json TEXT NOT NULL DEFAULT '{}',
  version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE access_keys (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  secret_hash BLOB NOT NULL UNIQUE,
  secret_prefix TEXT,
  group_id TEXT NOT NULL REFERENCES credential_groups(id)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  created_at TEXT NOT NULL,
  revoked_at TEXT
);

CREATE INDEX idx_access_keys_group ON access_keys(group_id);

CREATE TABLE credential_sources (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  template_id TEXT REFERENCES credential_templates(id)
    ON UPDATE CASCADE ON DELETE SET NULL,
  config_json TEXT NOT NULL DEFAULT '{}',
  code TEXT,
  state_json TEXT NOT NULL DEFAULT '{}',
  version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE secret_items (
  id TEXT PRIMARY KEY,
  owner_type TEXT NOT NULL,
  owner_id TEXT NOT NULL,
  name TEXT NOT NULL,
  ciphertext BLOB NOT NULL,
  nonce BLOB NOT NULL,
  key_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(owner_type, owner_id, name)
);

CREATE TABLE audit_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  actor TEXT NOT NULL,
  action TEXT NOT NULL,
  object_type TEXT NOT NULL,
  object_id TEXT,
  redacted_diff_json TEXT,
  created_at TEXT NOT NULL
);

INSERT OR IGNORE INTO schema_migrations(version, applied_at)
VALUES (1, datetime('now'));
