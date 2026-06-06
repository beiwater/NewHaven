#!/usr/bin/env node

import { spawn } from 'node:child_process'
import { existsSync } from 'node:fs'
import net from 'node:net'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { logToFile } from './lib/logger.mjs'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const isWindows = process.platform === 'win32'
const npmCmd = isWindows ? 'npm.cmd' : 'npm'
const goCmd = isWindows ? 'go.exe' : 'go'

const backendDir = path.join(__dirname, 'backend')
const clientDir = path.join(__dirname, 'client', 'atlas-foods-client')
const clientNodeModules = path.join(clientDir, 'node_modules')

const services = {
  backend: {
    label: 'Backend API',
    command: goCmd,
    args: ['run', './cmd/simapi/'],
    cwd: backendDir,
    url: 'http://127.0.0.1:8088',
    host: '127.0.0.1',
    port: 8088,
    child: null,
    status: 'stopped',
    startedAt: null,
  },
  frontend: {
    label: 'Frontend',
    command: npmCmd,
    args: ['run', 'dev', '--', '--host', '127.0.0.1'],
    cwd: clientDir,
    url: 'http://127.0.0.1:5173',
    host: '127.0.0.1',
    port: 5173,
    child: null,
    status: 'stopped',
    startedAt: null,
  },
  backendNext: {
    label: 'Backend Next',
    command: goCmd,
    args: ['run', './cmd/simapi-next/'],
    cwd: path.join(__dirname, 'backend-next'),
    url: 'http://127.0.0.1:8088',
    host: '127.0.0.1',
    port: 8088,
    child: null,
    status: 'stopped',
    startedAt: null,
  },
}

const logs = []
let shuttingDown = false
let renderTimer = null
let commandMode = false
let commandBuffer = ''

async function main() {
  if (!existsSync(path.join(backendDir, 'go.mod'))) {
    console.error('Backend folder was not found. Run this script from the project root.')
    process.exit(1)
  }

  if (!existsSync(path.join(clientDir, 'package.json'))) {
    console.error('Frontend folder was not found. Run this script from the project root.')
    process.exit(1)
  }

  if (!existsSync(clientNodeModules)) {
    pushLog('system', 'Frontend dependencies are missing. Run: cd client/atlas-foods-client && npm install')
  }

  await startService('backend')
  await startService('frontend')
  startInput()
  render()
}

async function startService(name) {
  const svc = services[name]
  if (svc.child) return

  const available = await isPortAvailable(svc.port, svc.host)
  if (!available) {
    svc.status = 'blocked'
    svc.startedAt = null
    pushLog('system', `${svc.label} port is already in use: ${svc.url}`)
    return
  }

  svc.status = 'starting'
  svc.startedAt = Date.now()
  pushLog('system', `Starting ${svc.label}...`)

  const child = spawn(svc.command, svc.args, {
    cwd: svc.cwd,
    env: { ...process.env },
    shell: false,
    detached: !isWindows,
  })

  svc.child = child
  svc.status = 'running'

  child.stdout.on('data', (data) => {
    for (const line of data.toString().split(/\r?\n/).filter(Boolean)) {
      pushLog(name, line)
    }
  })

  child.stderr.on('data', (data) => {
    for (const line of data.toString().split(/\r?\n/).filter(Boolean)) {
      pushLog(name, line)
    }
  })

  child.on('error', (error) => {
    svc.status = 'error'
    pushLog(name, `Failed to start: ${error.message}`)
  })

  child.on('exit', (code, signal) => {
    svc.child = null
    svc.status = shuttingDown ? 'stopped' : code === 0 ? 'stopped' : 'error'
    const reason = signal ? `signal ${signal}` : `code ${code ?? 'unknown'}`
    pushLog(name, `${svc.label} exited with ${reason}.`)
  })
}

function stopService(name) {
  const svc = services[name]
  if (!svc.child) {
    svc.status = 'stopped'
    return Promise.resolve()
  }

  svc.status = 'stopping'
  pushLog('system', `Stopping ${svc.label}...`)
  const child = svc.child

  return new Promise((resolve) => {
    const done = () => resolve()
    child.once('exit', done)

    if (isWindows) {
      spawn('taskkill', ['/pid', String(child.pid), '/T', '/F'], { stdio: 'ignore' })
      return
    }

    try {
      process.kill(-child.pid, 'SIGINT')
    } catch {
      child.kill('SIGINT')
    }

    setTimeout(() => {
      if (!svc.child) return
      try {
        process.kill(-child.pid, 'SIGTERM')
      } catch {
        child.kill('SIGTERM')
      }
    }, 2500).unref()
  })
}

async function restartService(name) {
  await stopService(name)
  await startService(name)
}

async function restartAll() {
  await Promise.all([stopService('frontend'), stopService('backend'), stopService('backendNext')])
  await startService('backend')
  await startService('frontend')
}

