import { getAppConfig } from '../config/app-config';
import type { Env } from '../types';

interface Bucket {
  count: number;
  resetAt: number;
}

const MAX_BUCKETS = 4096;
const buckets = new Map<string, Bucket>();

const clientKey = (request: Request) => request.headers.get('cf-connecting-ip') ?? 'local';

const pruneBuckets = (now: number) => {
  for (const [key, bucket] of buckets) {
    if (bucket.resetAt <= now || buckets.size > MAX_BUCKETS) buckets.delete(key);
  }
};

export const checkRateLimit = (request: Request, env: Env, now = Date.now()) => {
  const config = getAppConfig(env);
  const key = clientKey(request);
  const current = buckets.get(key);
  if (!current || current.resetAt <= now) {
    pruneBuckets(now);
    buckets.set(key, { count: 1, resetAt: now + config.rateLimitWindowSeconds * 1000 });
    return { allowed: true, retryAfterSeconds: 0 };
  }
  if (current.count >= config.rateLimitMaxRequests) {
    return { allowed: false, retryAfterSeconds: Math.ceil((current.resetAt - now) / 1000) };
  }
  buckets.set(key, { ...current, count: current.count + 1 });
  return { allowed: true, retryAfterSeconds: 0 };
};

export const resetRateLimitForTests = () => buckets.clear();
