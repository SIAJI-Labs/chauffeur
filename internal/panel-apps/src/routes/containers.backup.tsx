import * as React from "react";
import { Link, createFileRoute } from "@tanstack/react-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Eye,
  HardDrive,
  Loader2,
  MoreHorizontal,
  Plus,
  RotateCcw,
  Trash2,
} from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { OperationFeedback } from "@/components/operation-feedback";
import { formatSize } from "@/lib/format-size";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

export const Route = createFileRoute("/containers/backup")({
  component: BackupPage,
});

interface Backup {
  name: string;
  container: string;
  size: number;
  createdAt: string;
}

interface Container {
  name: string;
  engine: string;
  status: "running" | "stopped";
  hostPort: number;
  createdAt: string;
}

interface DatabaseBackup {
  name: string;
  description: string;
}

async function fetchBackups(): Promise<Array<Backup>> {
  const res = await fetch("/api/backups");
  if (!res.ok) throw new Error("Failed to fetch backups");
  return res.json();
}

async function fetchContainers(): Promise<Array<Container>> {
  const res = await fetch("/api/containers");
  if (!res.ok) throw new Error("Failed to fetch containers");
  return res.json();
}

async function fetchDatabases(containerName: string): Promise<Array<string>> {
  const res = await fetch(
    `/api/containers/${encodeURIComponent(containerName)}/databases`,
  );
  if (!res.ok) throw new Error("Failed to fetch databases");
  const data = await res.json();
  return data.databases;
}

async function createBackup(
  container: string,
  databases: Array<DatabaseBackup>,
): Promise<void> {
  const res = await fetch("/api/backups", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ container, databases }),
  });
  if (!res.ok) throw new Error("Failed to create backup");
}

async function restoreBackup(
  backupName: string,
  container: string,
): Promise<void> {
  const res = await fetch(
    `/api/backups/${encodeURIComponent(backupName)}/restore`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ container }),
    },
  );
  if (!res.ok) throw new Error("Failed to restore backup");
}

