import { useEffect, useState } from 'react'
import { listIssues, getIssue, patchIssue, createIssue } from './api.js'

const STATUS = ['', 'open', 'in_progress', 'closed']
const STATUS_LABEL = { '': '全部', open: '待處理', in_progress: '處理中', closed: '已關閉' }

export default function App() {
  const [filter, setFilter] = useState('')
  const [issues, setIssues] = useState([])
  const [selectedId, setSelectedId] = useState(null)
  const [detail, setDetail] = useState(null)
  const [err, setErr] = useState('')

  async function refresh(f = filter) {
    try {
      setErr('')
      setIssues(await listIssues(f))
    } catch (e) {
      setErr('後端連不上（先 go run ./backend，再 npm run dev）：' + e.message)
    }
  }

  useEffect(() => { refresh(filter) }, [filter])

  useEffect(() => {
    if (selectedId == null) { setDetail(null); return }
    getIssue(selectedId).then(setDetail).catch((e) => setErr(String(e.message || e)))
  }, [selectedId])

  async function onPatch(status) {
    if (selectedId == null) return
    const updated = await patchIssue(selectedId, { status })
    setDetail(updated)
    refresh()
  }

  async function onCreate(e) {
    e.preventDefault()
    const fd = new FormData(e.target)
    try {
      const created = await createIssue(fd)
      e.target.reset()
      refresh()
      setSelectedId(created.id)
    } catch (ex) {
      setErr(String(ex.message || ex))
    }
  }

  return (
    <div style={{ display: 'flex', gap: 24, padding: 24, fontFamily: 'system-ui' }}>
      <div style={{ width: 340 }}>
        <h2>RFI 議題列表（sample）</h2>
        <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
          {STATUS.map((s) => (
            <button key={s} onClick={() => setFilter(s)}
              style={{ fontWeight: filter === s ? 'bold' : 'normal' }}>
              {STATUS_LABEL[s]}
            </button>
          ))}
          <button onClick={() => refresh()}>重整</button>
        </div>
        {err && <p style={{ color: 'crimson' }}>{err}</p>}
        <ul style={{ listStyle: 'none', padding: 0 }}>
          {issues.map((it) => (
            <li key={it.id} onClick={() => setSelectedId(it.id)}
              style={{ cursor: 'pointer', border: '1px solid #ccc', marginBottom: 8, padding: 8,
                background: selectedId === it.id ? '#eef' : '#fff' }}>
              <b>#{it.id} {it.title}</b>
              <div>{STATUS_LABEL[it.status]} · {it.source} · {it.reporter}</div>
              {it.view_name && <div>視圖：{it.view_name}</div>}
            </li>
          ))}
        </ul>
        <h3>新增議題（手填版，外掛接上後改由 Revit 送）</h3>
        <form onSubmit={onCreate} style={{ display: 'grid', gap: 8 }}>
          <input name="title" placeholder="標題（必填）" required />
          <textarea name="description" placeholder="說明" rows={3} />
          <input name="reporter" placeholder="回報人" />
          <input name="revit_element_id" placeholder="Revit ElementId（可空）" />
          <input name="ifc_global_id" placeholder="IFC GlobalId（可空）" />
          <label>截圖 <input type="file" name="screenshot" accept="image/*" /></label>
          <button type="submit">送出</button>
        </form>
      </div>
      <div style={{ flex: 1 }}>
        <h2>議題詳情</h2>
        {!detail && <p>點左邊一筆議題。</p>}
        {detail && (
          <div>
            <h3>#{detail.id} {detail.title}</h3>
            <p>{detail.description}</p>
            <p>狀態：{STATUS_LABEL[detail.status]} ｜ 回報：{detail.reporter} ｜ 指派：{detail.assignee || '—'}</p>
            <p>綁定：Revit ElementId = {detail.revit_element_id || '—'}；IFC GlobalId = {detail.ifc_global_id || '—'}</p>
            <p>視圖：{detail.view_name || '—（舊議題無視圖，重建才有）'}</p>
            {detail.screenshot && (
              <img src={`/screenshots/${detail.screenshot}`} alt="issue screenshot"
                style={{ maxWidth: '100%', border: '1px solid #ccc' }} />
            )}
            <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
              <button onClick={() => onPatch('open')}>退回 open</button>
              <button onClick={() => onPatch('in_progress')}>設 in_progress</button>
              <button onClick={() => onPatch('closed')}>設 closed</button>
            </div>
            <p style={{ color: '#666', marginTop: 16 }}>
              第一版無 IFC viewer（依課堂決議先只做列表+詳情）。
              IFC 上傳/定位端點已在後端留好（POST /api/models），viewer 為延伸。
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
