import {
  Activity,
  BookOpen,
  Boxes,
  Braces,
  Bug,
  CircleAlert,
  Container,
  Cpu,
  Database,
  DatabaseBackup,
  FileCode2,
  FileText,
  FolderGit2,
  Gauge,
  GitBranch,
  Globe2,
  HardDrive,
  LayoutDashboard,
  ListChecks,
  Network,
  Search,
  Server,
  Settings2,
  ShieldCheck,
  TerminalSquare,
  Wrench,
} from "lucide-react";
import type { ReactNode } from "react";

export type DestinationStatus = "Available" | "Preview" | "Next" | "Planned";

export type SidebarDestination = {
  title: string;
  url: string;
  icon: ReactNode;
  status?: DestinationStatus;
};

export type SidebarGroup = SidebarDestination & {
  isActive?: boolean;
  items?: Array<SidebarDestination>;
};

const destination = (
  title: string,
  url: string,
  icon: ReactNode,
  status: DestinationStatus,
): SidebarDestination => ({ title, url, icon, status });

export const sidebarNavigation: {
  navMain: Array<SidebarGroup>;
  navSecondary: Array<SidebarDestination>;
} = {
  navMain: [
    {
      title: "Workspace",
      url: "/",
      icon: <LayoutDashboard />,
      items: [
        destination("Overview", "/", <LayoutDashboard />, "Available"),
        destination(
          "Recent activity",
          "/preview/recent-activity",
          <Activity />,
          "Preview",
        ),
        destination(
          "Command palette",
          "/preview/command-palette",
          <Search />,
          "Preview",
        ),
      ],
    },
    {
      title: "Projects",
      url: "/projects",
      icon: <FolderGit2 />,
      items: [
        destination("Project list", "/projects", <FolderGit2 />, "Available"),
        destination(
          "Link a project",
          "/preview/project-linking",
          <GitBranch />,
          "Next",
        ),
        destination(
          "Project overview",
          "/projects/commerce-api#overview",
          <FileCode2 />,
          "Available",
        ),
        destination("Logs", "/projects/commerce-api#logs", <FileText />, "Available"),
        destination(
          "Environment",
          "/projects/commerce-api#environment",
          <Braces />,
          "Available",
        ),
        destination(
          "Diagnostics",
          "/projects/commerce-api#diagnostics",
          <Wrench />,
          "Available",
        ),
        destination(
          "Worktrees",
          "/projects/commerce-api#worktrees",
          <GitBranch />,
          "Available",
        ),
      ],
    },
    {
      title: "Services",
      url: "/services",
      icon: <Server />,
      items: [
        destination("Installed services", "/services", <Server />, "Available"),
        destination(
          "Service details",
          "/services/PostgreSQL#admin",
          <FileCode2 />,
          "Available",
        ),
        destination(
          "Service catalog",
          "/preview/service-catalog",
          <Boxes />,
          "Preview",
        ),
        destination(
          "Databases and entities",
          "/services/PostgreSQL#databases",
          <Database />,
          "Available",
        ),
        destination(
          "Import, export, snapshots",
          "/services/PostgreSQL#databases",
          <DatabaseBackup />,
          "Preview",
        ),
        destination(
          "Entity actions",
          "/services/PostgreSQL#entities",
          <FileCode2 />,
          "Preview",
        ),
        destination(
          "Service logs",
          "/services/PostgreSQL#logs",
          <FileText />,
          "Available",
        ),
        destination(
          "Configuration",
          "/services/PostgreSQL#configuration",
          <Settings2 />,
          "Preview",
        ),
        destination(
          "Ports and dependencies",
          "/services/PostgreSQL#ports",
          <Network />,
          "Preview",
        ),
      ],
    },
    {
      title: "System",
      url: "/system",
      icon: <Settings2 />,
      items: [
        destination("System health", "/system", <ShieldCheck />, "Available"),
        destination(
          "Chauffeur runtime",
          "/preview/chauffeur-runtime",
          <Activity />,
          "Preview",
        ),
        destination("DNS and Nginx", "/preview/dns-nginx", <Globe2 />, "Preview"),
        destination("PHP runtimes", "/system/php", <Server />, "Available"),
        destination(
          "Node.js and Bun",
          "/preview/node-runtimes",
          <TerminalSquare />,
          "Preview",
        ),
        destination(
          "Tools and watchers",
          "/preview/system-tools",
          <Wrench />,
          "Preview",
        ),
        destination("Debug bridge", "/preview/debug-bridge", <Bug />, "Preview"),
        destination("Settings and updates", "/settings", <Settings2 />, "Available"),
      ],
    },
    {
      title: "Resources",
      url: "/containers",
      icon: <HardDrive />,
      items: [
        destination("Database containers", "/containers", <Container />, "Available"),
        destination("Backups", "/containers/backup", <DatabaseBackup />, "Available"),
        destination("Workers", "/preview/workers", <ListChecks />, "Preview"),
        destination("Resource usage", "/preview/resources", <Gauge />, "Preview"),
        destination(
          "CPU and memory",
          "/preview/resource-telemetry",
          <Cpu />,
          "Preview",
        ),
        destination(
          "Issues and diagnostics",
          "/preview/issues",
          <CircleAlert />,
          "Preview",
        ),
      ],
    },
  ],
  navSecondary: [destination("Documentation", "/docs", <BookOpen />, "Available")],
};
