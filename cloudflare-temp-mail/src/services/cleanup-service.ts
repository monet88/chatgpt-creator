import { getAppConfig } from '../config/app-config';
import { deleteMessageObjects, tryDeleteMessageObjects } from './message-storage';
import type { Env, MessageRow } from '../types';

export const tombstoneMailbox = async (env: Env, domain: string, user: string) => {
  const deletedAt = new Date().toISOString();
  await env.DB.prepare('UPDATE messages SET deleted_at = COALESCE(deleted_at, ?) WHERE domain = ? AND user = ? AND deleted_at IS NULL')
    .bind(deletedAt, domain, user)
    .run();
};

export const tombstoneMessage = async (env: Env, id: string) => {
  const deletedAt = new Date().toISOString();
  await env.DB.prepare('UPDATE messages SET deleted_at = COALESCE(deleted_at, ?) WHERE id = ?').bind(deletedAt, id).run();
};

export const purgeMessageObjects = async (env: Env, messages: Array<Pick<MessageRow, 'raw_key' | 'text_key' | 'html_key'>>) => {
  await Promise.all(messages.map((message) => deleteMessageObjects(env, message)));
};

export const runCleanup = async (env: Env) => {
  const config = getAppConfig(env);
  const cutoff = new Date(Date.now() - config.retentionDays * 24 * 60 * 60 * 1000).toISOString();
  const result = await env.DB.prepare(
    `SELECT * FROM messages
     WHERE purged_at IS NULL AND (deleted_at IS NOT NULL OR received_at < ?)
     ORDER BY received_at ASC
     LIMIT ?`,
  )
    .bind(cutoff, config.cleanupBatchSize)
    .all<MessageRow>();

  const messages = result.results;
  const purgedAt = new Date().toISOString();
  let purged = 0;
  for (const message of messages) {
    if (!(await tryDeleteMessageObjects(env, message))) continue;
    await env.DB.prepare('UPDATE messages SET deleted_at = COALESCE(deleted_at, ?), purged_at = ? WHERE id = ?')
      .bind(purgedAt, purgedAt, message.id)
      .run();
    purged += 1;
  }

  return { purged };
};
