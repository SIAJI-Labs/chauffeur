# Admin Panel Sidebar Layout Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add sidebar navigation layout to admin panel with placeholder items for future features (DNS, Services, etc.)

**Architecture:** Wrap existing dashboard in a sidebar layout component. Sidebar contains collapsible navigation with current and placeholder nav items. Theme toggle moves to sidebar header.

**Tech Stack:** React 19, Tailwind v4 (CSS variables), shadcn/ui pattern

---

### Task 1: Create sidebar layout component

**Files:**
- Create: `internal/panel-apps/src/components/layout/sidebar-layout.tsx`

**Step 1: Create the sidebar layout component**

```tsx
import { useState } from 'react'
import { cn } from '@/lib/utils'

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

interface NavItem {
  title: string
  href?: string
  disabled?: boolean
  icon?: React.ReactNode
}

const mainNavItems: NavItem[] = [
  { title: 'Dashboard', href: '/' },
  { title: 'Containers', href: '/containers' },
  { title: 'Backups', href: '/backups' },
  { title: 'Settings', href: '/settings' },
]

const futureNavItems: NavItem[] = [
  { title: 'DNS', disabled: true },
  { title: 'Services', disabled: true },
  { title: 'Other', disabled: true },
]

interface SidebarLayoutProps {
  children: React.ReactNode
}

export function SidebarLayout({ children }: SidebarLayoutProps) {
  const [collapsed, setCollapsed] = useState(false)
  const [theme, setTheme] = useState<'light' | 'dark'>('dark')

  const toggleTheme = () => setTheme(t => t === 'light' ? 'dark' : 'light')

  return (
    <div className="flex h-screen bg-[--color-background]">
      {/* Sidebar */}
      <aside
        className={cn(
          "flex flex-col border-r border-[--color-border] bg-[--color-card] transition-all duration-200",
          collapsed ? "w-16" : "w-64"
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-[--color-border]">
          {!collapsed && (
            <span className="font-semibold text-foreground">Chauffeur</span>
          )}
          <button
            onClick={() => setCollapsed(!collapsed)}
            className="p-1.5 rounded-md hover:bg-[--color-muted] text-muted-foreground"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              {collapsed ? (
                <path d="m9 18 6-6-6-6" />
              ) : (
                <path d="m15 18-6-6 6-6" />
              )}
            </svg>
          </button>
        </div>

        {/* Theme Toggle */}
        <div className="p-4 border-b border-[--color-border]">
          <button
            onClick={toggleTheme}
            className="p-2 rounded-md hover:bg-[--color-muted] text-muted-foreground w-full flex items-center justify-center"
            title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
          >
            {theme === 'dark' ? <SunIcon /> : <MoonIcon />}
          </button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-4 space-y-6 overflow-y-auto">
          {/* Main Nav */}
          <div className="space-y-1">
            {!collapsed && (
              <p className="px-2 text-xs font-medium text-muted-foreground uppercase tracking-wider mb-2">
                Main
              </p>
            )}
            {mainNavItems.map((item) => (
              <a
                key={item.title}
                href={item.href || '#'}
                className={cn(
                  "flex items-center gap-3 px-2 py-2 rounded-md text-sm transition-colors",
                  item.disabled
                    ? "text-muted-foreground cursor-not-allowed opacity-60"
                    : "text-foreground hover:bg-[--color-muted] hover:text-foreground",
                  collapsed && "justify-center"
                )}
                onClick={item.disabled ? (e) => e.preventDefault() : undefined}
              >
                <span className={cn(collapsed && "flex-1 text-center")}>
                  {item.title}
                </span>
              </a>
            ))}
          </div>

          {/* Future Items */}
          <div className="space-y-1">
            {!collapsed && (
              <p className="px-2 text-xs font-medium text-muted-foreground uppercase tracking-wider mb-2">
                Coming Soon
              </p>
            )}
            {futureNavItems.map((item) => (
              <div
                key={item.title}
                className={cn(
                  "flex items-center gap-3 px-2 py-2 rounded-md text-sm transition-colors text-muted-foreground cursor-not-allowed opacity-60",
                  collapsed && "justify-center"
                )}
              >
                <span className={cn(collapsed && "flex-1 text-center")}>
                  {item.title}
                </span>
              </div>
            ))}
          </div>
        </nav>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto">
        {children}
      </main>
    </div>
  )
}
```

**Step 2: Verify file creation**

Run: `ls -la internal/panel-apps/src/components/layout/`
Expected: `sidebar-layout.tsx` exists

---

### Task 2: Update App.tsx to use sidebar layout

**Files:**
- Modify: `internal/panel-apps/src/App.tsx`

**Step 1: Update App.tsx imports and wrap with SidebarLayout**

Replace the first 31 lines (imports through MoonIcon/SunIcon functions) with:

```tsx
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { SidebarLayout } from '@/components/layout/sidebar-layout'

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
      <SidebarLayout>
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
      </SidebarLayout>
    )
  }

  return (
    <SidebarLayout>
      <div className="container mx-auto max-w-4xl p-6 space-y-6">
        {/* ... rest of existing return content ... */}
```

**Step 2: Wrap the return statement content**

Wrap the return content (lines 139-260 in original) with `<SidebarLayout>` tags, removing the old header with theme toggle since it's now in the sidebar.

**Step 3: Remove unused theme import and toggle**

Remove from App.tsx:
- `import { useTheme } from '@/lib/theme'` (line 6)
- `const { theme, toggleTheme } = useTheme()` (line 81)
- Theme toggle button in the header (lines 146-148)

---

### Task 3: Create layout directory and ensure build

**Step 1: Create the layout directory if needed**

Run: `mkdir -p internal/panel-apps/src/components/layout`

**Step 2: Verify build**

Run: `cd internal/panel-apps && npm run build`
Expected: Successful build with no TypeScript errors

---

### Task 4: Commit changes

**Step 1: Stage and commit**

```bash
git add internal/panel-apps/src/components/layout/ internal/panel-apps/src/App.tsx
git commit -m "feat(panel): add sidebar navigation layout

- Add collapsible sidebar with main navigation
- Move theme toggle to sidebar header
- Add placeholder nav items for future features (DNS, Services, Other)
- Wrap dashboard in sidebar layout"
```
