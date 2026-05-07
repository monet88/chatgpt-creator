import { jsonError, jsonOk } from '../lib/http-response';
import { validateMailboxParams } from '../lib/validation';
import type { RouteContext } from '../lib/router';

interface OtpRow {
  otp: string | null;
  received_at: string;
}

export const getLatestOtp = async ({ env, params }: RouteContext) => {
  const mailbox = validateMailboxParams(params.domain, params.user);
  if (!mailbox) return jsonError(400, 'invalid_mailbox', 'Mailbox is invalid');
  const row = await env.DB.prepare(
    `SELECT otp, received_at FROM messages
     WHERE domain = ? AND user = ? AND otp IS NOT NULL AND deleted_at IS NULL AND purged_at IS NULL
     ORDER BY received_at DESC LIMIT 1`,
  )
    .bind(mailbox.domain, mailbox.user)
    .first<OtpRow>();
  return jsonOk({ email: mailbox.email, otp: row?.otp ?? null, status: row?.otp ? 'received' : 'pending', receivedAt: row?.received_at ?? null });
};
