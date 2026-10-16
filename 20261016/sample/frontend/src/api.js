// 後端位址：Vite proxy 下用相對路徑即可；直連後端時改這裡。
export const API_BASE = ''

let token = localStorage.getItem('rfi_token') || ''

export function setToken(t) {
  token = t
  if (t) localStorage.setItem('rfi_token', t)
  else localStorage.removeItem('rfi_token')
}

export function getToken() {
  return token
}

export function authHeaders() {
  const h = {}
  if (token) h['Authorization'] = 'Bearer ' + token
  return h
}

async function json(res) {
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || res.statusText)
  }
  return res.json()
}

export async function login(email, password) {
  const r = await fetch(`${API_BASE}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  const data = await json(r)
  setToken(data.token)
  return data
}

export async function register(email, password, name, role = 'client') {
  const r = await fetch(`${API_BASE}/api/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, name, role }),
  })
  return json(r)
}

export async function logout() {
  try {
    await fetch(`${API_BASE}/api/auth/logout`, {
      method: 'POST',
      headers: authHeaders(),
    })
  } catch (_) {}
  setToken('')
}

export async function me() {
  const r = await fetch(`${API_BASE}/api/auth/me`, { headers: authHeaders() })
  return json(r)
}

export async function listIssues({ status = '', archived = '', me = false } = {}) {
  const p = new URLSearchParams()
  if (status) p.set('status', status)
  if (archived) p.set('archived', archived)
  if (me) p.set('me', '1')
  const q = p.toString() ? `?${p.toString()}` : ''
  const r = await fetch(`${API_BASE}/api/issues${q}`, { headers: authHeaders() })
  return json(r)
}

export async function getIssue(id) {
  const r = await fetch(`${API_BASE}/api/issues/${id}`, { headers: authHeaders() })
  return json(r)
}

export async function patchIssue(id, body) {
  const r = await fetch(`${API_BASE}/api/issues/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(body),
  })
  return json(r)
}

export async function deleteIssue(id) {
  const r = await fetch(`${API_BASE}/api/issues/${id}`, {
    method: 'DELETE',
    headers: authHeaders(),
  })
  return json(r)
}

export async function createIssue(formData) {
  const r = await fetch(`${API_BASE}/api/issues`, {
    method: 'POST',
    headers: authHeaders(),
    body: formData,
  })
  return json(r)
}

export async function listUsers() {
  const r = await fetch(`${API_BASE}/api/users`, { headers: authHeaders() })
  return json(r)
}
