import { createFileRoute, useParams } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Circle,
  Copy,
  Eye,
  EyeOff,
  HardDrive,
  Pause,
  Play,
  Server,
  Terminal,
} from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { OperationFeedback } from "@/components/operation-feedback";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

const LOG_FOLLOW_RESUME_THRESHOLD_PX = 16;

export const Route = createFileRoute("/containers/$name")({
  component: ContainerDetailPage,
});

interface ContainerDetail {
  name: string;
  engine: string;
  status: "running" | "stopped";
  hostPort: number;
  createdAt: string;
  config: {
    databaseUser: string;
    databaseName: string;
    databasePassword: string;
  };
}

async function fetchContainer(name: string): Promise<ContainerDetail> {
  const res = await fetch(`/api/containers/${name}`);
  if (!res.ok) throw new Error("Failed to fetch container");
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

function ContainerDetailPage() {
  const { name } = useParams({ from: "/containers/$name" });
  const queryClient = useQueryClient();
  const [logs, setLogs] = useState<string>("");
  const [isFollowingLogs, setIsFollowingLogs] = useState(true);
  const [logsDisconnected, setLogsDisconnected] = useState(false);
  const [logConnectionNonce, setLogConnectionNonce] = useState(0);
  const [operationFeedback, setOperationFeedback] = useState<{
    status: "loading" | "success" | "error";
    message: string;
  } | null>(null);
  const [retryOperation, setRetryOperation] = useState<"start" | "stop" | null>(
    null,
  );
  const logsContainerRef = useRef<HTMLPreElement>(null);

  const scrollLogsToBottom = () => {
    if (logsContainerRef.current) {
      logsContainerRef.current.scrollTop =
        logsContainerRef.current.scrollHeight;
    }
  };

  const handleLogsScroll = () => {
    if (!logsContainerRef.current) return;

    const { scrollTop, scrollHeight, clientHeight } = logsContainerRef.current;
    const isNearBottom =
      scrollHeight - scrollTop - clientHeight <= LOG_FOLLOW_RESUME_THRESHOLD_PX;

    setIsFollowingLogs((prev) => (prev === isNearBottom ? prev : isNearBottom));
  };

  const {
    data: container,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["container", name],
    queryFn: () => fetchContainer(name),
    refetchInterval: 5000,
  });

  const startMutation = useMutation({
    mutationFn: () => startContainer(name),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["container", name] }),
  });

  const stopMutation = useMutation({
    mutationFn: () => stopContainer(name),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["container", name] }),
  });

  useEffect(() => {
    setLogs("");
    setIsFollowingLogs(true);
    setLogsDisconnected(false);
  }, [name]);

  useEffect(() => {
    if (container?.status !== "running") return;

    const eventSource = new EventSource(`/api/containers/${name}/logs`);

    eventSource.onmessage = (event) => {
      setLogsDisconnected(false);
      setLogs((prev) => {
        const newLogs = event.data.replace(/\\n/g, "\n").replace(/\\r/g, "\r");
        const lines = (prev + newLogs).split("\n").slice(-100);
        return lines.join("\n");
      });
    };

    eventSource.onerror = () => {
      setLogsDisconnected(true);
      eventSource.close();
    };

    return () => eventSource.close();
  }, [container?.status, logConnectionNonce, name]);

  useEffect(() => {
    if (isFollowingLogs) {
      scrollLogsToBottom();
    }
  }, [isFollowingLogs, logs]);

  const [copiedField, setCopiedField] = useState<string | null>(null);
  const [copyStatus, setCopyStatus] = useState("");
  const [isPasswordVisible, setIsPasswordVisible] = useState(false);

  const copyToClipboard = (text: string, field: string) => {
    void navigator.clipboard
      .writeText(text)
      .then(() => {
        setCopiedField(field);
        setCopyStatus(`${field} copied to clipboard.`);
        setTimeout(() => {
          setCopiedField(null);
          setCopyStatus("");
        }, 2000);
      })
      .catch(() => {
        setCopiedField(null);
        setCopyStatus(`Unable to copy ${field}.`);
      });
  };

  const runOperation = (type: "start" | "stop") => {
    const mutation = type === "start" ? startMutation : stopMutation;
    setRetryOperation(type);
    setOperationFeedback({
      status: "loading",
      message: `${type === "start" ? "Starting" : "Stopping"} ${name}...`,
    });
    mutation.mutate(undefined, {
      onSuccess: () => {
        setOperationFeedback({
          status: "success",
          message: `${name} ${type === "start" ? "started" : "stopped"}.`,
        });
        setRetryOperation(null);
      },
      onError: (operationError) => {
        setOperationFeedback({
          status: "error",
          message: `Unable to ${type} ${name}: ${operationError.message}`,
        });
      },
    });
  };

  if (isLoading) {
    return (
      <AppShell
        title="Loading container"
        breadcrumbs={[
          { label: "Services" },
          { label: "Containers", to: "/containers" },
          { label: name },
        ]}
        contentId="container-detail-content"
        skipLabel="Skip to container detail"
      >
        <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
          <Server className="size-8 text-muted-foreground mb-2" />
          <p className="text-sm text-muted-foreground">Loading container...</p>
        </div>
      </AppShell>
    );
  }

  if (error || !container) {
    return (
      <AppShell
        title={name}
        breadcrumbs={[
          { label: "Services" },
          { label: "Containers", to: "/containers" },
          { label: name },
        ]}
        contentId="container-detail-content"
        skipLabel="Skip to container detail"
      >
        <OperationFeedback
          status="error"
          message={`Unable to load ${name}: ${error?.message || "Container not found"}`}
          onRetry={() =>
            queryClient.refetchQueries({ queryKey: ["container", name] })
          }
        />
      </AppShell>
    );
  }

  return (
    <TooltipProvider>
      <AppShell
        title={container.name}
        breadcrumbs={[
          { label: "Services" },
          { label: "Containers", to: "/containers" },
          { label: container.name },
        ]}
        contentId="container-detail-content"
        skipLabel="Skip to container detail"
        shortcuts={
          container.status === "running" ? (
            <Button
              variant="outline"
              size="sm"
              className="header-button"
              onClick={() => runOperation("stop")}
              disabled={stopMutation.isPending}
              aria-busy={stopMutation.isPending}
            >
              <Pause aria-hidden="true" /> <span>Stop</span>
            </Button>
          ) : (
            <Button
              variant="outline"
              size="sm"
              className="header-button"
              onClick={() => runOperation("start")}
              disabled={startMutation.isPending}
              aria-busy={startMutation.isPending}
            >
              <Play aria-hidden="true" /> <span>Start</span>
            </Button>
          )
        }
      >
        <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-4">
          {operationFeedback ? (
            <OperationFeedback
              {...operationFeedback}
              onDismiss={() => setOperationFeedback(null)}
              onRetry={
                retryOperation ? () => runOperation(retryOperation) : undefined
              }
            />
          ) : null}
          {logsDisconnected ? (
            <OperationFeedback
              status="disconnected"
              message="Log stream disconnected. Retry to reconnect."
              onRetry={() => {
                setLogsDisconnected(false);
                setLogConnectionNonce((value) => value + 1);
              }}
            />
          ) : null}
          <div className="flex items-center">
            <Badge
              variant={container.status === "running" ? "default" : "secondary"}
            >
              <Circle
                className={`size-2 mr-1 ${container.status === "running" ? "fill-green-500 text-green-500" : "fill-gray-400 text-gray-400"}`}
              />
              {container.status}
            </Badge>
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
                  <span className="font-medium">
                    {container.hostPort || "-"}
                  </span>
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
                <span className="sr-only" role="status" aria-live="polite">
                  {copyStatus}
                </span>
                <div className="flex justify-between items-center">
                  <span className="text-muted-foreground">Host</span>
                  <div className="flex items-center gap-1">
                    <span className="font-medium">
                      localhost:{container.hostPort}
                    </span>
                    <Tooltip>
                      <TooltipTrigger
                        render={
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-6"
                            aria-label="Copy database host"
                          />
                        }
                        onClick={() =>
                          copyToClipboard(
                            `localhost:${container.hostPort}`,
                            "host",
                          )
                        }
                      >
                        <Copy
                          className={`size-3 ${copiedField === "host" ? "text-green-500" : ""}`}
                        />
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
                    <span className="font-medium">
                      {container.config.databaseName || "app"}
                    </span>
                    <Tooltip>
                      <TooltipTrigger
                        render={
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-6"
                            aria-label="Copy database name"
                          />
                        }
                        onClick={() =>
                          copyToClipboard(
                            container.config.databaseName || "app",
                            "database",
                          )
                        }
                      >
                        <Copy
                          className={`size-3 ${copiedField === "database" ? "text-green-500" : ""}`}
                        />
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
                    <span className="font-medium">
                      {container.config.databaseUser || "-"}
                    </span>
                    <Tooltip>
                      <TooltipTrigger
                        render={
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-6"
                            aria-label="Copy database user"
                          />
                        }
                        onClick={() =>
                          copyToClipboard(
                            container.config.databaseUser || "",
                            "user",
                          )
                        }
                      >
                        <Copy
                          className={`size-3 ${copiedField === "user" ? "text-green-500" : ""}`}
                        />
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
                    <span className="font-medium">
                      {container.config.databasePassword
                        ? isPasswordVisible
                          ? container.config.databasePassword
                          : "********"
                        : "-"}
                    </span>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-6"
                      aria-label={
                        isPasswordVisible
                          ? "Hide database password"
                          : "Reveal database password"
                      }
                      onClick={() =>
                        setIsPasswordVisible((visible) => !visible)
                      }
                    >
                      {isPasswordVisible ? (
                        <EyeOff className="size-3" />
                      ) : (
                        <Eye className="size-3" />
                      )}
                    </Button>
                    <Tooltip>
                      <TooltipTrigger
                        render={
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-6"
                            aria-label="Copy database password"
                          />
                        }
                        onClick={() =>
                          copyToClipboard(
                            container.config.databasePassword || "",
                            "password",
                          )
                        }
                      >
                        <Copy
                          className={`size-3 ${copiedField === "password" ? "text-green-500" : ""}`}
                        />
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
              {!isFollowingLogs && (
                <div className="mb-2 flex justify-end">
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => {
                      scrollLogsToBottom();
                      setIsFollowingLogs(true);
                    }}
                  >
                    Jump to latest
                  </Button>
                </div>
              )}
              <pre
                ref={logsContainerRef}
                onScroll={handleLogsScroll}
                className="h-64 overflow-auto rounded-md bg-muted p-2 text-xs font-mono leading-relaxed"
              >
                {logs ||
                  (container.status !== "running"
                    ? "Logs will appear here when the container is running..."
                    : "Waiting for logs...")}
              </pre>
            </CardContent>
          </Card>
        </div>
      </AppShell>
    </TooltipProvider>
  );
}
