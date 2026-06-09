const API_BASE = 'http://127.0.0.1:8088'

async function apiPost(path, body) {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}: ${res.statusText}`)
  return res.json()
}

async function apiDelete(path) {
  const res = await fetch(`${API_BASE}${path}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`HTTP ${res.status}: ${res.statusText}`)
  return res.json()
}

export async function adminMoneyGive(companyId, amount) {
  const delta = Math.abs(Number(amount))
  return apiPost(`/api/admin/companies/${companyId}/money`, { delta })
}

export async function adminMoneySet(companyId, amount) {
  return apiPost(`/api/admin/companies/${companyId}/money`, { set: Number(amount) })
}

export async function adminMoneyRemove(companyId, amount) {
  const delta = -Math.abs(Number(amount))
  return apiPost(`/api/admin/companies/${companyId}/money`, { delta })
}

export async function adminResourceGive(companyId, resourceId, amount) {
  return apiPost(`/api/admin/companies/${companyId}/inventory`, { resourceId: Number(resourceId), delta: Math.abs(Number(amount)) })
}

export async function adminResourceRemove(companyId, resourceId, amount) {
  return apiPost(`/api/admin/companies/${companyId}/inventory`, { resourceId: Number(resourceId), delta: -Math.abs(Number(amount)) })
}

export async function adminBuildingGive(companyId, buildingId, level = 1) {
  return apiPost(`/api/admin/companies/${companyId}/buildings`, { buildingId: Number(buildingId), level: Number(level) })
}

export async function adminBuildingRemove(companyId, buildingId) {
  return apiDelete(`/api/admin/companies/${companyId}/buildings/${buildingId}`)
}

export async function adminXpGive(companyId, amount) {
  return apiPost(`/api/admin/companies/${companyId}/xp`, { delta: Math.abs(Number(amount)) })
}

export async function adminXpSet(companyId, amount) {
  return apiPost(`/api/admin/companies/${companyId}/xp`, { set: Number(amount) })
}

export async function adminResearchSet(companyId, resourceId, level) {
  return apiPost(`/api/admin/companies/${companyId}/research`, { resourceId: Number(resourceId), level: Number(level) })
}

export async function adminExecutiveGive(companyId, name, title, level, rarity) {
  return apiPost(`/api/admin/companies/${companyId}/executives`, { name, title, level: Number(level), rarity })
}

export async function adminExecutiveRemove(companyId, executiveId) {
  return apiDelete(`/api/admin/companies/${companyId}/executives/${executiveId}`)
}
