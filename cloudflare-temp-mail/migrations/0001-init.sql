CREATE TABLE IF NOT EXISTS domains (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  domain TEXT NOT NULL UNIQUE,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE IF NOT EXISTS messages (
  id TEXT PRIMARY KEY,
  domain TEXT NOT NULL,
  user TEXT NOT NULL,
  sender TEXT NOT NULL,
  recipient TEXT NOT NULL,
  subject TEXT NOT NULL DEFAULT '',
  received_at TEXT NOT NULL,
  raw_key TEXT NOT NULL,
  text_key TEXT,
  html_key TEXT,
  otp TEXT,
  size INTEGER NOT NULL DEFAULT 0,
  deleted_at TEXT,
  purged_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_messages_mailbox ON messages(domain, user, received_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_retention ON messages(received_at, deleted_at, purged_at);
CREATE INDEX IF NOT EXISTS idx_domains_enabled ON domains(enabled, domain);

INSERT OR IGNORE INTO domains(domain, enabled) VALUES ('example.com', 1);
