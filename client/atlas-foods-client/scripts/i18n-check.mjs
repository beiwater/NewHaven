#!/usr/bin/env node
/**
 * i18n-check.mjs — Validate that all target locale files have the same keys as en-US.
 *
 * Usage:
 *   node scripts/i18n-check.mjs
 *
 * Exits 1 if any target locale is missing keys or has extra keys.
 */

import { readFileSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const LOCALES_DIR = resolve(__dirname, '..', 'src', 'i18n', 'locales')
const SOURCE = 'en-US'

function loadJson(filePath) {
  try {
    return JSON.parse(readFileSync(filePath, 'utf8'))
  } catch (err) {
    console.error(`ERROR: Cannot read ${filePath}: ${err.message}`)
    process.exit(1)
  }
}

function flattenKeys(obj, prefix = '') {
  const result = []
  for (const [key, value] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key
    if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
      result.push(...flattenKeys(value, fullKey))
    } else {
      result.push(fullKey)
    }
  }
  return result
}

const sourcePath = resolve(LOCALES_DIR, `${SOURCE}.json`)
const sourceData = loadJson(sourcePath)
const sourceKeys = new Set(flattenKeys(sourceData))

const targetFiles = []
const localeDir = LOCALES_DIR
// We don't have a proper directory listing, so we hardcode targets
const targets = ['zh-CN']

if (targets.length === 0) {
  console.log('No target locale files found. Nothing to check.')
  process.exit(0)
}

let hasErrors = false

for (const locale of targets) {
  const targetPath = resolve(localeDir, `${locale}.json`)
  const targetData = loadJson(targetPath)
  const targetKeys = new Set(flattenKeys(targetData))

  const missing = [...sourceKeys].filter((k) => !targetKeys.has(k))
  const extra = [...targetKeys].filter((k) => !sourceKeys.has(k))

  if (missing.length > 0 || extra.length > 0) {
    hasErrors = true
    console.log(`\n${locale}:`)
    for (const k of missing) {
      console.log(`  MISSING  ${k}`)
    }
    for (const k of extra) {
      console.log(`  EXTRA    ${k}`)
    }
  }
}

if (!hasErrors) {
  console.log(`All target locales match ${SOURCE}.`)
  process.exit(0)
} else {
  console.log(`\nValidation failed. Sync target locales with ${SOURCE}.json.`)
  process.exit(1)
}
