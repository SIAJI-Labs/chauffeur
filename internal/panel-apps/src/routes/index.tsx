import { Link, createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import {
  Activity,
  ArrowUpRight,
  ChevronRight,
  Clock3,
  Container,
  DatabaseBackup,
  ExternalLink,
  FileCode2,
  Gauge,
  ListChecks,
  Play,
  Search,
  Settings2,
  TerminalSquare,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { AppSidebar } from "@/components/app-sidebar";
import { AppNavbar } from "@/components/app-navbar";
import {
  commandPaletteEntries,
  featureGroups,
  healthHero,
  healthStates,
  onboardingSteps,
  operations,
  projects,
  quickActions,
  recentActivity,
  resourceSummary,
  services,
  systemHealth,
  workerStates,
  workspaceSignals,
  workspaceSummary,
} from "@/data/webui-fixtures";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";

export const Route = createFileRoute("/")({ component: DashboardPage });

function DashboardPage() {
  const [showFeedback, setShowFeedback] = useState(false);
  const today = new Intl.DateTimeFormat("en", {
    weekday: "long",
    month: "long",
    day: "numeric",
    year: "numeric",
  }).format(new Date());

  return (
    <SidebarProvider className="dashboard-frame">
      <a className="skip-link" href="#dashboard-content">
        Skip to workspace overview
      </a>
      <AppSidebar />
      <SidebarInset className="dashboard-shell">
        <AppNavbar
          title="Overview"
          breadcrumbs={[{ label: "Workspace", to: "/" }, { label: "Overview" }]}
          shortcuts={
            <>
              <nav
                className="header-shortcuts"
                aria-label="Workspace shortcuts"
              >
                <Button
                  variant="ghost"
                  size="sm"
                  className="shortcut-button"
                  render={<Link to="/containers" />}
                >
                  <Container aria-hidden="true" />
                  <span>Containers</span>
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="shortcut-button"
                  render={<Link to="/containers/backup" />}
                >
                  <DatabaseBackup aria-hidden="true" />
                  <span>Backups</span>
                </Button>
              </nav>
              <div className="live-indicator">
                <span aria-hidden="true" /> Runtime online
              </div>
              <Button
                variant="outline"
                size="sm"
                className="header-button"
                render={<Link to="/docs" />}
              >
                <TerminalSquare aria-hidden="true" />
                <span>CLI docs</span>
              </Button>
            </>
          }
        />

        <main
          className="dashboard-content min-h-0 flex-1"
          id="dashboard-content"
        >
          <ScrollArea className="dashboard-scroll-area">
            <div className="dashboard-section">
              <section className="dashboard-intro">
                <div>
                  <p className="section-kicker">{today}</p>
                  <h2>Your local workspace, clearly mapped.</h2>
                  <p className="intro-copy">
                    Review linked projects, runtime health, and recent
                    operations without leaving your development flow.
                  </p>
                </div>
                <Link to="/containers" className="text-link">
                  Manage containers <ArrowUpRight aria-hidden="true" />
                </Link>
              </section>

              <HealthHero />

              <section className="summary-grid" aria-label="Workspace summary">
                {workspaceSummary.map((item) => (
                  <SummaryCard key={item.label} item={item} />
                ))}
              </section>

              <OnboardingPanel />

              <section className="signal-grid" aria-label="Workspace signals">
                {workspaceSignals.map((signal) => (
                  <SignalCard key={signal.label} {...signal} />
                ))}
              </section>

              <div className="dashboard-observability-grid">
                <WorkersPanel />
                <SystemHealthPanel />
                <ResourcePanel />
              </div>

              <div className="content-grid">
                <section className="panel-card project-panel">
                  <div className="panel-heading">
                    <div>
                      <p className="section-kicker">Project desk</p>
                      <h3>Linked projects</h3>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      className="quiet-button"
                      render={<Link to="/docs" />}
                    >
                      <Settings2 /> Project guide
                    </Button>
                  </div>
                  <div className="project-list">
                    {projects.map((project) => (
                      <ProjectRow key={project.name} project={project} />
                    ))}
                  </div>
                  <div className="panel-footer">
                    <FileCode2 aria-hidden="true" /> Project actions remain
                    available through the CLI.
                  </div>
                </section>

                <section className="panel-card service-panel">
                  <div className="panel-heading">
                    <div>
                      <p className="section-kicker">Runtime</p>
                      <h3>Service pulse</h3>
                    </div>
                    <span className="pulse-mark" aria-hidden="true">
                      <Activity />
                    </span>
                  </div>
                  <div className="service-list">
                    {services.map((service) => {
                      const Icon = service.icon;
                      return (
                        <div className="service-row" key={service.name}>
                          <div className="service-icon">
                            <Icon />
                          </div>
                          <div className="service-copy">
                            <strong>{service.name}</strong>
                            <span>{service.detail}</span>
                          </div>
                          {service.updateAvailable && (
                            <Badge variant="outline" className="service-update">
                              Update
                            </Badge>
                          )}
                          <StatusDot state={service.state} />
                        </div>
                      );
                    })}
                  </div>
                  <Link to="/services" className="panel-link">
                    View all services <ChevronRight />
                  </Link>
                </section>
              </div>

              <div className="content-grid lower-grid">
                <section className="panel-card activity-panel">
                  <div className="panel-heading">
                    <div>
                      <p className="section-kicker">Activity</p>
                      <h3>Recent operations</h3>
                    </div>
                    <div className="activity-heading-actions">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setShowFeedback(true)}
                      >
                        Test feedback
                      </Button>
                      <Clock3 className="heading-icon" aria-hidden="true" />
                    </div>
                  </div>
                  <div className="activity-list">
                    {recentActivity.map((item) => (
                      <ActivityRow key={item.title} {...item} />
                    ))}
                  </div>
                </section>
                <section className="panel-card quick-panel">
                  <div className="panel-heading">
                    <div>
                      <p className="section-kicker">Next actions</p>
                      <h3>Keep moving</h3>
                    </div>
                    <Play className="heading-icon" />
                  </div>
                  {quickActions.map((action) => (
                    <QuickAction key={action.title} {...action} />
                  ))}
                </section>
              </div>

              <CommandPalettePreview />

              <section
                className="roadmap-panel"
                aria-labelledby="capability-map-title"
              >
                <div className="panel-heading">
                  <div>
                    <p className="section-kicker">Capability map</p>
                    <h3 id="capability-map-title">
                      Everything Chauffeur is building around your workspace
                    </h3>
                  </div>
                  <Badge variant="outline">Skeleton roadmap</Badge>
                </div>
                <div className="feature-groups">
                  {featureGroups.map((group) => (
                    <FeatureGroup key={group.label} group={group} />
                  ))}
                </div>
              </section>

              <section
                className="operations-panel"
                aria-labelledby="operations-title"
              >
                <div className="panel-heading">
                  <div>
                    <p className="section-kicker">Operations</p>
                    <h3 id="operations-title">
                      A clear path from signal to action
                    </h3>
                  </div>
                  <ListChecks className="heading-icon" />
                </div>
                <div className="operation-list">
                  {operations.map((operation) => (
                    <OperationRow key={operation.title} operation={operation} />
                  ))}
                </div>
              </section>
            </div>
          </ScrollArea>
        </main>
      </SidebarInset>
      {showFeedback ? (
        <div className="toast-preview" role="status" aria-live="polite">
          <div>
            <strong>Preview saved</strong>
            <span>No local operation was started.</span>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowFeedback(false)}
            aria-label="Dismiss feedback"
          >
            Dismiss
          </Button>
        </div>
      ) : null}
    </SidebarProvider>
  );
}

