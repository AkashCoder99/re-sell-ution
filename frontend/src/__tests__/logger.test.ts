import { describe, it, expect, vi, beforeEach } from 'vitest'
import { logInfo, logWarn, logError, logDebug } from '../utils/logger'

describe('logger', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('logInfo calls console.info', () => {
    const spy = vi.spyOn(console, 'info').mockImplementation(() => {})
    logInfo('test.event')
    expect(spy).toHaveBeenCalledTimes(1)
  })

  it('logInfo passes event name in the log message', () => {
    const spy = vi.spyOn(console, 'info').mockImplementation(() => {})
    logInfo('auth.login.success')
    const args = spy.mock.calls[0]
    expect(args[0]).toContain('auth.login.success')
  })

  it('logWarn calls console.warn', () => {
    const spy = vi.spyOn(console, 'warn').mockImplementation(() => {})
    logWarn('test.warn')
    expect(spy).toHaveBeenCalledTimes(1)
  })

  it('logError calls console.error', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})
    logError('test.error')
    expect(spy).toHaveBeenCalledTimes(1)
  })

  it('logDebug calls console.debug', () => {
    const spy = vi.spyOn(console, 'debug').mockImplementation(() => {})
    logDebug('test.debug')
    expect(spy).toHaveBeenCalledTimes(1)
  })

  it('logInfo includes extra context when provided', () => {
    const spy = vi.spyOn(console, 'info').mockImplementation(() => {})
    logInfo('auth.login', { user_id: '123' })
    const meta = spy.mock.calls[0][1] as Record<string, unknown>
    expect(meta).toMatchObject({ user_id: '123' })
  })

  it('logError includes error message in context', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})
    logError('api.failed', { error: 'network timeout' })
    const meta = spy.mock.calls[0][1] as Record<string, unknown>
    expect(meta).toMatchObject({ error: 'network timeout' })
  })
})
