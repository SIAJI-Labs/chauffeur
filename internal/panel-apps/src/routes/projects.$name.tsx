import * as React from "react";
import { Link, createFileRoute } from "@tanstack/react-router";
import {
  ArrowLeft,
  Bug,
  Check,
  CircleAlert,
  Clock3,
  Code2,
  Database,
  ExternalLink,
  GitBranch,
  Globe2,
  HardDrive,
  MoreHorizontal,
  Play,
  RefreshCw,
  RotateCw,
  Server,
  TerminalSquare,
  Wrench,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { AppNavbar } from "@/components/app-navbar";
import { AppSidebar } from "@/components/app-sidebar";
import { ResourceNotFound } from "@/components/resource-not-found";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { useHashTab } from "@/hooks/use-hash-tab";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  projectDetailActions,
  projectDetailLogLines,
  projectDetailTabs,
  projects,
  services,
  workerStates,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/projects/$name")({
  component: ProjectDetailPage,
});

type ProjectTab =
  | "overview"
  | "logs"
  | "environment"
  | "tinker"
  | "diagnostics"
  | "worktrees";

const projectTabValues = projectDetailTabs.map(
  (tab) => tab.value as ProjectTab,
);

type ProjectPreview =
  | "open-folder"
  | "open-terminal"
  | "manage-domains"
  | "nginx"
  | "doctor"
  | "restart-runtime"
  | "restart-server"
  | "sharing"
  | "xdebug"
  | "runtime-controls"
  | "workers"
  | "unlink"
  | "pause"
  | "pin"
  | "edit-environment"
  | "save-environment"
  | "restore-environment"
  | "duplicate-environment"
  | "tinker-controls"
  | "snippet"
  | "diagnostics"
  | "worktree-add"
  | "worktree-remove"
  | "worktree-isolation"
  | "worktree-database"
  | "worktree-workers"
  | "worktree-logs"
  | "worktree-environment"
  | "worktree-sharing"
  | "overflow";

