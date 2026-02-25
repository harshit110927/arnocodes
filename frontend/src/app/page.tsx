'use client'

import { useState } from 'react'
import { supabase } from '@/lib/supabase'
import { apiFetch } from '@/lib/api'

export default function Home() {
  const [apiResult, setAPIResult] = useState<string>('')

  const logout = async () => {
    await supabase.auth.signOut()
    window.location.href = '/login'
  }

  const checkProfile = async () => {
    try {
      const response = await apiFetch('/api/v1/profiles/me/status')
      const data = await response.json()
      setAPIResult(JSON.stringify(data, null, 2))
    } catch (error) {
      setAPIResult(`API call failed: ${String(error)}`)
    }
  }

  return (
    <main style={{ padding: '2rem', fontFamily: 'system-ui, sans-serif' }}>
      <h1>Welcome to ArnoCodes</h1>
      <p>Authenticated frontend using Supabase session + JWT forwarding.</p>

      <div style={{ marginTop: '1rem', display: 'flex', gap: '0.5rem' }}>
        <button onClick={checkProfile}>Call /api/v1/profiles/me/status</button>
        <button onClick={logout}>Logout</button>
      </div>

      {apiResult && (
        <pre style={{ marginTop: '1rem', background: '#111', color: '#eee', padding: '1rem' }}>{apiResult}</pre>
      )}
    </main>
  )
}
