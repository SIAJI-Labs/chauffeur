import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { useTheme } from '@/lib/theme'

function MoonIcon() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" />
    </svg>
  )
}

function SunIcon() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2" />
      <path d="M12 20v2" />
      <path d="m4.93 4.93 1.41 1.41" />
      <path d="m17.66 17.66 1.41 1.41" />
      <path d="M2 12h2" />
      <path d="M20 12h2" />
      <path d="m6.34 17.66-1.41 1.41" />
      <path d="m19.07 4.93-1.41 1.41" />
    </svg>
  )
}

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
  return res.json()
}

async function stopContainer(name: string) {
  const res = await fetch(`/api/containers/${name}/stop`, { method: 'POST' })
  if (!res.ok) throw new Error('Failed to stop container')
  return res.json()
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
  const { theme, toggleTheme } = useTheme()

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
      <div className="container mx-auto max-w-4xl p-6 space-y-6">
        <div className="space-y-2">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-4 w-64" />
        </div>
        <div className="grid grid-cols-4 gap-4">
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
        </div>
        <Skeleton className="h-64" />
      </div>
    )
  }

  return (
    <div className="container mx-auto max-w-4xl p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold">Chauffeur Panel</h1>
          <p className="text-sm text-muted-foreground">Database Container Management</p>
        </div>
        <Button variant="ghost" size="icon" onClick={toggleTheme}>
          {theme === 'dark' ? <SunIcon /> : <MoonIcon />}
        </Button>
      </div>

      <div className="grid grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total Containers</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{containers?.length ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-[--color-muted-foreground]">Running</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-[--color-primary]">{runningCount}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-[--color-muted-foreground]">Stopped</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-[--color-muted-foreground]">{stoppedCount}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-[--color-muted-foreground]">Backups</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{backups?.length ?? 0}</div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Containers</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {containers?.length === 0 ? (
            <p className="text-sm text-[--color-muted-foreground]">
              No containers configured. Create one using: <code className="text-xs bg-[--color-muted] px-1 py-0.5 rounded">chauf podman create</code>
            </p>
          ) : (
            <div className="space-y-3">
              {containers?.map((container) => (
                <div
                  key={container.name}
                  className="flex items-center justify-between p-3 rounded-lg border border-[--color-border] bg-[--color-background]"
                >
                  <div className="space-y-1">
                    <div className="font-medium">{container.name}</div>
                    <div className="flex items-center gap-3 text-xs text-muted-foreground">
                      <span className={container.status === 'running' ? 'text-green-500' : 'text-muted-foreground'}>
                        {container.status === 'running' ? '●' : '○'} {container.status}
                      </span>
                      <span>{container.engine}</span>
                      <span className="font-mono">:{container.hostPort}</span>
                    </div>
                  </div>
                  <div>
                    {loadingContainer === container.name ? (
                      <Button disabled variant="outline" size="sm">
                        <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                        </svg>
                      </Button>
                    ) : container.status === 'running' ? (
                      <Button size="sm" onClick={() => handleStop(container.name)}>
                        Stop
                      </Button>
                    ) : (
                      <Button variant="outline" size="sm" onClick={() => handleStart(container.name)}>
                        Start
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {backups && backups.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Recent Backups</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {backups.slice(0, 5).map((backup) => (
              <div
                key={backup.name}
                className="flex items-center justify-between p-2 rounded border border-[--color-border]"
              >
                <div className="space-y-0.5">
                  <div className="text-sm font-mono">{backup.name}</div>
                  <div className="text-xs text-[--color-muted-foreground]">
                    {formatBytes(backup.size)} • {new Date(backup.createdAt).toLocaleDateString()}
                  </div>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
