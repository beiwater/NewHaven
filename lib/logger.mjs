import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const logDir = path.join(__dirname, '..', 'log')

// Ensure log directory exists
if (!fs.existsSync(logDir)) {
  fs.mkdirSync(logDir, { recursive: true })
}

// Get today's log file path
function logFilePath() {
  const now = new Date()
  const date = now.toISOString().slice(0, 10) // YYYY-MM-DD
  return path.join(logDir, `dev-console-${date}.log`)
}

// Write a single log entry to the file
function writeEntry(time, source, message) {
  try {
    const line = `[${time}] [${source}] ${message}\n`
    fs.appendFileSync(logFilePath(), line, 'utf8')
  } catch {
    // Silently fail — don't let logging break the console
  }
}

export function logToFile(time, source, message) {
  writeEntry(time, source, message)
}
