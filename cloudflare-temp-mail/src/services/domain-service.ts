import { getAppConfig } from '../config/app-config';
import type { Env } from '../types';

export interface DomainRow {
  domain: string;
}

export const listEnabledDomains = async (env: Env) => {
  const result = await env.DB.prepare('SELECT domain FROM domains WHERE enabled = 1 ORDER BY domain ASC').all<DomainRow>();
  const databaseDomains = result.results.map((row) => row.domain.toLowerCase());
  const configDomains = getAppConfig(env).enabledDomains;
  return [...new Set([...databaseDomains, ...configDomains])];
};

export const isEnabledDomain = async (env: Env, domain: string) => {
  const domains = await listEnabledDomains(env);
  return domains.includes(domain.toLowerCase());
};
