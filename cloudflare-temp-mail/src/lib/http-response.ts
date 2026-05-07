export interface ApiEnvelope<T> {
  success: boolean;
  data: T | null;
  error: { code: string; message: string } | null;
  meta?: Record<string, unknown>;
}

const headers = { 'content-type': 'application/json; charset=utf-8' };

export const jsonOk = <T>(data: T, init: ResponseInit = {}, meta?: Record<string, unknown>) => {
  const body: ApiEnvelope<T> = { success: true, data, error: null, ...(meta ? { meta } : {}) };
  return new Response(JSON.stringify(body), { status: 200, ...init, headers });
};

export const jsonError = (status: number, code: string, message: string, init: ResponseInit = {}) => {
  const body: ApiEnvelope<never> = { success: false, data: null, error: { code, message } };
  return new Response(JSON.stringify(body), { status, ...init, headers: { ...headers, ...init.headers } });
};

export const textResponse = (body: string, contentType: string) =>
  new Response(body, { headers: { 'content-type': contentType } });
