import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'

interface Container {
  name: string
  engine: string
  status: 'running' | 'stopped'
  hostPort: number
  createdAt: string
}

interface Backup {
  name: string
  container: string
  size: number
  createdAt: string
}

async function fetchContainers(): Promise<Container[]> {
  const res = await fetch('/api/containers')
  if (!res.ok) throw new Error('Failed to fetch containers')
  return res.json()
}

async function fetchBackups(): Promise<Backup[]> {
  const res = await fetch('/api/backups')
  if (!res.ok) throw new Error('Failed to fetch backups')
  return res.json()
}

async function startContainer(name: string) {
  const res = await fetch(`/api/containers/${name}/start`, { method: 'POST' })
  if (!res.ok) throw new Error('Failed to start container')
}

async function stopContainer(name: string) {
  const res = await fetch(`/api/containers/${name}/stop`, { method: 'POST' })
  if (!res.ok) throw new Error('Failed to stop container')
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

export function App() {
  const [loadingContainer, setLoadingContainer] = useState<string | null>(null)

  const { data: containers, isLoading: containersLoading, refetch } = useQuery({
    queryKey: ['containers'],
    queryFn: fetchContainers,
    refetchInterval: 5000,
  })

  const { data: backups } = useQuery({
    queryKey: ['backups'],
    queryFn: fetchBackups,
  })

  const runningCount = containers?.filter(c => c.status === 'running').length ?? 0
  const stoppedCount = containers?.filter(c => c.status === 'stopped').length ?? 0

  const handleStart = async (name: string) => {
    setLoadingContainer(name)
    try {
      await startContainer(name)
      refetch()
    } catch (err) {
      console.error(err)
    } finally {
      setLoadingContainer(null)
    }
  }

  const handleStop = async (name: string) => {
    setLoadingContainer(name)
    try {
      await stopContainer(name)
      refetch()
    } catch (err) {
      console.error(err)
    } finally {
      setLoadingContainer(null)
    }
  }

  if (containersLoading) {
    return (
      <div className="loading">
        <div>Loading...</div>
      </div>
    )
  }

  return (
    <div className="app">
      <header className="header">
        <h1>Chauffeur Panel</h1>
        <p className="subtitle">Database Container Management</p>
      </header>

      <main className="main">
        <section className="stats">
          <div className="stat-card">
            <div className="stat-value">{containers?.length ?? 0}</div>
            <div className="stat-label">Total Containers</div>
          </div>
          <div className="stat-card running">
            <div className="stat-value">{runningCount}</div>
            <div className="stat-label">Running</div>
          </div>
          <div className="stat-card stopped">
            <div className="stat-value">{stoppedCount}</div>
            <div className="stat-label">Stopped</div>
          </div>
          <div className="stat-card">
            <div className="stat-value">{backups?.length ?? 0}</div>
            <div className="stat-label">Backups</div>
          </div>
        </section>

        <section className="section">
          <h2>Containers</h2>
          {containers?.length === 0 ? (
            <div className="empty">
              <p>No containers configured.</p>
              <p className="hint">Create a container using: chauf podman create</p>
            </div>
          ) : (
            <div className="container-list">
              {containers?.map(container => (
                <div key={container.name} className="container-card">
                  <div className="container-info">
                    <div className="container-name">{container.name}</div>
                    <div className="container-meta">
                      <span className={`status ${container.status}`}>
                        {container.status === 'running' ? '●' : '○'} {container.status}
                      </span>
                      <span className="engine">{container.engine}</span>
                      <span className="port">:{container.hostPort}</span>
                    </div>
                  </div>
                  <div className="container-actions">
                    {loadingContainer === container.name ? (
                      <button disabled className="btn loading">
                        <span className="spinner"></span>
                      </button>
                    ) : container.status === 'running' ? (
                      <button onClick={() => handleStop(container.name)} className="btn stop">
                        Stop
                      </button>
                    ) : (
                      <button onClick={() => handleStart(container.name)} className="btn start">
                        Start
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        {backups && backups.length > 0 && (
          <section className="section">
            <h2>Recent Backups</h2>
            <div className="backup-list">
              {backups.slice(0, 5).map(backup => (
                <div key={backup.name} className="backup-item">
                  <div className="backup-info">
                    <div className="backup-name">{backup.name}</div>
                    <div className="backup-meta">
                      <span>{formatBytes(backup.size)}</span>
                      <span>{new Date(backup.createdAt).toLocaleDateString()}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </section>
        )}
      </main>
    </div>
  )
}
