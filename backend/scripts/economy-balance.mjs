#!/usr/bin/env node

import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const TARGET_PROFIT_PER_HOUR = 300
const EXCHANGE_FEE = 0.04
const MAX_TARGET_DEVIATION = 35

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..')
const resources = JSON.parse(readFileSync(resolve(root, 'decompiled/data/resources.json'), 'utf8'))
const byID = new Map(resources.map((resource) => [resource.dbLetter || resource.id, resource]))

const workersByBuilding = {
  1: 2, 2: 3, 3: 4, 4: 5, 5: 6, 6: 3,
  7: 5, 8: 4, 9: 8, 10: 7, 11: 8, 12: 9,
}

function hourlyWage(kind) {
  return (workersByBuilding[kind] ?? 3) * 345
}

function inputCost(resource) {
  return Object.entries(resource.producedFrom ?? {}).reduce((total, [id, quantity]) => {
    const input = byID.get(Number(id))
    if (!input) throw new Error(`${resource.name} references missing input ${id}`)
    return total + input.basePrice * quantity
  }, 0)
}

const rows = resources
  .filter((resource) => (resource.dbLetter || resource.id) > 0)
  .map((resource) => {
    const rate = resource.producedPerHourRaw
    const wage = hourlyWage(resource.primaryBuildingKind)
    const materials = inputCost(resource)
    const profit = rate * (resource.basePrice * (1 - EXCHANGE_FEE) - materials) - wage
    return {
      id: resource.dbLetter || resource.id,
      resource: resource.name,
      producer: resource.primaryBuildingKind,
      rate,
      demand: resource.retailDemandPerHour,
      price: resource.basePrice,
      materials,
      wage,
      profit,
      delta: profit - TARGET_PROFIT_PER_HOUR,
    }
  })
  .sort((a, b) => a.id - b.id)

console.log('| Item | Producer | Rate/h | Demand/h | Market anchor | Inputs/h | Wage/h | Net/h |')
console.log('| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |')
for (const row of rows) {
  console.log(`| ${row.resource} | ${row.producer} | ${row.rate.toFixed(1)} | ${row.demand.toFixed(1)} | $${row.price.toFixed(2)} | $${row.materials.toFixed(2)} | $${row.wage.toFixed(2)} | $${row.profit.toFixed(2)} |`)
}

const failed = rows.filter((row) => (
  row.rate <= 0
  || row.demand <= 0
  || row.producer <= 0
  || Math.abs(row.delta) > MAX_TARGET_DEVIATION
))

if (process.argv.includes('--check')) {
  if (failed.length > 0) {
    console.error(`\nBalance check failed: ${failed.map((row) => `${row.resource} ($${row.profit.toFixed(2)}/h)`).join(', ')}`)
    process.exit(1)
  }
  console.log(`\nBalance check passed: all level-one producer routes are within ±$${MAX_TARGET_DEVIATION.toFixed(0)} of $${TARGET_PROFIT_PER_HOUR}/h after the 4% exchange fee.`)
}
