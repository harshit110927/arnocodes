import { supabase } from '@/lib/supabase'

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080'

type APIEnvelope<T> = {
  status: 'ok' | 'error'
  message: string
  data?: T
}

export type APIError = Error & {
  status?: number
  code?: string
  payload?: unknown
}

async function getAccessToken(): Promise<string> {
  const {
    data: { session },
  } = await supabase.auth.getSession()

  if (!session?.access_token) {
    if (typeof window !== 'undefined') {
      window.location.href = '/login'
    }
    const err = new Error('unauthenticated') as APIError
    err.status = 401
    err.code = 'UNAUTHORIZED'
    throw err
  }

  return session.access_token
}

export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = await getAccessToken()
  const headers = new Headers(init.headers ?? {})
  headers.set('Authorization', `Bearer ${token}`)

  if (!headers.has('Content-Type') && init.body) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers,
  })

  const contentType = response.headers.get('Content-Type') ?? ''
  const isJSON = contentType.includes('application/json')
  const payload = isJSON ? await response.json() : null

  if (!response.ok) {
    const err = new Error('request_failed') as APIError
    err.status = response.status
    err.payload = payload
    if (payload && typeof payload === 'object' && 'error' in payload && typeof payload.error === 'string') {
      err.code = payload.error
    }
    throw err
  }

  if (!payload) {
    return {} as T
  }

  if (typeof payload === 'object' && payload !== null && 'status' in payload) {
    const envelope = payload as APIEnvelope<T>
    return (envelope.data as T) ?? ({} as T)
  }

  return payload as T
}

export type ProfileStatus = {
  diagnostic_completed: boolean
  dashboard_unlocked: boolean
}

export type DashboardSummary = {
  streak_count: number
  questions_solved: number
  mastery_score: number
  topics_completed: number
  last_activity_at?: string
}

export type TopicMastery = {
  topic_id: string
  topic_name: string
  status: string
  mastery_score: number
}

export type ActivityItem = {
  source: string
  title: string
  difficulty?: string
  solved_at: string
}

export type PlatformConnection = {
  id: string
  platform: string
  platform_handle: string
  status: string
}

export type DashboardData = {
  summary: DashboardSummary
  topic_mastery: TopicMastery[]
  weak_topics: TopicMastery[]
  recent_activity: ActivityItem[]
}

export async function getProfileStatus() {
  return request<ProfileStatus>('/api/v1/profiles/me/status')
}

export async function getDashboard() {
  return request<DashboardData>('/api/v1/dashboard')
}

export type CourseTopic = {
  id: string
  name: string
  completion_status: string
  unlock_status: string
  mastery_score: number
}

export async function getCourse() {
  return request<CourseTopic[]>('/api/v1/course/structure')
}

export async function getPlatformConnections() {
  return request<PlatformConnection[]>('/api/v1/profiles/me/platform-connections')
}

export async function connectPlatform(platform: string, handle: string) {
  return request('/api/v1/profiles/me/platform-connections', {
    method: 'POST',
    body: JSON.stringify({ platform, handle }),
  })
}

export async function triggerPlatformSync() {
  return request('/api/v1/platform-sync/trigger', { method: 'POST' })
}

export async function logout() {
  await supabase.auth.signOut()
  window.location.href = '/login'
}
