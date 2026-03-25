package observability

import "net/http"

func DashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(observabilityDashboardHTML))
	}
}

const observabilityDashboardHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>ReSellution Observability</title>
  <style>
    :root {
      --bg: #f3efe7;
      --panel: rgba(255, 255, 255, 0.88);
      --panel-border: rgba(27, 42, 65, 0.12);
      --text: #1f2933;
      --muted: #52606d;
      --accent: #0f766e;
      --accent-soft: rgba(15, 118, 110, 0.12);
      --danger: #b42318;
      --shadow: 0 20px 50px rgba(15, 23, 42, 0.08);
    }

    * { box-sizing: border-box; }

    body {
      margin: 0;
      min-height: 100vh;
      font-family: "Avenir Next", "Segoe UI", sans-serif;
      color: var(--text);
      background:
        radial-gradient(circle at top left, rgba(15, 118, 110, 0.16), transparent 34%),
        radial-gradient(circle at bottom right, rgba(180, 35, 24, 0.12), transparent 28%),
        linear-gradient(135deg, #f7f3ec, #eef6f5 55%, #fffaf5);
    }

    .shell {
      width: min(1180px, calc(100% - 32px));
      margin: 32px auto 48px;
    }

    .hero, .panel {
      background: var(--panel);
      border: 1px solid var(--panel-border);
      border-radius: 24px;
      box-shadow: var(--shadow);
      backdrop-filter: blur(12px);
    }

    .hero {
      padding: 28px;
      margin-bottom: 20px;
    }

    .eyebrow {
      color: var(--accent);
      text-transform: uppercase;
      letter-spacing: 0.14em;
      font-size: 12px;
      margin-bottom: 12px;
      font-weight: 700;
    }

    h1 {
      margin: 0 0 10px;
      font-size: clamp(32px, 5vw, 52px);
      line-height: 0.96;
    }

    .subhead {
      margin: 0;
      max-width: 760px;
      color: var(--muted);
      line-height: 1.6;
    }

    .meta {
      margin-top: 16px;
      color: var(--muted);
      font-size: 14px;
    }

    .cards {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 14px;
      margin-bottom: 20px;
    }

    .card {
      padding: 18px 18px 20px;
      border-radius: 20px;
      background: var(--panel);
      border: 1px solid var(--panel-border);
      box-shadow: var(--shadow);
    }

    .card-label {
      color: var(--muted);
      font-size: 13px;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      margin-bottom: 10px;
    }

    .card-value {
      font-size: 32px;
      font-weight: 700;
      line-height: 1;
    }

    .grid {
      display: grid;
      grid-template-columns: 1.25fr 0.75fr;
      gap: 20px;
    }

    .panel {
      padding: 20px;
    }

    .panel h2 {
      margin: 0 0 14px;
      font-size: 20px;
    }

    table {
      width: 100%;
      border-collapse: collapse;
    }

    th, td {
      padding: 10px 0;
      border-bottom: 1px solid rgba(82, 96, 109, 0.14);
      text-align: left;
      font-size: 14px;
      vertical-align: top;
    }

    th {
      color: var(--muted);
      font-weight: 600;
    }

    .badge {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      min-width: 44px;
      padding: 4px 10px;
      border-radius: 999px;
      font-weight: 700;
      font-size: 12px;
      background: var(--accent-soft);
      color: var(--accent);
    }

    .badge.error {
      background: rgba(180, 35, 24, 0.12);
      color: var(--danger);
    }

    .status-list {
      display: grid;
      gap: 12px;
    }

    .status-row {
      display: grid;
      grid-template-columns: 68px 1fr auto;
      gap: 12px;
      align-items: center;
    }

    .bar {
      height: 10px;
      border-radius: 999px;
      background: rgba(15, 118, 110, 0.1);
      overflow: hidden;
    }

    .bar > span {
      display: block;
      height: 100%;
      background: linear-gradient(90deg, #0f766e, #14b8a6);
      border-radius: 999px;
    }

    .empty {
      color: var(--muted);
      font-size: 14px;
    }

    @media (max-width: 920px) {
      .grid {
        grid-template-columns: 1fr;
      }
    }
  </style>
</head>
<body>
  <main class="shell">
    <section class="hero">
      <div class="eyebrow">Backend Observability</div>
      <h1>ReSellution API pulse board</h1>
      <p class="subhead">A lightweight baseline dashboard for request volume, latency, status distribution, and route hotspots. It refreshes every 5 seconds from the backend metrics snapshot.</p>
      <div class="meta" id="meta">Loading metrics...</div>
    </section>

    <section class="cards">
      <article class="card">
        <div class="card-label">Total Requests</div>
        <div class="card-value" id="totalRequests">0</div>
      </article>
      <article class="card">
        <div class="card-label">Average Latency</div>
        <div class="card-value" id="avgLatency">0 ms</div>
      </article>
      <article class="card">
        <div class="card-label">Uptime</div>
        <div class="card-value" id="uptime">0s</div>
      </article>
      <article class="card">
        <div class="card-label">5xx Responses</div>
        <div class="card-value" id="serverErrors">0</div>
      </article>
    </section>

    <section class="grid">
      <article class="panel">
        <h2>Route Activity</h2>
        <table>
          <thead>
            <tr>
              <th>Route</th>
              <th>Method</th>
              <th>Status</th>
              <th>Requests</th>
              <th>Avg Latency</th>
            </tr>
          </thead>
          <tbody id="routeRows">
            <tr><td colspan="5" class="empty">No request data yet.</td></tr>
          </tbody>
        </table>
      </article>

      <article class="panel">
        <h2>Status Mix</h2>
        <div class="status-list" id="statusRows">
          <div class="empty">No status data yet.</div>
        </div>
      </article>
    </section>
  </main>

  <script>
    const refreshMs = 5000

    async function loadMetrics() {
      const response = await fetch('/metrics', { headers: { 'Accept': 'application/json' } })
      if (!response.ok) {
        throw new Error('Failed to load metrics')
      }
      return response.json()
    }

    function formatUptime(seconds) {
      if (seconds < 60) return seconds + 's'
      const mins = Math.floor(seconds / 60)
      const secs = seconds % 60
      if (mins < 60) return mins + 'm ' + secs + 's'
      const hours = Math.floor(mins / 60)
      const remMins = mins % 60
      return hours + 'h ' + remMins + 'm'
    }

    function latencyLookup(latencies) {
      const map = new Map()
      for (const entry of latencies || []) {
        map.set(entry.path + '|' + entry.method, entry.avg_latency_ms)
      }
      return map
    }

    function renderRouteRows(routes, latencies) {
      const tbody = document.getElementById('routeRows')
      const sorted = [...(routes || [])].sort((a, b) => b.requests - a.requests)
      if (!sorted.length) {
        tbody.innerHTML = '<tr><td colspan="5" class="empty">No request data yet.</td></tr>'
        return
      }

      const lookup = latencyLookup(latencies)
      tbody.innerHTML = sorted.slice(0, 12).map((entry) => {
        const statusClass = entry.status >= 500 ? 'badge error' : 'badge'
        const avgLatency = lookup.get(entry.path + '|' + entry.method) || 0
        return '<tr>' +
          '<td><strong>' + entry.path + '</strong></td>' +
          '<td>' + entry.method + '</td>' +
          '<td><span class="' + statusClass + '">' + entry.status + '</span></td>' +
          '<td>' + entry.requests + '</td>' +
          '<td>' + avgLatency.toFixed(2) + ' ms</td>' +
        '</tr>'
      }).join('')
    }

    function renderStatusRows(codes) {
      const container = document.getElementById('statusRows')
      const entries = Object.entries(codes || {}).sort((a, b) => Number(a[0]) - Number(b[0]))
      if (!entries.length) {
        container.innerHTML = '<div class="empty">No status data yet.</div>'
        return
      }

      const max = Math.max(...entries.map(([, value]) => value), 1)
      container.innerHTML = entries.map(([status, count]) => {
        const width = Math.max((count / max) * 100, 4)
        const tone = Number(status) >= 500 ? 'badge error' : 'badge'
        return '<div class="status-row">' +
          '<span class="' + tone + '">' + status + '</span>' +
          '<div class="bar"><span style="width:' + width + '%"></span></div>' +
          '<strong>' + count + '</strong>' +
        '</div>'
      }).join('')
    }

    async function refresh() {
      try {
        const metrics = await loadMetrics()
        document.getElementById('totalRequests').textContent = metrics.total_requests || 0
        document.getElementById('avgLatency').textContent = (metrics.avg_latency_ms || 0) + ' ms'
        document.getElementById('uptime').textContent = formatUptime(metrics.uptime_seconds || 0)

        const serverErrors = Object.entries(metrics.requests_by_code || {})
          .filter(([status]) => Number(status) >= 500)
          .reduce((total, [, count]) => total + count, 0)
        document.getElementById('serverErrors').textContent = serverErrors

        renderRouteRows(metrics.routes, metrics.latency_by_route)
        renderStatusRows(metrics.requests_by_code)

        document.getElementById('meta').textContent =
          'Updated ' + new Date().toLocaleTimeString() + ' | Correlation IDs available via X-Correlation-ID / X-Request-ID'
      } catch (error) {
        document.getElementById('meta').textContent = 'Metrics refresh failed: ' + error.message
      }
    }

    refresh()
    setInterval(refresh, refreshMs)
  </script>
</body>
</html>
`