function ProjectDetailPage() {
  const { name } = Route.useParams();
  const project = projects.find((item) => item.name === name);
  const [activeTab, setActiveTab] = useHashTab(projectTabValues, "overview");
  const [preview, setPreview] = React.useState<ProjectPreview | null>(null);

  if (!project) return <ResourceNotFound kind="project" name={name} />;

  return (
    <SidebarProvider className="dashboard-frame">
      <a className="skip-link" href="#project-detail-content">
        Skip to project detail
      </a>
      <AppSidebar />
      <SidebarInset className="dashboard-shell">
        <AppNavbar
          title={project.name}
          breadcrumbs={[
            { label: "Workspace", to: "/" },
            { label: "Projects", to: "/projects" },
            { label: project.name },
          ]}
          shortcuts={
            <Button
              variant="outline"
              size="sm"
              className="header-button"
              render={<Link to="/projects" />}
            >
              <ArrowLeft aria-hidden="true" />
              <span>Projects</span>
            </Button>
          }
        />
        <main
          className="dashboard-content min-h-0 flex-1"
          id="project-detail-content"
        >
          <ScrollArea className="dashboard-scroll-area">
            <div className="project-detail-page">
              <ProjectDetailHeader project={project} onPreview={setPreview} />
              <ProjectActionBar onPreview={setPreview} project={project} />
              <div
                className="project-detail-tabs"
                role="tablist"
                aria-label="Project detail sections"
                aria-orientation="horizontal"
              >
                {projectDetailTabs.map((tab) => (
                  <button
                    className="project-detail-tab"
                    data-active={activeTab === tab.value}
                    id={`project-tab-${tab.value}`}
                    key={tab.value}
                    onClick={() => setActiveTab(tab.value as ProjectTab)}
                    onKeyDown={(event) => {
                      if (
                        !["ArrowLeft", "ArrowRight", "Home", "End"].includes(
                          event.key,
                        )
                      ) {
                        return;
                      }
                      event.preventDefault();
                      const currentIndex = projectDetailTabs.findIndex(
                        (item) => item.value === tab.value,
                      );
                      const nextIndex =
                        event.key === "Home"
                          ? 0
                          : event.key === "End"
                            ? projectDetailTabs.length - 1
                            : (currentIndex +
                                (event.key === "ArrowRight" ? 1 : -1) +
                                projectDetailTabs.length) %
                              projectDetailTabs.length;
                      const nextTab = projectDetailTabs[nextIndex]
                        .value as ProjectTab;
                      setActiveTab(nextTab);
                      requestAnimationFrame(() =>
                        document
                          .getElementById(`project-tab-${nextTab}`)
                          ?.focus(),
                      );
                    }}
                    role="tab"
                    tabIndex={activeTab === tab.value ? 0 : -1}
                    aria-selected={activeTab === tab.value}
                    aria-controls={`project-panel-${tab.value}`}
                    type="button"
                  >
                    {tab.label}
                    {tab.value === "logs" ? <span>4</span> : null}
                  </button>
                ))}
              </div>
              {projectDetailTabs
                .filter((tab) => tab.value !== activeTab)
                .map((tab) => (
                  <div
                    aria-labelledby={`project-tab-${tab.value}`}
                    hidden
                    id={`project-panel-${tab.value}`}
                    key={tab.value}
                    role="tabpanel"
                  />
                ))}
              <div
                className="project-detail-tab-panel"
                id={`project-panel-${activeTab}`}
                role="tabpanel"
                aria-labelledby={`project-tab-${activeTab}`}
                tabIndex={0}
              >
                {activeTab === "overview" ? (
                  <ProjectOverview project={project} onPreview={setPreview} />
                ) : null}
                {activeTab === "logs" ? <ProjectLogs /> : null}
                {activeTab === "environment" ? (
                  <ProjectEnvironment onPreview={setPreview} />
                ) : null}
                {activeTab === "tinker" ? (
                  <ProjectTinker onPreview={setPreview} />
                ) : null}
                {activeTab === "diagnostics" ? (
                  <ProjectDiagnostics onPreview={setPreview} />
                ) : null}
                {activeTab === "worktrees" ? (
                  <ProjectWorktrees project={project} onPreview={setPreview} />
                ) : null}
              </div>
            </div>
          </ScrollArea>
        </main>
      </SidebarInset>
      <ProjectPreviewDialog
        preview={preview}
        onClose={() => setPreview(null)}
        onPreview={setPreview}
      />
    </SidebarProvider>
  );
}

function ProjectDetailHeader({
  project,
  onPreview,
}: {
  project: (typeof projects)[number];
  onPreview: (preview: ProjectPreview) => void;
}) {
  const isHealthy = project.state === "healthy";
  return (
    <section
      className="project-detail-header"
      aria-labelledby="project-detail-title"
    >
      <div className="project-detail-heading">
        <div>
          <div className="project-detail-kicker">
            <span className={`project-detail-status ${project.state}`}>
              <span aria-hidden="true" />{" "}
              {isHealthy ? "Healthy" : "Needs attention"}
            </span>
            <Badge variant="outline">{project.type}</Badge>
            <Badge variant="outline">PHP {project.php}</Badge>
          </div>
          <h2 id="project-detail-title">{project.name}</h2>
          <p>{project.path}</p>
        </div>
        <div className="project-detail-heading-actions">
          <Button
            variant="outline"
            size="sm"
            render={<a href={project.url} target="_blank" rel="noreferrer" />}
          >
            <Globe2 aria-hidden="true" />
            Open project
            <ExternalLink aria-hidden="true" />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => onPreview("overflow")}
            aria-label="More project actions"
          >
            <MoreHorizontal aria-hidden="true" />
          </Button>
        </div>
      </div>
      <div className="project-detail-meta">
        <span>
          <Globe2 aria-hidden="true" /> {project.url}
        </span>
        <span>
          <Server aria-hidden="true" /> PHP {project.php} · Node {project.node}
        </span>
        <span>
          <ShieldIcon /> {project.https} HTTPS
        </span>
        <span>
          <GitBranch aria-hidden="true" /> {project.worktrees}
        </span>
      </div>
    </section>
  );
}

