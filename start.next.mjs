#!/usr/bin/env node

import { spawn } from 'node:child_process'
import { existsSync } from 'node:fs'
import net from 'node:net'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { logToFile } from './lib/logger.mjs'
import {
  adminMoneyGive, adminMoneySet, adminMoneyRemove,
  adminResourceGive, adminResourceRemove,
  adminBuildingGive, adminBuildingRemove,
  adminXpGive, adminXpSet,
  adminResearchSet,
  adminExecutiveGive, adminExecutiveRemove,
} from './lib/admin.mjs'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const isWindows = process.platform === 'win32'
const npmCmd = isWindows ? 'npm.cmd' : 'npm'
const goCmd = isWindows ? 'go.exe' : 'go'

const backendDir = path.join(__dirname, 'backend')
const clientDir = path.join(__dirname, 'client', 'atlas-foods-client')
const clientNodeModules = path.join(clientDir, 'node_modules')
const snapshotDir = path.join(__dirname, 'backend-next', 'data')

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
let autosaveEnabled = true
let autosaveInterval = 300000 // 5 minutes in ms
let autosaveTimer = null
let lastSaveTime = null


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

  await forceFreePort(8088)
  await startService('backendNext')
  await startService('frontend')
  startInput()
  startAutosave()

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
    shell: isWindows,
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
      pushLog('system', '  :money give|set|remove <companyId> <amount>')
      pushLog('system', '  :resource give|remove <companyId> <resourceId> <amount>')
      pushLog('system', '  :building give <companyId> <buildingId> [level] | remove')
      pushLog('system', '  :xp give|set <companyId> <amount>')
      pushLog('system', '  :research set <companyId> <resourceId> <level>')
      pushLog('system', '  :executive give|remove <companyId> ...')
      pushLog('system', '  :autosave on|off|interval <N>  - Toggle/configure autosave')
      pushLog('system', '  :save <name>           - Save snapshot as <name>')
      pushLog('system', '  :load <name>           - Load snapshot from <name>')
      pushLog('system', '  :saves                 - List named snapshots')
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

    case 'money':
      if (args[0] === 'give' && args[1] && args[2]) {
        try { await adminMoneyGive(args[1], args[2]); pushLog('system', `Money given to company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else if (args[0] === 'set' && args[1] && args[2]) {
        try { await adminMoneySet(args[1], args[2]); pushLog('system', `Money set for company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else if (args[0] === 'remove' && args[1] && args[2]) {
        try { await adminMoneyRemove(args[1], args[2]); pushLog('system', `Money removed from company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else {
        pushLog('system', 'Usage: :money give|set|remove <companyId> <amount>')
      }
      break

    case 'resource':
      if (args[0] === 'give' && args[1] && args[2] && args[3]) {
        try { await adminResourceGive(args[1], args[2], args[3]); pushLog('system', `Resource ${args[2]} given to company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else if (args[0] === 'remove' && args[1] && args[2] && args[3]) {
        try { await adminResourceRemove(args[1], args[2], args[3]); pushLog('system', `Resource ${args[2]} removed from company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else {
        pushLog('system', 'Usage: :resource give|remove <companyId> <resourceId> <amount>')
      }
      break

    case 'building':
      if (args[0] === 'give' && args[1] && args[2]) {
        try { await adminBuildingGive(args[1], args[2], args[3]); pushLog('system', `Building ${args[2]} given to company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else if (args[0] === 'remove' && args[1] && args[2]) {
        try { await adminBuildingRemove(args[1], args[2]); pushLog('system', `Building ${args[2]} removed from company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else {
        pushLog('system', 'Usage: :building give <companyId> <buildingId> [level] | :building remove <companyId> <buildingId>')
      }
      break

    case 'xp':
      if (args[0] === 'give' && args[1] && args[2]) {
        try { await adminXpGive(args[1], args[2]); pushLog('system', `XP given to company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else if (args[0] === 'set' && args[1] && args[2]) {
        try { await adminXpSet(args[1], args[2]); pushLog('system', `XP set for company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else {
        pushLog('system', 'Usage: :xp give|set <companyId> <amount>')
      }
      break

    case 'research':
      if (args[0] === 'set' && args[1] && args[2] && args[3]) {
        try { await adminResearchSet(args[1], args[2], args[3]); pushLog('system', `Research level ${args[3]} set for resource ${args[2]} on company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else {
        pushLog('system', 'Usage: :research set <companyId> <resourceId> <level>')
      }
      break

    case 'executive':
      if (args[0] === 'give' && args[1] && args[2] && args[3] && args[4] && args[5]) {
        try { await adminExecutiveGive(args[1], args[2], args[3], args[4], args[5]); pushLog('system', `Executive added to company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else if (args[0] === 'remove' && args[1] && args[2]) {
        try { await adminExecutiveRemove(args[1], args[2]); pushLog('system', `Executive ${args[2]} removed from company ${args[1]}.`) }
        catch (e) { pushLog('system', `Failed: ${e.message}`) }
      } else {
        pushLog('system', 'Usage: :executive give <companyId> <name> <title> <level> <rarity> | :executive remove <companyId> <executiveId>')
      }
      break
    case 'autosave':
      if (args[0] === 'on') {
        autosaveEnabled = true
        startAutosave()
        pushLog('system', 'Autosave enabled.')
      } else if (args[0] === 'off') {
        autosaveEnabled = false
        stopAutosave()
        pushLog('system', 'Autosave disabled.')
      } else if (args[0] === 'interval' && args[1]) {
        const minutes = Math.max(1, parseInt(args[1], 10) || 5)
        autosaveInterval = minutes * 60000
        if (autosaveEnabled) startAutosave()
        pushLog('system', `Autosave interval set to ${minutes} min.`)
      } else {
        pushLog('system', `Autosave: ${autosaveEnabled ? 'on' : 'off'} (every ${autosaveInterval / 60000} min)`)
        pushLog('system', 'Usage: :autosave on|off|interval <minutes>')
      }
      break

    case 'save':
      await saveNamedSnapshot(args[0])
      break

    case 'load':
      await loadNamedSnapshot(args[0])
      break

    case 'saves':
      await listNamedSnapshots()
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
    lastSaveTime = Date.now()
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
  const snapPath = path.join(snapshotDir, 'snapshot.json')
  try {
    const { unlinkSync } = await import('node:fs')
    unlinkSync(snapPath)
    pushLog('system', 'Snapshot file deleted.')
  } catch (err) {
    pushLog('system', `Clear failed: ${err.message}`)
  }
}

function snapshotsDir() {
  return path.join(snapshotDir, 'snapshots')
}

async function saveNamedSnapshot(name) {
  if (!name) {
    pushLog('system', 'Usage: :save <name>')
    return
  }
  // First flush current state to primary snapshot
  try {
    const res = await fetch('http://127.0.0.1:8088/api/admin/snapshot/save', { method: 'POST' })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
  } catch (err) {
    pushLog('system', `Failed to save state: ${err.message}`)
    return
  }

  // Copy primary snapshot to named file
  const dir = snapshotsDir()
  const src = path.join(snapshotDir, 'snapshot.json')
  const dst = path.join(dir, `${name}.json`)

  const { copyFileSync, mkdirSync, existsSync } = await import('node:fs')
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true })
  try {
    copyFileSync(src, dst)
    pushLog('system', `Saved snapshot: ${name}`)
    lastSaveTime = Date.now()
  } catch (err) {
    pushLog('system', `Failed to copy snapshot: ${err.message}`)
  }
}

async function loadNamedSnapshot(name) {
  if (!name) {
    pushLog('system', 'Usage: :load <name>')
    return
  }
  const src = path.join(snapshotsDir(), `${name}.json`)
  const dst = path.join(snapshotDir, 'snapshot.json')

  const { copyFileSync, existsSync } = await import('node:fs')
  if (!existsSync(src)) {
    pushLog('system', `Snapshot not found: ${name}`)
    return
  }

  try {
    copyFileSync(src, dst)
  } catch (err) {
    pushLog('system', `Failed to copy snapshot: ${err.message}`)
    return
  }

  // Load into backend
  try {
    const res = await fetch('http://127.0.0.1:8088/api/admin/snapshot/load', { method: 'POST' })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    pushLog('system', `Loaded snapshot: ${name}`)
  } catch (err) {
    pushLog('system', `Failed to load snapshot: ${err.message}`)
  }
}

async function listNamedSnapshots() {
  const dir = snapshotsDir()
  const { readdirSync, existsSync } = await import('node:fs')
  if (!existsSync(dir)) {
    pushLog('system', 'No named snapshots.')
    return
  }
  try {
    const files = readdirSync(dir).filter(f => f.endsWith('.json')).map(f => f.replace(/\.json$/, ''))
    if (files.length === 0) {
      pushLog('system', 'No named snapshots.')
    } else {
      pushLog('system', `Snapshots: ${files.join(', ')}`)
    }
  } catch (err) {
    pushLog('system', `Failed to list snapshots: ${err.message}`)
  }
}

async function forceFreePort(port) {
  // Windows: find and kill any process holding the port
  if (!isWindows) return
  try {
    const { execSync } = await import('node:child_process')
    const out = execSync(`netstat -ano | findstr :${port}`, { encoding: 'utf8', timeout: 3000 })
    const pids = new Set()
    for (const line of out.split(/\r?\n/)) {
      const m = line.trim().match(/LISTENING\s+(\d+)$/)
      if (m) pids.add(m[1])
    }
    for (const pid of pids) {
      try { execSync(`taskkill /F /PID ${pid}`, { stdio: 'ignore', timeout: 2000 }) } catch {}
      pushLog('system', `Killed stale process ${pid} on port ${port}`)
    }
    if (pids.size > 0) {
      await new Promise(r => setTimeout(r, 500))
    }
  } catch { /* netstat failed or no process found */ }
}

async function toggleBackendNext() {
  const bn = services['backendNext']
  if (!bn) {
    pushLog('system', 'Backend-next service not defined.')
    return
  }

  if (bn.child) {
    await stopService('backendNext')
    pushLog('system', 'Backend-next stopped.')
  } else {
    await forceFreePort(8088)
    // Also stop the old backend tracker if it still thinks it's running
    if (services['backend'].child) {
      await stopService('backend')
    }
    services['backend'].status = 'stopped'
    await startService('backendNext')
  }
}

function startAutosave() {
  stopAutosave()
  if (!autosaveEnabled) return
  autosaveTimer = setInterval(() => {
    saveSnapshot()
  }, autosaveInterval)
  pushLog('system', `Autosave enabled (every ${autosaveInterval / 60000} min)`)
}

function stopAutosave() {
  if (autosaveTimer) {
    clearInterval(autosaveTimer)
    autosaveTimer = null
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
  stopAutosave()
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
  const logRows = Math.max(8, height - 14 - (commandMode ? 1 : 0))
  const line = '─'.repeat(width)

  process.stdout.write('\x1b[?25l\x1b[H\x1b[2J')
  writeLine(color('New Haven Dev Console', 'gold', true), width)
  writeLine('Browser-based economy sim | API + Vite launcher', width)
  writeLine(line, width)

  for (const key of ['backendNext', 'frontend', 'backend']) {
    const svc = services[key]
    if (!svc) continue
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

  // System status line
  const dbMode = process.env.SIM_API_DATABASE_URL ? color('PostgreSQL', 'cyan', true) : color('Memory (file)', 'gold', true)
  const asStatus = autosaveEnabled ? color('on', 'green') : color('off', 'gray')
  const asInterval = `${autosaveInterval / 60000}min`
  const saveTime = lastSaveTime
    ? new Date(lastSaveTime).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    : '--'
  writeLine(`DB: ${dbMode}  |  Autosave: ${asStatus} (${asInterval})  |  Last save: ${saveTime}`, width)

  writeLine(`Login tip: use dev / dev  |  Logs: log/dev-console-${new Date().toISOString().slice(0, 10)}.log  |  :help for commands`, width)

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
