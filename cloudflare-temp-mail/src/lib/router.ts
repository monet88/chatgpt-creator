import { API_PREFIX } from '../config/app-config';
import type { Env, ExecutionContextLike } from '../types';
import { jsonError } from './http-response';

export interface RouteContext {
  request: Request;
  env: Env;
  ctx: ExecutionContextLike;
  params: Record<string, string>;
}

type Handler = (context: RouteContext) => Promise<Response> | Response;

interface Route {
  method: string;
  pattern: RegExp;
  keys: string[];
  handler: Handler;
}

const compileRoute = (path: string) => {
  const keys: string[] = [];
  const source = `${API_PREFIX}${path}`.replace(/:[^/]+/g, (match) => {
    keys.push(match.slice(1));
    return '([^/]+)';
  });
  return { pattern: new RegExp(`^${source}/?$`), keys };
};

export class Router {
  private routes: Route[] = [];

  add(method: string, path: string, handler: Handler) {
    const route = compileRoute(path);
    this.routes = [...this.routes, { method, handler, ...route }];
    return this;
  }

  async handle(request: Request, env: Env, ctx: ExecutionContextLike) {
    const url = new URL(request.url);
    for (const route of this.routes) {
      const match = url.pathname.match(route.pattern);
      if (!match || route.method !== request.method) continue;
      const params = Object.fromEntries(route.keys.map((key, index) => [key, decodeURIComponent(match[index + 1])]));
      return route.handler({ request, env, ctx, params });
    }
    return jsonError(404, 'not_found', 'Route not found');
  }
}