function ProjectActionBar({
  onPreview,
  project,
}: {
  onPreview: (preview: ProjectPreview) => void;
  project: (typeof projects)[number];
}) {
  const previewByAction: Record<string, ProjectPreview | undefined> = {
    "Open folder": "open-folder",
    "Open terminal": "open-terminal",
    "Manage domains": "manage-domains",
    "Nginx configuration": "nginx",
    "Project doctor": "doctor",
  };
  return (
    <section className="project-action-bar" aria-label="Project actions">
      {projectDetailActions.map((action) => {
        const isAvailable = action.status === "Available";
        return (
          <Button
            className={`project-action ${action.tone}`}
            key={action.label}
            variant={isAvailable ? "outline" : "ghost"}
            render={
              isAvailable ? (
                <a href={project.url} target="_blank" rel="noreferrer" />
              ) : undefined
            }
            onClick={() => {
              const nextPreview = previewByAction[action.label];
              if (nextPreview) onPreview(nextPreview);
            }}
          >
            {action.label}
            <Badge variant="outline">{action.status}</Badge>
          </Button>
        );
      })}
      <Button
        className="project-action"
        variant="ghost"
        onClick={() => onPreview("restart-runtime")}
        aria-label="Restart project runtime"
      >
        <RotateCw aria-hidden="true" /> Restart
        <Badge variant="outline">Planned</Badge>
      </Button>
      <Button
        className="project-action"
        variant="ghost"
        onClick={() => onPreview("restart-server")}
      >
        <RefreshCw aria-hidden="true" /> Dev server
        <Badge variant="outline">Planned</Badge>
      </Button>
      <Button
        className="project-action"
        variant="ghost"
        onClick={() => onPreview("sharing")}
      >
        <Globe2 aria-hidden="true" /> Share
        <Badge variant="outline">Planned</Badge>
      </Button>
      <Button
        className="project-action"
        variant="ghost"
        onClick={() => onPreview("xdebug")}
      >
        <Bug aria-hidden="true" /> Xdebug
        <Badge variant="outline">Planned</Badge>
      </Button>
      <span className="project-permission-note">
        <Badge variant="outline" className="project-local-badge">
          Local
        </Badge>
        process controls
        <Badge variant="outline" className="project-remote-badge">
          Remote
        </Badge>
        requires backend
      </span>
    </section>
  );
}