async function executeCommand(input) {
  const parts = input.trim().split(/\s+/)
  const cmd = parts[0]?.toLowerCase()
  const args = parts.slice(1)

  switch (cmd) {
    case 'help':
      pushLog('system', 'Available commands:')
      pushLog('system', '  :users              - List users')
      pushLog('system', '  :user delete <id>   - Delete user by ID')
      pushLog('system', '  :data save          - Save snapshot')
      pushLog('system', '  :data load          - Load snapshot')
      pushLog('system', '  :data clear         - Clear snapshot file')
      pushLog('system', '  :bn | :next         - Toggle backend-next')
      pushLog('system', '  :q | :quit          - Shutdown')
      pushLog('system', '  :help               - Show this help')
      break
    case 'users':
    case 'user':
      if (args[0] === 'delete' && args[1]) {
        await deleteUser(args[1])
      } else if (args[0] === 'list' || args.length === 0) {
        await listUsers()
      } else {
        pushLog('system', 'Usage: :user list | :user delete <id>')
      }
      break
    case 'data':
      if (args[0] === 'save') await saveSnapshot()
      else if (args[0] === 'load') await loadSnapshot()
      else if (args[0] === 'clear') await clearSnapshot()
      else pushLog('system', 'Usage: :data save | :data load | :data clear')
      break
    case 'bn':
    case 'next':
      await toggleBackendNext()
      break
    case 'q':
    case 'quit':
      shutdown()
      break
    default:
      pushLog('system', `Unknown command: ${cmd}. Type :help for available commands.`)
  }
}

async function listUsers() {
  try {
    const res = await fetch('http://127.0.0.1:8088/api/admin/players')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    pushLog('system', `Users: ${JSON.stringify(data)}`)
  } catch (err) {
    pushLog('system', `Failed to list users: ${err.message}`)
  }
}

async function deleteUser(id) {
  try {
    const res = await fetch(`http://127.0.0.1:8088/api/admin/players/${id}`, { method: 'DELETE' })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    pushLog('system', `User ${id} deleted.`)
  } catch (err) {
    pushLog('system', `Failed to delete user: ${err.message}`)
  }
}

async function saveSnapshot() {
  try {
    const res = await fetch('http://127.0.0.1:8088/api/admin/snapshot/save', { method: 'POST' })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    pushLog('system', 'Snapshot saved.')
  } catch (err) {
    pushLog('system', `Save failed: ${err.message}`)
  }
}

async function loadSnapshot() {
  try {
    const res = await fetch('http://127.0.0.1:8088/api/admin/snapshot/load', { method: 'POST' })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    pushLog('system', 'Snapshot loaded.')
  } catch (err) {
    pushLog('system', `Load failed: ${err.message}`)
  }
}

async function clearSnapshot() {
  const snapPath = path.join(__dirname, 'data', 'snapshot.json')
  try {
    const { unlinkSync } = await import('node:fs')
    unlinkSync(snapPath)
    pushLog('system', 'Snapshot file deleted.')
  } catch (err) {
    pushLog('system', `Clear failed: ${err.message}`)
  }
}

async function toggleBackendNext() {
  const existing = services['backend']
  const bn = services['backendNext']
  if (!bn) {
    pushLog('system', 'Backend-next service not defined.')
    return
  }

  if (bn.child) {
    await stopService('backendNext')
    pushLog('system', 'Backend-next stopped.')
  } else {
    // Stop original backend if running (same port)
    if (existing && existing.child) {
      pushLog('system', 'Stopping original backend before starting backend-next...')
      await stopService('backend')
    }
    await startService('backendNext')
    pushLog('system', 'Backend-next started.')
  }
}

async function shutdown() {
  if (shuttingDown) return
  shuttingDown = true
  pushLog('system', 'Shutting down services...')
  render()
  await Promise.all([stopService('frontend'), stopService('backend'), stopService('backendNext')])
  cleanupInput()
  if (renderTimer) clearTimeout(renderTimer)
  process.stdout.write('\x1b[?25h\x1b[0m\n')
  process.exit(0)
}

function startInput() {
  if (!process.stdin.isTTY) return
  process.stdin.setRawMode(true)
  process.stdin.resume()
  process.stdin.setEncoding('utf8')
  process.stdin.on('data', handleKey)
}

function cleanupInput() {
  if (!process.stdin.isTTY) return
  process.stdin.off('data', handleKey)
  process.stdin.setRawMode(false)
  process.stdin.pause()
}

