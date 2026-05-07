import type { Env } from '../types';

export const issueMailbox = async (env: Env, domain: string, user: string) => {
  const email = `${user}@${domain}`;
  const issuedAt = new Date().toISOString();
  await env.DB.prepare('INSERT INTO mailboxes(domain, user, email, issued_at) VALUES (?, ?, ?, ?)')
    .bind(domain, user, email, issuedAt)
    .run();
  return { email, user, domain };
};

export const revokeMailbox = async (env: Env, domain: string, user: string) => {
  await env.DB.prepare('DELETE FROM mailboxes WHERE domain = ? AND user = ?').bind(domain, user).run();
};

export const isIssuedMailbox = async (env: Env, domain: string, user: string) => {
  const row = await env.DB.prepare('SELECT 1 FROM mailboxes WHERE domain = ? AND user = ? LIMIT 1').bind(domain, user).first();
  return Boolean(row);
};
