const LOCAL_PART_PATTERN = /^[a-z0-9._%+-]{1,64}$/i;
const DOMAIN_PATTERN = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)+$/i;

export const parseEmailAddress = (value: string) => {
  const trimmed = value.trim().toLowerCase();
  const at = trimmed.lastIndexOf('@');
  if (at <= 0 || at === trimmed.length - 1) return null;

  const user = trimmed.slice(0, at);
  const domain = trimmed.slice(at + 1);
  if (!LOCAL_PART_PATTERN.test(user) || !DOMAIN_PATTERN.test(domain)) return null;
  return { user, domain, email: `${user}@${domain}` };
};

export const isValidDomain = (value: string) => DOMAIN_PATTERN.test(value.trim().toLowerCase());
export const isValidLocalPart = (value: string) => LOCAL_PART_PATTERN.test(value.trim());
