import { listEnabledDomains } from '../services/domain-service';
import { jsonOk } from '../lib/http-response';
import type { RouteContext } from '../lib/router';

export const listDomains = async ({ env }: RouteContext) => {
  const domains = await listEnabledDomains(env);
  return jsonOk({ domains });
};

export const listRandomDomains = async ({ request, env }: RouteContext) => {
  const domains = await listEnabledDomains(env);
  const requestedLimit = Number(new URL(request.url).searchParams.get('limit') ?? '10') || 10;
  const limit = Math.max(1, Math.min(requestedLimit, 25));
  const shuffled = domains.map((domain) => ({ domain, sort: crypto.getRandomValues(new Uint32Array(1))[0] })).sort((a, b) => a.sort - b.sort);
  return jsonOk({ domains: shuffled.slice(0, limit).map((item) => item.domain) });
};
