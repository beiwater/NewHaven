/**
 * WebSocket hook - disabled until backend implements WS endpoint.
 * The backend currently returns 404 for all WS paths.
 */

export function useMarketWebSocket() {
  // Not implemented
}

export function useProductionWebSocket() {
  // Not implemented
}

export function sendWS(type: string, payload?: Record<string, unknown>) {
  void type
  void payload
  // Not implemented
}
