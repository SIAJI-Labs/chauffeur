import * as React from "react";
import { Link } from "@tanstack/react-router";
import { NavMain } from "@/components/nav-main";
import { NavSecondary } from "@/components/nav-secondary";
import { sidebarNavigation } from "@/data/sidebar-destinations";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

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
        <NavMain items={sidebarNavigation.navMain} />
        <NavSecondary items={sidebarNavigation.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <div className="sidebar-footer-note">
          <span className="live-dot" aria-hidden="true" /> Workspace online
        </div>
      </SidebarFooter>
    </Sidebar>
  );
}
