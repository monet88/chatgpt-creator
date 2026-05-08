import PostalMime from 'postal-mime';

export interface ParsedMessage {
  subject: string;
  text: string;
  html: string;
}

const decodeRawMessage = async (raw: ReadableStream<Uint8Array> | ArrayBuffer | string) => {
  if (typeof raw === 'string') return raw;
  if (raw instanceof ArrayBuffer) return new TextDecoder().decode(raw);
  return new Response(raw).text();
};

export const parseMimeMessage = async (raw: ReadableStream<Uint8Array> | ArrayBuffer | string): Promise<ParsedMessage> => {
  const rawText = await decodeRawMessage(raw);
  try {
    const parsed = await new PostalMime().parse(rawText);
    return {
      subject: parsed.subject ?? '',
      text: parsed.text ?? '',
      html: parsed.html ?? '',
    };
  } catch {
    return { subject: '', text: rawText, html: '' };
  }
};

export const rawToText = decodeRawMessage;
