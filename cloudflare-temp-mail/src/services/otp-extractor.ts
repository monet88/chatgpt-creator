const OTP_PATTERN = /(?<!\d)(\d{4,8})(?!\d)/;

export const extractOtp = (...values: Array<string | undefined | null>) => {
  const joined = values.filter(Boolean).join(' ');
  return joined.match(OTP_PATTERN)?.[1] ?? null;
};
