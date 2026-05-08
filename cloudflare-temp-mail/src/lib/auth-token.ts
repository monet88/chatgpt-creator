import { API_PREFIX, getAppConfig } from '../config/app-config';
import type { Env } from '../types';

export const isApiRequest = (request: Request) => new URL(request.url).pathname.startsWith(API_PREFIX);

export const isAuthorized = (request: Request, env: Env) => {
  if (!isApiRequest(request)) return true;
  if (env.AUTH_DISABLED === 'true') return true;
  const token = getAppConfig(env).apiToken;
  if (!token) return false;
  const header = request.headers.get('authorization') ?? '';
  return header === `Bearer ${token}`;
};