function HealthHero() {
  const AttentionIcon = healthStates[1].icon;
  return (
    <section className="health-hero" aria-labelledby="health-hero-title">
      <div className="health-hero-main">
        <span className="health-hero-icon" aria-hidden="true">
          <AttentionIcon />
        </span>
        <div className="health-hero-copy">
          <p className="section-kicker">Workspace health</p>
          <h3 id="health-hero-title">{healthHero.title}</h3>
          <p>{healthHero.detail}</p>
        </div>
        <Badge variant="outline" className="health-hero-status">
          Attention
        </Badge>
        <Button
          variant="outline"
          size="sm"
          className="health-hero-action"
          render={<Link to="/docs" />}
        >
          {healthHero.action}
          <ArrowUpRight aria-hidden="true" />
        </Button>
      </div>
      <div className="health-state-list" aria-label="Health state examples">
        {healthStates.map((state) => {
          const Icon = state.icon;
          return (
            <div className={`health-state ${state.tone}`} key={state.label}>
              <Icon aria-hidden="true" />
              <span>
                <strong>{state.label}</strong>
                <small>{state.detail}</small>
              </span>
            </div>
          );
        })}
      </div>
    </section>
  );
}

function SummaryCard({ item }: { item: (typeof workspaceSummary)[number] }) {
  const Icon = item.icon;
  return (
    <article className="summary-card">
      <div className="summary-card-top">
        <span>{item.label}</span>
        <Icon aria-hidden="true" />
      </div>
      <strong>{item.value}</strong>
      <small>{item.detail}</small>
    </article>
  );
}

