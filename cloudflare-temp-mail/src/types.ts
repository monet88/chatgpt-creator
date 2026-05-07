export interface Env {
  DB: D1Database;
  MAIL_BUCKET: R2Bucket;
  ENABLED_DOMAINS?: string;
  RETENTION_DAYS?: string;
  MAX_MESSAGE_BYTES?: string;
  PAGE_LIMIT?: string;
  CLEANUP_BATCH_SIZE?: string;
  API_TOKEN?: string;
}

export interface MessageRow {
  id: string;
  domain: string;
  user: string;
  sender: string;
  recipient: string;
  subject: string;
  received_at: string;
  raw_key: string;
  text_key: string | null;
  html_key: string | null;
  otp: string | null;
  size: number;
  deleted_at: string | null;
  purged_at: string | null;
}

export interface ExecutionContextLike {
  waitUntil(promise: Promise<unknown>): void;
}
