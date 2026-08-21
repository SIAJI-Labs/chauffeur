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

export type DestinationStatus = "Available" | "Next" | "Planned";

export type SidebarDestination = {
  title: string;
  url: string;
  icon: ReactNode;
  status?: DestinationStatus;
  action?: "command-palette";
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
        destination("Recent activity", "/activity", <Activity />, "Available"),
        {
          ...destination("Command palette", "/", <Search />, "Available"),
          action: "command-palette" as const,
        },
      ],
    },
    {
      title: "Projects",
      url: "/projects",
      icon: <FolderGit2 />,
      items: [
        destination("Project list", "/projects", <FolderGit2 />, "Available"),
        destination("Link a project", "/projects/link", <GitBranch />, "Next"),
        destination(
          "Project overview",
          "/projects/overview",
          <FileCode2 />,
          "Planned",
        ),
        destination("Logs", "/projects/logs", <FileText />, "Planned"),
        destination(
          "Environment",
          "/projects/environment",
          <Braces />,
          "Planned",
        ),
        destination(
          "Diagnostics",
          "/projects/diagnostics",
          <Wrench />,
          "Planned",
        ),
        destination(
          "Worktrees",
          "/projects/worktrees",
          <GitBranch />,
          "Planned",
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
          "/services/details",
          <FileCode2 />,
          "Planned",
        ),
        destination(
          "Service catalog",
          "/services/catalog",
          <Boxes />,
          "Available",
        ),
        destination(
          "Databases and entities",
          "/services/databases",
          <Database />,
          "Planned",
        ),
        destination(
          "Import, export, snapshots",
          "/services/data-lifecycle",
          <DatabaseBackup />,
          "Planned",
        ),
        destination(
          "Entity actions",
          "/services/entities",
          <FileCode2 />,
          "Planned",
        ),
        destination("Service logs", "/services/logs", <FileText />, "Planned"),
        destination(
          "Configuration",
          "/services/configuration",
          <Settings2 />,
          "Planned",
        ),
        destination(
          "Ports and dependencies",
          "/services/dependencies",
          <Network />,
          "Planned",
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
          "/system/runtime",
          <Activity />,
          "Available",
        ),
        destination(
          "DNS and Nginx",
          "/system/network",
          <Globe2 />,
          "Available",
        ),
        destination("PHP runtimes", "/system/php", <Server />, "Available"),
        destination(
          "Node.js and Bun",
          "/system/node",
          <TerminalSquare />,
          "Available",
        ),
        destination(
          "Tools and watchers",
          "/system/tools",
          <Wrench />,
          "Available",
        ),
        destination("Debug bridge", "/system/debug", <Bug />, "Available"),
        destination(
          "Settings and updates",
          "/settings",
          <Settings2 />,
          "Available",
        ),
      ],
    },
    {
      title: "Resources",
      url: "/containers",
      icon: <HardDrive />,
      items: [
        destination(
          "Database containers",
          "/containers",
          <Container />,
          "Available",
        ),
        destination(
          "Backups",
          "/containers/backup",
          <DatabaseBackup />,
          "Available",
        ),
        destination(
          "Workers",
          "/resources/workers",
          <ListChecks />,
          "Available",
        ),
        destination(
          "Resource usage",
          "/resources/usage",
          <Gauge />,
          "Available",
        ),
        destination(
          "CPU and memory",
          "/resources/telemetry",
          <Cpu />,
          "Available",
        ),
        destination(
          "Issues and diagnostics",
          "/resources/issues",
          <CircleAlert />,
          "Available",
        ),
      ],
    },
  ],
  navSecondary: [
    destination("Documentation", "/docs", <BookOpen />, "Available"),
  ],
};
