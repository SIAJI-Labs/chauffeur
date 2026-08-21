"use client";

import * as React from "react";
import { Link, useRouterState } from "@tanstack/react-router";
import { ChevronRight } from "lucide-react";
import type { SidebarGroup as SidebarNavigationGroup } from "@/data/sidebar-destinations";

import { Collapsible, CollapsibleContent } from "@/components/ui/collapsible";
import {
  SidebarGroup as SidebarGroupPrimitive,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from "@/components/ui/sidebar";

export function NavMain({ items }: { items: Array<SidebarNavigationGroup> }) {
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
      : pathname === destinationPath ||
          pathname.startsWith(`${destinationPath}/`);
  };
  const isSubitemActive = (
    subItem: NonNullable<SidebarNavigationGroup["items"]>[number],
  ) => subItem.action !== "command-palette" && isPathActive(subItem.url);
  const isGroupActive = (item: SidebarNavigationGroup) =>
    isPathActive(item.url) ||
    item.items?.some((subItem) => isSubitemActive(subItem)) === true;

  React.useEffect(() => {
    setOpenGroups((current) => {
      const next = { ...current };
      for (const item of items) {
        if (item.isActive || isGroupActive(item)) next[item.title] = true;
      }
      return next;
    });
  }, [items, pathname]);

  return (
    <SidebarGroupPrimitive>
      <SidebarMenu>
        {items.map((item) => {
          const hasSubitems = Boolean(item.items?.length);
          const groupActive = isGroupActive(item);
          const groupOpen =
            openGroups[item.title] ?? (item.isActive === true || groupActive);
          const groupContentId = `sidebar-group-${item.title
            .toLowerCase()
            .replaceAll(/[^a-z0-9]+/g, "-")}`;
          return (
            <Collapsible
              key={item.title}
              open={groupOpen}
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
                isActive={groupActive}
                aria-current={groupActive && !hasSubitems ? "page" : undefined}
                aria-expanded={hasSubitems ? groupOpen : undefined}
                aria-controls={hasSubitems ? groupContentId : undefined}
                render={hasSubitems ? undefined : <Link to={item.url} />}
                onClick={
                  hasSubitems
                    ? () =>
                        setOpenGroups((current) => ({
                          ...current,
                          [item.title]: !groupOpen,
                        }))
                    : undefined
                }
              >
                {item.icon}
                <span>{item.title}</span>
                {hasSubitems ? (
                  <ChevronRight
                    aria-hidden="true"
                    className={`ms-auto size-4 transition-transform ${
                      groupOpen ? "rotate-90" : ""
                    }`}
                  />
                ) : null}
              </SidebarMenuButton>
              {item.items?.length ? (
                <CollapsibleContent id={groupContentId}>
                  <SidebarMenuSub>
                    {item.items.map((subItem) => (
                      <SidebarMenuSubItem key={subItem.title}>
                        <SidebarMenuSubButton
                          isActive={isSubitemActive(subItem)}
                          aria-current={
                            isSubitemActive(subItem) ? "page" : undefined
                          }
                          render={<Link to={subItem.url} />}
                          onClick={(event) => {
                            if (subItem.action !== "command-palette") return;
                            event.preventDefault();
                            window.dispatchEvent(
                              new Event("chauffeur:open-command-palette"),
                            );
                          }}
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
              ) : null}
            </Collapsible>
          );
        })}
      </SidebarMenu>
    </SidebarGroupPrimitive>
  );
}
