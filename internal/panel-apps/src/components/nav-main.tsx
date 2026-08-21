"use client";

import * as React from "react";
import { Link, useRouterState } from "@tanstack/react-router";
import { ChevronRight } from "lucide-react";
import type { SidebarGroup as SidebarNavigationGroup } from "@/data/sidebar-destinations";

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  SidebarGroup as SidebarGroupPrimitive,
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
  items: Array<SidebarNavigationGroup>;
}) {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });
  const [openGroups, setOpenGroups] = React.useState<Record<string, boolean>>(
    {},
  );
  const isPathActive = (url: string) => {
    const destinationPath = url.split(/[?#]/, 1)[0] || "/";
    return destinationPath === "/"
      ? pathname === "/"
      : pathname === destinationPath || pathname.startsWith(`${destinationPath}/`);
  };

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
    <SidebarGroupPrimitive>
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
                render={<Link to={item.url} />}
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
                              isPathActive(subItem.url)
                            }
                            aria-current={
                              isPathActive(subItem.url)
                                ? "page"
                                : undefined
                            }
                            render={<Link to={subItem.url} />}
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
    </SidebarGroupPrimitive>
  );
}
