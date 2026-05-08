import { getAppConfig } from '../config/app-config';
import { parseEmailAddress } from '../lib/email-address';
import { isEnabledDomain } from '../services/domain-service';
import { isIssuedMailbox } from '../services/mailbox-service';
import { parseMimeMessage, rawToText } from '../services/message-parser';
import { createMessageId, deleteMessageObjects, storeMessageContent } from '../services/message-storage';
import { extractOtp } from '../services/otp-extractor';
import type { Env } from '../types';

export interface InboundEmailMessage {
  from: string;
  to: string;
  raw: ReadableStream<Uint8Array> | ArrayBuffer | string;
  rawSize?: number;
  setReject(reason: string): void;
}

export const handleEmail = async (message: InboundEmailMessage, env: Env) => {
  const recipient = parseEmailAddress(message.to);
  if (!recipient || !(await isEnabledDomain(env, recipient.domain))) {
    message.setReject('Unknown recipient domain');
    return;
  }
  if (!(await isIssuedMailbox(env, recipient.domain, recipient.user))) {
    message.setReject('Unknown mailbox');
    return;
  }

  const maxBytes = getAppConfig(env).maxMessageBytes;
  if ((message.rawSize ?? 0) > maxBytes) {
    message.setReject('Message too large');
    return;
  }

  const id = createMessageId();
  const rawText = await rawToText(message.raw);
  const rawBytes = new TextEncoder().encode(rawText).byteLength;
  if (rawBytes > maxBytes) {
    message.setReject('Message too large');
    return;
  }
  const parsed = await parseMimeMessage(rawText);
  const stored = await storeMessageContent(env, recipient.domain, recipient.user, id, rawText, parsed.text, parsed.html);
  const receivedAt = new Date().toISOString();
  const otp = extractOtp(parsed.subject, parsed.text, parsed.html);

  try {
    await env.DB.prepare(
      `INSERT INTO messages(id, domain, user, sender, recipient, subject, received_at, raw_key, text_key, html_key, otp, size)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    )
      .bind(
        id,
        recipient.domain,
        recipient.user,
        message.from,
        recipient.email,
        parsed.subject,
        receivedAt,
        stored.rawKey,
        stored.textKey,
        stored.htmlKey,
        otp,
        message.rawSize ?? rawBytes,
      )
      .run();
  } catch (error) {
    try {
      await deleteMessageObjects(env, { raw_key: stored.rawKey, text_key: stored.textKey, html_key: stored.htmlKey });
    } catch {
      // Preserve DB failure as primary cause; scheduled cleanup cannot find orphaned objects.
    }
    throw error;
  }
};
