import { createFileRoute, useParams } from "@tanstack/react-router"
import { useEffect, useRef, useState } from "react"
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip"
import {
  Circle,
  Copy,
  HardDrive,
  Pause,
  Play,
  Server,
  Terminal,
} from "lucide-react"

export const Route = createFileRoute("/containers/$name")({
  component: ContainerDetailPage,
})

interface ContainerDetail {
  name: string
  engine: string
  status: "running" | "stopped"
  hostPort: number
  createdAt: string
  config: {
    databaseUser: string
    databaseName: string
    databasePassword: string
  }
}

async function fetchContainer(name: string): Promise<ContainerDetail> {
  const res = await fetch(`/api/containers/${name}`)
  if (!res.ok) throw new Error("Failed to fetch container")
  return res.json()
}

async function startContainer(name: string): Promise<void> {
  const res = await fetch(`/api/containers/${name}/start`, { method: "POST" })
  if (!res.ok) throw new Error("Failed to start container")
}

async function stopContainer(name: string): Promise<void> {
  const res = await fetch(`/api/containers/${name}/stop`, { method: "POST" })
  if (!res.ok) throw new Error("Failed to stop container")
}

function ContainerDetailPage() {
  const { name } = useParams({ from: "/containers/$name" })
  const queryClient = useQueryClient()
  const [logs, setLogs] = useState<string>("")
  const logsRef = useRef<HTMLPreElement>(null)

  const { data: container, isLoading, error } = useQuery({
    queryKey: ["container", name],
    queryFn: () => fetchContainer(name),
    refetchInterval: 5000,
  })

  const startMutation = useMutation({
    mutationFn: () => startContainer(name),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["container", name] }),
  })

  const stopMutation = useMutation({
    mutationFn: () => stopContainer(name),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["container", name] }),
  })

  useEffect(() => {
    if (container?.status !== "running") return

    const eventSource = new EventSource(`/api/containers/${name}/logs`)

    eventSource.onmessage = (event) => {
      setLogs((prev) => {
        const newLogs = event.data.replace(/\\n/g, "\n").replace(/\\r/g, "\r")
        const lines = (prev + newLogs).split("\n").slice(-100)
        return lines.join("\n")
      })
    }

    eventSource.onerror = () => {
      eventSource.close()
    }

    return () => eventSource.close()
  }, [container?.status, name])

  useEffect(() => {
    if (logsRef.current) {
      logsRef.current.scrollTop = logsRef.current.scrollHeight
    }
  }, [logs])

  const [copiedField, setCopiedField] = useState<string | null>(null)

  const copyToClipboard = (text: string, field: string) => {
    navigator.clipboard.writeText(text)
    setCopiedField(field)
    setTimeout(() => setCopiedField(null), 2000)
  }

  if (isLoading) {
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
                  <BreadcrumbLink href="#">Podman Service</BreadcrumbLink>
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
                  <BreadcrumbLink href="/containers">Container List</BreadcrumbLink>
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
                  <BreadcrumbPage className="font-normal text-foreground">{name}</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </header>
          <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
            <Server className="size-8 text-muted-foreground mb-2" />
            <p className="text-sm text-muted-foreground">Loading container...</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    )
  }

  if (error || !container) {
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
                  <BreadcrumbLink href="#">Podman Service</BreadcrumbLink>
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
                  <BreadcrumbLink href="/containers">Container List</BreadcrumbLink>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </header>
          <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
            <p className="text-sm text-destructive">Error: {error?.message || "Container not found"}</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    )
  }

  return (
    <TooltipProvider>
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="#">Podman Service</BreadcrumbLink>
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
                <BreadcrumbLink href="/containers">Container List</BreadcrumbLink>
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
                <BreadcrumbPage className="font-normal text-foreground">{name}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <h1 className="text-lg font-semibold">{container.name}</h1>
              <Badge variant={container.status === "running" ? "default" : "secondary"}>
                <Circle className={`size-2 mr-1 ${container.status === "running" ? "fill-green-500 text-green-500" : "fill-gray-400 text-gray-400"}`} />
                {container.status}
              </Badge>
            </div>
            <div className="flex items-center gap-2">
              {container.status === "running" ? (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => stopMutation.mutate()}
                  disabled={stopMutation.isPending}
                >
                  <Pause className="size-4 mr-1" />
                  Stop
                </Button>
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => startMutation.mutate()}
                  disabled={startMutation.isPending}
                >
                  <Play className="size-4 mr-1" />
                  Start
                </Button>
              )}
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                  <HardDrive className="size-4" />
                  Container Info
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Engine</span>
                  <span className="font-medium">{container.engine}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Host Port</span>
                  <span className="font-medium">{container.hostPort || "-"}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Created</span>
                  <span className="font-medium">
                    {new Date(container.createdAt).toLocaleDateString()}
                  </span>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                  <Server className="size-4" />
                  Database Connection
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <div className="flex justify-between items-center">
                  <span className="text-muted-foreground">Host</span>
                  <div className="flex items-center gap-1">
                    <span className="font-medium">localhost:{container.hostPort}</span>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-6"
                          onClick={() => copyToClipboard(`localhost:${container.hostPort}`, "host")}
                        >
                          <Copy className={`size-3 ${copiedField === "host" ? "text-green-500" : ""}`} />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        {copiedField === "host" ? "Copied!" : "Copy"}
                      </TooltipContent>
                    </Tooltip>
                  </div>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-muted-foreground">Database</span>
                  <div className="flex items-center gap-1">
                    <span className="font-medium">{container.config?.databaseName || "app"}</span>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-6"
                          onClick={() => copyToClipboard(container.config?.databaseName || "app", "database")}
                        >
                          <Copy className={`size-3 ${copiedField === "database" ? "text-green-500" : ""}`} />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        {copiedField === "database" ? "Copied!" : "Copy"}
                      </TooltipContent>
                    </Tooltip>
                  </div>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-muted-foreground">User</span>
                  <div className="flex items-center gap-1">
                    <span className="font-medium">{container.config?.databaseUser || "-"}</span>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-6"
                          onClick={() => copyToClipboard(container.config?.databaseUser || "", "user")}
                        >
                          <Copy className={`size-3 ${copiedField === "user" ? "text-green-500" : ""}`} />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        {copiedField === "user" ? "Copied!" : "Copy"}
                      </TooltipContent>
                    </Tooltip>
                  </div>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-muted-foreground">Password</span>
                  <div className="flex items-center gap-1">
                    <span className="font-medium">{container.config?.databasePassword || "-"}</span>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-6"
                          onClick={() => copyToClipboard(container.config?.databasePassword || "", "password")}
                        >
                          <Copy className={`size-3 ${copiedField === "password" ? "text-green-500" : ""}`} />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        {copiedField === "password" ? "Copied!" : "Copy"}
                      </TooltipContent>
                    </Tooltip>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <Terminal className="size-4" />
                Logs
                {container.status !== "running" && (
                  <Badge variant="secondary" className="ml-2 text-xs">
                    Container must be running
                  </Badge>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <pre
                ref={logsRef}
                className="h-64 overflow-auto rounded-md bg-muted p-2 text-xs font-mono leading-relaxed"
              >
                {logs || (container.status !== "running" ? "Logs will appear here when the container is running..." : "Waiting for logs...")}
              </pre>
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
    </TooltipProvider>
  )
}