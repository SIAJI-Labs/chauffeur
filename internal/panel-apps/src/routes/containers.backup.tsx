import * as React from "react"
import { createFileRoute } from "@tanstack/react-router"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { AppSidebar } from "@/components/app-sidebar"
import { Separator } from "@/components/ui/separator"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { HardDrive, Plus, RotateCcw } from "lucide-react"

export const Route = createFileRoute("/containers/backup")({
  component: BackupPage,
})

interface Backup {
  name: string
  container: string
  size: number
  createdAt: string
}

interface Container {
  name: string
  engine: string
  status: "running" | "stopped"
  hostPort: number
  createdAt: string
}

async function fetchBackups(): Promise<Backup[]> {
  const res = await fetch("/api/backups")
  if (!res.ok) throw new Error("Failed to fetch backups")
  return res.json()
}

async function fetchContainers(): Promise<Container[]> {
  const res = await fetch("/api/containers")
  if (!res.ok) throw new Error("Failed to fetch containers")
  return res.json()
}

async function createBackup(container: string, description: string): Promise<void> {
  const res = await fetch("/api/backups", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ container, description }),
  })
  if (!res.ok) throw new Error("Failed to create backup")
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

function BackupPage() {
  const queryClient = useQueryClient()
  const [isCreateOpen, setIsCreateOpen] = React.useState(false)
  const [selectedContainer, setSelectedContainer] = React.useState("")
  const [description, setDescription] = React.useState("")

  const { data: backups = [], isLoading: backupsLoading } = useQuery({
    queryKey: ["backups"],
    queryFn: fetchBackups,
    refetchInterval: 30000,
  })

  const { data: containers = [], isLoading: containersLoading } = useQuery({
    queryKey: ["containers"],
    queryFn: fetchContainers,
  })

  const createMutation = useMutation({
    mutationFn: () => createBackup(selectedContainer, description),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["backups"] })
      setIsCreateOpen(false)
      setSelectedContainer("")
      setDescription("")
    },
  })

  const runningContainers = containers.filter((c) => c.status === "running")

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="/containers">Podman Service</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator className="hidden md:block">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  className="lucide lucide-chevron-right"
                >
                  <path d="m9 18 6-6-6-6" />
                </svg>
              </BreadcrumbSeparator>
              <BreadcrumbItem>
                <BreadcrumbPage className="font-normal text-foreground">Backup</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4">
          <div className="flex items-center justify-between">
            <h1 className="text-lg font-semibold">Database Backups</h1>
            <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
              <DialogTrigger>
                <Button size="sm">
                  <Plus className="size-4 mr-1" />
                  Create Backup
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Create New Backup</DialogTitle>
                  <DialogDescription>
                    Select a running container to backup its database.
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Container</label>
                    <select
                      className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
                      value={selectedContainer}
                      onChange={(e) => setSelectedContainer(e.target.value)}
                    >
                      <option value="">Select a container...</option>
                      {runningContainers.map((c) => (
                        <option key={c.name} value={c.name}>
                          {c.name} ({c.engine})
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Description (optional)</label>
                    <Input
                      placeholder="e.g., before schema migration"
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsCreateOpen(false)}>
                    Cancel
                  </Button>
                  <Button
                    onClick={() => createMutation.mutate()}
                    disabled={!selectedContainer || createMutation.isPending}
                  >
                    {createMutation.isPending ? "Creating..." : "Create Backup"}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          {backupsLoading || containersLoading ? (
            <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
              <HardDrive className="size-8 text-muted-foreground mb-2" />
              <p className="text-sm text-muted-foreground">Loading backups...</p>
            </div>
          ) : backups.length === 0 ? (
            <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
              <HardDrive className="size-8 text-muted-foreground mb-2" />
              <p className="text-sm text-muted-foreground">No backups found</p>
              <p className="text-xs text-muted-foreground mt-1">
                Create a backup from a running container
              </p>
            </div>
          ) : (
            <div className="rounded-xl border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Container</TableHead>
                    <TableHead>Size</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {backups.map((backup) => (
                    <TableRow key={backup.name}>
                      <TableCell className="font-medium">{backup.name}</TableCell>
                      <TableCell>
                        <Badge variant="outline">{backup.container}</Badge>
                      </TableCell>
                      <TableCell>{formatSize(backup.size)}</TableCell>
                      <TableCell className="text-muted-foreground">
                        {new Date(backup.createdAt).toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button variant="ghost" size="sm">
                            <a href={`/containers/restore?backup=${encodeURIComponent(backup.name)}`}>
                              <RotateCcw className="size-4" />
                            </a>
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}