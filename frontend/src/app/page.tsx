'use client'

import { FormEvent, useEffect, useMemo, useState } from 'react'
import {
  APIError,
  DashboardData,
  PlatformConnection,
  getCourse,
  getDashboard,
  getPlatformConnections,
  getProfileStatus,
  connectPlatform,
  triggerPlatformSync,
  logout,
} from '@/lib/api'

type ThemeMode = 'midnight' | 'forest' | 'slate' | 'obsidian' | 'aurora'

const THEMES: ThemeMode[] = ['midnight', 'forest', 'slate', 'obsidian', 'aurora']

export default function Home() {
  const [theme, setTheme] = useState<ThemeMode>('midnight')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [status, setStatus] = useState<{ diagnostic_completed: boolean; dashboard_unlocked: boolean } | null>(null)
  const [dashboard, setDashboard] = useState<DashboardData | null>(null)
  const [connections, setConnections] = useState<PlatformConnection[]>([])
  const [courseTopics, setCourseTopics] = useState<Array<{ id: string; name: string; completion_status: string }>>([])
  const [platform, setPlatform] = useState('leetcode')
  const [handle, setHandle] = useState('')

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
  }, [theme])

  const load = async () => {
    setLoading(true)
    setError('')

    try {
      const profileStatus = await getProfileStatus()
      setStatus(profileStatus)

      const [courseResp, platformResp] = await Promise.all([getCourse(), getPlatformConnections()])
      setCourseTopics(courseResp ?? [])
      setConnections(platformResp)

      if (profileStatus.dashboard_unlocked) {
        const dashboardResp = await getDashboard()
        setDashboard(dashboardResp)
      } else {
        setDashboard(null)
      }
    } catch (e) {
      const apiErr = e as APIError
      if (apiErr.code === 'DIAGNOSTIC_REQUIRED') {
        setStatus({ diagnostic_completed: false, dashboard_unlocked: false })
      } else {
        setError(`Failed to load data${apiErr.status ? ` (HTTP ${apiErr.status})` : ''}`)
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  const summaryCards = useMemo(() => {
    if (!dashboard?.summary) return []
    return [
      { label: 'Streak', value: dashboard.summary.streak_count },
      { label: 'Solved', value: dashboard.summary.questions_solved },
      { label: 'Mastery', value: `${Math.round(dashboard.summary.mastery_score)}%` },
      { label: 'Completed', value: dashboard.summary.topics_completed },
    ]
  }, [dashboard])

  const onConnect = async (e: FormEvent) => {
    e.preventDefault()
    await connectPlatform(platform, handle)
    setHandle('')
    await load()
  }

  const onSync = async () => {
    await triggerPlatformSync()
    await load()
  }

  return (
    <main className='ac-shell'>
      <header className='ac-topbar'>
        <div className='ac-logo'>AlgoIQ</div>
        <div className='ac-topbar-actions'>
          <select value={theme} onChange={(e) => setTheme(e.target.value as ThemeMode)}>
            {THEMES.map((themeOption) => (
              <option key={themeOption} value={themeOption}>
                {themeOption}
              </option>
            ))}
          </select>
          <button onClick={() => void load()}>Refresh</button>
          <button onClick={() => void logout()}>Logout</button>
        </div>
      </header>

      <section className='ac-layout'>
        <aside className='ac-sidebar'>
          <h3>Progress</h3>
          <p>Diagnostic: {status?.diagnostic_completed ? 'Completed' : 'Pending'}</p>
          <p>Dashboard: {status?.dashboard_unlocked ? 'Unlocked' : 'Locked'}</p>
          <h4>Topics ({courseTopics.length})</h4>
          <ul>
            {courseTopics.slice(0, 10).map((topic) => (
              <li key={topic.id}>
                {topic.name} <span>{topic.completion_status}</span>
              </li>
            ))}
          </ul>
        </aside>

        <div className='ac-content'>
          <h1>ArnoCodes Dashboard</h1>
          <p className='muted'>Reference UI adapted from `frontend/reference.html` and wired to backend APIs.</p>

          {loading && <p>Loading dashboard...</p>}
          {error && <p className='error'>{error}</p>}

          {!loading && !status?.dashboard_unlocked && (
            <div className='ac-card lock'>
              <h2>Dashboard Locked</h2>
              <p>Complete the diagnostic flow first. The backend currently returns `DIAGNOSTIC_REQUIRED` until unlock.</p>
            </div>
          )}

          {!!summaryCards.length && (
            <div className='ac-grid'>
              {summaryCards.map((card) => (
                <div key={card.label} className='ac-card'>
                  <p className='muted'>{card.label}</p>
                  <h2>{card.value}</h2>
                </div>
              ))}
            </div>
          )}

          <div className='ac-grid two'>
            <div className='ac-card'>
              <h3>Weak Topics</h3>
              <ul>
                {(dashboard?.weak_topics ?? []).slice(0, 6).map((topic) => (
                  <li key={topic.topic_id}>
                    {topic.topic_name} <span>{Math.round(topic.mastery_score)}%</span>
                  </li>
                ))}
              </ul>
            </div>

            <div className='ac-card'>
              <h3>Recent Activity</h3>
              <ul>
                {(dashboard?.recent_activity ?? []).slice(0, 6).map((item, idx) => (
                  <li key={`${item.title}-${idx}`}>
                    {item.title} <span>{item.source}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>

          <div className='ac-card'>
            <h3>Platform Connections</h3>
            <form className='ac-form' onSubmit={onConnect}>
              <select value={platform} onChange={(e) => setPlatform(e.target.value)}>
                <option value='leetcode'>leetcode</option>
                <option value='gfg'>gfg</option>
                <option value='codeforces'>codeforces</option>
                <option value='hackerrank'>hackerrank</option>
                <option value='codechef'>codechef</option>
              </select>
              <input value={handle} onChange={(e) => setHandle(e.target.value)} placeholder='handle' required />
              <button type='submit'>Connect</button>
              <button type='button' onClick={() => void onSync()}>Trigger Sync</button>
            </form>
            <ul>
              {connections.map((conn) => (
                <li key={conn.id}>
                  {conn.platform} - {conn.platform_handle} ({conn.status})
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>
    </main>
  )
}
