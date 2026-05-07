import { API_PREFIX, getAppConfig } from '../config/app-config';
import type { Env } from '../types';

export const isApiRequest = (request: Request) => new URL(request.url).pathname.startsWith(API_PREFIX);

export const isAuthorized = (request: Request, env: Env) => {
  const token = getAppConfig(env).apiToken;
  if (!token || !isApiRequest(request)) return true;
  const header = request.headers.get('authorization') ?? '';
  return header === `Bearer ${token}`;
};
