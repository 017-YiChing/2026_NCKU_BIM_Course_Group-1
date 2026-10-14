// 後端位址：Vite proxy 下用相對路徑即可；直連後端時改這裡。
export const API_BASE = ''

export async function listIssues(status = '', archived = '') {
  const p = new URLSearchParams()
  if (status) p.set('status', status)
  if (archived) p.set('archived', archived) // '1' 只看已封存，'all' 全看，預設只看未封存
  const q = p.toString() ? `?${p.toString()}` : ''
  const r = await fetch(`${API_BASE}/api/issues${q}`)
  if (!r.ok) throw new Error(await r.text())
  return r.json()
}

export async function getIssue(id) {
  const r = await fetch(`${API_BASE}/api/issues/${id}`)
  if (!r.ok) throw new Error(await r.text())
  return r.json()
}

export async function patchIssue(id, body) {
  const r = await fetch(`${API_BASE}/api/issues/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!r.ok) throw new Error(await r.text())
  return r.json()
}

export async function createIssue(formData) {
  const r = await fetch(`${API_BASE}/api/issues`, { method: 'POST', body: formData })
  if (!r.ok) throw new Error(await r.text())
  return r.json()
}