function ProjectPreviewDialog({
  preview,
  onClose,
  onPreview,
}: {
  preview: ProjectPreview | null;
  onClose: () => void;
  onPreview: (preview: ProjectPreview) => void;
}) {
  if (!preview) return null;
  const copy: Record<
    Exclude<ProjectPreview, "overflow">,
    { title: string; description: string; action: string }
  > = {
    "open-folder": {
      title: "Open folder preview",
      description:
        "Review the local path and terminal handoff before opening a project folder.",
      action: "Open folder",
    },
    "open-terminal": {
      title: "Open terminal preview",
      description:
        "Choose the project shell and working directory before launching a terminal.",
      action: "Open terminal",
    },
    "manage-domains": {
      title: "Manage domains preview",
      description:
        "Review local, LAN, tunnel, and public domain options without changing Nginx.",
      action: "Manage domains",
    },
    nginx: {
      title: "Nginx configuration preview",
      description:
        "Inspect the generated server block and certificate state before applying it.",
      action: "Open Nginx configuration",
    },
    doctor: {
      title: "Project doctor preview",
      description:
        "Run a planned diagnostic pass across runtime, domain, database, and worker checks.",
      action: "Run project doctor",
    },
    "restart-runtime": {
      title: "Restart runtime preview",
      description:
        "Confirm the PHP-FPM profile, affected projects, and expected downtime before restarting.",
      action: "Restart runtime",
    },
    "restart-server": {
      title: "Restart development server preview",
      description:
        "Review the server command, port, and current process before restarting it.",
      action: "Restart development server",
    },
    sharing: {
      title: "Sharing preview",
      description:
        "Select a LAN, tunnel, or public share target and review its expiration settings.",
      action: "Configure sharing",
    },
    xdebug: {
      title: "Xdebug preview",
      description:
        "Choose whether Xdebug is enabled and which mode the project should use.",
      action: "Configure Xdebug",
    },
    "runtime-controls": {
      title: "Runtime controls preview",
      description:
        "Choose PHP and Node versions for this project without changing its current profile.",
      action: "Save runtime selection",
    },
    workers: {
      title: "Worker controls preview",
      description:
        "Review queue, Horizon, scheduler, Reverb, Stripe, Vite, and custom worker states.",
      action: "Save worker selection",
    },
    unlink: {
      title: "Unlink project confirmation",
      description:
        "Confirm the project path and understand that unlinking removes it from the workspace inventory without deleting files.",
      action: "Confirm unlink",
    },
    pause: {
      title: "Pause or resume project preview",
      description:
        "Review the project services that would be paused and the resume conditions before changing its state.",
      action: "Save project state",
    },
    pin: {
      title: "Pin or unpin project preview",
      description:
        "Choose whether this project stays at the top of the workspace inventory.",
      action: "Save pin state",
    },
    "edit-environment": {
      title: "Edit environment preview",
      description:
        "Review the variable name, value visibility, source, and worktree scope before editing.",
      action: "Open environment editor",
    },
    "save-environment": {
      title: "Save environment confirmation",
      description:
        "Confirm the changed variables and affected services before saving a project environment.",
      action: "Confirm environment save",
    },
    "restore-environment": {
      title: "Restore environment confirmation",
      description:
        "Restore the last known environment snapshot without overwriting the current preview.",
      action: "Restore environment",
    },
    "duplicate-environment": {
      title: "Duplicate variable warning",
      description:
        "A variable with the same name already exists in the inherited scope. Review precedence before saving.",
      action: "Review duplicate variable",
    },
    "tinker-controls": {
      title: "Tinker workspace controls preview",
      description:
        "Review fullscreen, split direction, code, output, SQL, dump, and error regions without executing code.",
      action: "Save workspace layout",
    },
    snippet: {
      title: "Saved snippet preview",
      description:
        "Preview saving, loading, and deleting a Tinker snippet. Draft code stays local to this static screen.",
      action: "Save snippet",
    },
    diagnostics: {
      title: "Diagnostics lenses preview",
      description:
        "Review sample records and planned filters for dumps, queries, jobs, views, mail, cache, events, and HTTP.",
      action: "Run diagnostic lenses",
    },
    "worktree-add": {
      title: "Add worktree preview",
      description:
        "Choose a branch, domain, runtime, and database isolation policy before creating a worktree.",
      action: "Create worktree",
    },
    "worktree-remove": {
      title: "Remove worktree confirmation",
      description:
        "Confirm the worktree path and understand that its project files remain available in the main checkout.",
      action: "Confirm remove",
    },
    "worktree-isolation": {
      title: "Worktree isolation preview",
      description:
        "Choose whether the worktree receives an isolated database and runtime configuration.",
      action: "Save isolation setting",
    },
    "worktree-database": {
      title: "Worktree database drop confirmation",
      description:
        "Confirm the isolated database target before dropping it. No database operation is performed.",
      action: "Confirm database drop",
    },
    "worktree-workers": {
      title: "Worktree workers preview",
      description:
        "Review worker state and project-specific queue configuration for this worktree.",
      action: "Save worker configuration",
    },
    "worktree-logs": {
      title: "Worktree logs preview",
      description:
        "Choose the worktree log sources and review disconnected or empty states before following output.",
      action: "Open worktree logs",
    },
    "worktree-environment": {
      title: "Worktree environment preview",
      description:
        "Review inherited and worktree-specific variables without changing the project environment.",
      action: "Open worktree environment",
    },
    "worktree-sharing": {
      title: "Worktree sharing preview",
      description:
        "Review the worktree domain, LAN exposure, and tunnel settings before sharing it.",
      action: "Configure worktree sharing",
    },
  };
  const overflowActions = [
    { label: "Pause or resume project", preview: "pause" as const },
    { label: "Pin project", preview: "pin" as const },
    { label: "Restart runtime", preview: "restart-runtime" as const },
    { label: "Manage domains and sharing", preview: "sharing" as const },
    { label: "Unlink project", preview: "unlink" as const },
  ];
  const content =
    preview === "overflow"
      ? {
          title: "Project actions",
          description:
            "Preview of the project action menu. Each operation remains static until backend support is available.",
          action: "Choose an action",
        }
      : copy[preview];
  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="project-preview-dialog">
        <DialogHeader>
          <DialogTitle>{content.title}</DialogTitle>
          <DialogDescription>{content.description}</DialogDescription>
        </DialogHeader>
        {preview === "overflow" ? (
          <div className="project-overflow-list">
            {overflowActions.map((action) => (
              <div key={action.label}>
                <span>{action.label}</span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onPreview(action.preview)}
                >
                  Review
                </Button>
              </div>
            ))}
          </div>
        ) : (
          <div className="project-preview-body">
            <Badge variant="outline" className="project-next-badge">
              Static preview only
            </Badge>
            <div>
              <span>Operation plan</span>
              <strong>{content.action}</strong>
              <small>No command, process, or configuration is changed.</small>
            </div>
            {preview === "sharing" ? (
              <div className="project-preview-options">
                <span>Share target</span>
                <div>
                  <Badge variant="outline">LAN</Badge>
                  <Badge variant="outline">Tunnel</Badge>
                  <Badge variant="outline">Public</Badge>
                </div>
              </div>
            ) : null}
            {preview === "xdebug" ? (
              <div className="project-preview-options">
                <span>Xdebug mode</span>
                <div>
                  <Button variant="outline" size="sm" disabled>
                    Off
                  </Button>
                  <Button variant="outline" size="sm" disabled>
                    Debug
                  </Button>
                  <Button variant="outline" size="sm" disabled>
                    Develop
                  </Button>
                </div>
              </div>
            ) : null}
          </div>
        )}
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Close preview
          </Button>
          <Button disabled>{content.action}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function ProjectOverview({
  project,
  onPreview,
}: {
  project: (typeof projects)[number];
  onPreview: (preview: ProjectPreview) => void;
}) {
  return (
    <div className="project-overview-grid">
      <section className="project-detail-card project-runtime-card">
        <DetailCardHeading
          kicker="Runtime"
          title="Project services"
          icon={Server}
        />
        <div className="project-service-grid">
          {services.slice(0, 3).map((service) => (
            <div className="project-service-card" key={service.name}>
              <span
                className={`project-service-dot ${service.state}`}
                aria-hidden="true"
              />
              <div>
                <strong>{service.name}</strong>
                <small>{service.detail}</small>
              </div>
              <Badge variant="outline">
                {service.state === "healthy" ? "Active" : "Review"}
              </Badge>
            </div>
          ))}
        </div>
        <div className="project-runtime-footer">
          <span>
            <TerminalSquare aria-hidden="true" /> Shared PHP-FPM profile
          </span>
          <Badge variant="outline">PHP {project.php}</Badge>
        </div>
        <div className="project-runtime-controls">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPreview("runtime-controls")}
          >
            Configure PHP and Node
          </Button>
          <Badge variant="outline">Preview</Badge>
        </div>
      </section>
      <section className="project-detail-card">
        <DetailCardHeading
          kicker="Data"
          title="Database connection"
          icon={Database}
        />
        <div className="project-database-summary">
          <div>
            <strong>commerce</strong>
            <span>PostgreSQL · 24 tables · 84 MB</span>
          </div>
          <Badge variant="outline" className="project-health-badge">
            Connected
          </Badge>
        </div>
        <div className="project-dsn-row">
          <code>postgres://localhost:5432/commerce</code>
          <Button variant="outline" size="sm" disabled>
            Copy DSN
          </Button>
        </div>
      </section>
      <section className="project-detail-card">
        <DetailCardHeading
          kicker="Requests"
          title="Recent timing"
          icon={Clock3}
        />
        <div className="project-timing-value">
          118<span>ms</span>
        </div>
        <div
          className="project-timing-bars"
          aria-label="Request timing examples"
        >
          <span className="fast" />
          <span className="medium" />
          <span className="slow" />
          <span className="fast" />
        </div>
        <small className="project-card-note">
          p95 response time · last 15 minutes
        </small>
      </section>
      <section className="project-detail-card">
        <DetailCardHeading
          kicker="Workers"
          title="Background execution"
          icon={GitBranch}
        />
        <div className="project-worker-list">
          {workerStates.map((worker) => (
            <div key={worker.name}>
              <span>
                <strong>{worker.name}</strong>
                <small>{worker.detail}</small>
              </span>
              <Badge
                variant="outline"
                className={`worker-state-badge ${worker.state}`}
              >
                {worker.state}
              </Badge>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onPreview("workers")}
              >
                Configure
              </Button>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}

function ProjectLogs() {
  const [followOutput, setFollowOutput] = React.useState(false);
  const [copied, setCopied] = React.useState(false);
  return (
    <section className="project-tab-card" aria-labelledby="project-logs-title">
      <div className="project-tab-card-heading">
        <DetailCardHeading
          kicker="Static output"
          title="Application logs"
          icon={HardDrive}
        />
        <div className="project-log-actions">
          <Button
            variant={followOutput ? "default" : "outline"}
            size="sm"
            onClick={() => setFollowOutput((value) => !value)}
            aria-pressed={followOutput}
          >
            {followOutput ? "Following output" : "Follow output"}
          </Button>
          <Button variant="outline" size="sm" onClick={() => setCopied(true)}>
            {copied ? "Copied sample" : "Copy logs"}
          </Button>
        </div>
      </div>
      <div className="project-log-viewer" id="project-logs-title">
        {projectDetailLogLines.map((line) => (
          <div
            className="project-log-line"
            key={`${line.time}-${line.message}`}
          >
            <time>{line.time}</time>
            <span className={`project-log-level ${line.level.toLowerCase()}`}>
              {line.level}
            </span>
            <code>{line.message}</code>
          </div>
        ))}
      </div>
      <div className="project-log-state-strip" aria-label="Log viewer states">
        <Badge variant="outline" className="project-log-source-active">
          Application
        </Badge>
        <Badge variant="outline">PHP-FPM</Badge>
        <Badge variant="outline">Development server</Badge>
        <Badge variant="outline">Workers</Badge>
      </div>
      <div className="project-log-state-strip" aria-label="Log viewer states">
        <Badge variant="outline" className="project-log-state-active">
          Loaded
        </Badge>
        <Badge variant="outline">Loading</Badge>
        <Badge variant="outline">Empty</Badge>
        <Badge variant="outline">Disconnected</Badge>
        <Badge variant="outline">Failed</Badge>
      </div>
      <p className="project-static-note">
        Static sample output. Follow and copy demonstrate the planned feedback
        flow; live streaming is not connected.
      </p>
    </section>
  );
}

function ProjectEnvironment({
  onPreview,
}: {
  onPreview: (preview: ProjectPreview) => void;
}) {
  const variables = [
    ["APP_ENV", "local", "Project"],
    ["APP_URL", "https://commerce-api.test:8443", "Project"],
    ["DB_CONNECTION", "postgres", "Inherited"],
    ["DB_PASSWORD", "••••••••••••", "Secret"],
  ];
  return (
    <section
      className="project-tab-card"
      aria-labelledby="project-environment-title"
    >
      <div className="project-tab-card-heading">
        <DetailCardHeading
          kicker="Configuration"
          title="Environment variables"
          icon={Code2}
        />
        <div className="project-environment-actions">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPreview("edit-environment")}
          >
            Edit preview
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onPreview("restore-environment")}
          >
            Restore preview
          </Button>
        </div>
      </div>
      <div className="project-environment-table" id="project-environment-title">
        {variables.map(([name, value, source]) => (
          <div className="project-environment-row" key={name}>
            <code>{name}</code>
            <span>{value}</span>
            <Badge variant="outline">{source}</Badge>
          </div>
        ))}
      </div>
      <div className="project-environment-footer">
        <Badge variant="outline">Worktree-specific state: Preview</Badge>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("save-environment")}
        >
          Save confirmation
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onPreview("duplicate-environment")}
        >
          Duplicate warning
        </Button>
      </div>
      <p className="project-static-note">
        Secrets are masked. Saving and restoring values requires backend
        support.
      </p>
    </section>
  );
}

