import { createFileRoute } from "@tanstack/react-router"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { AppSidebar } from "@/components/app-sidebar"
import { Separator } from "@/components/ui/separator"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { BookOpen, ExternalLink, FileText, Globe, MessageCircle } from "lucide-react"

export const Route = createFileRoute("/docs")({
  component: DocsPage,
})

function DocsPage() {
  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="/docs">Documentation</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator className="hidden md:block">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  className="lucide lucide-chevron-right"
                >
                  <path d="m9 18 6-6-6-6" />
                </svg>
              </BreadcrumbSeparator>
              <BreadcrumbItem>
                <BreadcrumbPage className="font-normal text-foreground">Overview</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold">Documentation</h1>
            <p className="text-muted-foreground mt-1">
              Welcome to the Chauffeur Panel documentation
            </p>
          </div>

          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <BookOpen className="size-4" />
                  Getting Started
                </CardTitle>
                <CardDescription>Learn the basics of Chauffeur</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <a href="#" className="block text-sm text-muted-foreground hover:text-foreground">
                  Introduction
                </a>
                <a href="#" className="block text-sm text-muted-foreground hover:text-foreground">
                  Installation
                </a>
                <a href="#" className="block text-sm text-muted-foreground hover:text-foreground">
                  Quick Start Guide
                </a>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <FileText className="size-4" />
                  Core Concepts
                </CardTitle>
                <CardDescription>Understand the fundamentals</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <a href="#" className="block text-sm text-muted-foreground hover:text-foreground">
                  Podman Containers
                </a>
                <a href="#" className="block text-sm text-muted-foreground hover:text-foreground">
                  Database Backups
                </a>
                <a href="#" className="block text-sm text-muted-foreground hover:text-foreground">
                  Site Management
                </a>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <ExternalLink className="size-4" />
                  References
                </CardTitle>
                <CardDescription>API and CLI references</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <a href="#" className="block text-sm text-muted-foreground hover:text-foreground">
                  CLI Commands
                </a>
                <a href="#" className="block text-sm text-muted-foreground hover:text-foreground">
                  Configuration
                </a>
                <a href="#" className="block text-sm text-muted-foreground hover:text-foreground">
                  API Reference
                </a>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Chauffeur Panel v0.1.0</CardTitle>
              <CardDescription>Web-based admin panel for managing podman database containers</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex flex-wrap gap-2">
                <Badge variant="outline">Go Backend</Badge>
                <Badge variant="outline">React + TanStack</Badge>
                <Badge variant="outline">shadcn/ui</Badge>
                <Badge variant="outline">Tailwind v4</Badge>
              </div>
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <a href="#" className="flex items-center gap-1 hover:text-foreground">
                  <Globe className="size-4" />
                  GitHub
                </a>
                <a href="#" className="flex items-center gap-1 hover:text-foreground">
                  <MessageCircle className="size-4" />
                  Discussions
                </a>
              </div>
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}