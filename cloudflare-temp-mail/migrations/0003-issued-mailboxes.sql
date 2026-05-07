CREATE TABLE IF NOT EXISTS mailboxes (
  domain TEXT NOT NULL,
  user TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  issued_at TEXT NOT NULL,
  PRIMARY KEY (domain, user)
);

CREATE INDEX IF NOT EXISTS idx_mailboxes_issued_at ON mailboxes(issued_at);
