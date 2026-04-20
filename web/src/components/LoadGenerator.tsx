import { useState, useRef } from 'react'

interface Results {
  total: number
  successes: number
  failures: number
  avgLatency: number
}

export default function LoadGenerator() {
  const [requests, setRequests] = useState(50)
  const [concurrency, setConcurrency] = useState(5)
  const [pattern, setPattern] = useState<'browse' | 'order' | 'mixed'>('mixed')
  const [running, setRunning] = useState(false)
  const [results, setResults] = useState<Results | null>(null)
  const [progress, setProgress] = useState(0)
  const abortRef = useRef<AbortController | null>(null)

  async function runLoad() {
    setRunning(true)
    setResults(null)
    setProgress(0)

    const controller = new AbortController()
    abortRef.current = controller
    const signal = controller.signal

    const latencies: number[] = []
    let successes = 0
    let failures = 0
    let completed = 0

    const endpoints = getEndpoints(pattern)

    async function doRequest() {
      while (completed < requests && !signal.aborted) {
        const url = endpoints[Math.floor(Math.random() * endpoints.length)]
        const start = performance.now()
        try {
          const res = await fetch(url.path, {
            method: url.method,
            signal,
            headers: url.body ? { 'Content-Type': 'application/json' } : undefined,
            body: url.body ? JSON.stringify(url.body) : undefined,
          })
          if (res.ok) successes++
          else failures++
        } catch {
          if (!signal.aborted) failures++
        }
        latencies.push(performance.now() - start)
        completed++
        setProgress(Math.round((completed / requests) * 100))
      }
    }

    const workers = Array.from({ length: Math.min(concurrency, requests) }, () => doRequest())
    await Promise.all(workers)

    if (!signal.aborted) {
      setResults({
        total: completed,
        successes,
        failures,
        avgLatency: latencies.length > 0
          ? Math.round(latencies.reduce((a, b) => a + b, 0) / latencies.length)
          : 0,
      })
    }
    setRunning(false)
  }

  function stop() {
    abortRef.current?.abort()
    setRunning(false)
  }

  return (
    <div className="load-generator">
      <h2>Load Generator</h2>
      <p className="subtitle">Generate traffic to observe in your dashboards</p>

      <div className="form-grid">
        <label>
          Requests
          <input type="number" min={1} max={1000} value={requests}
            onChange={e => setRequests(Number(e.target.value))} disabled={running} />
        </label>
        <label>
          Concurrency
          <input type="number" min={1} max={50} value={concurrency}
            onChange={e => setConcurrency(Number(e.target.value))} disabled={running} />
        </label>
        <label>
          Pattern
          <select value={pattern} onChange={e => setPattern(e.target.value as typeof pattern)} disabled={running}>
            <option value="browse">Browse (GET only)</option>
            <option value="order">Orders (POST + GET)</option>
            <option value="mixed">Mixed</option>
          </select>
        </label>
      </div>

      <div className="actions">
        {!running ? (
          <button className="btn btn-primary" onClick={runLoad}>Generate Traffic</button>
        ) : (
          <button className="btn btn-danger" onClick={stop}>Stop</button>
        )}
      </div>

      {running && (
        <div className="progress-bar">
          <div className="progress-fill" style={{ width: `${progress}%` }} />
          <span>{progress}%</span>
        </div>
      )}

      {results && (
        <div className="results">
          <h3>Results</h3>
          <div className="results-grid">
            <div className="stat">
              <span className="stat-value">{results.total}</span>
              <span className="stat-label">Total</span>
            </div>
            <div className="stat">
              <span className="stat-value success">{results.successes}</span>
              <span className="stat-label">Success</span>
            </div>
            <div className="stat">
              <span className="stat-value failure">{results.failures}</span>
              <span className="stat-label">Failed</span>
            </div>
            <div className="stat">
              <span className="stat-value">{results.avgLatency}ms</span>
              <span className="stat-label">Avg Latency</span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

interface Endpoint {
  method: string
  path: string
  body?: unknown
}

function getEndpoints(pattern: 'browse' | 'order' | 'mixed'): Endpoint[] {
  const browse: Endpoint[] = [
    { method: 'GET', path: '/products' },
    { method: 'GET', path: '/products/prod-1' },
    { method: 'GET', path: '/products/prod-2' },
    { method: 'GET', path: '/inventory' },
    { method: 'GET', path: '/orders' },
  ]
  const order: Endpoint[] = [
    { method: 'POST', path: '/orders', body: { items: [{ product_id: 'prod-1', quantity: 1 }] } },
    { method: 'POST', path: '/orders', body: { items: [{ product_id: 'prod-2', quantity: 2 }] } },
    { method: 'GET', path: '/orders' },
  ]
  if (pattern === 'browse') return browse
  if (pattern === 'order') return order
  return [...browse, ...order]
}
