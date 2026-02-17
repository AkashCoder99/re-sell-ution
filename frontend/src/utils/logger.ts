const DEBUG_LOGS_ENABLED =
  import.meta.env.DEV || import.meta.env.VITE_ENABLE_DEBUG_LOGS === 'true'

type LogMeta = Record<string, unknown> | undefined

function withTime(meta: LogMeta): LogMeta {
  return {
    ...meta,
    ts: new Date().toISOString()
  }
}

export function logDebug(message: string, meta?: LogMeta): void {
  if (!DEBUG_LOGS_ENABLED) return
  console.debug(`[debug] ${message}`, withTime(meta))
}

export function logInfo(message: string, meta?: LogMeta): void {
  if (!DEBUG_LOGS_ENABLED) return
  console.info(`[info] ${message}`, withTime(meta))
}

export function logWarn(message: string, meta?: LogMeta): void {
  console.warn(`[warn] ${message}`, withTime(meta))
}

export function logError(message: string, meta?: LogMeta): void {
  console.error(`[error] ${message}`, withTime(meta))
}
