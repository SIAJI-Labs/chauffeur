import * as React from "react";
import { Link, createFileRoute } from "@tanstack/react-router";
import {
  ArrowUpRight,
  ChevronRight,
  CircleAlert,
  FolderGit2,
  GitBranch,
  Globe2,
  Link2,
  Search,
  Server,
} from "lucide-react";
import { AppNavbar } from "@/components/app-navbar";
import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ScrollArea } from "@/components/ui/scroll-area";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import {
  projectGroupOptions,
  projectLinkingSteps,
  projectListStates,
  projectSortOptions,
  projects,
  workspaceManagementSteps,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/projects")({
  component: ProjectsPage,
});

type ProjectGroup = "workspace" | "framework";
type ProjectSort = "name" | "status" | "recent";

function ProjectsPage() {
  const [search, setSearch] = React.useState("");
  const [groupBy, setGroupBy] = React.useState<ProjectGroup>("workspace");
  const [sortBy, setSortBy] = React.useState<ProjectSort>("name");
  const [isLinkDialogOpen, setIsLinkDialogOpen] = React.useState(false);
  const [isWorkspaceDialogOpen, setIsWorkspaceDialogOpen] =
    React.useState(false);
  const deferredSearch = React.useDeferredValue(search);

  const visibleProjects = projects
    .filter((project) => {
      const query = deferredSearch.trim().toLowerCase();
      if (!query) return true;
      return [project.name, project.type, project.workspace, project.url]
        .join(" ")
        .toLowerCase()
        .includes(query);
    })
    .sort((left, right) => {
      if (sortBy === "status") return left.state.localeCompare(right.state);
      if (sortBy === "recent") {
        return (
          new Date(right.updatedAt).getTime() -
          new Date(left.updatedAt).getTime()
        );
      }
      return left.name.localeCompare(right.name);
    });

  const projectGroups = visibleProjects.reduce<
    Record<string, Array<(typeof projects)[number]>>
  >((groups, project) => {
    const key = groupBy === "workspace" ? project.workspace : project.type;
    groups[key] ??= [];
    groups[key].push(project);
    return groups;
  }, {});

  return (
    <SidebarProvider className="dashboard-frame">
      <a className="skip-link" href="#projects-content">
        Skip to project list
      </a>
      <AppSidebar />
      <SidebarInset className="dashboard-shell">
        <AppNavbar
          title="Projects"
          breadcrumbs={[{ label: "Workspace", to: "/" }, { label: "Projects" }]}
          shortcuts={
            <Button
              variant="outline"
              size="sm"
              className="header-button"
              render={<Link to="/" />}
            >
              <ArrowUpRight aria-hidden="true" />
              <span>Overview</span>
            </Button>
          }
        />
        <main
          className="dashboard-content min-h-0 flex-1"
          id="projects-content"
        >
          <ScrollArea className="dashboard-scroll-area">
            <div className="projects-page">
              <section className="projects-intro">
                <div>
                  <p className="section-kicker">Linked workspaces</p>
                  <h2>Projects that know how to run locally.</h2>
                  <p>
                    Review paths, runtimes, domains, and HTTPS readiness before
                    opening a project detail workspace.
                  </p>
                </div>
                <div className="projects-intro-actions">
                  <Button onClick={() => setIsLinkDialogOpen(true)}>
                    <Link2 aria-hidden="true" />
                    Link a project
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => setIsWorkspaceDialogOpen(true)}
                  >
                    Manage workspaces
                  </Button>
                </div>
              </section>

              <ProjectStateStrip />

              <section
                className="projects-list-panel"
                aria-labelledby="project-list-title"
              >
                <div className="projects-list-heading">
                  <div>
                    <p className="section-kicker">Project inventory</p>
                    <h3 id="project-list-title">
                      {visibleProjects.length} linked project
                      {visibleProjects.length === 1 ? "" : "s"}
                    </h3>
                  </div>
                  <Badge variant="outline">Static preview</Badge>
                </div>
                <div className="project-list-controls">
                  <label className="project-search">
                    <span className="sr-only">Search projects</span>
                    <Search aria-hidden="true" />
                    <Input
                      value={search}
                      onChange={(event) => setSearch(event.target.value)}
                      placeholder="Search projects, frameworks, or domains"
                    />
                  </label>
                  <Select
                    value={groupBy}
                    onValueChange={(value) => {
                      if (value) setGroupBy(value as ProjectGroup);
                    }}
                  >
                    <SelectTrigger
                      aria-label="Group projects by"
                      className="project-select"
                    >
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {projectGroupOptions.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          Group: {option.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <Select
                    value={sortBy}
                    onValueChange={(value) => {
                      if (value) setSortBy(value as ProjectSort);
                    }}
                  >
                    <SelectTrigger
                      aria-label="Sort projects by"
                      className="project-select"
                    >
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {projectSortOptions.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          Sort: {option.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="sr-only" aria-live="polite">
                  Showing {visibleProjects.length} of {projects.length} projects
                </div>

                {visibleProjects.length === 0 ? (
                  <div className="project-filter-empty">
                    <FolderGit2 aria-hidden="true" />
                    <strong>No matching projects</strong>
                    <span>Try a different project, framework, or domain.</span>
                  </div>
                ) : (
                  <div className="project-groups">
                    {Object.entries(projectGroups).map(
                      ([group, groupProjects]) => (
                        <section className="project-group" key={group}>
                          <div className="project-group-heading">
                            <h4>{group}</h4>
                            <span>
                              {groupProjects.length} project
                              {groupProjects.length === 1 ? "" : "s"}
                            </span>
                          </div>
                          <div className="project-inventory-list">
                            {groupProjects.map((project) => (
                              <ProjectInventoryRow
                                key={project.name}
                                project={project}
                              />
                            ))}
                          </div>
                        </section>
                      ),
                    )}
                  </div>
                )}
              </section>

              <section
                className="projects-next-panel"
                aria-labelledby="projects-next-title"
              >
                <div>
                  <p className="section-kicker">Next layer</p>
                  <h3 id="projects-next-title">
                    A project detail workspace is planned
                  </h3>
                  <p>
                    Logs, environment, worktrees, diagnostics, and runtime
                    controls will live behind each project row.
                  </p>
                </div>
                <div className="projects-next-items">
                  <span>
                    <Server aria-hidden="true" /> Runtime profiles
                  </span>
                  <span>
                    <GitBranch aria-hidden="true" /> Worktrees
                  </span>
                  <span>
                    <CircleAlert aria-hidden="true" /> Diagnostics
                  </span>
                </div>
              </section>
            </div>
          </ScrollArea>
        </main>
      </SidebarInset>

      <Dialog open={isLinkDialogOpen} onOpenChange={setIsLinkDialogOpen}>
        <DialogContent className="project-link-dialog">
          <DialogHeader>
            <DialogTitle>Link a project</DialogTitle>
            <DialogDescription>
              Preview of the future project linking flow. No folder or runtime
              changes are made in the static phase.
            </DialogDescription>
          </DialogHeader>
          <div className="project-link-steps">
            {projectLinkingSteps.map((step) => (
              <div className="project-link-step" key={step.number}>
                <span>{step.number}</span>
                <div>
                  <strong>{step.title}</strong>
                  <p>{step.detail}</p>
                </div>
                <Badge variant="outline">Planned</Badge>
              </div>
            ))}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsLinkDialogOpen(false)}
            >
              Close preview
            </Button>
            <Button disabled>Continue to folder picker</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog
        open={isWorkspaceDialogOpen}
        onOpenChange={setIsWorkspaceDialogOpen}
      >
        <DialogContent className="workspace-management-dialog">
          <DialogHeader>
            <DialogTitle>Manage workspaces</DialogTitle>
            <DialogDescription>
              Static preview of workspace creation, rename, organization, and
              delete safeguards. No folders or project membership change.
            </DialogDescription>
          </DialogHeader>
          <div className="project-link-steps">
            {workspaceManagementSteps.map((step, index) => (
              <div className="project-link-step" key={step.title}>
                <span>{String(index + 1).padStart(2, "0")}</span>
                <div>
                  <strong>{step.title}</strong>
                  <p>{step.detail}</p>
                </div>
                <Badge variant="outline">Planned</Badge>
              </div>
            ))}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsWorkspaceDialogOpen(false)}
            >
              Close preview
            </Button>
            <Button disabled>Continue to workspace editor</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SidebarProvider>
  );
}

function ProjectStateStrip() {
  return (
    <section
      className="project-state-strip"
      aria-label="Project list state examples"
    >
      {projectListStates.map((state) => {
        const Icon = state.icon;
        return (
          <div
            className={`project-state-example ${state.tone}`}
            key={state.label}
          >
            <Icon aria-hidden="true" />
            <span>
              <strong>{state.label}</strong>
              <small>{state.detail}</small>
            </span>
          </div>
        );
      })}
    </section>
  );
}

function ProjectInventoryRow({
  project,
}: {
  project: (typeof projects)[number];
}) {
  const isHealthy = project.state === "healthy";
  return (
    <article className="project-inventory-row">
      <span
        className={`project-inventory-status ${project.state}`}
        aria-label={isHealthy ? "Healthy" : "Needs attention"}
      />
      <div className="project-inventory-main">
        <div className="project-inventory-title">
          <strong>{project.name}</strong>
          <Badge variant="outline">{project.type}</Badge>
          <Badge
            variant="outline"
            className={`project-state-badge ${project.state}`}
          >
            {isHealthy ? "Healthy" : "Attention"}
          </Badge>
        </div>
        <span className="project-inventory-path">{project.path}</span>
        <a
          className="project-inventory-url"
          href={project.url}
          target="_blank"
          rel="noreferrer"
        >
          <Globe2 aria-hidden="true" />
          {project.url}
        </a>
      </div>
      <div className="project-inventory-meta">
        <span>PHP {project.php}</span>
        <span>Node {project.node}</span>
        <span>{project.https} HTTPS</span>
        <small>
          <GitBranch aria-hidden="true" /> {project.worktrees}
        </small>
      </div>
      <Button
        variant="outline"
        size="sm"
        className="project-inventory-action"
        render={<Link to="/projects/$name" params={{ name: project.name }} />}
        aria-label={`Open ${project.name} project detail`}
      >
        Details
        <ChevronRight aria-hidden="true" />
      </Button>
    </article>
  );
}