function OnboardingPanel() {
  return (
    <section className="onboarding-panel" aria-labelledby="onboarding-title">
      <div className="onboarding-heading">
        <div>
          <p className="section-kicker">Start here</p>
          <h3 id="onboarding-title">Build a workspace that is ready to run</h3>
          <p>
            A static preview of the future project setup flow. Existing
            container actions remain available below.
          </p>
        </div>
        <Badge variant="outline">Static preview</Badge>
      </div>
      <div className="onboarding-steps">
        {onboardingSteps.map((step) => (
          <OnboardingStep key={step.number} step={step} />
        ))}
      </div>
    </section>
  );
}

function OnboardingStep({ step }: { step: (typeof onboardingSteps)[number] }) {
  const isNext = step.status === "next";
  return (
    <article className="onboarding-step">
      <span className="onboarding-number" aria-hidden="true">
        {step.number}
      </span>
      <div className="onboarding-step-copy">
        <div className="onboarding-step-title">
          <strong>{step.title}</strong>
          <Badge variant="outline" className={`feature-status ${step.status}`}>
            {isNext ? "Next" : "Planned"}
          </Badge>
        </div>
        <p>{step.detail}</p>
        {isNext ? (
          <Button
            variant="outline"
            size="sm"
            className="onboarding-action"
            render={<Link to="/docs" />}
          >
            {step.action}
            <ArrowUpRight aria-hidden="true" />
          </Button>
        ) : (
          <Button
            variant="outline"
            size="sm"
            className="onboarding-action"
            disabled
          >
            {step.action}
          </Button>
        )}
      </div>
    </article>
  );
}

function WorkersPanel() {
  return (
    <section className="observability-panel" aria-labelledby="workers-title">
      <div className="panel-heading">
        <div>
          <p className="section-kicker">Workers</p>
          <h3 id="workers-title">Background jobs</h3>
        </div>
        <Activity className="heading-icon" />
      </div>
      <div className="worker-list">
        {workerStates.map((worker) => {
          const Icon = worker.icon;
          return (
            <div className="worker-row" key={worker.name}>
              <span className={`worker-icon ${worker.state}`}>
                <Icon aria-hidden="true" />
              </span>
              <span className="worker-copy">
                <strong>{worker.name}</strong>
                <small>{worker.detail}</small>
              </span>
              <span className={`worker-state ${worker.state}`}>
                {worker.state}
              </span>
            </div>
          );
        })}
      </div>
    </section>
  );
}

function SystemHealthPanel() {
  return (
    <section
      className="observability-panel"
      aria-labelledby="system-health-title"
    >
      <div className="panel-heading">
        <div>
          <p className="section-kicker">System</p>
          <h3 id="system-health-title">Health checks</h3>
        </div>
        <Badge variant="outline" className="health-check-badge">
          4 / 4 ready
        </Badge>
      </div>
      <div className="system-health-list">
        {systemHealth.map((item) => {
          const Icon = item.icon;
          return (
            <div className="system-health-row" key={item.name}>
              <Icon aria-hidden="true" />
              <span>
                <strong>{item.name}</strong>
                <small>{item.detail}</small>
              </span>
              <span className="health-check-dot" aria-label="Healthy" />
            </div>
          );
        })}
      </div>
    </section>
  );
}

function ResourcePanel() {
  return (
    <section className="observability-panel" aria-labelledby="resource-title">
      <div className="panel-heading">
        <div>
          <p className="section-kicker">Resources</p>
          <h3 id="resource-title">Local capacity</h3>
        </div>
        <Gauge className="heading-icon" />
      </div>
      <div className="resource-list">
        {resourceSummary.map((item) => {
          const Icon = item.icon;
          return (
            <div className="resource-row" key={item.label}>
              <div className="resource-label">
                <span>
                  <Icon aria-hidden="true" /> {item.label}
                </span>
                <strong>{item.value}</strong>
              </div>
              <div className="resource-meter" aria-hidden="true">
                <span
                  className={`resource-meter-fill ${item.label.toLowerCase()}`}
                />
              </div>
              <small>{item.detail}</small>
            </div>
          );
        })}
      </div>
    </section>
  );
}

