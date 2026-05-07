CREATE INDEX IF NOT EXISTS idx_messages_cleanup_deleted ON messages(deleted_at, purged_at);
CREATE INDEX IF NOT EXISTS idx_messages_cleanup_mailbox ON messages(domain, user, deleted_at, purged_at);
