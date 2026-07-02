export interface HealthResponse {
  console: 'operational' | 'down'
  core: 'operational' | 'down'
}

export async function fetchHealth(): Promise<HealthResponse> {
  const res = await fetch('/api/health')
  if (!res.ok) {
    throw new Error('Health check failed')
  }
  return res.json()
}
