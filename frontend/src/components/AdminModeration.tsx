import { useEffect, useState } from 'react'
import {
  addModerationAction,
  getReport,
  listReports,
  updateReportStatus
} from '../api/admin'
import type { ModerationReport, ReportStatus } from '../api/admin'
import { IconBack } from './Icons'

interface AdminModerationProps {
  token: string
  onBack: () => void
}

const PAGE_SIZE = 10
const statuses: (ReportStatus | 'all')[] = ['open', 'in_review', 'resolved', 'rejected', 'all']

export default function AdminModeration({ token, onBack }: AdminModerationProps) {
  const [status, setStatus] = useState<ReportStatus | 'all'>('open')
  const [reports, setReports] = useState<ModerationReport[]>([])
  const [selected, setSelected] = useState<ModerationReport | null>(null)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [note, setNote] = useState('')
  const [actionLoading, setActionLoading] = useState('')

  async function loadReports(nextStatus = status, nextPage = page) {
    setLoading(true)
    setError('')
    try {
      const res = await listReports(token, { status: nextStatus, page: nextPage, limit: PAGE_SIZE })
      setReports(res.reports)
      setPage(res.page)
      setTotalPages(res.total_pages)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load reports')
      setReports([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadReports(status, page)
  }, [token, status, page])

  async function refreshSelected(reportId: string) {
    const res = await getReport(token, reportId)
    setSelected(res.report)
  }

  async function handleStatus(nextStatus: ReportStatus) {
    if (!selected) return
    setActionLoading(nextStatus)
    setError('')
    try {
      const res = await updateReportStatus(token, selected.id, {
        status: nextStatus,
        resolution_note: note.trim() || undefined
      })
      setSelected(res.report)
      setNote('')
      await loadReports(status, page)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to update report')
    } finally {
      setActionLoading('')
    }
  }

  async function handleAddNote() {
    if (!selected || !note.trim()) return
    setActionLoading('note')
    setError('')
    try {
      await addModerationAction(token, selected.id, {
        action_type: 'note',
        payload: { note: note.trim() }
      })
      setNote('')
      await refreshSelected(selected.id)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to add note')
    } finally {
      setActionLoading('')
    }
  }

  return (
    <div className="admin-moderation">
      <button type="button" className="back-link-btn" onClick={onBack}>
        <IconBack className="back-link-icon" aria-hidden />
        <span>Back to Profile</span>
      </button>
      <h2>Moderation</h2>

      <div className="my-listings-tabs">
        {statuses.map((item) => (
          <button
            key={item}
            type="button"
            className={`my-listings-tab ${status === item ? 'active' : ''}`}
            onClick={() => {
              setStatus(item)
              setPage(1)
              setSelected(null)
            }}
          >
            {item.replace('_', ' ')}
          </button>
        ))}
      </div>

      {error && <p className="error-message">{error}</p>}
      {loading ? (
        <p className="my-listings-loading">Loading...</p>
      ) : (
        <div className="admin-moderation-layout">
          <ul className="admin-report-list">
            {reports.map((report) => (
              <li key={report.id}>
                <button
                  type="button"
                  className={`admin-report-row ${selected?.id === report.id ? 'active' : ''}`}
                  onClick={() => {
                    setSelected(report)
                    setNote('')
                  }}
                >
                  <span>{report.target_type}</span>
                  <strong>{report.status.replace('_', ' ')}</strong>
                  <small>{new Date(report.created_at).toLocaleString()}</small>
                </button>
              </li>
            ))}
            {reports.length === 0 && <li className="admin-empty">No reports in this queue.</li>}
          </ul>

          <section className="admin-report-detail">
            {selected ? (
              <>
                <h3>Report {selected.id.slice(0, 8)}</h3>
                <p><strong>Target:</strong> {selected.target_type}</p>
                <p><strong>Reason:</strong> {selected.reason_text || selected.reason_code || 'No reason provided'}</p>
                <p><strong>Reporter:</strong> {selected.reporter_user_id}</p>
                <label>
                  Moderation note
                  <textarea
                    value={note}
                    onChange={(event) => setNote(event.target.value)}
                    rows={4}
                    placeholder="Add a note or resolution reason"
                  />
                </label>
                <div className="button-group">
                  <button type="button" className="secondary" onClick={handleAddNote} disabled={!note.trim() || !!actionLoading}>
                    {actionLoading === 'note' ? 'Adding...' : 'Add note'}
                  </button>
                  <button type="button" onClick={() => void handleStatus('in_review')} disabled={!!actionLoading}>
                    Review
                  </button>
                  <button type="button" onClick={() => void handleStatus('resolved')} disabled={!!actionLoading}>
                    Resolve
                  </button>
                  <button type="button" className="secondary" onClick={() => void handleStatus('rejected')} disabled={!!actionLoading}>
                    Reject
                  </button>
                </div>

                <h4>Action log</h4>
                <ul className="admin-action-log">
                  {(selected.moderation_actions || []).map((action) => (
                    <li key={action.id}>
                      <strong>{action.action_type}</strong>
                      <span>{new Date(action.created_at).toLocaleString()}</span>
                    </li>
                  ))}
                  {(!selected.moderation_actions || selected.moderation_actions.length === 0) && (
                    <li>No actions yet.</li>
                  )}
                </ul>
              </>
            ) : (
              <p>Select a report to review.</p>
            )}
          </section>
        </div>
      )}

      {totalPages > 1 && (
        <div className="my-listings-pagination">
          <button type="button" className="my-listings-page-btn" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            Previous
          </button>
          <span className="my-listings-page-info">Page {page} of {totalPages}</span>
          <button type="button" className="my-listings-page-btn" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            Next
          </button>
        </div>
      )}
    </div>
  )
}
