export const uiScript = `const state = { domains: [], email: '', domain: '', user: '', messages: [], otp: null };
const $ = (id) => document.getElementById(id);
let lastReadButton = null;
const toast = (message) => {
  const node = $('toast');
  node.textContent = message;
  node.classList.add('show');
  setTimeout(() => node.classList.remove('show'), 1800);
};
const showError = (message) => {
  $('inline-error').textContent = message;
  toast(message);
};
const clearError = () => { $('inline-error').textContent = ''; };
const setStatus = (status) => { $('status-badge').textContent = status; };
const api = async (path, options = {}) => {
  const response = await fetch('/api/v1' + path, { headers: { 'content-type': 'application/json' }, ...options });
  const body = await response.json();
  if (!body.success) throw new Error(body.error?.message || 'Request failed');
  return body.data;
};
const splitEmail = (email) => {
  const [user, domain] = email.split('@');
  return { user, domain };
};
const renderMessages = () => {
  $('message-count').textContent = state.messages.length + ' messages';
  $('inbox-body').replaceChildren(...(state.messages.length ? state.messages.map((message) => {
    const row = document.createElement('tr');
    const date = new Date(message.receivedAt).toLocaleString();
    row.innerHTML = '<td></td><td></td><td></td><td></td>';
    row.children[0].textContent = message.from;
    row.children[1].textContent = message.subject || '(no subject)';
    row.children[2].textContent = date;
    const read = document.createElement('button');
    read.type = 'button';
    read.textContent = 'Read';
    read.addEventListener('click', () => readMessage(message.id, read));
    row.children[3].append(read);
    return row;
  }) : [Object.assign(document.createElement('tr'), { innerHTML: '<td colspan="4">Inbox empty. Hit refresh after email arrives.</td>' })]));
};
const refreshOtp = async () => {
  if (!state.email) return;
  const data = await api('/email/' + state.domain + '/' + state.user + '/otp');
  state.otp = data.otp;
  $('copy-otp').textContent = data.otp || 'No OTP yet';
  setStatus(data.otp ? 'RECEIVED' : 'WAITING');
};
const refreshInbox = async () => {
  if (!state.email) return toast('Generate address first');
  setStatus('WAITING');
  const data = await api('/email/' + state.domain + '/' + state.user + '/messages');
  state.messages = data.messages;
  renderMessages();
  await refreshOtp();
};
const readMessage = async (id, trigger) => {
  try {
    clearError();
    lastReadButton = trigger;
    const message = await api('/email/' + state.domain + '/' + state.user + '/messages/' + id);
    $('modal-from').textContent = message.from + ' → ' + message.to;
    $('modal-subject').textContent = message.subject || '(no subject)';
    $('modal-body').textContent = message.body || message.html || '(empty message)';
    $('message-modal').showModal();
  } catch (error) {
    showError(error.message);
  }
};
const generateEmail = async () => {
  const domain = $('domain-select').value;
  const data = await api('/email/generate', { method: 'POST', body: JSON.stringify({ domain }) });
  Object.assign(state, { email: data.email, ...splitEmail(data.email), messages: [], otp: null });
  $('email-output').value = data.email;
  $('copy-otp').textContent = 'No OTP yet';
  renderMessages();
  setStatus('READY');
  toast('Address generated');
};
const deleteMailbox = async () => {
  if (!state.email) return;
  const data = await api('/email/' + state.domain + '/' + state.user, { method: 'DELETE' });
  state.messages = [];
  state.otp = null;
  renderMessages();
  $('copy-otp').textContent = 'No OTP yet';
  setStatus('READY');
  toast('Deleted ' + data.deleted + ' messages');
};
const copyText = async (value, label) => {
  if (!value) return;
  try {
    await navigator.clipboard.writeText(value);
    toast(label + ' copied');
  } catch {
    showError('Copy blocked. Select text and copy manually.');
  }
};
const init = async () => {
  const { domains } = await api('/domains');
  state.domains = domains;
  $('domain-select').replaceChildren(...domains.map((domain) => Object.assign(document.createElement('option'), { value: domain, textContent: domain })));
  $('generate-email').addEventListener('click', generateEmail);
  $('refresh-inbox').addEventListener('click', refreshInbox);
  $('delete-mailbox').addEventListener('click', deleteMailbox);
  $('copy-email').addEventListener('click', () => copyText(state.email, 'Email'));
  $('copy-otp').addEventListener('click', () => copyText(state.otp, 'OTP'));
  $('close-modal').addEventListener('click', () => $('message-modal').close());
  $('message-modal').addEventListener('close', () => lastReadButton?.focus());
};
init().catch((error) => toast(error.message));`;
