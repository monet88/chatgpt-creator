import { getAppConfig } from '../config/app-config';
import type { Env } from '../types';
import { isValidDomain, isValidLocalPart, parseEmailAddress } from './email-address';

export const clampPagination = (request: Request, env: Env) => {
  const url = new URL(request.url);
  const page = Math.max(1, Number(url.searchParams.get('page') ?? '1') || 1);
  const requestedLimit = Math.max(1, Number(url.searchParams.get('limit') ?? '25') || 25);
  const maxLimit = getAppConfig(env).pageLimit;
  return { page, limit: Math.min(requestedLimit, maxLimit), offset: (page - 1) * Math.min(requestedLimit, maxLimit) };
};

export const readJson = async <T>(request: Request): Promise<T | null> => {
  try {
    return (await request.json()) as T;
  } catch {
    return null;
  }
};

export const validateMailboxParams = (domain: string, user: string) => {
  if (!isValidDomain(domain) || !isValidLocalPart(user)) return null;
  return { domain: domain.toLowerCase(), user: user.toLowerCase(), email: `${user.toLowerCase()}@${domain.toLowerCase()}` };
};

export const parseMailboxEmail = (email: string | null) => (email ? parseEmailAddress(email) : null);
