import { Fragment } from "react";
import { Link, useLocation } from "@tanstack/react-router";
import {
  Boxes,
  FolderGit2,
  LayoutDashboard,
  Server,
  Settings2,
} from "lucide-react";
import type { ReactNode } from "react";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";

export type NavbarPath =
  | "/"
  | "/projects"
  | "/services"
  | "/system"
  | "/containers"
  | "/containers/backup"
  | "/docs";

export type BreadcrumbEntry = {
  label: string;
  to?: NavbarPath;
};

export type AppNavbarProps = {
  title: string;
  breadcrumbs: Array<BreadcrumbEntry>;
  shortcuts?: ReactNode;
};

const mobileNavigation = [
  { label: "Overview", to: "/", icon: LayoutDashboard },
  { label: "Projects", to: "/projects", icon: FolderGit2 },
  { label: "Services", to: "/services", icon: Server },
  { label: "System", to: "/system", icon: Settings2 },
  { label: "Resources", to: "/containers", icon: Boxes },
] as const;

export function AppNavbar({ title, breadcrumbs, shortcuts }: AppNavbarProps) {
  const location = useLocation();

  return (
    <>
      <header className="app-header">
        <div className="header-left">
          <SidebarTrigger className="sidebar-trigger" />
          <Separator orientation="vertical" className="header-rule" />
          <div className="navbar-context">
            <Breadcrumb>
              <BreadcrumbList>
                {breadcrumbs.map((item, index) => {
                  const current = index === breadcrumbs.length - 1;
                  return (
                    <Fragment key={`${item.label}-${index}`}>
                      {index > 0 && <BreadcrumbSeparator />}
                      <BreadcrumbItem>
                        {item.to && !current ? (
                          <BreadcrumbLink render={<Link to={item.to} />}>
                            {item.label}
                          </BreadcrumbLink>
                        ) : (
                          <BreadcrumbPage>{item.label}</BreadcrumbPage>
                        )}
                      </BreadcrumbItem>
                    </Fragment>
                  );
                })}
              </BreadcrumbList>
            </Breadcrumb>
            <h1>{title}</h1>
          </div>
        </div>
        {shortcuts && <div className="header-actions">{shortcuts}</div>}
      </header>
      <nav className="mobile-bottom-nav" aria-label="Primary navigation">
        {mobileNavigation.map((item) => {
          const Icon = item.icon;
          const active =
            item.to === "/"
              ? location.pathname === "/"
              : location.pathname === item.to ||
                location.pathname.startsWith(`${item.to}/`);
          return (
            <Link
              key={item.to}
              to={item.to}
              className="mobile-bottom-nav-link"
              aria-current={active ? "page" : undefined}
            >
              <Icon aria-hidden="true" />
              <span>{item.label}</span>
            </Link>
          );
        })}
      </nav>
    </>
  );
}
