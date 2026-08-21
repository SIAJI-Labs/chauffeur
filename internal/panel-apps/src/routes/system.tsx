import {
  Link,
  Outlet,
  createFileRoute,
  useRouterState,
} from "@tanstack/react-router";
import {
  Activity,
  ArrowUpRight,
  Bug,
  Check,
  ChevronRight,
  CircleAlert,
  Code2,
  Cpu,
  RefreshCw,
  Server,
  Settings2,
  TerminalSquare,
  Wrench,
} from "lucide-react";
import { AppNavbar } from "@/components/app-navbar";
import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import {
  debugBridgeLensFixtures,
  phpRuntimeFixtures,
  systemComponentFixtures,
  systemSettingFixtures,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/system")({
  component: SystemPage,
});

function SystemPage() {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });
  if (pathname !== "/system") return <Outlet />;

  return (
    <SidebarProvider className="dashboard-frame">
      <a className="skip-link" href="#system-content">
        Skip to system overview
      </a>
      <AppSidebar />
      <SidebarInset className="dashboard-shell">
        <AppNavbar
          title="System"
          breadcrumbs={[{ label: "Workspace", to: "/" }, { label: "System" }]}
          shortcuts={
            <Button
              variant="outline"
              size="sm"
              className="header-button"
              render={<Link to="/docs" />}
            >
              <TerminalSquare aria-hidden="true" />
              <span>System docs</span>
            </Button>
          }
        />
        <main className="dashboard-content min-h-0 flex-1" id="system-content">
          <ScrollArea className="dashboard-scroll-area">
            <div className="system-page">
              <section className="system-intro">
                <div>
                  <p className="section-kicker">Runtime foundation</p>
                  <h2>Know what is running before you open a project.</h2>
                  <p>
                    Chauffeur, gateway, runtimes, tools, and automation in one
                    static health surface.
                  </p>
                </div>
                <Badge variant="outline" className="system-preview-badge">
                  Static system view
                </Badge>
              </section>

              <DebugBridgePreview />

              <section
                className="system-summary-grid"
                aria-label="System summary"
              >
                <SystemSummary
                  label="Components"
                  value="8"
                  detail="Core and toolchain"
                  icon={Settings2}
                />
                <SystemSummary
                  label="Healthy"
                  value="3"
                  detail="Ready to serve projects"
                  icon={Check}
                />
                <SystemSummary
                  label="Attention"
                  value="4"
                  detail="Review runtime signals"
                  icon={CircleAlert}
                />
                <SystemSummary
                  label="Updates"
                  value="2"
                  detail="Tools ready to update"
                  icon={RefreshCw}
                />
              </section>

              <section
                className="system-components-panel"
                aria-labelledby="system-components-title"
              >
                <div className="system-panel-heading">
                  <div>
                    <p className="section-kicker">System components</p>
                    <h3 id="system-components-title">Local runtime map</h3>
                  </div>
                  <Badge variant="outline">8 components</Badge>
                </div>
                <div className="system-component-grid">
                  {systemComponentFixtures.map((component) => (
                    <SystemComponentCard
                      key={component.name}
                      component={component}
                    />
                  ))}
                </div>
              </section>

              <div className="system-runtime-grid">
                <section
                  className="system-subpanel"
                  aria-labelledby="php-runtimes-title"
                >
                  <SystemPanelHeading
                    kicker="PHP"
                    title="Runtime versions"
                    icon={Server}
                  />
                  <div className="php-runtime-list" id="php-runtimes-title">
                    {phpRuntimeFixtures.map((runtime) => (
                      <div key={runtime.version}>
                        <span className={`php-runtime-dot ${runtime.state}`} />
                        <span>
                          <strong>PHP {runtime.version}</strong>
                          <small>{runtime.detail}</small>
                        </span>
                        <Badge
                          variant="outline"
                          className={
                            runtime.state === "active"
                              ? "system-good-badge"
                              : "system-warning-badge"
                          }
                        >
                          {runtime.update}
                        </Badge>
                        <div className="php-runtime-actions">
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            render={<Link to="/system/php" />}
                            aria-label={`Open PHP ${runtime.version} details`}
                          >
                            <ArrowUpRight aria-hidden="true" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            disabled
                            aria-label={`Set PHP ${runtime.version} as default, planned`}
                          >
                            <Check aria-hidden="true" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            disabled
                            aria-label={`Manage PHP ${runtime.version}, planned`}
                          >
                            <Settings2 aria-hidden="true" />
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                  <div className="system-button-row">
                    <Button variant="outline" size="sm" disabled>
                      Install PHP version
                    </Button>
                    <Button variant="ghost" size="sm" disabled>
                      Start FPM
                    </Button>
                    <Button variant="ghost" size="sm" disabled>
                      Stop FPM
                    </Button>
                    <Button variant="ghost" size="sm" disabled>
                      Update
                    </Button>
                    <Button variant="ghost" size="sm" disabled>
                      Rebuild FPM
                    </Button>
                    <Button variant="ghost" size="sm" disabled>
                      Remove
                    </Button>
                    <Button variant="ghost" size="sm" disabled>
                      Xdebug
                    </Button>
                  </div>
                </section>
                <section
                  className="system-subpanel"
                  aria-labelledby="toolchain-title"
                >
                  <SystemPanelHeading
                    kicker="Toolchain"
                    title="Node and tools"
                    icon={Wrench}
                  />
                  <div className="system-tool-list" id="toolchain-title">
                    <div>
                      <TerminalSquare aria-hidden="true" />
                      <span>
                        <strong>Node.js 22.12</strong>
                        <small>System-managed · default</small>
                      </span>
                      <Badge variant="outline" className="system-good-badge">
                        Ready
                      </Badge>
                    </div>
                    <div>
                      <Code2 aria-hidden="true" />
                      <span>
                        <strong>Bun</strong>
                        <small>Not installed · install preview</small>
                      </span>
                      <Badge variant="outline">Planned</Badge>
                    </div>
                    <div>
                      <Wrench aria-hidden="true" />
                      <span>
                        <strong>mkcert</strong>
                        <small>Certificate tool · available</small>
                      </span>
                      <Badge variant="outline" className="system-good-badge">
                        Ready
                      </Badge>
                    </div>
                    <div>
                      <ArrowUpRight aria-hidden="true" />
                      <span>
                        <strong>ngrok</strong>
                        <small>Update available · v3.18</small>
                      </span>
                      <Badge variant="outline" className="system-warning-badge">
                        Update
                      </Badge>
                    </div>
                  </div>
                  <div className="system-button-row">
                    <Button variant="outline" size="sm" disabled>
                      Check for updates
                    </Button>
                    <Button variant="ghost" size="sm" disabled>
                      Manage versions
                    </Button>
                    <Button variant="ghost" size="sm" disabled>
                      Install Bun
                    </Button>
                    <Button variant="ghost" size="sm" disabled>
                      Set default
                    </Button>
                    <Button variant="ghost" size="sm" disabled>
                      Remove version
                    </Button>
                  </div>
                </section>
              </div>

              <section
                className="system-settings-panel"
                aria-labelledby="system-settings-title"
              >
                <div className="system-panel-heading">
                  <div>
                    <p className="section-kicker">Preferences and automation</p>
                    <h3 id="system-settings-title">
                      Make the local runtime fit your flow
                    </h3>
                  </div>
                  <Badge variant="outline">Planned controls</Badge>
                </div>
                <div className="system-setting-grid">
                  {systemSettingFixtures.map((setting) => {
                    const Icon = setting.icon;
                    return (
                      <div key={setting.label}>
                        <span className="system-setting-icon">
                          <Icon aria-hidden="true" />
                        </span>
                        <span>
                          <strong>{setting.label}</strong>
                          <small>{setting.detail}</small>
                        </span>
                        <Badge
                          variant="outline"
                          className={
                            setting.state === "Next"
                              ? "system-next-badge"
                              : "system-planned-badge"
                          }
                        >
                          {setting.state}
                        </Badge>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          disabled
                          aria-label={`${setting.label}, planned`}
                        >
                          <ChevronRight aria-hidden="true" />
                        </Button>
                      </div>
                    );
                  })}
                </div>
              </section>

              <section
                className="system-execution-panel"
                aria-labelledby="execution-title"
              >
                <div className="system-panel-heading">
                  <div>
                    <p className="section-kicker">Execution mode</p>
                    <h3 id="execution-title">Where background work runs</h3>
                  </div>
                  <Cpu className="heading-icon" />
                </div>
                <div className="system-execution-options">
                  <div className="selected">
                    <Activity aria-hidden="true" />
                    <span>
                      <strong>Host mode</strong>
                      <small>
                        Fast local workers with direct filesystem access.
                      </small>
                    </span>
                    <Badge variant="outline">Current</Badge>
                  </div>
                  <div>
                    <Server aria-hidden="true" />
                    <span>
                      <strong>Container mode</strong>
                      <small>
                        Isolated workers with reproducible dependencies.
                      </small>
                    </span>
                    <Badge variant="outline">Planned</Badge>
                  </div>
                </div>
              </section>
            </div>
          </ScrollArea>
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}

function SystemSummary({
  label,
  value,
  detail,
  icon: Icon,
}: {
  label: string;
  value: string;
  detail: string;
  icon: typeof Settings2;
}) {
  return (
    <article className="system-summary-card">
      <div>
        <span>{label}</span>
        <Icon aria-hidden="true" />
      </div>
      <strong>{value}</strong>
      <small>{detail}</small>
    </article>
  );
}

function DebugBridgePreview() {
  return (
    <section
      className="debug-bridge-panel"
      aria-labelledby="debug-bridge-title"
    >
      <div className="system-panel-heading">
        <div>
          <p className="section-kicker">Diagnostics bridge</p>
          <h3 id="debug-bridge-title">Inspect the signals behind a request</h3>
          <p className="debug-bridge-description">
            Preview the bridge that will connect project output to focused
            diagnostic lenses.
          </p>
        </div>
        <Bug className="heading-icon" aria-hidden="true" />
      </div>
      <div className="debug-bridge-controls">
        <div className="debug-bridge-status">
          <span className="debug-bridge-dot" />
          <span>
            <strong>Bridge disabled</strong>
            <small>Enablement requires the project runtime integration.</small>
          </span>
          <Badge variant="outline">Planned</Badge>
        </div>
        <div className="debug-bridge-actions">
          <Button variant="outline" size="sm" disabled>
            Enable bridge
          </Button>
          <Button variant="ghost" size="sm" disabled>
            Disable bridge
          </Button>
          <Button variant="ghost" size="sm" disabled>
            Choose mode
          </Button>
        </div>
      </div>
      <div className="debug-lens-grid">
        {debugBridgeLensFixtures.map((lens) => {
          const Icon = lens.icon;
          return (
            <div className="debug-lens-card" key={lens.name}>
              <span className="debug-lens-icon">
                <Icon aria-hidden="true" />
              </span>
              <span>
                <strong>{lens.name}</strong>
                <small>{lens.detail}</small>
              </span>
              <Badge
                variant="outline"
                className={
                  lens.state === "Available"
                    ? "system-good-badge"
                    : "system-planned-badge"
                }
              >
                {lens.state}
              </Badge>
            </div>
          );
        })}
      </div>
    </section>
  );
}

function SystemComponentCard({
  component,
}: {
  component: (typeof systemComponentFixtures)[number];
}) {
  const Icon = component.icon;
  const healthy = component.state === "healthy";
  return (
    <article className={`system-component-card ${component.state}`}>
      <div className="system-component-top">
        <span className={`system-component-icon ${component.state}`}>
          <Icon aria-hidden="true" />
        </span>
        <Badge
          variant="outline"
          className={
            healthy
              ? "system-good-badge"
              : component.state === "idle"
                ? "system-idle-badge"
                : "system-warning-badge"
          }
        >
          {healthy
            ? "Healthy"
            : component.state === "idle"
              ? "Idle"
              : "Attention"}
        </Badge>
      </div>
      <strong>{component.name}</strong>
      <span>
        {component.category} · {component.version}
      </span>
      <p>{component.detail}</p>
      <div className="system-component-actions">
        <Button variant="ghost" size="sm" disabled>
          View details
        </Button>
        {component.name === "Nginx" || component.name === "DNS" ? (
          <>
            <Button variant="ghost" size="sm" disabled>
              Logs
            </Button>
            <Button variant="ghost" size="sm" disabled>
              Config
            </Button>
          </>
        ) : null}
        {component.name === "Watchers" ? (
          <Button variant="ghost" size="sm" disabled>
            Start watcher
          </Button>
        ) : null}
      </div>
    </article>
  );
}

function SystemPanelHeading({
  kicker,
  title,
  icon: Icon,
}: {
  kicker: string;
  title: string;
  icon: typeof Server;
}) {
  return (
    <div className="system-panel-heading">
      <div>
        <p className="section-kicker">{kicker}</p>
        <h3>{title}</h3>
      </div>
      <Icon className="heading-icon" aria-hidden="true" />
    </div>
  );
}
