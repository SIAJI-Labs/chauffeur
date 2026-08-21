import {
  Activity,
  Bell,
  Boxes,
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
  FolderGit2,
  FolderOpen,
  Gauge,
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
  },
  {
    group: "Services",
    label: "Inspect PHP 8.3",
    detail: "Review runtime health and ports",
    shortcut: "G S",
    icon: Server,
  },
  {
    group: "Actions",
    label: "Run workspace doctor",
    detail: "Check routes, certificates, and tools",
    shortcut: "D",
    icon: Wrench,
  },
  {
    group: "Toggles",
    label: "Toggle sidebar",
    detail: "Change workspace navigation density",
    shortcut: "[",
    icon: Settings2,
  },
];

export const previewDestinationFixtures = {
  "recent-activity": {
    eyebrow: "Workspace signal",
    title: "Recent activity",
    description:
      "A single timeline for project, service, runtime, and backup events.",
    status: "Preview",
    icon: Activity,
    sections: [
      { label: "Healthy", value: "2 events", detail: "Certificates and project routes" },
      { label: "Attention", value: "1 event", detail: "client-portal needs HTTPS" },
      { label: "Disconnected", value: "No live feed", detail: "Activity streaming requires backend support" },
    ],
  },
  "command-palette": {
    eyebrow: "Workspace shortcut",
    title: "Command palette",
    description: "Search pages, projects, services, toggles, and future actions from one keyboard-first surface.",
    status: "Preview",
    icon: Search,
    sections: [
      { label: "Shortcut", value: "Ctrl / Cmd + K", detail: "Open the global palette" },
      { label: "Available", value: "Navigation", detail: "Open Projects, Services, and System" },
      { label: "Planned", value: "Mutations", detail: "Actions stay previews until backend support exists" },
    ],
  },
  "project-linking": {
    eyebrow: "Projects / onboarding",
    title: "Link a project",
    description: "Preview the folder, runtime, and local route review used before linking a project.",
    status: "Next",
    icon: FolderGit2,
    sections: projectLinkingSteps.map((step) => ({ label: step.number, value: step.title, detail: step.detail })),
  },
  "service-catalog": {
    eyebrow: "Services / discovery",
    title: "Service catalog",
    description: "Browse service presets without installing or changing any local service.",
    status: "Preview",
    icon: Boxes,
    sections: serviceCatalogFixtures.slice(0, 3).map((service) => ({ label: service.category, value: service.name, detail: `${service.state} · ${service.detail}` })),
  },
  "chauffeur-runtime": {
    eyebrow: "System / core",
    title: "Chauffeur runtime",
    description: "Review the panel and CLI runtime connection without changing the host process.",
    status: "Preview",
    icon: Activity,
    sections: [{ label: "Connected", value: "v0.1.0", detail: "Runtime health is represented by static fixtures" }, { label: "Planned", value: "Restart and update", detail: "Requires backend" }],
  },
  "dns-nginx": {
    eyebrow: "System / gateway",
    title: "DNS and Nginx",
    description: "Inspect local resolver and gateway health, logs, and configuration previews.",
    status: "Preview",
    icon: Globe2,
    sections: [{ label: "DNS", value: "Ready", detail: ".test resolver fixture" }, { label: "Nginx", value: "Online", detail: "Local routes and TLS gateway" }, { label: "Configuration", value: "Planned", detail: "Editing requires backend" }],
  },
  "node-runtimes": {
    eyebrow: "System / toolchain",
    title: "Node.js and Bun",
    description: "Compare installed JavaScript runtimes and their project assignment states.",
    status: "Preview",
    icon: TerminalSquare,
    sections: [{ label: "Ready", value: "Node.js 22.12", detail: "System-managed default" }, { label: "Planned", value: "Bun", detail: "Not installed · install preview" }],
  },
  "system-tools": {
    eyebrow: "System / toolchain",
    title: "Tools and watchers",
    description: "Review certificate, tunnel, and file-watcher availability before connecting controls.",
    status: "Preview",
    icon: Wrench,
    sections: [{ label: "Ready", value: "mkcert", detail: "Certificate tool available" }, { label: "Update", value: "ngrok", detail: "Version 3.18 is available" }, { label: "Idle", value: "File watcher", detail: "Start action requires backend" }],
  },
  "debug-bridge": {
    eyebrow: "System / diagnostics",
    title: "Debug bridge",
    description: "Choose diagnostic lenses without enabling a bridge or exposing application data.",
    status: "Preview",
    icon: Bug,
    sections: debugBridgeLensFixtures.slice(0, 3).map((lens) => ({ label: lens.state, value: lens.name, detail: lens.detail })),
  },
  workers: {
    eyebrow: "Resources / automation",
    title: "Workers",
    description: "See active, sleeping, failed, and stopped worker profiles across the workspace.",
    status: "Preview",
    icon: ListChecks,
    sections: workerStates.map((worker) => ({ label: worker.state, value: worker.name, detail: worker.detail })),
  },
  resources: {
    eyebrow: "Resources / capacity",
    title: "Resource usage",
    description: "Review static CPU, memory, disk, and reclaimable storage signals.",
    status: "Preview",
    icon: Gauge,
    sections: resourceSummary.map((resource) => ({ label: resource.label, value: resource.value, detail: resource.detail })),
  },
  "resource-telemetry": {
    eyebrow: "Resources / telemetry",
    title: "CPU and memory",
    description: "Inspect the telemetry surface planned for runtime and container pressure.",
    status: "Preview",
    icon: Cpu,
    sections: resourceSummary.slice(0, 2).map((resource) => ({ label: resource.label, value: resource.value, detail: resource.detail })),
  },
  issues: {
    eyebrow: "Resources / diagnostics",
    title: "Issues and diagnostics",
    description: "Collect actionable workspace warnings without running a live doctor command.",
    status: "Preview",
    icon: CircleAlert,
    sections: [{ label: "Attention", value: "SSL disabled", detail: "client-portal needs certificate setup" }, { label: "Planned", value: "Workspace doctor", detail: "Checks require backend support" }],
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
