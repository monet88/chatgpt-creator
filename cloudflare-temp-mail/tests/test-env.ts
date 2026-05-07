import type { Env, MessageRow } from '../src/types';

export interface TestState {
  domains: string[];
  messages: MessageRow[];
  objects: Map<string, string>;
}

class FakeStatement {
  private values: unknown[] = [];

  constructor(private sql: string, private state: TestState) {}

  bind(...values: unknown[]) {
    this.values = values;
    return this;
  }

  async all<T>() {
    const sql = this.sql;
    if (sql.includes('FROM domains')) return { results: this.state.domains.map((domain) => ({ domain })) as T[] };
    if (sql.includes('WHERE domain = ? AND user = ?') && sql.includes('ORDER BY received_at DESC')) {
      const [domain, user] = this.values;
      return { results: this.state.messages.filter((message) => message.domain === domain && message.user === user && !message.deleted_at && !message.purged_at) as T[] };
    }
    if (sql.includes('WHERE domain = ? AND user = ? AND purged_at IS NULL')) {
      const [domain, user] = this.values;
      return { results: this.state.messages.filter((message) => message.domain === domain && message.user === user && !message.purged_at) as T[] };
    }
    if (sql.includes('purged_at IS NULL') && sql.includes('ORDER BY received_at ASC')) {
      const [cutoff, limit] = this.values as [string, number];
      return { results: this.state.messages.filter((message) => !message.purged_at && (Boolean(message.deleted_at) || message.received_at < cutoff)).slice(0, limit) as T[] };
    }
    return { results: [] as T[] };
  }

  async first<T>() {
    if (this.sql.includes('SELECT otp')) {
      const [domain, user] = this.values;
      return (this.state.messages.find((message) => message.domain === domain && message.user === user && message.otp && !message.deleted_at && !message.purged_at) ?? null) as T | null;
    }
    if (this.sql.includes('WHERE id = ? AND domain = ? AND user = ?')) {
      const [id, domain, user] = this.values;
      return (this.state.messages.find((message) => message.id === id && message.domain === domain && message.user === user && !message.deleted_at && !message.purged_at) ?? null) as T | null;
    }
    if (this.sql.includes('WHERE id = ?')) {
      const [id] = this.values;
      return (this.state.messages.find((message) => message.id === id && !message.purged_at) ?? null) as T | null;
    }
    return null;
  }

  async run() {
    if (this.sql.includes('INSERT INTO messages')) {
      const [id, domain, user, sender, recipient, subject, receivedAt, rawKey, textKey, htmlKey, otp, size] = this.values;
      this.state.messages = [...this.state.messages, { id, domain, user, sender, recipient, subject, received_at: receivedAt, raw_key: rawKey, text_key: textKey, html_key: htmlKey, otp, size, deleted_at: null, purged_at: null } as MessageRow];
    }
    if (this.sql.includes('UPDATE messages SET deleted_at = COALESCE(deleted_at, ?) WHERE domain = ? AND user = ?')) {
      const [deletedAt, domain, user] = this.values;
      this.state.messages = this.state.messages.map((message) => (message.domain === domain && message.user === user ? { ...message, deleted_at: message.deleted_at ?? String(deletedAt) } : message));
    }
    if (this.sql.includes('UPDATE messages SET deleted_at = COALESCE(deleted_at, ?) WHERE id = ?')) {
      const [deletedAt, id] = this.values;
      this.state.messages = this.state.messages.map((message) => (message.id === id ? { ...message, deleted_at: message.deleted_at ?? String(deletedAt) } : message));
    }
    if (this.sql.includes('purged_at = ? WHERE id = ?')) {
      const [deletedAt, purgedAt, id] = this.values;
      this.state.messages = this.state.messages.map((message) => (message.id === id ? { ...message, deleted_at: message.deleted_at ?? String(deletedAt), purged_at: String(purgedAt) } : message));
    }
    return { success: true };
  }
}

export const createTestEnv = (initial?: Partial<TestState>) => {
  const state: TestState = {
    domains: initial?.domains ?? ['example.com'],
    messages: initial?.messages ?? [],
    objects: initial?.objects ?? new Map(),
  };
  const env = {
    ENABLED_DOMAINS: 'example.com',
    RETENTION_DAYS: '3',
    PAGE_LIMIT: '50',
    CLEANUP_BATCH_SIZE: '100',
    DB: { prepare: (sql: string) => new FakeStatement(sql, state) },
    MAIL_BUCKET: {
      put: async (key: string, value: string) => state.objects.set(key, value),
      get: async (key: string) => (state.objects.has(key) ? { text: async () => state.objects.get(key) ?? '' } : null),
      delete: async (key: string) => state.objects.delete(key),
    },
  } as unknown as Env;
  return { env, state };
};

export const createCtx = () => ({ waitUntil: (promise: Promise<unknown>) => void promise });
