import { useState, useEffect, useRef } from 'react'

export function usePolling<T>(fetcher: () => Promise<T>, intervalMs = 5000) {
  const [data, setData] = useState<T | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const fetcherRef = useRef(fetcher)
  fetcherRef.current = fetcher

  useEffect(() => {
    let active = true

    async function poll() {
      try {
        const result = await fetcherRef.current()
        if (active) {
          setData(result)
          setError(null)
        }
      } catch (e) {
        if (active) setError(e instanceof Error ? e.message : 'Unknown error')
      } finally {
        if (active) setLoading(false)
      }
    }

    poll()
    const id = setInterval(poll, intervalMs)
    return () => { active = false; clearInterval(id) }
  }, [intervalMs])

  return { data, error, loading }
}
