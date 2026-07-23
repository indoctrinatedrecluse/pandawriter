import { LogDebug, LogError, LogInfo } from '../wailsjs/runtime/runtime'

type LogLevel = 'debug' | 'info' | 'error'

function formatDetail(detail: unknown): string {
  if (detail === undefined) return ''
  if (detail instanceof Error) return `\n${detail.stack ?? detail.message}`

  try {
    return `\n${JSON.stringify(detail)}`
  } catch {
    return `\n${String(detail)}`
  }
}

function write(level: LogLevel, message: string, detail?: unknown) {
  const entry = `[PandaWriter] ${message}${formatDetail(detail)}`

  if (level === 'error') console.error(entry)
  else if (level === 'debug') console.debug(entry)
  else console.info(entry)

  try {
    if (level === 'error') LogError(entry)
    else if (level === 'debug') LogDebug(entry)
    else LogInfo(entry)
  } catch {
    // The Wails runtime is unavailable during an early frontend startup failure.
    // The browser console above remains available for that case.
  }
}

export function logDebug(message: string, detail?: unknown) {
  write('debug', message, detail)
}

export function logInfo(message: string, detail?: unknown) {
  write('info', message, detail)
}

export function logError(message: string, detail?: unknown) {
  write('error', message, detail)
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.stack ?? error.message : String(error)
}

export function showDevelopmentFailure(root: HTMLElement, error: unknown) {
  if (!import.meta.env.DEV) return

  const container = document.createElement('section')
  container.className = 'development-error'
  const heading = document.createElement('h1')
  heading.textContent = 'PandaWriter could not start'
  const explanation = document.createElement('p')
  explanation.textContent = 'Open the Wails development console or Web Inspector for the full diagnostic trace.'
  const details = document.createElement('pre')
  details.textContent = errorMessage(error)
  container.append(heading, explanation, details)
  root.replaceChildren(container)
}

export function installDevelopmentDiagnostics(root: HTMLElement) {
  window.addEventListener('error', (event) => {
    logError('Unhandled frontend error', event.error ?? event.message)
    showDevelopmentFailure(root, event.error ?? event.message)
  })

  window.addEventListener('unhandledrejection', (event) => {
    logError('Unhandled frontend promise rejection', event.reason)
    showDevelopmentFailure(root, event.reason)
  })
}
