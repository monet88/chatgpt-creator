import type { Env, MessageRow } from '../types';

export interface StoredMessageContent {
  id: string;
  rawKey: string;
  textKey: string | null;
  htmlKey: string | null;
}

export const createMessageId = () => crypto.randomUUID();

export const buildMessageKeys = (domain: string, user: string, id: string) => ({
  rawKey: `emails/${domain}/${user}/${id}/raw.eml`,
  textKey: `emails/${domain}/${user}/${id}/body.txt`,
  htmlKey: `emails/${domain}/${user}/${id}/body.html`,
});

export const storeMessageContent = async (
  env: Env,
  domain: string,
  user: string,
  id: string,
  raw: string,
  text: string,
  html: string,
): Promise<StoredMessageContent> => {
  const keys = buildMessageKeys(domain, user, id);
  await env.MAIL_BUCKET.put(keys.rawKey, raw, { httpMetadata: { contentType: 'message/rfc822' } });
  const textKey = text ? keys.textKey : null;
  const htmlKey = html ? keys.htmlKey : null;
  if (textKey) await env.MAIL_BUCKET.put(textKey, text, { httpMetadata: { contentType: 'text/plain; charset=utf-8' } });
  if (htmlKey) await env.MAIL_BUCKET.put(htmlKey, html, { httpMetadata: { contentType: 'text/html; charset=utf-8' } });
  return { id, rawKey: keys.rawKey, textKey, htmlKey };
};

export const readObjectText = async (env: Env, key: string | null) => {
  if (!key) return '';
  return (await env.MAIL_BUCKET.get(key))?.text() ?? '';
};

export const deleteMessageObjects = async (env: Env, message: Pick<MessageRow, 'raw_key' | 'text_key' | 'html_key'>) => {
  const keys = [message.raw_key, message.text_key, message.html_key].filter((key): key is string => Boolean(key));
  await Promise.all(keys.map((key) => env.MAIL_BUCKET.delete(key)));
};

export const tryDeleteMessageObjects = async (env: Env, message: Pick<MessageRow, 'raw_key' | 'text_key' | 'html_key'>) => {
  try {
    await deleteMessageObjects(env, message);
    return true;
  } catch {
    return false;
  }
};
