export default function Home() {
  return (
    <main style={{ padding: '2rem', fontFamily: 'system-ui, sans-serif' }}>
      <h1>Welcome to ArnoCodes</h1>
      <p>Monorepo with Backend, Frontend, and AI services</p>
      
      <div style={{ marginTop: '2rem' }}>
        <h2>Services:</h2>
        <ul>
          <li>Backend - Go API</li>
          <li>Frontend - Next.js + React + TypeScript</li>
          <li>AI - Python RAG Service</li>
        </ul>
      </div>
    </main>
  )
}
