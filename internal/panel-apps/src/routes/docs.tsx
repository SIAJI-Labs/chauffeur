import { createFileRoute } from "@tanstack/react-router";
import {
  Activity,
  ArrowRight,
  BookOpen,
  Container,
  ExternalLink,
  FileText,
  FolderGit2,
  Globe,
  MessageCircle,
} from "lucide-react";
import { AppShell } from "@/components/app-shell";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export const Route = createFileRoute("/docs")({
  component: DocsPage,
});

function DocsPage() {
  return (
    <AppShell
      title="Documentation"
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Documentation" },
      ]}
      contentId="docs-content"
      skipLabel="Skip to documentation"
    >
      <div className="flex flex-1 flex-col gap-6 overflow-y-auto p-6">
        <div>
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
              <PlannedDocItem>Introduction</PlannedDocItem>
              <PlannedDocItem>Installation</PlannedDocItem>
              <PlannedDocItem>Quick Start Guide</PlannedDocItem>
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
              <PlannedDocItem>Podman Containers</PlannedDocItem>
              <PlannedDocItem>Database Backups</PlannedDocItem>
              <PlannedDocItem>Site Management</PlannedDocItem>
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
              <PlannedDocItem>CLI Commands</PlannedDocItem>
              <PlannedDocItem>Configuration</PlannedDocItem>
              <PlannedDocItem>API Reference</PlannedDocItem>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Product surface</CardTitle>
            <CardDescription>
              The planned panel follows the path from local project to healthy
              runtime.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="grid gap-3 md:grid-cols-3">
              <RoadmapItem
                icon={FolderGit2}
                title="Project workspace"
                detail="Link folders, assign PHP runtimes, and publish local domains."
                status="Next"
              />
              <RoadmapItem
                icon={Container}
                title="Runtime operations"
                detail="Operate databases, workers, logs, ports, and service health."
                status="In progress"
              />
              <RoadmapItem
                icon={Activity}
                title="Diagnostics"
                detail="Turn signals into guided checks for certificates and dependencies."
                status="Planned"
              />
            </div>
            <div className="flex flex-wrap items-center gap-2 border-t pt-4 text-sm text-muted-foreground">
              <Badge variant="outline">Signal</Badge>
              <ArrowRight className="size-4" aria-hidden="true" />
              <Badge variant="outline">Operation</Badge>
              <ArrowRight className="size-4" aria-hidden="true" />
              <Badge variant="outline">Verification</Badge>
              <span className="basis-full text-xs md:basis-auto md:ml-2">
                Every future action should leave a visible, reviewable result.
              </span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Chauffeur Panel v0.1.0</CardTitle>
            <CardDescription>
              Web-based admin panel for managing podman database containers
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap gap-2">
              <Badge variant="outline">Go Backend</Badge>
              <Badge variant="outline">React + TanStack</Badge>
              <Badge variant="outline">shadcn/ui</Badge>
              <Badge variant="outline">Tailwind v4</Badge>
            </div>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <a
                href="https://github.com/siegg/chauffeur"
                target="_blank"
                rel="noreferrer"
                className="flex items-center gap-1 hover:text-foreground"
              >
                <Globe className="size-4" />
                GitHub
              </a>
              <a
                href="https://github.com/siegg/chauffeur/discussions"
                target="_blank"
                rel="noreferrer"
                className="flex items-center gap-1 hover:text-foreground"
              >
                <MessageCircle className="size-4" />
                Discussions
              </a>
            </div>
          </CardContent>
        </Card>
      </div>
    </AppShell>
  );
}

function PlannedDocItem({ children }: { children: string }) {
  return (
    <span className="docs-planned-item">
      <span>{children}</span>
      <Badge variant="outline">Planned</Badge>
    </span>
  );
}

function RoadmapItem({
  icon: Icon,
  title,
  detail,
  status,
}: {
  icon: typeof Activity;
  title: string;
  detail: string;
  status: string;
}) {
  return (
    <div className="space-y-3 border bg-muted/20 p-4">
      <div className="flex items-center justify-between gap-3">
        <Icon className="size-5 text-primary" aria-hidden="true" />
        <Badge variant="outline">{status}</Badge>
      </div>
      <div>
        <h3 className="font-medium">{title}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{detail}</p>
      </div>
    </div>
  );
}
