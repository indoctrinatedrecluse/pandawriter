import './style.css'
import { mount } from 'svelte'
import App from './App.svelte'
import { installDevelopmentDiagnostics, logError, logInfo, showDevelopmentFailure } from './logger'

const target = document.getElementById('app')
let app: ReturnType<typeof mount> | undefined

if (!target) {
  throw new Error('The #app mount target is missing')
}

installDevelopmentDiagnostics(target)

try {
  logInfo('Frontend bootstrap started')
  app = mount(App, { target })
  logInfo('Frontend bootstrap completed')
} catch (error) {
  logError('Frontend bootstrap failed', error)
  showDevelopmentFailure(target, error)
}

export default app
