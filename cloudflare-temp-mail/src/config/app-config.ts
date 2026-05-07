import type { Env } from '../types';

export const API_PREFIX = '/api/v1';
export const DEFAULT_RETENTION_DAYS = 3;
export const DEFAULT_MAX_MESSAGE_BYTES = 1024 * 1024;
export const DEFAULT_PAGE_LIMIT = 50;
export const DEFAULT_CLEANUP_BATCH_SIZE = 100;

const parsePositiveInt = (value: string | undefined, fallback: number) => {
  const parsed = Number(value);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
};

export const getAppConfig = (env: Env) => ({
  enabledDomains: (env.ENABLED_DOMAINS ?? '')
    .split(',')
    .map((domain) => domain.trim().toLowerCase())
    .filter(Boolean),
  retentionDays: parsePositiveInt(env.RETENTION_DAYS, DEFAULT_RETENTION_DAYS),
  maxMessageBytes: parsePositiveInt(env.MAX_MESSAGE_BYTES, DEFAULT_MAX_MESSAGE_BYTES),
  pageLimit: parsePositiveInt(env.PAGE_LIMIT, DEFAULT_PAGE_LIMIT),
  cleanupBatchSize: parsePositiveInt(env.CLEANUP_BATCH_SIZE, DEFAULT_CLEANUP_BATCH_SIZE),
  apiToken: env.API_TOKEN?.trim() || '',
});