function ProjectTinker({
  onPreview,
}: {
  onPreview: (preview: ProjectPreview) => void;
}) {
  const [fullscreen, setFullscreen] = React.useState(false);
  const [splitDirection, setSplitDirection] = React.useState<"side" | "stack">(
    "side",
  );
  const snippet = `Order::query()
  ->where('status', 'pending')
  ->latest()
  ->limit(10)
  ->get();`;
  return (
    <section
      className="project-tab-card"
      aria-labelledby="project-tinker-title"
    >
      <div className="project-tab-card-heading">
        <DetailCardHeading
          kicker="Diagnostics"
          title="Tinker workspace"
          icon={Code2}
        />
        <div className="project-tinker-actions">
          <Button
            variant={fullscreen ? "default" : "outline"}
            size="sm"
            onClick={() => setFullscreen((value) => !value)}
            aria-pressed={fullscreen}
          >
            {fullscreen ? "Exit fullscreen" : "Fullscreen"}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              setSplitDirection((value) =>
                value === "side" ? "stack" : "side",
              )
            }
          >
            Split {splitDirection === "side" ? "side" : "stacked"}
          </Button>
          <Badge variant="outline">Planned</Badge>
        </div>
      </div>
      <div
        className={`project-tinker-grid ${fullscreen ? "is-fullscreen" : ""} ${splitDirection}`}
        id="project-tinker-title"
      >
        <textarea
          aria-label="Tinker code editor preview"
          defaultValue={snippet}
          spellCheck={false}
        />
        <div className="project-tinker-output">
          <div className="project-tinker-output-heading">
            <span>Output</span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPreview("tinker-controls")}
            >
              <Play aria-hidden="true" /> Run preview
            </Button>
          </div>
          <p>Run a snippet to inspect records, SQL, dumps, or errors.</p>
          <div className="project-tinker-output-regions">
            <div>
              <strong>Output</strong>
              <span>2 records · sample only</span>
            </div>
            <div>
              <strong>SQL</strong>
              <span>select * from orders limit 10</span>
            </div>
            <div>
              <strong>Dump</strong>
              <span>OrderCollection</span>
            </div>
            <div>
              <strong>Error</strong>
              <span>No error in preview</span>
            </div>
          </div>
        </div>
      </div>
      <div className="project-snippet-list">
        <div className="project-snippet-heading">
          <span>
            <strong>Saved snippets</strong>
            <small>Save, load, or delete local drafts</small>
          </span>
          <Badge variant="outline">2 previews</Badge>
        </div>
        <div className="project-snippet-row">
          <span>
            <strong>Pending orders</strong>
            <small>Updated 4 minutes ago</small>
          </span>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onPreview("snippet")}
          >
            Load
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onPreview("snippet")}
          >
            Delete
          </Button>
        </div>
        <div className="project-snippet-row">
          <span>
            <strong>Slow checkout query</strong>
            <small>Draft persisted locally</small>
          </span>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onPreview("snippet")}
          >
            Load
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onPreview("snippet")}
          >
            Delete
          </Button>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("snippet")}
        >
          Save current draft
        </Button>
      </div>
    </section>
  );
}

