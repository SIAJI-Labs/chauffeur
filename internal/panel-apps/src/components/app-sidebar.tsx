import * as React from "react";
import { Link } from "@tanstack/react-router";
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
import { NavMain } from "@/components/nav-main";
import { NavSecondary } from "@/components/nav-secondary";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

const data = {
  navMain: [
    {
      title: "Workspace",
      url: "/",
      icon: <LayoutDashboard />,
      items: [
        {
          title: "Overview",
          url: "/",
          icon: <LayoutDashboard />,
          status: "Available",
        },
        {
          title: "Recent activity",
          url: "#activity",
          icon: <Activity />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Command palette",
          url: "#command-palette",
          icon: <Search />,
          status: "Planned",
          disabled: true,
        },
      ],
    },
    {
      title: "Projects",
      url: "/projects",
      icon: <FolderGit2 />,
      items: [
        {
          title: "Project list",
          url: "/projects",
          icon: <FolderGit2 />,
          status: "Available",
        },
        {
          title: "Link a project",
          url: "#project-linking",
          icon: <GitBranch />,
          status: "Next",
          disabled: true,
        },
        {
          title: "Project overview",
          url: "#project-overview",
          icon: <FileCode2 />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Logs",
          url: "#project-logs",
          icon: <FileText />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Environment",
          url: "#project-environment",
          icon: <Braces />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Diagnostics",
          url: "#project-diagnostics",
          icon: <Wrench />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Worktrees",
          url: "#project-worktrees",
          icon: <GitBranch />,
          status: "Planned",
          disabled: true,
        },
      ],
    },
    {
      title: "Services",
      url: "/services",
      icon: <Server />,
      items: [
        {
          title: "Installed services",
          url: "/services",
          icon: <Server />,
          status: "Available",
        },
        {
          title: "Service details",
          url: "#service-details",
          icon: <FileCode2 />,
          status: "Next",
          disabled: true,
        },
        {
          title: "Service catalog",
          url: "#service-catalog",
          icon: <Boxes />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Databases and entities",
          url: "#databases",
          icon: <Database />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Import, export, snapshots",
          url: "#database-lifecycle",
          icon: <DatabaseBackup />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Entity actions",
          url: "#entity-actions",
          icon: <FileCode2 />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Service logs",
          url: "#service-logs",
          icon: <FileText />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Configuration",
          url: "#service-configuration",
          icon: <Settings2 />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Ports and dependencies",
          url: "#service-dependencies",
          icon: <Network />,
          status: "Planned",
          disabled: true,
        },
      ],
    },
    {
      title: "System",
      url: "/system",
      icon: <Settings2 />,
      items: [
        {
          title: "System health",
          url: "/system",
          icon: <ShieldCheck />,
          status: "Available",
        },
        {
          title: "Chauffeur runtime",
          url: "#chauffeur-runtime",
          icon: <Activity />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "DNS and Nginx",
          url: "#dns-nginx",
          icon: <Globe2 />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "PHP runtimes",
          url: "/system/php",
          icon: <Server />,
          status: "Available",
        },
        {
          title: "Node.js and Bun",
          url: "#node-runtimes",
          icon: <TerminalSquare />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Tools and watchers",
          url: "#system-tools",
          icon: <Wrench />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Debug bridge",
          url: "#debug-bridge",
          icon: <Bug />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Settings and updates",
          url: "#system-settings",
          icon: <Settings2 />,
          status: "Planned",
          disabled: true,
        },
      ],
    },
    {
      title: "Resources",
      url: "/containers",
      icon: <HardDrive />,
      items: [
        {
          title: "Database containers",
          url: "/containers",
          icon: <Container />,
          status: "Available",
        },
        {
          title: "Backups",
          url: "/containers/backup",
          icon: <DatabaseBackup />,
          status: "Available",
        },
        {
          title: "Workers",
          url: "#workers",
          icon: <ListChecks />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Resource usage",
          url: "#resources",
          icon: <Gauge />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "CPU and memory",
          url: "#resource-telemetry",
          icon: <Cpu />,
          status: "Planned",
          disabled: true,
        },
        {
          title: "Issues and diagnostics",
          url: "#issues",
          icon: <CircleAlert />,
          status: "Planned",
          disabled: true,
        },
      ],
    },
  ],
  navSecondary: [{ title: "Documentation", url: "/docs", icon: <BookOpen /> }],
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar variant="inset" {...props}>
      <SidebarHeader className="app-sidebar-header">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" render={<Link to="/" />}>
              <span className="brand-glyph" aria-hidden="true">
                C
              </span>
              <span className="brand-copy">
                <strong>Chauffeur</strong>
                <small>Local PHP workspace</small>
              </span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <div className="sidebar-footer-note">
          <span className="live-dot" aria-hidden="true" /> Workspace online
        </div>
      </SidebarFooter>
    </Sidebar>
  );
}
