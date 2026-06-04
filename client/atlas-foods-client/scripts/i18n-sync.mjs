#!/usr/bin/env node
/**
 * i18n-sync.mjs — Sync target locale files with en-US source.
 *
 * Usage:
 *   node scripts/i18n-sync.mjs
 *
 * Adds missing keys from en-US.json to target files.
 * Preserves existing translations.
 * Marks missing translations with [TODO <locale>] prefix.
 */

import { readFileSync, writeFileSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const LOCALES_DIR = resolve(__dirname, '..', 'src', 'i18n', 'locales')
const SOURCE = 'en-US'
const PLACEHOLDER_RE = /^\{\{[^}]+\}\}$/

function loadJson(filePath) {
  return JSON.parse(readFileSync(filePath, 'utf8'))
}

function deepMerge(target, source, localeLabel) {
  const result = { ...target }
  for (const [key, value] of Object.entries(source)) {
    if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
      // Nested object: recurse
      result[key] = deepMerge(
        typeof result[key] === 'object' && result[key] !== null ? result[key] : {},
        value,
        localeLabel,
      )
    } else if (result[key] === undefined) {
      // Missing key: add placeholder
      if (typeof value === 'string' && PLACEHOLDER_RE.test(value)) {
        // Preserve interpolation placeholders as-is
        result[key] = `[TODO ${localeLabel}] ${value}`
      } else {
        result[key] = `[TODO ${localeLabel}] ${value}`
      }
    }
    // Existing key: preserve translation
  }
  return result
}

function deepClean(obj, source) {
  // Remove keys that don't exist in source
  const result = {}
  for (const [key, value] of Object.entries(obj)) {
    if (source[key] !== undefined) {
      if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
        result[key] = deepClean(value, source[key] || {})
      } else {
        result[key] = value
      }
    }
    // else: drop extra key
  }
  return result
}

const sourcePath = resolve(LOCALES_DIR, `${SOURCE}.json`)
const sourceData = loadJson(sourcePath)

const targets = ['zh-CN']

for (const locale of targets) {
  const targetPath = resolve(LOCALES_DIR, `${locale}.json`)
  const targetData = loadJson(targetPath)

  // Merge missing keys, preserve existing
  const merged = deepMerge(targetData, sourceData, locale)
  // Clean up extra keys not in source
  const cleaned = deepClean(merged, sourceData)

  const output = JSON.stringify(cleaned, null, 2) + '\n'
  writeFileSync(targetPath, output, 'utf8')
  console.log(`Synced ${locale}.json`)
}

console.log('Sync complete.')