async function deleteBackup(name: string): Promise<void> {
  const res = await fetch(`/api/backups/${encodeURIComponent(name)}`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Failed to delete backup");
}

function BackupPage() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = React.useState(false);
  const [isViewOpen, setIsViewOpen] = React.useState(false);
  const [isRestoreOpen, setIsRestoreOpen] = React.useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = React.useState(false);
  const [selectedContainer, setSelectedContainer] = React.useState("");
  const [selectedDatabases, setSelectedDatabases] = React.useState<Set<string>>(
    new Set(),
  );
  const [databaseDescriptions, setDatabaseDescriptions] = React.useState<
    Record<string, string>
  >({});
  const [filterContainer, setFilterContainer] = React.useState("");
  const [selectedBackup, setSelectedBackup] = React.useState<Backup | null>(
    null,
  );
  const [restoreTargetContainer, setRestoreTargetContainer] =
    React.useState("");
  const [feedback, setFeedback] = React.useState<{
    status: "loading" | "success" | "error";
    message: string;
  } | null>(null);
  const [retryAction, setRetryAction] = React.useState<(() => void) | null>(
    null,
  );

  const {
    data: backups = [],
    isLoading: backupsLoading,
    error: backupsError,
  } = useQuery({
    queryKey: ["backups"],
    queryFn: fetchBackups,
    refetchInterval: 30000,
  });

  const {
    data: containers = [],
    isLoading: containersLoading,
    error: containersError,
  } = useQuery({
    queryKey: ["containers"],
    queryFn: fetchContainers,
  });

  const {
    data: databases = [],
    isLoading: databasesLoading,
    error: databasesError,
  } = useQuery({
    queryKey: ["databases", selectedContainer],
    queryFn: () => fetchDatabases(selectedContainer),
    enabled: selectedContainer !== "",
  });

  const queryError = backupsError || containersError;

  const retryQueries = () => {
    setFeedback({ status: "loading", message: "Retrying backup data..." });
    void Promise.all([
      queryClient.refetchQueries({ queryKey: ["backups"] }),
      queryClient.refetchQueries({ queryKey: ["containers"] }),
      ...(selectedContainer
        ? [
            queryClient.refetchQueries({
              queryKey: ["databases", selectedContainer],
            }),
          ]
        : []),
    ]).then(() => setFeedback(null));
  };

  const createMutation = useMutation({
    mutationFn: () => {
      const dbs: Array<DatabaseBackup> = Array.from(selectedDatabases).map(
        (name) => ({
          name,
          description: databaseDescriptions[name] || "",
        }),
      );
      console.log("Creating backup:", {
        container: selectedContainer,
        databases: dbs,
      });
      return createBackup(selectedContainer, dbs);
    },
    onSuccess: () => {
      setFeedback({ status: "success", message: "Backup created." });
      setRetryAction(null);
      queryClient.invalidateQueries({ queryKey: ["backups"] });
      setIsCreateOpen(false);
      setSelectedContainer("");
      setSelectedDatabases(new Set());
      setDatabaseDescriptions({});
    },
    onError: (error) => {
      console.error("Backup failed:", error);
      setFeedback({
        status: "error",
        message: `Unable to create backup: ${error.message}`,
      });
    },
  });

  const restoreMutation = useMutation({
    mutationFn: () =>
      restoreBackup(selectedBackup!.name, restoreTargetContainer),
    onSuccess: () => {
      setFeedback({ status: "success", message: "Backup restored." });
      setRetryAction(null);
      queryClient.invalidateQueries({ queryKey: ["backups"] });
      setIsRestoreOpen(false);
      setSelectedBackup(null);
      setRestoreTargetContainer("");
    },
    onError: (error) => {
      setFeedback({
        status: "error",
        message: `Unable to restore backup: ${error.message}`,
      });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (name: string) => deleteBackup(name),
    onSuccess: () => {
      setFeedback({ status: "success", message: "Backup deleted." });
      setRetryAction(null);
      queryClient.invalidateQueries({ queryKey: ["backups"] });
      setIsDeleteOpen(false);
      setSelectedBackup(null);
    },
    onError: (error) => {
      setFeedback({
        status: "error",
        message: `Unable to delete backup: ${error.message}`,
      });
    },
  });

  React.useEffect(() => {
    const error = queryError || databasesError;
    if (error) {
      setFeedback({
        status: "error",
        message: `Unable to load backup data: ${error.message}`,
      });
      setRetryAction(() => retryQueries);
    }
  }, [backupsError, containersError, databasesError, queryError]);

  const runCreate = () => {
    setFeedback({ status: "loading", message: "Creating backup..." });
    setRetryAction(() => runCreate);
    createMutation.mutate();
  };

  const runRestore = () => {
    setFeedback({ status: "loading", message: "Restoring backup..." });
    setRetryAction(() => runRestore);
    restoreMutation.mutate();
  };

  const runDelete = (name: string) => {
    setFeedback({ status: "loading", message: "Deleting backup..." });
    setRetryAction(() => () => runDelete(name));
    deleteMutation.mutate(name);
  };

  const runningContainers = containers.filter((c) => c.status === "running");

  const filteredBackups = filterContainer
    ? backups.filter((b) => b.container === filterContainer)
    : backups;

  const openRestoreDialog = (backup: Backup) => {
    setSelectedBackup(backup);
    setRestoreTargetContainer(backup.container);
    setIsRestoreOpen(true);
  };

  const openViewDialog = (backup: Backup) => {
    setSelectedBackup(backup);
    setIsViewOpen(true);
  };

  const openDeleteDialog = (backup: Backup) => {
    setSelectedBackup(backup);
    setIsDeleteOpen(true);
  };

  const handleContainerChange = (name: string) => {
    setSelectedContainer(name);
    setSelectedDatabases(new Set());
    setDatabaseDescriptions({});
  };

  const toggleDatabase = (db: string) => {
    const newSelected = new Set(selectedDatabases);
    if (newSelected.has(db)) {
      newSelected.delete(db);
    } else {
      newSelected.add(db);
    }
    setSelectedDatabases(newSelected);
  };

  const handleSelectAll = () => {
    if (selectedDatabases.size === databases.length) {
      setSelectedDatabases(new Set());
    } else {
      setSelectedDatabases(new Set(databases));
    }
  };

  const handleDescriptionChange = (db: string, value: string) => {
    setDatabaseDescriptions((prev) => ({ ...prev, [db]: value }));
  };

  const resetCreateDialog = () => {
    setIsCreateOpen(false);
    setSelectedContainer("");
    setSelectedDatabases(new Set());
    setDatabaseDescriptions({});
  };

  return (
    <TooltipProvider>
      <AppShell
        title="Backups"
        breadcrumbs={[
          { label: "Services" },
          { label: "Containers", to: "/containers" },
          { label: "Backups" },
        ]}
        contentId="backups-content"
        skipLabel="Skip to backups"
        shortcuts={
          <Button
            variant="outline"
            size="sm"
            className="header-button"
            render={<Link to="/containers" />}
          >
            <HardDrive aria-hidden="true" /> <span>Containers</span>
          </Button>
        }
      >
        <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-4">
          {feedback ? (
            <OperationFeedback
              {...feedback}
              onDismiss={() => setFeedback(null)}
              onRetry={retryAction ?? undefined}
            />
          ) : null}
          <div className="flex items-center justify-between">
            <h1 className="text-lg font-semibold">Database Backups</h1>
            <Dialog
              open={isCreateOpen}
              onOpenChange={(open) => !open && resetCreateDialog()}
            >
              <DialogTrigger render={<Button size="sm" />}>
                <Plus className="size-4 mr-1" />
                Create Backup
              </DialogTrigger>
              <DialogContent className="max-w-lg max-h-[85vh] overflow-y-auto">
                <DialogHeader>
                  <DialogTitle>Create New Backup</DialogTitle>
                  <DialogDescription>
                    Select a running container and choose which databases to
                    backup.
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Container</label>
                    <Select
                      value={selectedContainer}
                      onValueChange={(value) =>
                        handleContainerChange(value || "")
                      }
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Select a container..." />
                      </SelectTrigger>
                      <SelectContent>
                        {runningContainers.map((c) => (
                          <SelectItem key={c.name} value={c.name}>
                            {c.name} ({c.engine})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {selectedContainer && (
                    <div className="space-y-3">
                      <div className="flex items-center justify-between">
                        <label className="text-sm font-medium">Databases</label>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={handleSelectAll}
                          className="h-auto p-0 text-xs text-muted-foreground hover:text-foreground"
                        >
                          {selectedDatabases.size === databases.length
                            ? "Deselect all"
                            : "Select all"}
                        </Button>
                      </div>

                      {databasesLoading ? (
                        <div className="flex items-center justify-center py-4">
                          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
                        </div>
                      ) : databases.length === 0 ? (
                        <p className="text-sm text-muted-foreground py-2">
                          No databases found
                        </p>
                      ) : (
                        <div className="rounded-md border max-h-48 overflow-y-auto">
                          {databases.map((db) => (
                            <div key={db} className="border-b last:border-b-0">
                              <div className="flex items-center gap-3 p-3">
                                <Checkbox
                                  id={db}
                                  checked={selectedDatabases.has(db)}
                                  onCheckedChange={() => toggleDatabase(db)}
                                />
                                <label
                                  htmlFor={db}
                                  className="text-sm font-medium cursor-pointer flex-1"
                                >
                                  {db}
                                </label>
                              </div>
                              {selectedDatabases.has(db) && (
                                <div className="px-3 pb-3 pl-9">
                                  <Input
                                    placeholder="Description (optional)"
                                    value={databaseDescriptions[db] || ""}
                                    onChange={(e) =>
                                      handleDescriptionChange(
                                        db,
                                        e.target.value,
                                      )
                                    }
                                    className="h-8 text-xs"
                                  />
                                </div>
                              )}
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={resetCreateDialog}>
                    Cancel
                  </Button>
                  <Button
                    onClick={runCreate}
                    disabled={
                      selectedDatabases.size === 0 || createMutation.isPending
                    }
                    aria-busy={createMutation.isPending}
                  >
                    {createMutation.isPending
                      ? "Creating..."
                      : `Backup ${selectedDatabases.size} database${selectedDatabases.size !== 1 ? "s" : ""}`}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          <div className="flex items-center gap-2">
            <label className="text-sm font-medium text-muted-foreground">
              Filter by container:
            </label>
            <Select
              value={filterContainer}
              onValueChange={(v) => setFilterContainer(v || "")}
            >
              <SelectTrigger className="w-48">
                <SelectValue placeholder="All containers" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">All containers</SelectItem>
                {[...new Set(backups.map((b) => b.container))].map(
                  (container) => (
                    <SelectItem key={container} value={container}>
                      {container}
                    </SelectItem>
                  ),
                )}
              </SelectContent>
            </Select>
          </div>

          <Dialog
            open={isViewOpen}
            onOpenChange={(open) =>
              !open && (setIsViewOpen(false), setSelectedBackup(null))
            }
          >
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Backup Details</DialogTitle>
                <DialogDescription>{selectedBackup?.name}</DialogDescription>
              </DialogHeader>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Container</span>
                  <span className="font-medium">
                    {selectedBackup?.container}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Size</span>
                  <span className="font-medium">
                    {selectedBackup && formatSize(selectedBackup.size)}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Created</span>
                  <span className="font-medium">
                    {selectedBackup &&
                      new Date(selectedBackup.createdAt).toLocaleString()}
                  </span>
                </div>
              </div>
              <DialogFooter className="gap-2 sm:gap-0">
                <Button
                  variant="outline"
                  onClick={() => {
                    if (selectedBackup) openRestoreDialog(selectedBackup);
                  }}
                >
                  <RotateCcw className="size-4 mr-1" />
                  Restore
                </Button>
                <AlertDialog
                  open={isDeleteOpen}
                  onOpenChange={(open) => !open && setIsDeleteOpen(false)}
                >
                  <AlertDialogTrigger
                    render={<Button variant="destructive" size="sm" />}
                    onClick={() =>
                      selectedBackup && openDeleteDialog(selectedBackup)
                    }
                  >
                    <Trash2 className="size-4 mr-1" />
                    Delete
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Delete Backup?</AlertDialogTitle>
                      <AlertDialogDescription>
                        This action cannot be undone. The backup file will be
                        permanently deleted.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Cancel</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={() =>
                          selectedBackup && runDelete(selectedBackup.name)
                        }
                        disabled={deleteMutation.isPending}
                      >
                        {deleteMutation.isPending ? "Deleting..." : "Delete"}
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          <Dialog
            open={isRestoreOpen}
            onOpenChange={(open) => !open && setIsRestoreOpen(false)}
          >
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Restore Backup</DialogTitle>
                <DialogDescription>
                  This will overwrite the current database data. The container
                  must be running.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="rounded-md bg-muted p-3 text-sm">
                  <div className="font-medium">{selectedBackup?.name}</div>
                  <div className="text-muted-foreground text-xs mt-1">
                    {selectedBackup && formatSize(selectedBackup.size)} -{" "}
                    {selectedBackup &&
                      new Date(selectedBackup.createdAt).toLocaleString()}
                  </div>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">
                    Target Container
                  </label>
                  <Select
                    value={restoreTargetContainer}
                    onValueChange={(v) => setRestoreTargetContainer(v || "")}
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select a container..." />
                    </SelectTrigger>
                    <SelectContent>
                      {runningContainers.map((c) => (
                        <SelectItem key={c.name} value={c.name}>
                          {c.name} ({c.engine})
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setIsRestoreOpen(false)}
                >
                  Cancel
                </Button>
                <Button
                  variant="destructive"
                  onClick={runRestore}
                  disabled={
                    !restoreTargetContainer || restoreMutation.isPending
                  }
                  aria-busy={restoreMutation.isPending}
                >
                  {restoreMutation.isPending ? "Restoring..." : "Restore"}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {queryError ? (
            <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
              <HardDrive className="size-8 text-muted-foreground mb-2" />
              <p className="text-sm text-muted-foreground">
                Backup data is unavailable.
              </p>
              <p className="text-xs text-muted-foreground mt-1">
                Use Retry above to reconnect.
              </p>
            </div>
          ) : backupsLoading || containersLoading ? (
            <div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center">
              <HardDrive className="size-8 text-muted-foreground mb-2" />
              <p className="text-sm text-muted-foreground">
                Loading backups...
              </p>
            </div>
          ) : filteredBackups.length === 0 ? (
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
                  {filteredBackups.map((backup) => (
                    <TableRow key={backup.name}>
                      <TableCell className="font-medium">
                        {backup.name}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{backup.container}</Badge>
                      </TableCell>
                      <TableCell>{formatSize(backup.size)}</TableCell>
                      <TableCell className="text-muted-foreground">
                        {new Date(backup.createdAt).toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Tooltip>
                            <TooltipTrigger
                              render={
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="size-8"
                                  aria-label={`View details for ${backup.name}`}
                                />
                              }
                              onClick={() => openViewDialog(backup)}
                            >
                              <Eye className="size-4" />
                            </TooltipTrigger>
                            <TooltipContent>View details</TooltipContent>
                          </Tooltip>
                          <DropdownMenu>
                            <DropdownMenuTrigger
                              render={
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="size-8"
                                  aria-label={`Actions for ${backup.name}`}
                                />
                              }
                            >
                              <MoreHorizontal className="size-4" />
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem
                                onClick={() => openRestoreDialog(backup)}
                              >
                                <RotateCcw className="size-4 mr-2" />
                                Restore
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="text-destructive focus:text-destructive"
                                onClick={() => openDeleteDialog(backup)}
                              >
                                <Trash2 className="size-4 mr-2" />
                                Delete
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
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
    </TooltipProvider>
  );
}