function CommandPalettePreview() {
  return (
    <section
      className="command-palette-panel"
      aria-labelledby="command-palette-title"
    >
      <div className="command-palette-heading">
        <div>
          <p className="section-kicker">Keyboard first</p>
          <h3 id="command-palette-title">Find any workspace action</h3>
          <p>
            Preview of the command surface for pages, services, toggles, and
            diagnostics.
          </p>
        </div>
        <Badge variant="outline">Planned</Badge>
      </div>
      <div
        className="command-palette-window"
        aria-label="Command palette preview"
      >
        <div className="command-search">
          <Search aria-hidden="true" />
          <span>Search projects, services, or actions</span>
          <kbd>Cmd K</kbd>
        </div>
        <div className="command-entry-list">
          {commandPaletteEntries.map((entry) => {
            const Icon = entry.icon;
            return (
              <div className="command-entry" key={entry.label}>
                <Icon aria-hidden="true" />
                <span>
                  <strong>{entry.label}</strong>
                  <small>{entry.detail}</small>
                </span>
                <em>{entry.shortcut}</em>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}

function SignalCard({
  label,
  value,
  detail,
  tone,
  icon: Icon,
}: {
  label: string;
  value: string;
  detail: string;
  tone: string;
  icon: LucideIcon;
}) {
  return (
    <article className={`signal-card signal-${tone}`}>
      <div className="signal-top">
        <span>{label}</span>
        <span className="signal-icon">
          <Icon aria-hidden="true" />
        </span>
      </div>
      <strong>{value}</strong>
      <small>{detail}</small>
    </article>
  );
}

function ProjectRow({ project }: { project: (typeof projects)[number] }) {
  return (
    <div className="project-row">
      <div
        className={`project-state ${project.state}`}
        aria-label={project.state === "healthy" ? "Healthy" : "Needs attention"}
      >
        <span />
      </div>
      <div className="project-main">
        <div className="project-title">
          <strong>{project.name}</strong>
          <Badge variant="outline">{project.type}</Badge>
        </div>
        <span className="project-path">{project.path}</span>
        <a
          className="project-url"
          href={project.url}
          target="_blank"
          rel="noreferrer"
        >
          {project.url} <ExternalLink aria-hidden="true" />
        </a>
      </div>
      <div className="project-meta">
        <span>PHP {project.php}</span>
        <small>{project.note}</small>
      </div>
    </div>
  );
}

function StatusDot({ state }: { state: string }) {
  const label =
    state === "healthy"
      ? "Running"
      : state === "attention"
        ? "Needs attention"
        : "Stopped";
  return (
    <span className={`status-dot ${state}`}>
      <span aria-hidden="true" />
      <span className="status-label">{label}</span>
    </span>
  );
}

function ActivityRow({
  icon: Icon,
  tone,
  title,
  detail,
  time,
}: {
  icon: LucideIcon;
  tone: string;
  title: string;
  detail: string;
  time: string;
}) {
  return (
    <div className="activity-row">
      <div className={`activity-icon ${tone}`}>
        <Icon />
      </div>
      <div>
        <strong>{title}</strong>
        <span>{detail}</span>
      </div>
      <time>{time}</time>
    </div>
  );
}

function QuickAction({
  to,
  icon: Icon,
  title,
  detail,
}: {
  to: "/docs" | "/containers" | "/containers/backup";
  icon: LucideIcon;
  title: string;
  detail: string;
}) {
  return (
    <Link to={to} className="quick-action">
      <span className="quick-icon">
        <Icon aria-hidden="true" />
      </span>
      <span>
        <strong>{title}</strong>
        <small>{detail}</small>
      </span>
      <ChevronRight aria-hidden="true" />
    </Link>
  );
}

function FeatureGroup({ group }: { group: (typeof featureGroups)[number] }) {
  const Icon = group.icon;
  return (
    <section
      className={`feature-group feature-${group.tone}`}
      aria-labelledby={`feature-${group.label}`}
    >
      <div className="feature-group-heading">
        <span className="feature-group-icon">
          <Icon aria-hidden="true" />
        </span>
        <h4 id={`feature-${group.label}`}>{group.label}</h4>
      </div>
      <div className="feature-list">
        {group.features.map((feature) => (
          <div className="feature-item" key={feature.name}>
            <div className="feature-item-copy">
              <strong>{feature.name}</strong>
              <span>{feature.detail}</span>
            </div>
            <Badge
              variant={feature.state === "available" ? "default" : "outline"}
              className={`feature-status ${feature.state}`}
            >
              {feature.state === "available"
                ? "Available"
                : feature.state === "next"
                  ? "Next"
                  : "Planned"}
            </Badge>
          </div>
        ))}
      </div>
    </section>
  );
}

function OperationRow({
  operation,
}: {
  operation: (typeof operations)[number];
}) {
  const Icon = operation.icon;
  return (
    <div className="operation-row">
      <div className={`operation-icon ${operation.tone}`}>
        <Icon aria-hidden="true" />
      </div>
      <div className="operation-copy">
        <div className="operation-title">
          <strong>{operation.title}</strong>
          <Badge
            variant="outline"
            className={`operation-status ${operation.state.toLowerCase()}`}
          >
            {operation.state}
          </Badge>
        </div>
        <span>{operation.detail}</span>
      </div>
      <Button
        variant="outline"
        size="sm"
        className="operation-button"
        render={<Link to={operation.to} />}
      >
        {operation.action}
        <ArrowUpRight aria-hidden="true" />
      </Button>
    </div>
  );
}
