import { Link, createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Circle, HardDrive, Pause, Play, Server } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { OperationFeedback } from "@/components/operation-feedback";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute("/containers/_index")({
  component: ContainersPage,
});

interface Container {
  name: string;
  engine: string;
  status: "running" | "stopped";
  hostPort: number;
  createdAt: string;
}

async function fetchContainers(): Promise<Array<Container>> {
  const res = await fetch("/api/containers");
  if (!res.ok) throw new Error("Failed to fetch containers");
  return res.json();
}

async function startContainer(name: string): Promise<void> {
  const res = await fetch(`/api/containers/${name}/start`, { method: "POST" });
  if (!res.ok) throw new Error("Failed to start container");
}

async function stopContainer(name: string): Promise<void> {
  const res = await fetch(`/api/containers/${name}/stop`, { method: "POST" });
  if (!res.ok) throw new Error("Failed to stop container");
}

function ContainersPage() {
  const queryClient = useQueryClient();
  const [feedback, setFeedback] = useState<{
    status: "loading" | "success" | "error";
    message: string;
  } | null>(null);
  const [retryOperation, setRetryOperation] = useState<{
    type: "start" | "stop";
    name: string;
  } | null>(null);

  const {
    data: containers = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ["containers"],
    queryFn: fetchContainers,
    refetchInterval: 5000,
  });

  const startMutation = useMutation({
    mutationFn: startContainer,
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["containers"] }),
  });

  const stopMutation = useMutation({
    mutationFn: stopContainer,
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["containers"] }),
  });

  const runOperation = (type: "start" | "stop", name: string) => {
    const mutation = type === "start" ? startMutation : stopMutation;
    setRetryOperation({ type, name });
    setFeedback({
      status: "loading",
      message: `${type === "start" ? "Starting" : "Stopping"} ${name}...`,
    });
    mutation.mutate(name, {
      onSuccess: () => {
        setFeedback({
          status: "success",
          message: `${name} ${type === "start" ? "started" : "stopped"}.`,
        });
        setRetryOperation(null);
      },
      onError: (operationError) => {
        setFeedback({
          status: "error",
          message: `Unable to ${type} ${name}: ${operationError.message}`,
        });
      },
    });
  };

  return (
    <AppShell
      title="Containers"
      breadcrumbs={[{ label: "Services" }, { label: "Containers" }]}
      contentId="containers-content"
      skipLabel="Skip to containers"
      shortcuts={
        <Button
          variant="outline"
          size="sm"
          className="header-button"
          render={<Link to="/containers/backup" />}
        >
          <HardDrive aria-hidden="true" /> <span>Backups</span>
        </Button>
      }
    >
      <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-4">
        {feedback ? (
          <OperationFeedback
            {...feedback}
            onDismiss={() => setFeedback(null)}
            onRetry={
              retryOperation
                ? () => runOperation(retryOperation.type, retryOperation.name)
                : undefined
            }
          />
        ) : null}
        {isLoading && (
          <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
            <Server className="size-8 text-muted-foreground mb-2" />
            <p className="text-sm text-muted-foreground">
              Loading containers...
            </p>
          </div>
        )}

        {error && (
          <OperationFeedback
            status="error"
            message={`Unable to load containers: ${error.message}`}
            onRetry={() =>
              queryClient.refetchQueries({ queryKey: ["containers"] })
            }
          />
        )}

        {!isLoading && !error && containers.length === 0 && (
          <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
            <Server className="size-8 text-muted-foreground mb-2" />
            <p className="text-sm text-muted-foreground">No containers found</p>
            <p className="text-xs text-muted-foreground mt-1">
              Create a container using the chauffeur CLI
            </p>
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
                      <Badge
                        variant={
                          container.status === "running"
                            ? "default"
                            : "secondary"
                        }
                      >
                        <Circle
                          className={`size-2 mr-1 ${container.status === "running" ? "fill-green-500 text-green-500" : "fill-gray-400 text-gray-400"}`}
                        />
                        {container.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-medium">
                      {container.name}
                    </TableCell>
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
                            aria-label={`Stop ${container.name}`}
                            title={`Stop ${container.name}`}
                            aria-busy={stopMutation.isPending}
                            onClick={() => runOperation("stop", container.name)}
                            disabled={stopMutation.isPending}
                          >
                            <Pause className="size-4" />
                          </Button>
                        ) : (
                          <Button
                            variant="ghost"
                            size="sm"
                            aria-label={`Start ${container.name}`}
                            title={`Start ${container.name}`}
                            aria-busy={startMutation.isPending}
                            onClick={() =>
                              runOperation("start", container.name)
                            }
                            disabled={startMutation.isPending}
                          >
                            <Play className="size-4" />
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          aria-label={`Open ${container.name}`}
                          title={`Open ${container.name}`}
                        >
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
    </AppShell>
  );
}