function ProjectDiagnostics({
  onPreview,
}: {
  onPreview: (preview: ProjectPreview) => void;
}) {
  const checks = [
    ["Local route", "commerce-api.test resolves", "pass"],
    ["HTTPS certificate", "Certificate is ready", "pass"],
    ["PHP runtime", "PHP 8.3 FPM is active", "pass"],
    ["Queue worker", "Reverb connection needs review", "warning"],
  ];
  return (
    <section
      className="project-tab-card"
      aria-labelledby="project-diagnostics-title"
    >
      <div className="project-tab-card-heading">
        <DetailCardHeading
          kicker="Doctor"
          title="Project diagnostics"
          icon={Wrench}
        />
        <Badge variant="outline">Sample + planned</Badge>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("diagnostics")}
        >
          Run again
        </Button>
      </div>
      <div className="project-diagnostic-lenses" aria-label="Diagnostic lenses">
        {[
          ["Dumps", "12"],
          ["Queries", "24"],
          ["Jobs", "4"],
          ["Views", "8"],
          ["Mail", "2"],
          ["Cache", "6"],
          ["Events", "9"],
          ["HTTP", "18"],
        ].map(([label, count]) => (
          <Button
            key={label}
            variant="outline"
            size="sm"
            onClick={() => onPreview("diagnostics")}
          >
            {label} <Badge variant="outline">{count}</Badge>
          </Button>
        ))}
      </div>
      <div className="project-diagnostic-list" id="project-diagnostics-title">
        {checks.map(([name, detail, state]) => (
          <div className="project-diagnostic-row" key={name}>
            {state === "pass" ? (
              <Check aria-hidden="true" />
            ) : (
              <CircleAlert aria-hidden="true" />
            )}
            <span>
              <strong>{name}</strong>
              <small>{detail}</small>
            </span>
            <Badge variant="outline" className={state}>
              {state === "pass" ? "Pass" : "Attention"}
            </Badge>
          </div>
        ))}
      </div>
    </section>
  );
}

