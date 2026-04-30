const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

async function request<TResponse>(
  path: string,
  options: RequestInit & { token: string }
): Promise<TResponse> {
  const { token, ...fetchOptions } = options
  const response = await fetch(`${API_BASE}${path}`, {
    ...fetchOptions,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      ...(fetchOptions.headers || {})
    }
  })
  const data: unknown = await response.json().catch(() => ({}))
  if (!response.ok) {
    const message =
      typeof data === 'object' &&
      data !== null &&
      'error' in data &&
      typeof (data as { error?: unknown }).error === 'string'
        ? (data as { error: string }).error
        : 'Request failed'
    throw new Error(message)
  }
  return data as TResponse
}

export type ReportStatus = 'open' | 'in_review' | 'resolved' | 'rejected'
export type ModerationActionType =
  | 'assign'
  | 'note'
  | 'hide_listing'
  | 'warn_user'
  | 'ban_user'
  | 'close_report'

export interface ModerationAction {
  id: string
  report_id: string
  actor_user_id: string
  action_type: ModerationActionType
  action_payload?: Record<string, unknown>
  created_at: string
}

export interface ModerationReport {
  id: string
  reporter_user_id: string
  target_type: 'listing' | 'user' | 'message'
  target_listing_id?: string
  target_user_id?: string
  target_message_id?: string
  reason_code?: string
  reason_text?: string
  status: ReportStatus
  priority: number
  assigned_admin_id?: string
  resolution_note?: string
  resolved_at?: string
  created_at: string
  updated_at: string
  moderation_actions?: ModerationAction[]
}

export interface ReportsResponse {
  reports: ModerationReport[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export function listReports(
  token: string,
  params: { status?: ReportStatus | 'all'; page?: number; limit?: number } = {}
): Promise<ReportsResponse> {
  const sp = new URLSearchParams()
  if (params.status) sp.set('status', params.status)
  if (params.page) sp.set('page', String(params.page))
  if (params.limit) sp.set('limit', String(params.limit))
  const qs = sp.toString()
  return request<ReportsResponse>(`/api/v1/admin/reports${qs ? '?' + qs : ''}`, { token })
}

export function getReport(token: string, reportId: string): Promise<{ report: ModerationReport }> {
  return request<{ report: ModerationReport }>(`/api/v1/admin/reports/${reportId}`, { token })
}

export function updateReportStatus(
  token: string,
  reportId: string,
  payload: { status: ReportStatus; resolution_note?: string }
): Promise<{ report: ModerationReport }> {
  return request<{ report: ModerationReport }>(`/api/v1/admin/reports/${reportId}`, {
    method: 'PATCH',
    token,
    body: JSON.stringify(payload)
  })
}

export function addModerationAction(
  token: string,
  reportId: string,
  payload: { action_type: ModerationActionType; payload?: Record<string, unknown> }
): Promise<{ action: ModerationAction }> {
  return request<{ action: ModerationAction }>(`/api/v1/admin/reports/${reportId}/actions`, {
    method: 'POST',
    token,
    body: JSON.stringify(payload)
  })
}
