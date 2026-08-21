import {
  Activity,
  Bell,
  Braces,
  Bug,
  Check,
  CircleAlert,
  Clock3,
  Container,
  Cpu,
  Database,
  DatabaseBackup,
  Eye,
  FileCode2,
  FileText,
  FolderGit2,
  FolderOpen,
  Gauge,
  GitBranch,
  Globe2,
  HardDrive,
  Link2,
  ListChecks,
  LoaderCircle,
  Mail,
  Network,
  PauseCircle,
  Play,
  RefreshCw,
  Search,
  Server,
  Settings2,
  ShieldCheck,
  TerminalSquare,
  Wrench,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

export type FixtureStatus =
  | "healthy"
  | "attention"
  | "idle"
  | "available"
  | "next"
  | "planned";

export const projects = [
  {
    name: "commerce-api",
    updatedAt: "2026-08-20T14:30:00Z",
    path: "~/Projects/commerce-api",
    url: "https://commerce-api.test:8443",
    php: "8.3",
    type: "Laravel",
    state: "healthy" as const,
    note: "14 requests · last 15 min",
    workspace: "Commerce",
    node: "22.12",
    https: "Ready",
    worktrees: "main + 1 worktree",
  },
  {
    name: "client-portal",
    updatedAt: "2026-08-18T09:15:00Z",
    path: "~/Projects/client-portal",
    url: "http://client-portal.test:8080",
    php: "8.4",
    type: "Generic PHP",
    state: "attention" as const,
    note: "SSL is not enabled",
    workspace: "Client work",
    node: "20.19",
    https: "Needs setup",
    worktrees: "main",
  },
];

export const projectDetailTabs = [
  { value: "overview", label: "Overview" },
  { value: "logs", label: "Logs" },
  { value: "environment", label: "Environment" },
  { value: "tinker", label: "Tinker" },
  { value: "diagnostics", label: "Diagnostics" },
  { value: "worktrees", label: "Worktrees" },
];

export const projectDetailLogLines = [
  { time: "09:42:18", level: "INFO", message: "GET /api/orders 200 42ms" },
  { time: "09:42:13", level: "INFO", message: "GET /dashboard 200 118ms" },
  {
    time: "09:41:57",
    level: "WARN",
    message: "Certificate renewal is pending",
  },
  {
    time: "09:41:32",
    level: "INFO",
    message: "Queue worker processed order.created",
  },
];

export const projectDetailActions = [
  { label: "Open in browser", status: "Available", tone: "available" },
  { label: "Open folder", status: "Planned", tone: "planned" },
  { label: "Open terminal", status: "Planned", tone: "planned" },
  { label: "Manage domains", status: "Planned", tone: "planned" },
  { label: "Nginx configuration", status: "Planned", tone: "planned" },
  { label: "Project doctor", status: "Next", tone: "next" },
];

export const services = [
  {
    name: "nginx",
    detail: "workspace gateway",
    state: "healthy" as const,
    icon: Globe2,
    updateAvailable: false,
  },
  {
    name: "PHP 8.3",
    detail: "shared FPM pool",
    state: "healthy" as const,
    icon: Server,
    updateAvailable: false,
  },
  {
    name: "PHP 8.4",
    detail: "installed · stopped · update ready",
    state: "idle" as const,
    icon: Server,
    updateAvailable: true,
  },
  {
    name: "Podman",
    detail: "runtime unavailable",
    state: "attention" as const,
    icon: Database,
    updateAvailable: false,
  },
];

export const serviceOverviewFixtures = [
  {
    name: "PostgreSQL",
    category: "Databases",
    version: "16.4",
    projects: "2 projects",
    ports: "5432",
    dependencies: "Ready",
    update: "Available",
    state: "healthy",
    icon: Database,
  },
  {
    name: "Redis",
    category: "Cache",
    version: "7.4",
    projects: "1 project",
    ports: "6379",
    dependencies: "Ready",
    update: "Current",
    state: "healthy",
    icon: Server,
  },
  {
    name: "Mailpit",
    category: "Mail",
    version: "1.21",
    projects: "0 projects",
    ports: "8025",
    dependencies: "Not configured",
    update: "Planned",
    state: "idle",
    icon: Mail,
  },
  {
    name: "Meilisearch",
    category: "Search",
    version: "1.12",
    projects: "1 project",
    ports: "7700",
    dependencies: "Port conflict",
    update: "Update ready",
    state: "attention",
    icon: Search,
  },
  {
    name: "Queue workers",
    category: "Workers",
    version: "Project profile",
    projects: "2 projects",
    ports: "Host mode",
    dependencies: "1 warning",
    update: "Current",
    state: "attention",
    icon: Activity,
  },
];

export const serviceCatalogFixtures = [
  {
    name: "MySQL",
    category: "Databases",
    detail: "Local relational database with project pairing.",
    state: "Available",
    icon: Database,
  },
  {
    name: "MinIO",
    category: "Storage",
    detail: "S3-compatible object storage for local workflows.",
    state: "Planned",
    icon: HardDrive,
  },
  {
    name: "Meilisearch",
    category: "Search",
    detail: "Fast local search index with project-aware access.",
    state: "Available",
    icon: Search,
  },
  {
    name: "Mailpit",
    category: "Mail",
    detail: "Capture local email with a safe browser inbox.",
    state: "Available",
    icon: Mail,
  },
  {
    name: "RabbitMQ",
    category: "Queues",
    detail: "Message broker for distributed local workers.",
    state: "Planned",
    icon: Activity,
  },
  {
    name: "Object storage",
    category: "Storage",
    detail: "Preview storage presets before installing a service.",
    state: "Planned",
    icon: HardDrive,
  },
];

export const serviceCategories = [
  "All services",
  "Databases",
  "Cache",
  "Queues",
  "Search",
  "Mail",
  "Storage",
  "Workers",
];

export const serviceDetailTabs = [
  { value: "admin", label: "Admin" },
  { value: "databases", label: "Databases" },
  { value: "entities", label: "Entities" },
  { value: "logs", label: "Logs" },
  { value: "environment", label: "Environment" },
  { value: "configuration", label: "Configuration" },
  { value: "tools", label: "Tools" },
  { value: "ports", label: "Ports" },
];

export const serviceDetailActions = [
  "Start",
  "Stop",
  "Restart",
  "Pin",
  "Update",
  "Upgrade",
  "Migrate",
  "Rollback",
  "Reinstall",
  "Remove",
];

export const serviceDatabaseFixtures = [
  {
    name: "commerce",
    size: "84 MB",
    owner: "commerce-api",
    status: "Connected",
  },
  {
    name: "portal",
    size: "42 MB",
    owner: "client-portal",
    status: "Connected",
  },
  {
    name: "commerce_test",
    size: "12 MB",
    owner: "commerce-api",
    status: "Testing",
  },
];

export const serviceEntityFixtures = [
  {
    kind: "Tables",
    count: "24",
    detail: "orders, users, products, migrations",
  },
  { kind: "Views", count: "6", detail: "order_summary, revenue_by_day" },
  { kind: "Extensions", count: "3", detail: "uuid-ossp, pgcrypto, citext" },
];

export const serviceSnapshotFixtures = [
  {
    name: "commerce-2026-08-21",
    size: "84 MB",
    age: "18 min ago",
    state: "Ready",
  },
  {
    name: "commerce-2026-08-20",
    size: "82 MB",
    age: "Yesterday",
    state: "Ready",
  },
];

export const serviceImportStateFixtures = [
  {
    label: "Importing",
    detail: "Restoring commerce-2026-08-21",
    tone: "loading",
  },
  {
    label: "Warning",
    detail: "Target database already exists",
    tone: "warning",
  },
  { label: "Error", detail: "Snapshot checksum mismatch", tone: "error" },
  { label: "Issue", detail: "1 table needs manual review", tone: "issue" },
];

export const serviceLogFixtures = [
  {
    time: "09:43:18",
    level: "INFO",
    message: "database system is ready to accept connections",
  },
  {
    time: "09:42:51",
    level: "INFO",
    message: "connection authorized · commerce-api",
  },
  {
    time: "09:41:07",
    level: "WARN",
    message: "checkpoint completion is close to the configured limit",
  },
];

export const workspaceSignals = [
  {
    label: "Projects",
    value: "2",
    detail: "1 needs attention",
    tone: "violet",
    icon: FolderGit2,
  },
  {
    label: "Healthy routes",
    value: "1 / 2",
    detail: "HTTPS ready",
    tone: "green",
    icon: Globe2,
  },
  {
    label: "PHP runtimes",
    value: "2",
    detail: "8.3 active · 8.4 idle",
    tone: "amber",
    icon: Server,
  },
  {
    label: "Open issues",
    value: "1",
    detail: "SSL is disabled",
    tone: "rose",
    icon: CircleAlert,
  },
];

export const projectListStates = [
  {
    label: "Healthy",
    detail: "Routes and runtime checks pass.",
    tone: "healthy",
    icon: Check,
  },
  {
    label: "Attention",
    detail: "A project needs review.",
    tone: "attention",
    icon: CircleAlert,
  },
  {
    label: "Paused",
    detail: "Excluded from active checks.",
    tone: "paused",
    icon: PauseCircle,
  },
  {
    label: "Loading",
    detail: "Discovering project metadata.",
    tone: "loading",
    icon: LoaderCircle,
  },
  {
    label: "Empty",
    detail: "No projects in this workspace.",
    tone: "empty",
    icon: FolderOpen,
  },
  {
    label: "Error",
    detail: "Could not read project metadata.",
    tone: "error",
    icon: CircleAlert,
  },
];

export const projectGroupOptions = [
  { value: "workspace", label: "Workspace" },
  { value: "framework", label: "Framework" },
];

export const projectSortOptions = [
  { value: "name", label: "Name" },
  { value: "status", label: "Status" },
  { value: "recent", label: "Recent activity" },
];

export const projectLinkingSteps = [
  {
    number: "01",
    title: "Choose a local folder",
    detail: "Select a project path and confirm its workspace identity.",
  },
  {
    number: "02",
    title: "Assign a runtime",
    detail: "Pick PHP, Node, and service profiles for the project.",
  },
  {
    number: "03",
    title: "Verify the route",
    detail: "Preview the domain, HTTPS certificate, and health checks.",
  },
];

export const workspaceManagementSteps = [
  {
    title: "Create or rename a workspace",
    detail: "Preview the workspace name, path scope, and project membership.",
  },
  {
    title: "Collapse and reorder groups",
    detail:
      "Choose the order and visibility of workspace groups in the sidebar.",
  },
  {
    title: "Delete a workspace",
    detail:
      "Confirm that deletion removes organization only, never local project files.",
  },
];

export const healthHero = {
  status: "attention" as const,
  title: "One workspace issue needs attention",
  detail:
    "client-portal is reachable, but its local HTTPS certificate is not enabled.",
  action: "Review project issue",
};

export const healthStates = [
  {
    label: "Healthy",
    detail: "All workspace checks pass.",
    tone: "healthy",
    icon: Check,
  },
  {
    label: "Attention",
    detail: "A project needs review.",
    tone: "attention",
    icon: CircleAlert,
  },
  {
    label: "Failure",
    detail: "A runtime is unavailable.",
    tone: "failure",
    icon: CircleAlert,
  },
  {
    label: "Update",
    detail: "A newer runtime is ready.",
    tone: "update",
    icon: RefreshCw,
  },
  {
    label: "Requires backend",
    detail: "Remote operation is not connected in the static phase.",
    tone: "requires-backend",
    icon: Settings2,
  },
];

export const workspaceSummary = [
  {
    label: "Projects",
    value: "2",
    detail: "1 needs attention",
    icon: FolderGit2,
  },
  {
    label: "Services",
    value: "4",
    detail: "2 active · 1 stopped",
    icon: Server,
  },
  {
    label: "Databases",
    value: "3",
    detail: "2 connected projects",
    icon: Database,
  },
  {
    label: "Backups",
    value: "6",
    detail: "Latest 18 min ago",
    icon: DatabaseBackup,
  },
  {
    label: "Open issues",
    value: "1",
    detail: "SSL is disabled",
    icon: CircleAlert,
  },
];

export const workerStates = [
  {
    name: "Queue worker",
    detail: "active · 18 jobs/min",
    state: "active",
    icon: Activity,
  },
  {
    name: "Scheduler",
    detail: "sleeping · next run in 12 min",
    state: "sleeping",
    icon: Container,
  },
  {
    name: "Reverb",
    detail: "failed · connection refused",
    state: "failed",
    icon: CircleAlert,
  },
  {
    name: "Vite",
    detail: "stopped · not configured for this project",
    state: "stopped",
    icon: TerminalSquare,
  },
];

export const systemHealth = [
  { name: "Nginx", detail: "gateway online", state: "healthy", icon: Globe2 },
  {
    name: "PHP-FPM",
    detail: "8.3 pool active",
    state: "healthy",
    icon: Server,
  },
  {
    name: "DNS",
    detail: "local resolver ready",
    state: "healthy",
    icon: Globe2,
  },
  {
    name: "Chauffeur",
    detail: "runtime connected",
    state: "healthy",
    icon: Activity,
  },
];

export const systemComponentFixtures = [
  {
    name: "Chauffeur runtime",
    category: "Core",
    detail: "Panel and CLI runtime connected",
    version: "v0.1.0",
    state: "healthy",
    icon: Activity,
  },
  {
    name: "Nginx",
    category: "Gateway",
    detail: "Local routes and TLS gateway online",
    version: "1.27.4",
    state: "healthy",
    icon: Globe2,
  },
  {
    name: "DNS",
    category: "Networking",
    detail: "Local resolver ready for .test domains",
    version: "dnsmasq",
    state: "healthy",
    icon: Network,
  },
  {
    name: "PHP-FPM",
    category: "Runtime",
    detail: "8.3 active · 8.4 installed",
    version: "2 versions",
    state: "attention",
    icon: Server,
  },
  {
    name: "Node.js / Bun",
    category: "Toolchain",
    detail: "Node active · Bun not installed",
    version: "Node 22.12",
    state: "attention",
    icon: TerminalSquare,
  },
  {
    name: "Tools",
    category: "Toolchain",
    detail: "mkcert ready · ngrok update available",
    version: "2 updates",
    state: "attention",
    icon: Wrench,
  },
  {
    name: "Watchers",
    category: "Automation",
    detail: "File watcher waiting to start",
    version: "Idle",
    state: "idle",
    icon: RefreshCw,
  },
  {
    name: "Workers",
    category: "Automation",
    detail: "Host mode · 1 worker failing",
    version: "2 profiles",
    state: "attention",
    icon: ListChecks,
  },
];

export const phpRuntimeFixtures = [
  {
    version: "8.3",
    detail: "Active FPM pool · 2 projects",
    state: "active",
    update: "Current",
  },
  {
    version: "8.4",
    detail: "Installed · no projects assigned",
    state: "idle",
    update: "Update ready",
  },
];

export const phpDetailTabs = [
  { value: "logs", label: "Logs" },
  { value: "configuration", label: "INI / Configuration" },
  { value: "ports", label: "Ports" },
  { value: "extensions", label: "Extensions" },
];

export const phpLogFixtures = [
  {
    time: "09:43:18",
    level: "INFO",
    message: "pool www ready · 2 active workers",
  },
  {
    time: "09:42:51",
    level: "INFO",
    message: "commerce-api request completed in 42ms",
  },
  {
    time: "09:41:07",
    level: "WARN",
    message: "max_children is close to the configured limit",
  },
];

export const phpIniFixtures = [
  { key: "memory_limit", value: "512M", source: "Chauffeur preset" },
  { key: "upload_max_filesize", value: "64M", source: "Project default" },
  { key: "opcache.enable", value: "On", source: "Runtime" },
  { key: "xdebug.mode", value: "off", source: "Default" },
];

export const phpPortFixtures = [
  {
    port: "9000",
    protocol: "FastCGI",
    detail: "PHP-FPM pool · localhost",
    state: "Ready",
  },
  {
    port: "8000",
    protocol: "HTTP",
    detail: "Development server preview",
    state: "Planned",
  },
];

export const phpExtensionFixtures = [
  {
    name: "pdo_pgsql",
    version: "8.3.12",
    state: "Installed",
    detail: "PostgreSQL client support",
  },
  {
    name: "intl",
    version: "8.3.12",
    state: "Installed",
    detail: "Internationalization helpers",
  },
  {
    name: "xdebug",
    version: "3.3.2",
    state: "Available",
    detail: "Debug bridge currently disabled",
  },
  {
    name: "imagick",
    version: "Not installed",
    state: "Planned",
    detail: "Image processing extension",
  },
];

export const systemSettingFixtures = [
  {
    label: "Autostart",
    detail: "Start Chauffeur with the desktop session",
    state: "Planned",
    icon: Play,
  },
  {
    label: "Idle suspension",
    detail: "Pause services after inactivity",
    state: "Planned",
    icon: Clock3,
  },
  {
    label: "Notifications",
    detail: "Show health and operation feedback",
    state: "Planned",
    icon: Bell,
  },
  {
    label: "LAN exposure",
    detail: "Share the dashboard on the local network",
    state: "Planned",
    icon: Network,
  },
  {
    label: "Remote dashboard",
    detail: "Configure LAN access and remote setup",
    state: "Planned",
    icon: Network,
  },
  {
    label: "Debug bridge",
    detail: "Enable the bridge and inspect diagnostic lenses",
    state: "Planned",
    icon: Bug,
  },
  {
    label: "Version updates",
    detail: "Review version, changelog, and terminal update",
    state: "Next",
    icon: RefreshCw,
  },
];

export const debugBridgeLensFixtures = [
  {
    name: "Dumps",
    detail: "Render dump payloads beside request output",
    state: "Available",
    icon: Eye,
  },
  {
    name: "Queries",
    detail: "Collect SQL timing and bindings",
    state: "Planned",
    icon: Database,
  },
  {
    name: "Requests",
    detail: "Inspect headers, route, and response timing",
    state: "Available",
    icon: Network,
  },
  {
    name: "Jobs",
    detail: "Trace queue payloads and failures",
    state: "Planned",
    icon: Activity,
  },
];

export const resourceSummary = [
  { label: "CPU", value: "18%", detail: "4 cores available", icon: Cpu },
  {
    label: "Memory",
    value: "4.8 GB",
    detail: "of 16 GB allocated",
    icon: Gauge,
  },
  { label: "Disk", value: "62%", detail: "112 GB of 180 GB", icon: HardDrive },
  {
    label: "Reclaimable",
    value: "1.8 GB",
    detail: "unused images and logs",
    icon: Database,
  },
];

export const commandPaletteEntries = [
  {
    group: "Pages",
    label: "Open Projects",
    detail: "View linked project workspaces",
    shortcut: "G P",
    icon: FolderGit2,
    to: "/projects" as const,
  },
  {
    group: "Services",
    label: "Inspect PHP 8.3",
    detail: "Review runtime health and ports",
    shortcut: "G S",
    icon: Server,
    to: "/system/php" as const,
  },
  {
    group: "Actions",
    label: "Run workspace doctor",
    detail: "Check routes, certificates, and tools",
    shortcut: "D",
    icon: Wrench,
    to: undefined,
  },
  {
    group: "Toggles",
    label: "Toggle sidebar",
    detail: "Change workspace navigation density",
    shortcut: "[",
    icon: Settings2,
    to: undefined,
  },
];

export const projectLinkPageFixture = {
  samplePath: "~/Projects/new-project",
  workspace: "Client work",
  runtime: "PHP 8.3 · Node 22.12",
  route: "new-project.test",
  https: "Requires backend",
  steps: projectLinkingSteps,
};

export const initializedPageFixtures = {
  runtime: {
    eyebrow: "System / core runtime",
    title: "Chauffeur runtime",
    description: "The local process boundary for the panel and CLI.",
    status: "Healthy",
    icon: Activity,
    metrics: [
      {
        label: "Connection",
        value: "Connected",
        detail: "Panel and CLI fixture",
      },
      { label: "Version", value: "v0.1.0", detail: "Current local build" },
      { label: "Controls", value: "Unavailable", detail: "Requires backend" },
    ],
    sections: [
      {
        title: "Runtime details",
        detail: "Process health, version, and local socket state belong here.",
      },
      {
        title: "Lifecycle",
        detail: "Restart, update, and diagnostics controls are planned.",
      },
    ],
  },
  network: {
    eyebrow: "System / gateway",
    title: "DNS and Nginx",
    description:
      "Local name resolution, routes, certificates, and gateway output.",
    status: "Healthy",
    icon: Globe2,
    metrics: [
      { label: "DNS", value: "Ready", detail: ".test resolver" },
      { label: "Nginx", value: "Online", detail: "Local gateway" },
      { label: "TLS", value: "1 attention", detail: "client-portal" },
    ],
    sections: [
      {
        title: "Resolver status",
        detail: "Show resolver health and domain registration state.",
      },
      {
        title: "Gateway logs",
        detail:
          "Static log states are ready; live streaming remains connected only where supported.",
      },
      {
        title: "Configuration",
        detail:
          "Configuration editing is planned and requires backend support.",
      },
    ],
  },
  node: {
    eyebrow: "System / toolchain",
    title: "Node.js and Bun",
    description: "JavaScript runtime versions and project assignment context.",
    status: "Attention",
    icon: TerminalSquare,
    metrics: [
      { label: "Node.js", value: "22.12", detail: "System-managed default" },
      { label: "Bun", value: "Not installed", detail: "Install planned" },
      { label: "Projects", value: "2", detail: "Using Node profiles" },
    ],
    sections: [
      {
        title: "Installed runtimes",
        detail: "Version cards, default selection, and manager ownership.",
      },
      {
        title: "Runtime actions",
        detail:
          "Install, remove, and default controls require backend support.",
      },
    ],
  },
  tools: {
    eyebrow: "System / automation",
    title: "Tools and watchers",
    description:
      "Certificate tools, tunnels, file watchers, and local automation.",
    status: "Attention",
    icon: Wrench,
    metrics: [
      { label: "Ready", value: "mkcert", detail: "Certificate tool" },
      { label: "Update", value: "ngrok", detail: "Version 3.18 available" },
      { label: "Watchers", value: "Idle", detail: "Start requires backend" },
    ],
    sections: [
      {
        title: "Tool availability",
        detail:
          "Show installed, missing, system-managed, and update-ready tools.",
      },
      {
        title: "Watcher profiles",
        detail:
          "Host and container execution modes are shown without starting a process.",
      },
    ],
  },
  debug: {
    eyebrow: "System / diagnostics",
    title: "Debug bridge",
    description: "Diagnostic lenses for request, query, dump, and job signals.",
    status: "Planned",
    icon: Bug,
    metrics: [
      { label: "Bridge", value: "Disabled", detail: "No process is attached" },
      { label: "Available", value: "2 lenses", detail: "Dumps and requests" },
      { label: "Planned", value: "2 lenses", detail: "Queries and jobs" },
    ],
    sections: [
      {
        title: "Diagnostic lenses",
        detail: "Show lens availability, sample records, and planned states.",
      },
      {
        title: "Bridge controls",
        detail: "Enable and disable actions require backend support.",
      },
    ],
  },
  workers: {
    eyebrow: "Resources / automation",
    title: "Workers",
    description:
      "Background execution profiles grouped by project and runtime.",
    status: "Attention",
    icon: ListChecks,
    metrics: [
      { label: "Active", value: "1", detail: "Queue worker" },
      { label: "Sleeping", value: "1", detail: "Scheduler" },
      { label: "Failed", value: "1", detail: "Reverb" },
    ],
    sections: [
      {
        title: "Execution states",
        detail:
          "Active, sleeping, failed, and stopped worker fixtures are represented below.",
      },
      {
        title: "Worker controls",
        detail:
          "Start, stop, and execution-mode actions require backend support.",
      },
    ],
  },
  usage: {
    eyebrow: "Resources / capacity",
    title: "Resource usage",
    description:
      "A workspace-level view of current local capacity and reclaimable storage.",
    status: "Available",
    icon: Gauge,
    metrics: resourceSummary.map((resource) => ({
      label: resource.label,
      value: resource.value,
      detail: resource.detail,
    })),
    sections: [
      {
        title: "Capacity summary",
        detail: "Use the cards above for the current static fixture snapshot.",
      },
      {
        title: "Collection state",
        detail:
          "Live telemetry collection is disconnected in the static phase.",
      },
    ],
  },
  telemetry: {
    eyebrow: "Resources / telemetry",
    title: "CPU and memory",
    description:
      "Detailed pressure indicators for local runtimes and containers.",
    status: "Disconnected",
    icon: Cpu,
    metrics: resourceSummary.slice(0, 2).map((resource) => ({
      label: resource.label,
      value: resource.value,
      detail: resource.detail,
    })),
    sections: [
      {
        title: "Static trend",
        detail:
          "A chart and time-range controls belong here; no live samples are collected yet.",
      },
      {
        title: "Disconnected state",
        detail: "Reconnect requires a telemetry backend and is not simulated.",
      },
    ],
  },
  issues: {
    eyebrow: "Resources / diagnostics",
    title: "Issues and diagnostics",
    description: "Warnings and failures that need review across the workspace.",
    status: "Attention",
    icon: CircleAlert,
    metrics: [
      { label: "Attention", value: "1", detail: "SSL is disabled" },
      { label: "Resolved", value: "2", detail: "Recent fixture events" },
      { label: "Doctor", value: "Planned", detail: "Requires backend" },
    ],
    sections: [
      {
        title: "Issue list",
        detail:
          "Show warning, failure, resolved, and empty states with clear next steps.",
      },
      {
        title: "Run diagnostics",
        detail:
          "Diagnostic execution is not connected and will not run from this page.",
      },
    ],
  },
} as const;

export const dedicatedProjectPageFixtures = {
  overview: {
    eyebrow: "Projects / overview",
    title: "Project overview",
    description:
      "Compare runtime, connected services, workers, and request health for a selected project.",
    status: "Static workspace data",
    icon: FileCode2,
    metrics: [
      {
        label: "Selected project",
        value: "commerce-api",
        detail: "Laravel · Commerce",
      },
      {
        label: "Runtime",
        value: "PHP 8.3",
        detail: "Node 22.12 · HTTPS ready",
      },
      { label: "Requests", value: "14", detail: "Last 15 minutes" },
    ],
    sections: [
      {
        title: "Runtime profile",
        detail:
          "PHP and Node version selectors belong here. Changes require backend support.",
      },
      {
        title: "Connected services",
        detail:
          "PostgreSQL and queue worker relationships are represented by workspace fixtures.",
      },
      {
        title: "Worker health",
        detail:
          "Running, sleeping, failed, and stopped worker states stay visible beside the project.",
      },
    ],
  },
  logs: {
    eyebrow: "Projects / observability",
    title: "Project logs",
    description:
      "Review application, runtime, development-server, and worker output for a selected project.",
    status: "Static logs",
    icon: FileText,
    metrics: [
      { label: "Source", value: "Application", detail: "4 fixture lines" },
      { label: "Warnings", value: "1", detail: "Certificate renewal pending" },
      {
        label: "Connection",
        value: "Disconnected",
        detail: "Live stream requires backend",
      },
    ],
    sections: [
      {
        title: "Log source filters",
        detail:
          "Application, PHP-FPM, dev server, queue, and framework worker sources belong here.",
      },
      {
        title: "Viewer states",
        detail:
          "Loading, empty, failed, and disconnected log viewer treatments are ready for real data.",
      },
      {
        title: "Copy and follow",
        detail:
          "Copy and follow controls remain non-submitting until a live log source is connected.",
      },
    ],
    rows: projectDetailLogLines.map((line) => ({
      label: `${line.time} · ${line.level}`,
      value: line.message,
      detail: "commerce-api · application fixture",
      status: line.level === "WARN" ? "Warning" : "Healthy",
    })),
  },
  environment: {
    eyebrow: "Projects / configuration",
    title: "Project environment",
    description:
      "Review masked environment values, inheritance, and worktree-specific configuration.",
    status: "Requires backend",
    icon: Braces,
    metrics: [
      { label: "Variables", value: "18", detail: "Fixture inventory" },
      {
        label: "Masked",
        value: "6",
        detail: "Secrets never shown in clear text",
      },
      { label: "Warnings", value: "1", detail: "Duplicate variable" },
    ],
    sections: [
      {
        title: "Environment table",
        detail:
          "Source, inherited value, worktree scope, and masked value columns belong here.",
      },
      {
        title: "Edit and restore",
        detail:
          "Edit, save, restore, and duplicate-variable actions require backend support.",
      },
    ],
    rows: [
      {
        label: "APP_ENV",
        value: "local",
        detail: "Project source",
        status: "Available",
      },
      {
        label: "APP_KEY",
        value: "••••••••",
        detail: "Project source · secret masked",
        status: "Masked",
      },
      {
        label: "DB_CONNECTION",
        value: "pgsql",
        detail: "Inherited from workspace",
        status: "Inherited",
      },
    ],
  },
  diagnostics: {
    eyebrow: "Projects / diagnostics",
    title: "Project diagnostics",
    description:
      "Organize health lenses for routes, dumps, queries, jobs, views, mail, cache, events, and HTTP.",
    status: "Planned",
    icon: Wrench,
    metrics: [
      { label: "Available lenses", value: "2", detail: "Routes and requests" },
      {
        label: "Needs backend",
        value: "6",
        detail: "Runtime diagnostic lenses",
      },
      { label: "Open issues", value: "1", detail: "HTTPS attention" },
    ],
    sections: [
      {
        title: "Diagnostic lenses",
        detail:
          "Lens cards, count badges, filters, and sample records belong on this page.",
      },
      {
        title: "Run diagnostics",
        detail:
          "The doctor action is disabled and explicitly requires backend support.",
      },
    ],
  },
  worktrees: {
    eyebrow: "Projects / source control",
    title: "Project worktrees",
    description:
      "Compare branches, domains, runtimes, databases, and worker context for a selected project.",
    status: "Static inventory",
    icon: GitBranch,
    metrics: [
      { label: "Main branch", value: "main", detail: "Primary worktree" },
      { label: "Worktrees", value: "1", detail: "Feature branch fixture" },
      {
        label: "Isolation",
        value: "Planned",
        detail: "Database isolation requires backend",
      },
    ],
    sections: [
      {
        title: "Branch selector",
        detail:
          "Main and feature worktree selection belongs here with project-specific metadata.",
      },
      {
        title: "Worktree actions",
        detail:
          "Add, remove, drop database, and sharing actions remain unavailable.",
      },
    ],
    rows: [
      {
        label: "main",
        value: "Primary",
        detail: "PHP 8.3 · commerce-api.test",
        status: "Active",
      },
      {
        label: "feature/checkout",
        value: "Worktree",
        detail: "PHP 8.3 · isolated route planned",
        status: "Planned",
      },
    ],
  },
} as const;

export const dedicatedServicePageFixtures = {
  details: {
    eyebrow: "Services / details",
    title: "Service details",
    description:
      "Inspect the selected service health, version, project usage, and local connection context.",
    status: "Static inventory",
    icon: FileCode2,
    metrics: [
      { label: "Service", value: "PostgreSQL", detail: "Databases" },
      { label: "Version", value: "16.4", detail: "Current fixture" },
      { label: "Projects", value: "2", detail: "Connected projects" },
    ],
    sections: [
      {
        title: "Health header",
        detail:
          "Status, version, category, ports, and dependency summary belong here.",
      },
      {
        title: "Service actions",
        detail:
          "Start, stop, restart, update, migrate, rollback, and remove require backend support.",
      },
    ],
    rows: [
      {
        label: "Health",
        value: "Healthy",
        detail: "Dependency checks pass",
        status: "Available",
      },
      {
        label: "Connection",
        value: "localhost:5432",
        detail: "Credentials remain masked",
        status: "Masked",
      },
    ],
  },
  databases: {
    eyebrow: "Services / data",
    title: "Databases and entities",
    description:
      "Review databases, entity groups, owners, sizes, and connection status for a selected service.",
    status: "Static inventory",
    icon: Database,
    metrics: [
      { label: "Databases", value: "3", detail: "Connected and testing" },
      { label: "Entities", value: "33", detail: "Tables, views, extensions" },
      { label: "Testing", value: "1", detail: "Isolated fixture" },
    ],
    sections: [
      {
        title: "Database cards",
        detail:
          "Name, size, owner project, status, and masked DSN actions belong here.",
      },
      {
        title: "Entity groups",
        detail:
          "Tables, views, extensions, and entity status are represented by fixtures.",
      },
      {
        title: "Lifecycle actions",
        detail: "Create, drop, and copy actions require backend support.",
      },
    ],
    rows: serviceDatabaseFixtures.map((database) => ({
      label: database.name,
      value: database.size,
      detail: `${database.owner} · ${database.status}`,
      status: database.status,
    })),
  },
  lifecycle: {
    eyebrow: "Services / data lifecycle",
    title: "Import, export, and snapshots",
    description:
      "Review safe data lifecycle states before connecting import, export, or restore operations.",
    status: "Requires backend",
    icon: DatabaseBackup,
    metrics: [
      { label: "Snapshots", value: "2", detail: "Ready fixtures" },
      { label: "Import", value: "Warning", detail: "Target already exists" },
      { label: "Restore", value: "Planned", detail: "No mutation performed" },
    ],
    sections: [
      {
        title: "Snapshot inventory",
        detail: "Age, size, checksum, and restore warning details belong here.",
      },
      {
        title: "Progress and failure",
        detail:
          "Importing, warning, error, and issue states are represented without fake progress.",
      },
    ],
    rows: serviceSnapshotFixtures.map((snapshot) => ({
      label: snapshot.name,
      value: snapshot.size,
      detail: snapshot.age,
      status: snapshot.state,
    })),
  },
  entities: {
    eyebrow: "Services / schema",
    title: "Entity actions",
    description:
      "Organize create, inspect, and delete actions for service entities without applying schema changes.",
    status: "Requires backend",
    icon: FileCode2,
    metrics: [
      { label: "Tables", value: "24", detail: "Orders, users, products" },
      { label: "Views", value: "6", detail: "Read-only fixtures" },
      { label: "Extensions", value: "3", detail: "Installed fixtures" },
    ],
    sections: [
      {
        title: "Entity groups",
        detail:
          "Group entities by kind with counts, descriptions, and dependency notes.",
      },
      {
        title: "Create and delete",
        detail:
          "Final-position controls are disabled until a schema API exists.",
      },
    ],
    rows: serviceEntityFixtures.map((entity) => ({
      label: entity.kind,
      value: entity.count,
      detail: entity.detail,
      status: "Available",
    })),
  },
  logs: {
    eyebrow: "Services / observability",
    title: "Service logs",
    description:
      "Review service output with source filters and clear connection state.",
    status: "Disconnected",
    icon: FileText,
    metrics: [
      {
        label: "Source",
        value: "PostgreSQL",
        detail: "Selected service fixture",
      },
      { label: "Lines", value: "3", detail: "Static log lines" },
      {
        label: "Connection",
        value: "Disconnected",
        detail: "Live stream requires backend",
      },
    ],
    sections: [
      {
        title: "Log sources",
        detail:
          "Service output, startup, health, and dependency log filters belong here.",
      },
      {
        title: "Viewer states",
        detail:
          "Loading, empty, failed, and disconnected states remain explicit.",
      },
    ],
    rows: serviceLogFixtures.map((line) => ({
      label: `${line.time} · ${line.level}`,
      value: line.message,
      detail: "PostgreSQL service fixture",
      status: line.level === "WARN" ? "Warning" : "Healthy",
    })),
  },
  configuration: {
    eyebrow: "Services / tuning",
    title: "Service configuration",
    description:
      "Review tuning values, ownership, and restart impact before making a service change.",
    status: "Requires backend",
    icon: Settings2,
    metrics: [
      { label: "Parameters", value: "12", detail: "Fixture configuration" },
      { label: "Changes", value: "0", detail: "No changes staged" },
      { label: "Restart", value: "Planned", detail: "Impact review required" },
    ],
    sections: [
      {
        title: "Tuning groups",
        detail:
          "Connection, memory, logging, and project pairing configuration belong here.",
      },
      {
        title: "Save and reset",
        detail:
          "Save, reset, and restart controls require backend support and remain disabled.",
      },
    ],
    rows: [
      {
        label: "max_connections",
        value: "100",
        detail: "Service preset",
        status: "Current",
      },
      {
        label: "shared_buffers",
        value: "128MB",
        detail: "Chauffeur preset",
        status: "Current",
      },
      {
        label: "log_min_duration_statement",
        value: "500ms",
        detail: "Workspace override",
        status: "Planned",
      },
    ],
  },
  dependencies: {
    eyebrow: "Services / networking",
    title: "Ports and dependencies",
    description:
      "Review service ports, conflicts, project links, and a static dependency map.",
    status: "Attention",
    icon: Network,
    metrics: [
      { label: "Ports", value: "5432", detail: "PostgreSQL localhost" },
      { label: "Dependencies", value: "Ready", detail: "Runtime map fixture" },
      { label: "Conflicts", value: "0", detail: "No current conflict" },
    ],
    sections: [
      {
        title: "Port inventory",
        detail:
          "Protocol, host binding, project usage, and conflict warnings belong here.",
      },
      {
        title: "Dependency map",
        detail:
          "Show service-to-project relationships without implying orchestration.",
      },
    ],
    rows: [
      {
        label: "5432",
        value: "Ready",
        detail: "PostgreSQL · localhost",
        status: "Available",
      },
      {
        label: "commerce-api",
        value: "Connected",
        detail: "Primary database",
        status: "Healthy",
      },
      {
        label: "client-portal",
        value: "Connected",
        detail: "Preview environment",
        status: "Healthy",
      },
    ],
  },
} as const;

export const onboardingSteps = [
  {
    number: "01",
    title: "Link a project",
    detail:
      "Register a local folder and give it a recognizable workspace identity.",
    status: "next" as const,
    action: "Plan project flow",
  },
  {
    number: "02",
    title: "Choose a runtime",
    detail:
      "Select the PHP version and service profile the project should use.",
    status: "planned" as const,
    action: "Planned",
  },
  {
    number: "03",
    title: "Verify the route",
    detail:
      "Confirm the domain, HTTPS certificate, and runtime health before coding.",
    status: "planned" as const,
    action: "Planned",
  },
];

export const featureGroups = [
  {
    label: "Projects",
    icon: FolderGit2,
    tone: "violet",
    features: [
      {
        name: "Link local projects",
        detail: "Register folders and keep their runtime context together.",
        state: "next" as const,
      },
      {
        name: "Local domains",
        detail: "Generate routes, aliases, and HTTPS certificates per project.",
        state: "planned" as const,
      },
      {
        name: "Runtime profiles",
        detail: "Choose PHP versions and project-specific service presets.",
        state: "planned" as const,
      },
    ],
  },
  {
    label: "Runtime",
    icon: Server,
    tone: "green",
    features: [
      {
        name: "Database containers",
        detail: "Start, stop, inspect, and connect to local databases.",
        state: "available" as const,
      },
      {
        name: "Logs and diagnostics",
        detail: "Follow service output and surface actionable failures.",
        state: "next" as const,
      },
      {
        name: "Workers and queues",
        detail: "Track background processes beside the project they serve.",
        state: "planned" as const,
      },
    ],
  },
  {
    label: "Data",
    icon: Database,
    tone: "amber",
    features: [
      {
        name: "Backup and restore",
        detail:
          "Create snapshots and recover a database without leaving the panel.",
        state: "available" as const,
      },
      {
        name: "Environment vault",
        detail:
          "Review connection details and keep secrets out of project files.",
        state: "planned" as const,
      },
      {
        name: "Resource telemetry",
        detail: "See CPU, memory, ports, and disk pressure at a glance.",
        state: "planned" as const,
      },
    ],
  },
];

export const operations = [
  {
    title: "Inspect runtime health",
    detail: "Review container status, ports, and recent output.",
    state: "Available",
    tone: "green",
    icon: Activity,
    to: "/containers" as const,
    action: "Open containers",
  },
  {
    title: "Recover a database",
    detail: "Choose a snapshot, preview its metadata, and restore safely.",
    state: "Available",
    tone: "green",
    icon: RefreshCw,
    to: "/containers/backup" as const,
    action: "Open backups",
  },
  {
    title: "Link a project workspace",
    detail: "Select a folder, assign a runtime, and publish a local route.",
    state: "Next",
    tone: "amber",
    icon: Link2,
    to: "/docs" as const,
    action: "View the plan",
  },
  {
    title: "Run a workspace doctor",
    detail: "Check ports, certificates, runtimes, and service dependencies.",
    state: "Planned",
    tone: "violet",
    icon: Wrench,
    to: "/docs" as const,
    action: "View the plan",
  },
];

export const recentActivity = [
  {
    icon: Check,
    tone: "success",
    title: "Project linked",
    detail: "commerce-api · 8.3 shared FPM",
    time: "4 min ago",
  },
  {
    icon: ShieldCheck,
    tone: "success",
    title: "Certificate ready",
    detail: "commerce-api.test + 1 alias",
    time: "5 min ago",
  },
  {
    icon: CircleAlert,
    tone: "warning",
    title: "SSL skipped",
    detail: "client-portal · mkcert not enabled",
    time: "8 min ago",
  },
];

export const activityPageFixtures = [
  ...recentActivity,
  {
    icon: RefreshCw,
    tone: "info",
    title: "Runtime check queued",
    detail: "Chauffeur runtime · static fixture",
    time: "12 min ago",
  },
];

export const quickActions = [
  {
    to: "/docs" as const,
    icon: Search,
    title: "Diagnose a project",
    detail: "Review doctor and health commands",
  },
  {
    to: "/containers" as const,
    icon: Database,
    title: "Manage containers",
    detail: "Start or stop local databases",
  },
  {
    to: "/containers/backup" as const,
    icon: HardDrive,
    title: "Open backups",
    detail: "Create or restore database backups",
  },
];

export type FeatureGroupFixture = (typeof featureGroups)[number];
export type OperationFixture = (typeof operations)[number];
export type LucideFixtureIcon = LucideIcon;
export { Container, DatabaseBackup, FileCode2 };
