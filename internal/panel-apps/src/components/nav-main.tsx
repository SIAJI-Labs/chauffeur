"use client";

import * as React from "react";
import { Link, useRouterState } from "@tanstack/react-router";
import { ChevronRight } from "lucide-react";

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  SidebarGroup,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from "@/components/ui/sidebar";

export function NavMain({
  items,
}: {
  items: Array<{
    title: string;
    url: string;
    icon: React.ReactNode;
    isActive?: boolean;
    items?: Array<{
      title: string;
      url: string;
      icon?: React.ReactNode;
      status?: string;
      disabled?: boolean;
    }>;
    disabled?: boolean;
  }>;
}) {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });
  const [openGroups, setOpenGroups] = React.useState<Record<string, boolean>>(
    {},
  );
  const isPathActive = (url: string) =>
    url === "/"
      ? pathname === "/"
      : pathname === url || pathname.startsWith(`${url}/`);

  React.useEffect(() => {
    setOpenGroups((current) => {
      const next = { ...current };
      for (const item of items) {
        if (item.isActive || isPathActive(item.url)) next[item.title] = true;
      }
      return next;
    });
  }, [items, pathname]);

  return (
    <SidebarGroup>
      <SidebarMenu>
        {items.map((item) => {
          const groupActive = isPathActive(item.url);
          return (
            <Collapsible
              key={item.title}
              open={
                openGroups[item.title] ??
                (item.isActive === true || groupActive)
              }
              onOpenChange={(open) =>
                setOpenGroups((current) => ({
                  ...current,
                  [item.title]: open,
                }))
              }
              render={<SidebarMenuItem />}
            >
              <SidebarMenuButton
                tooltip={item.title}
                isActive={groupActive && !item.items?.length}
                aria-current={
                  groupActive && !item.items?.length ? "page" : undefined
                }
                disabled={item.disabled}
                render={item.disabled ? undefined : <Link to={item.url} />}
              >
                {item.icon}
                <span>{item.title}</span>
              </SidebarMenuButton>
              {item.items?.length ? (
                <>
                  <CollapsibleTrigger
                    render={
                      <SidebarMenuAction className="aria-expanded:rotate-90" />
                    }
                  >
                    <ChevronRight className="size-4" />
                    <span className="sr-only">Toggle</span>
                  </CollapsibleTrigger>
                  <CollapsibleContent>
                    <SidebarMenuSub>
                      {item.items.map((subItem) => (
                        <SidebarMenuSubItem key={subItem.title}>
                          <SidebarMenuSubButton
                            isActive={
                              !subItem.disabled && isPathActive(subItem.url)
                            }
                            aria-current={
                              !subItem.disabled && isPathActive(subItem.url)
                                ? "page"
                                : undefined
                            }
                            aria-disabled={subItem.disabled || undefined}
                            tabIndex={subItem.disabled ? -1 : undefined}
                            className={
                              subItem.disabled ? "planned-nav-item" : undefined
                            }
                            render={
                              subItem.disabled ? (
                                <span />
                              ) : (
                                <Link to={subItem.url} />
                              )
                            }
                          >
                            {subItem.icon}
                            <span className="nav-subitem-title">
                              {subItem.title}
                            </span>
                            {subItem.status ? (
                              <span
                                className={`sidebar-status-badge ${subItem.status.toLowerCase().replaceAll(" ", "-")}`}
                              >
                                {subItem.status}
                              </span>
                            ) : null}
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                      ))}
                    </SidebarMenuSub>
                  </CollapsibleContent>
                </>
              ) : null}
            </Collapsible>
          );
        })}
      </SidebarMenu>
    </SidebarGroup>
  );
}
