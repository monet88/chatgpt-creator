import { jsonError, jsonOk } from '../lib/http-response';
import { tombstoneMessage } from '../services/cleanup-service';
import { deleteMessageObjects } from '../services/message-storage';
import type { RouteContext } from '../lib/router';
import type { MessageRow } from '../types';

export const deleteMessage = async ({ env, ctx, params }: RouteContext) => {
  const row = await env.DB.prepare('SELECT * FROM messages WHERE id = ? AND domain = ? AND user = ? AND purged_at IS NULL')
    .bind(params.id, params.domain.toLowerCase(), params.user.toLowerCase())
    .first<MessageRow>();
  if (!row) return jsonOk({ deleted: 0 });
  await tombstoneMessage(env, params.id);
  ctx.waitUntil(deleteMessageObjects(env, row));
  return jsonOk({ deleted: 1 });
};