function handleKey(key) {
  if (commandMode) {
    if (key === '\u001b' || key === '\u0003') {
      // Escape or Ctrl+C cancels command mode
      commandMode = false
      commandBuffer = ''
      scheduleRender()
      return
    }
    if (key === '\r' || key === '\n') {
      // Enter executes command
      const cmd = commandBuffer
      commandMode = false
      commandBuffer = ''
      scheduleRender()
      if (cmd.trim()) executeCommand(cmd.trim())
      return
    }
    if (key === '\x7f' || key === '\b') {
      // Backspace
      commandBuffer = commandBuffer.slice(0, -1)
      scheduleRender()
      return
    }
    // Printable character (ignore escape sequences)
    if (key.length === 1 && key.charCodeAt(0) >= 32) {
      commandBuffer += key
      scheduleRender()
    }
    return
  }

  // Normal mode
  if (key === '\u0003' || key.toLowerCase() === 'q') {
    shutdown()
    return
  }
  if (key === ':') {
    commandMode = true
    commandBuffer = ''
    scheduleRender()
    return
  }
  if (key.toLowerCase() === 'b') restartService('backend')
  if (key.toLowerCase() === 'f') restartService('frontend')
  if (key.toLowerCase() === 'n') toggleBackendNext()
  if (key.toLowerCase() === 'r') restartAll()
  if (key.toLowerCase() === 'c') {
    logs.length = 0
    pushLog('system', 'Logs cleared.')
  }
}

function pushLog(source, message) {
  const time = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  logs.push({ time, source, message })
  while (logs.length > 200) logs.shift()
  // Also write to daily log file
  logToFile(time, source, message)
  scheduleRender()
}

function scheduleRender() {
  if (renderTimer) return
  renderTimer = setTimeout(() => {
    renderTimer = null
    render()
  }, 80)
}

function render() {
  const width = process.stdout.columns || 100
  const height = process.stdout.rows || 32
  const logRows = Math.max(8, height - 13 - (commandMode ? 1 : 0))
  const line = '─'.repeat(width)

  process.stdout.write('\x1b[?25l\x1b[H\x1b[2J')
  writeLine(color('New Haven Dev Console', 'gold', true), width)
  writeLine('Browser-based economy sim | API + Vite launcher', width)
  writeLine(line, width)

  for (const key of ['backend', 'frontend', 'backendNext']) {
    const svc = services[key]
    const uptime = svc.startedAt && svc.child ? formatDuration(Date.now() - svc.startedAt) : '--'
    writeLine(`${statusBadge(svc.status)} ${svc.label.padEnd(12)} ${svc.url.padEnd(24)} uptime ${uptime}`, width)
  }

  writeLine(line, width)
  writeLine('Keys: q quit | r restart all | b backend | f frontend | n backend-next | c clear logs | : command', width)
  writeLine(line, width)

  const visibleLogs = logs.slice(-logRows)
  for (const entry of visibleLogs) {
    const prefix = `${entry.time} ${sourceLabel(entry.source)}`
    writeLine(`${prefix} ${entry.message}`, width)
  }

  for (let i = visibleLogs.length; i < logRows; i++) writeLine('', width)
  writeLine(line, width)
  writeLine(`Login tip: use dev / dev | Logs: log/dev-console-${new Date().toISOString().slice(0, 10)}.log`, width)

  if (commandMode) {
    const prompt = `: ${commandBuffer}`
    writeLine(prompt, width)
  }
}

function writeLine(text, width) {
  const plain = stripAnsi(text)
  if (plain.length <= width) {
    process.stdout.write(`${text}${' '.repeat(width - plain.length)}\n`)
    return
  }
  process.stdout.write(`${truncateAnsi(text, width)}\n`)
}

function sourceLabel(source) {
  if (source === 'backend') return color('[api]', 'cyan')
  if (source === 'frontend') return color('[web]', 'green')
  if (source === 'backendNext') return color('[nxt]', 'magenta')
  return color('[sys]', 'gold')
}

function statusBadge(status) {
  if (status === 'running') return color('● running ', 'green', true)
  if (status === 'starting') return color('● starting', 'gold', true)
  if (status === 'stopping') return color('● stopping', 'gold', true)
  if (status === 'blocked') return color('● blocked ', 'red', true)
  if (status === 'error') return color('● error   ', 'red', true)
  return color('● stopped ', 'gray', true)
}

function isPortAvailable(port, host) {
  return new Promise((resolve) => {
    const server = net.createServer()
    server.once('error', () => resolve(false))
    server.once('listening', () => {
      server.close(() => resolve(true))
    })
    server.listen(port, host)
  })
}

function formatDuration(ms) {
  const total = Math.floor(ms / 1000)
  const min = Math.floor(total / 60)
  const sec = total % 60
  return `${String(min).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
}

function color(text, name, bold = false) {
  const codes = {
    gray: 90,
    red: 31,
    green: 32,
    gold: 33,
    cyan: 36,
  }
  const prefix = `\x1b[${bold ? '1;' : ''}${codes[name] ?? 37}m`
  return `${prefix}${text}\x1b[0m`
}

function stripAnsi(text) {
  return text.replace(/\x1b\[[0-9;]*m/g, '')
}

function truncateAnsi(text, width) {
  const plain = stripAnsi(text)
  if (plain.length <= width) return text
  return `${plain.slice(0, Math.max(0, width - 1))}…`
}

process.on('SIGINT', shutdown)
process.on('SIGTERM', shutdown)
process.on('exit', () => {
  process.stdout.write('\x1b[?25h\x1b[0m')
})

main()