function ProjectWorktrees({
  project,
  onPreview,
}: {
  project: (typeof projects)[number];
  onPreview: (preview: ProjectPreview) => void;
}) {
  return (
    <section
      className="project-tab-card"
      aria-labelledby="project-worktrees-title"
    >
      <div className="project-tab-card-heading">
        <DetailCardHeading
          kicker="Branches"
          title="Worktrees"
          icon={GitBranch}
        />
        <Badge variant="outline">Branch: main</Badge>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("worktree-add")}
        >
          Add worktree
        </Button>
      </div>
      <div className="project-worktree-list" id="project-worktrees-title">
        <div className="project-worktree-row active">
          <GitBranch aria-hidden="true" />
          <span>
            <strong>main</strong>
            <small>{project.path} · default environment</small>
          </span>
          <Badge variant="outline">Current</Badge>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onPreview("worktree-isolation")}
          >
            Isolation
          </Button>
        </div>
        <div className="project-worktree-row">
          <GitBranch aria-hidden="true" />
          <span>
            <strong>feature/checkout</strong>
            <small>Isolated domain planned · shared database</small>
          </span>
          <Badge variant="outline">Planned</Badge>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onPreview("worktree-remove")}
          >
            Remove
          </Button>
        </div>
      </div>
      <div className="project-worktree-footer">
        <Badge variant="outline">Domain: feature-checkout.test</Badge>
        <Badge variant="outline">Runtime: PHP 8.3</Badge>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onPreview("worktree-database")}
        >
          Drop isolated database preview
        </Button>
      </div>
      <div className="project-worktree-capabilities">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("worktree-workers")}
        >
          Workers
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("worktree-logs")}
        >
          Logs
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("worktree-environment")}
        >
          Environment
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("worktree-sharing")}
        >
          Sharing
        </Button>
      </div>
    </section>
  );
}

function DetailCardHeading({
  kicker,
  title,
  icon: Icon,
}: {
  kicker: string;
  title: string;
  icon: LucideIcon;
}) {
  return (
    <div className="detail-card-heading">
      <div>
        <p className="section-kicker">{kicker}</p>
        <h3>{title}</h3>
      </div>
      <Icon className="heading-icon" aria-hidden="true" />
    </div>
  );
}

function ShieldIcon() {
  return (
    <span className="project-shield-icon" aria-hidden="true">
      HTTPS
    </span>
  );
}
