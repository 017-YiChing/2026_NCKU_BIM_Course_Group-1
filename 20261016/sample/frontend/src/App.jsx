import { useEffect, useState } from 'react'
import {
  login, register, logout, me, setToken, getToken,
  listIssues, getIssue, patchIssue, deleteIssue, createIssue, listUsers,
} from './api.js'

const STATUS = ['open', 'in_progress', 'closed']
const STATUS_LABEL = { open: '待處理', in_progress: '處理中', closed: '已關閉' }
const BADGE_TEXT = { published: '已發布', draft: '草稿', overdue: '逾期', open: '待處理', in_progress: '處理中', closed: '已關閉' }

export default function App() {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  const [issues, setIssues] = useState([])
  const [selectedId, setSelectedId] = useState(null)
  const [detail, setDetail] = useState(null)
  const [err, setErr] = useState('')
  const [filter, setFilter] = useState('')
  const [viewMy, setViewMy] = useState(false)
  const [showArchived, setShowArchived] = useState(false)
  const [users, setUsers] = useState([])
  const [showCreate, setShowCreate] = useState(false)

  useEffect(() => {
    const t = getToken()
    if (t) {
      me().then(setUser).catch(() => setToken('')).finally(() => setLoading(false))
    } else {
      setLoading(false)
    }
  }, [])

  async function refresh() {
    try {
      setErr('')
      setIssues(await listIssues({ status: filter, archived: showArchived ? 'all' : '', me: viewMy }))
    } catch (e) {
      setErr('後端連不上：' + e.message)
    }
  }

  useEffect(() => { if (!loading) refresh() }, [loading, filter, viewMy, showArchived])

  useEffect(() => {
    if (selectedId == null) { setDetail(null); return }
    getIssue(selectedId).then(setDetail).catch((e) => setErr(String(e.message || e)))
  }, [selectedId])

  async function onLogin(email, password) {
    try {
      const data = await login(email, password)
      setUser(data.user)
      refresh()
    } catch (e) {
      setErr('登入失敗：' + e.message)
    }
  }

  async function onRegister(email, password, name, role) {
    try {
      await register(email, password, name, role)
      setErr('註冊成功，請登入')
    } catch (e) {
      setErr('註冊失敗：' + e.message)
    }
  }

  async function onLogout() {
    await logout()
    setUser(null)
    setDetail(null)
    setSelectedId(null)
    refresh()
  }

  async function onPatch(updates) {
    if (selectedId == null) return
    try {
      const updated = await patchIssue(selectedId, updates)
      setDetail(updated)
      refresh()
    } catch (e) {
      setErr('更新失敗：' + e.message)
    }
  }

  async function onCreate(formData) {
    try {
      const created = await createIssue(formData)
      setSelectedId(created.id)
      setShowCreate(false)
      refresh()
    } catch (e) {
      setErr('建立失敗：' + e.message)
    }
  }

  async function onDelete() {
    if (!window.confirm('確定要刪除這筆議題嗎？')) return
    try {
      await deleteIssue(selectedId)
      setSelectedId(null)
      setDetail(null)
      refresh()
    } catch (e) {
      setErr('刪除失敗：' + e.message)
    }
  }

  if (loading) {
    return (
      <div className="rfi-login">
        <div className="rfi-login-card">
          <h2>RFI 議題追蹤</h2>
          <p className="hint">Ocean Waves · 多使用者 Demo</p>
          <div className="rfi-field">
            <label>電子郵件</label>
            <input id="loginEmail" />
          </div>
          <div className="rfi-field">
            <label>密碼</label>
            <input id="loginPassword" type="password" />
          </div>
          <div className="rfi-roles">
            <div className="rfi-role-chip selected" onClick={() => { window.__role = 'owner' }}>負責人</div>
            <div className="rfi-role-chip" onClick={() => { window.__role = 'client' }}>客戶</div>
          </div>
          <button className="btn-primary" onClick={async () => {
            const email = document.getElementById('loginEmail').value
            const password = document.getElementById('loginPassword').value
            const role = window.__role || 'client'
            await onLogin(email, password)
          }}>登入</button>
          <p style={{ fontSize: 12, color: '#6b7a85', marginTop: 16 }}>
            預設帳號：james@example.com / bob@example.com / sally@example.com，密碼皆為 password123
          </p>
          <hr style={{ border: 'none', borderTop: '1px solid #d7dde2', margin: '24px 0' }} />
          <div style={{ textAlign: 'left' }}>
            <label style={{ display: 'block', marginBottom: 6, fontSize: 13, color: '#6b7a85' }}>新註冊</label>
            <input id="regEmail" style={{ width: '100%', padding: 8, border: '1px solid #d7dde2', borderRadius: 6, marginBottom: 8 }} placeholder="Email" />
            <input id="regPassword" type="password" style={{ width: '100%', padding: 8, border: '1px solid #d7dde2', borderRadius: 6, marginBottom: 8 }} placeholder="密碼" />
            <input id="regName" style={{ width: '100%', padding: 8, border: '1px solid #d7dde2', borderRadius: 6, marginBottom: 8 }} placeholder="姓名" />
            <select id="regRole" style={{ width: '100%', padding: 8, border: '1px solid #d7dde2', borderRadius: 6, marginBottom: 16 }}>
              <option value="client">客戶</option>
              <option value="owner">負責人</option>
            </select>
            <button className="btn-camel" onClick={async () => {
              const el = (id) => document.getElementById(id)
              await onRegister(el('regEmail').value, el('regPassword').value, el('regName').value, el('regRole').value)
            }}>建立帳號</button>
          </div>
          {err && <p style={{ color: '#c0392b', marginTop: 16, fontSize: 13 }}>{err}</p>}
        </div>
      </div>
    )
  }

  const roleLabel = (r) => r === 'owner' ? '負責人' : r === 'client' ? '客戶' : r
  const avatarClass = (r) => 'rfi-avatar' + (r === 'owner' ? ' owner' : '')

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      {/* Topbar */}
      <header className="rfi-topbar">
        <div className="rfi-brand">
          <div className="rfi-logo">R</div>
          <div>RFI 議題追蹤圖台</div>
        </div>
        <ul className="rfi-nav">
          <li><a href="#" style={{ color: '#b8c9d0', textDecoration: 'none' }}>RFI 請求</a></li>
          <li><a href="#" style={{ color: '#b8c9d0', textDecoration: 'none' }}>行事曆</a></li>
          <li><a href="#" style={{ color: '#b8c9d0', textDecoration: 'none' }}>客戶</a></li>
          <li><a href="#" style={{ color: '#b8c9d0', textDecoration: 'none' }}>團隊</a></li>
        </ul>
        <div className="rfi-topbar-right">
          <div className="rfi-avatar">{user.name.charAt(0)}</div>
          <div className="rfi-user-info">{user.name} · {roleLabel(user.role)}</div>
          <button className="btn-logout" onClick={onLogout}>登出</button>
        </div>
      </header>

      {/* Toolbar */}
      <div className="rfi-toolbar">
        <button className={filter ? 'rfi-pill' : 'rfi-pill active'} onClick={() => setFilter('')}>全部</button>
        {STATUS.map(s => (
          <button key={s} className={filter === s ? 'rfi-pill active' : 'rfi-pill'} onClick={() => setFilter(s)}>
            {STATUS_LABEL[s]}
          </button>
        ))}
        <button className={viewMy ? 'rfi-pill active camel' : 'rfi-pill'} onClick={() => setViewMy(!viewMy)}>
          {viewMy ? '取消「我的議題」' : '我的議題'}
        </button>
        <button className={showArchived ? 'rfi-pill active' : 'rfi-pill'} onClick={() => setShowArchived(!showArchived)}>
          {showArchived ? '隱藏封存' : '顯示封存'}
        </button>
        <div className="rfi-divider"></div>
        <div className="rfi-search"><input id="searchInput" placeholder="搜尋標題…" /></div>
        <div className="rfi-toolbar-actions">
          <button className="btn-refresh" onClick={refresh}>重整</button>
          <button className="btn-add" onClick={() => setShowCreate(true)}>新增議題</button>
        </div>
      </div>
      {err && <div className="rfi-err">{err}</div>}

      {/* Grid */}
      <main className="rfi-main">
        <div className="rfi-grid">
          {issues.map(it => (
            <div
              key={it.id}
              className={`rfi-card ${it.id === 1 ? 'highlight' : ''} ${user?.role === 'owner' ? '' : ''}`}
              onClick={() => setSelectedId(it.id)}
            >
              <div className="rfi-card-header">
                <div className={avatarClass(it.reporterRole)}>{it.reporter ? it.reporter.charAt(0) : '?'}</div>
                <div className="rfi-user-info">
                  <div className="rfi-name">{it.reporter || '—'}</div>
                  <div className="rfi-role">{roleLabel(it.reporterRole)}</div>
                </div>
                <div className="rfi-card-actions">{it.source}</div>
              </div>
              <div className="rfi-card-body">
                <div className="rfi-title">{it.title}</div>
                <div className="rfi-due">到期：{it.due || '未設定'}</div>
                <div className="rfi-progress-wrap">
                  <div className="rfi-progress-label">
                    <span>{Math.round((it.progress || 0))}%</span>
                  </div>
                  <div className="rfi-progress-track">
                    <div className="rfi-progress-fill" style={{ width: `${Math.min(Math.max(it.progress || 0, 0), 100)}%` }}></div>
                  </div>
                </div>
                <div className="rfi-stats">
                  <div className="rfi-stat"><span className="rfi-stat-num">{it.stats?.approved ?? 0}</span><span className="rfi-stat-lbl">核准</span></div>
                  <div className="rfi-stat"><span className="rfi-stat-num">{it.stats?.completed ?? 0}</span><span className="rfi-stat-lbl">已完成</span></div>
                  <div className="rfi-stat"><span className="rfi-stat-num">{it.stats?.todo ?? 0}</span><span className="rfi-stat-lbl">待辦</span></div>
                </div>
              </div>
              <div className="rfi-card-footer">
                <span className={`rfi-badge ${it.status}`}>{BADGE_TEXT[it.status] || it.status}</span>
                <button className="btn-review" onClick={(e) => { e.stopPropagation(); setSelectedId(it.id) }}>檢視</button>
              </div>
            </div>
          ))}
          <div className="rfi-card empty" onClick={() => setShowCreate(true)}>
            <div style={{ fontSize: 40, marginBottom: 8 }}>📄</div>
            <div style={{ fontWeight: 700 }}>新增議題</div>
          </div>
        </div>
      </main>

      {/* Create modal */}
      {showCreate && (
        <div className="rfi-modal-backdrop">
          <div className="rfi-modal create">
            <div className="rfi-modal-header">
              <div style={{ fontWeight: 700 }}>新增 RFI 議題</div>
              <button className="rfi-btn-close" onClick={() => setShowCreate(false)}>✕</button>
            </div>
            <div className="rfi-modal-body">
              <form className="rfi-form" onSubmit={(e) => { e.preventDefault(); onCreate(new FormData(e.target)) }}>
                <label>標題（必填）</label>
                <input name="title" required />
                <label>說明</label>
                <textarea name="description" rows={3} />
                <label>指派人</label>
                <select name="assignee">
                  <option value="">— 未指派 —</option>
                  {users.map(u => (
                    <option key={u.id} value={u.id}>{u.name} ({roleLabel(u.role)})</option>
                  ))}
                </select>
                <label>截圖</label>
                <input type="file" name="screenshot" accept="image/*" />
                <div className="rfi-form-row">
                  <button type="button" className="rfi-btn-cancel" onClick={() => setShowCreate(false)}>取消</button>
                  <button type="submit" className="rfi-btn-submit">送出</button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Detail modal */}
      {detail && (
        <div className="rfi-modal-backdrop">
          <div className="rfi-modal detail">
            <div className="rfi-modal-header">
              <div style={{ fontWeight: 700 }}>{detail.title}</div>
              <button className="rfi-btn-close" onClick={() => setSelectedId(null)}>✕</button>
            </div>
            <div className="rfi-modal-body rfi-detail">
              <p><strong>建立者：</strong>{detail.reporter || '—'} ({roleLabel(detail.reporterRole)})</p>
              <p><strong>指派人：</strong>{detail.assignee || '—'}</p>
              <p><strong>到期日：</strong>{detail.due || '未設定'}</p>
              <p><strong>狀態：</strong><span className={`rfi-badge ${detail.status}`}>{BADGE_TEXT[detail.status] || detail.status}</span></p>
              <p><strong>說明：</strong>{detail.description}</p>
              {detail.screenshot && <img src={`/screenshots/${detail.screenshot}`} alt="screenshot" />}
              <hr />
              <p style={{ fontSize: 12, color: '#6b7a85' }}>
                {detail.revit_element_id ? `Revit ElementId：${detail.revit_element_id}` : ''}
                {detail.ifc_global_id ? ` · IFC：${detail.ifc_global_id}` : ''}
              </p>
              <p style={{ fontSize: 12, color: '#6b7a85' }}>{detail.view_name ? `視圖：${detail.view_name}` : ''}</p>
            </div>
            <div className="rfi-modal-footer">
              <button className="rfi-btn-modal" onClick={() => onPatch({ status: 'open' })}>待處理</button>
              <button className="rfi-btn-modal" onClick={() => onPatch({ status: 'in_progress' })}>處理中</button>
              <button className="rfi-btn-modal" onClick={() => onPatch({ status: 'closed' })}>已關閉</button>
              <button className="rfi-btn-modal" onClick={() => onPatch({ archived: !detail.archived })}>{detail.archived ? '取消封存' : '封存'}</button>
              {canDelete(detail, user) && <button className="rfi-btn-delete" onClick={onDelete}>刪除</button>}
              <button className="rfi-btn-modal" onClick={() => setSelectedId(null)}>關閉</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )

  function canDelete(it, u) {
    if (!u) return false
    if (u.role === 'owner' || u.role === 'admin') return true
    return it.reporter_id && it.reporter_id === user.id
  }
}
