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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Circle,
  HardDrive,
  Pause,
  Play,
  Server,
} from "lucide-react"

export const Route = createFileRoute("/containers")({
  component: ContainersPage,
})

interface Container {
  name: string
  engine: string
  status: "running" | "stopped"
  hostPort: number
  createdAt: string
}

async function fetchContainers(): Promise<Array<Container>> {
  const res = await fetch("/api/containers")
  if (!res.ok) throw new Error("Failed to fetch containers")
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

function ContainersPage() {
  const queryClient = useQueryClient()
  
  const { data: containers = [], isLoading, error } = useQuery({
    queryKey: ["containers"],
    queryFn: fetchContainers,
    refetchInterval: 5000,
  })

  const startMutation = useMutation({
    mutationFn: startContainer,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["containers"] }),
  })

  const stopMutation = useMutation({
    mutationFn: stopContainer,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["containers"] }),
  })

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
                <BreadcrumbPage className="font-normal text-foreground">Container List</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4">
          <div className="flex items-center justify-between">
            <h1 className="text-lg font-semibold">Database Containers</h1>
          </div>
          
          {isLoading && (
            <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
              <Server className="size-8 text-muted-foreground mb-2" />
              <p className="text-sm text-muted-foreground">Loading containers...</p>
            </div>
          )}
          
          {error && (
            <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
              <p className="text-sm text-destructive">Error: {error.message}</p>
            </div>
          )}
          
          {!isLoading && !error && containers.length === 0 && (
            <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
              <Server className="size-8 text-muted-foreground mb-2" />
              <p className="text-sm text-muted-foreground">No containers found</p>
              <p className="text-xs text-muted-foreground mt-1">Create a container using the chauffeur CLI</p>
            </div>
          )}
          
          {!isLoading && !error && containers.length > 0 && (
            <div className="rounded-xl border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Status</TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead>Engine</TableHead>
                    <TableHead>Host Port</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {containers.map((container) => (
                    <TableRow key={container.name}>
                      <TableCell>
                        <Badge variant={container.status === "running" ? "default" : "secondary"}>
                          <Circle className={`size-2 mr-1 ${container.status === "running" ? "fill-green-500 text-green-500" : "fill-gray-400 text-gray-400"}`} />
                          {container.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-medium">{container.name}</TableCell>
                      <TableCell>{container.engine}</TableCell>
                      <TableCell>{container.hostPort || "-"}</TableCell>
                      <TableCell className="text-muted-foreground">
                        {new Date(container.createdAt).toLocaleDateString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          {container.status === "running" ? (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => stopMutation.mutate(container.name)}
                              disabled={stopMutation.isPending}
                            >
                              <Pause className="size-4" />
                            </Button>
                          ) : (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => startMutation.mutate(container.name)}
                              disabled={startMutation.isPending}
                            >
                              <Play className="size-4" />
                            </Button>
                          )}
                          <Button variant="ghost" size="sm">
                            <a href={`/containers/${container.name}`}>
                              <HardDrive className="size-4" />
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