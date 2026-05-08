import { jsonError, jsonOk } from '../lib/http-response';
import { validateMailboxParams, clampPagination } from '../lib/validation';
import { revokeMailbox } from '../services/mailbox-service';
import { readObjectText } from '../services/message-storage';
import type { RouteContext } from '../lib/router';
import type { MessageRow } from '../types';

const toMessageSummary = (row: MessageRow) => ({
  id: row.id,
  from: row.sender,
  to: row.recipient,
  subject: row.subject,
  receivedAt: row.received_at,
  otp: row.otp,
  size: row.size,
});

export const listMailbox = async ({ request, env, params }: RouteContext) => {
  const mailbox = validateMailboxParams(params.domain, params.user);
  if (!mailbox) return jsonError(400, 'invalid_mailbox', 'Mailbox is invalid');
  const { page, limit, offset } = clampPagination(request, env);
  const result = await env.DB.prepare(
    `SELECT * FROM messages
     WHERE domain = ? AND user = ? AND deleted_at IS NULL AND purged_at IS NULL
     ORDER BY received_at DESC LIMIT ? OFFSET ?`,
  )
    .bind(mailbox.domain, mailbox.user, limit, offset)
    .all<MessageRow>();
  return jsonOk({ messages: result.results.map(toMessageSummary) }, {}, { page, limit });
};

export const readMailboxMessage = async ({ env, params }: RouteContext) => {
  const mailbox = validateMailboxParams(params.domain, params.user);
  if (!mailbox) return jsonError(400, 'invalid_mailbox', 'Mailbox is invalid');
  const row = await env.DB.prepare(
    `SELECT * FROM messages
     WHERE id = ? AND domain = ? AND user = ? AND deleted_at IS NULL AND purged_at IS NULL`,
  )
    .bind(params.id, mailbox.domain, mailbox.user)
    .first<MessageRow>();
  if (!row) return jsonError(404, 'message_not_found', 'Message not found');
  const [text, html] = await Promise.all([readObjectText(env, row.text_key), readObjectText(env, row.html_key)]);
  return jsonOk({ ...toMessageSummary(row), body: text, html });
};

export const deleteMailbox = async ({ env, params }: RouteContext) => {
  const mailbox = validateMailboxParams(params.domain, params.user);
  if (!mailbox) return jsonError(400, 'invalid_mailbox', 'Mailbox is invalid');
  const deletedAt = new Date().toISOString();
  const result = await env.DB.prepare(
    'UPDATE messages SET deleted_at = COALESCE(deleted_at, ?) WHERE domain = ? AND user = ? AND deleted_at IS NULL AND purged_at IS NULL',
  )
    .bind(deletedAt, mailbox.domain, mailbox.user)
    .run();
  await revokeMailbox(env, mailbox.domain, mailbox.user);
  return jsonOk({ deleted: result.meta?.changes ?? 0 });
};
