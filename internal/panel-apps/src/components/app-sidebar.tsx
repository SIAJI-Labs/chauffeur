import * as React from "react"

import {
  BookOpen,
  Container,
  FileCode,
  Globe,
  HardDrive,
  Home,
  LayoutDashboard,
  List,
  Network,
  RotateCcw,
  ScrollText,
  Server,
  Settings,
  Settings2,
  Shield,
} from "lucide-react"
import { NavMain } from "@/components/nav-main"
import { NavSecondary } from "@/components/nav-secondary"
import { NavUser } from "@/components/nav-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

const data = {
  user: {
    name: "Chauffeur",
    email: "admin@chauffeur.dev",
    avatar: "",
  },
  navMain: [
    {
      title: "Podman Service",
      url: "#",
      icon: (
        <Server className="size-4" />
      ),
      isActive: true,
      items: [
        {
          title: "Container List",
          url: "/containers",
          icon: <Container className="size-4" />,
        },
        {
          title: "Backup",
          url: "/containers/backup",
          icon: <HardDrive className="size-4" />,
        },
        {
          title: "Restore",
          url: "#",
          icon: <RotateCcw className="size-4" />,
        },
        {
          title: "Logs",
          url: "#",
          icon: <ScrollText className="size-4" />,
        },
      ],
    },
    {
      title: "Chauf Service",
      url: "#",
      icon: (
        <Globe className="size-4" />
      ),
      items: [
        {
          title: "Site List",
          url: "#",
          icon: <List className="size-4" />,
        },
        {
          title: "Config",
          url: "#",
          icon: <FileCode className="size-4" />,
        },
        {
          title: "DNS",
          url: "#",
          icon: <Network className="size-4" />,
        },
        {
          title: "SSL",
          url: "#",
          icon: <Shield className="size-4" />,
        },
      ],
    },
    {
      title: "Settings",
      url: "#",
      icon: (
        <Settings className="size-4" />
      ),
      items: [
        {
          title: "General",
          url: "#",
          icon: <Settings2 className="size-4" />,
        },
        {
          title: "Appearance",
          url: "#",
          icon: <LayoutDashboard className="size-4" />,
        },
        {
          title: "About",
          url: "#",
          icon: <BookOpen className="size-4" />,
        },
      ],
    },
  ],
  navSecondary: [
    {
      title: "Documentation",
      url: "/docs",
      icon: (
        <BookOpen className="size-4" />
      ),
    },
  ],
}
export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar variant="inset" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" render={<a href="/" />}>
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                <Home className="size-4" />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">Dashboard</span>
                <span className="truncate text-xs">Overview</span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
    </Sidebar>
  )
}
