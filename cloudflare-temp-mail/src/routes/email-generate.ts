import { jsonError, jsonOk } from '../lib/http-response';
import { randomLocalPart } from '../lib/name-generator';
import { readJson } from '../lib/validation';
import { listEnabledDomains } from '../services/domain-service';
import { issueMailbox } from '../services/mailbox-service';
import type { RouteContext } from '../lib/router';

interface GenerateEmailBody {
  domain?: string;
}

export const generateEmail = async ({ request, env }: RouteContext) => {
  const body = request.method === 'POST' ? await readJson<GenerateEmailBody>(request) : null;
  const domains = await listEnabledDomains(env);
  const domain = (body?.domain ?? domains[0] ?? '').toLowerCase();
  if (!domain || !domains.includes(domain)) return jsonError(400, 'invalid_domain', 'Domain is not enabled');
  const user = randomLocalPart();
  return jsonOk(await issueMailbox(env, domain, user));
};
