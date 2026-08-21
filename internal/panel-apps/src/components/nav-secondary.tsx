"use client"

import * as React from "react"
import { Link, useRouterState } from "@tanstack/react-router"

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavSecondary({
  items,
  ...props
}: {
  items: Array<{
    title: string
    url: string
    icon: React.ReactNode
  }>
} & React.ComponentPropsWithoutRef<typeof SidebarGroup>) {
  const pathname = useRouterState({ select: (state) => state.location.pathname })

  return (
    <SidebarGroup {...props}>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton
                size="sm"
                isActive={pathname === item.url}
                aria-current={pathname === item.url ? "page" : undefined}
                render={<Link to={item.url} />}
              >
                {item.icon}
                <span>{item.title}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}
