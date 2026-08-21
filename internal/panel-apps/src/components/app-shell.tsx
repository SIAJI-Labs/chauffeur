import type { ReactNode } from "react";

import type { BreadcrumbEntry } from "@/components/app-navbar";
import { AppNavbar } from "@/components/app-navbar";
import { AppSidebar } from "@/components/app-sidebar";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";

type AppShellProps = {
  title: string;
  breadcrumbs: Array<BreadcrumbEntry>;
  contentId: string;
  skipLabel: string;
  shortcuts?: ReactNode;
  children: ReactNode;
};

export function AppShell({
  title,
  breadcrumbs,
  contentId,
  skipLabel,
  shortcuts,
  children,
}: AppShellProps) {
  return (
    <SidebarProvider className="dashboard-frame">
      <a className="skip-link" href={`#${contentId}`}>
        {skipLabel}
      </a>
      <AppSidebar />
      <SidebarInset className="dashboard-shell">
        <AppNavbar
          title={title}
          breadcrumbs={breadcrumbs}
          shortcuts={shortcuts}
        />
        <main className="dashboard-content min-h-0 flex-1" id={contentId}>
          {children}
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}
