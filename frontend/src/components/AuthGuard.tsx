'use client'

import { useEffect, useState } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { supabase } from '@/lib/supabase'

const PUBLIC_PATHS = new Set(['/login', '/signup'])

export default function AuthGuard({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()
  const [ready, setReady] = useState(false)

  useEffect(() => {
    let isMounted = true

    const verify = async () => {
      const {
        data: { session },
      } = await supabase.auth.getSession()

      if (!isMounted) return

      if (!session && !PUBLIC_PATHS.has(pathname)) {
        router.replace('/login')
        return
      }
      if (session && PUBLIC_PATHS.has(pathname)) {
        router.replace('/')
        return
      }
      setReady(true)
    }

    verify()

    const { data: listener } = supabase.auth.onAuthStateChange(() => {
      verify()
    })

    return () => {
      isMounted = false
      listener.subscription.unsubscribe()
    }
  }, [pathname, router])

  if (!ready) {
    return <main style={{ padding: '2rem', fontFamily: 'system-ui, sans-serif' }}>Loading...</main>
  }

  return <>{children}</>
}
