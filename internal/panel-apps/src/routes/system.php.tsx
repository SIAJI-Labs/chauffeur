import * as React from "react";
import { Link, createFileRoute } from "@tanstack/react-router";
import {
  ArrowLeft,
  Check,
  CircleAlert,
  Code2,
  Copy,
  Download,
  ExternalLink,
  FileCode2,
  Network,
  Play,
  RefreshCw,
  Server,
  Settings2,
} from "lucide-react";
import { AppNavbar } from "@/components/app-navbar";
import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { useHashTab } from "@/hooks/use-hash-tab";
import {
  phpDetailTabs,
  phpExtensionFixtures,
  phpIniFixtures,
  phpLogFixtures,
  phpPortFixtures,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/system/php")({
  component: PhpRuntimePage,
});

type PhpTab = "logs" | "configuration" | "ports" | "extensions";

const phpTabValues = phpDetailTabs.map((tab) => tab.value as PhpTab);

function PhpRuntimePage() {
  const [activeTab, setActiveTab] = useHashTab(phpTabValues, "logs");

  return (
    <SidebarProvider className="dashboard-frame">
      <a className="skip-link" href="#php-content">
        Skip to PHP runtime detail
      </a>
      <AppSidebar />
      <SidebarInset className="dashboard-shell">
        <AppNavbar
          title="PHP 8.3"
          breadcrumbs={[
            { label: "Workspace", to: "/" },
            { label: "System", to: "/system" },
            { label: "PHP 8.3" },
          ]}
          shortcuts={
            <Button
              variant="outline"
              size="sm"
              className="header-button"
              render={<Link to="/system" />}
            >
              <ArrowLeft aria-hidden="true" />
              <span>System</span>
            </Button>
          }
        />
        <main className="dashboard-content min-h-0 flex-1" id="php-content">
          <ScrollArea className="dashboard-scroll-area">
            <div className="php-detail-page">
              <section
                className="php-detail-header"
                aria-labelledby="php-title"
              >
                <div className="php-detail-heading">
                  <div>
                    <div className="php-detail-kicker">
                      <span className="php-detail-status">
                        <span aria-hidden="true" /> Active
                      </span>
                      <Badge variant="outline">Runtime</Badge>
                      <Badge variant="outline">Default</Badge>
                    </div>
                    <h2 id="php-title">PHP 8.3</h2>
                    <p>Shared FPM pool · 2 projects · installed locally</p>
                  </div>
                  <Button variant="outline" size="sm" disabled>
                    <ExternalLink aria-hidden="true" /> Open phpinfo preview
                  </Button>
                </div>
                <div className="php-detail-meta">
                  <span>
                    <Server aria-hidden="true" /> FPM pool www
                  </span>
                  <span>
                    <Network aria-hidden="true" /> FastCGI :9000
                  </span>
                  <span>
                    <Check aria-hidden="true" /> 2 active projects
                  </span>
                  <span>
                    <Code2 aria-hidden="true" /> 8.3.12
                  </span>
                </div>
              </section>

              <section
                className="php-action-bar"
                aria-label="PHP runtime actions"
              >
                <Button variant="outline" disabled>
                  <Play aria-hidden="true" /> Start
                </Button>
                <Button variant="ghost" disabled>
                  Stop
                </Button>
                <Button variant="ghost" disabled>
                  <RefreshCw aria-hidden="true" /> Update
                </Button>
                <Button variant="ghost" disabled>
                  Rebuild
                </Button>
                <Button variant="ghost" disabled>
                  Remove
                </Button>
                <Button variant="ghost" disabled>
                  Xdebug
                </Button>
                <Badge variant="outline">Static controls</Badge>
              </section>

              <div
                className="php-detail-tabs"
                role="tablist"
                aria-label="PHP runtime sections"
                aria-orientation="horizontal"
              >
                {phpDetailTabs.map((tab) => (
                  <button
                    className="php-detail-tab"
                    data-active={activeTab === tab.value}
                    id={`php-tab-${tab.value}`}
                    key={tab.value}
                    onClick={() => setActiveTab(tab.value as PhpTab)}
                    onKeyDown={(event) => {
                      if (
                        !["ArrowLeft", "ArrowRight", "Home", "End"].includes(
                          event.key,
                        )
                      ) {
                        return;
                      }
                      event.preventDefault();
                      const currentIndex = phpDetailTabs.findIndex(
                        (item) => item.value === tab.value,
                      );
                      const nextIndex =
                        event.key === "Home"
                          ? 0
                          : event.key === "End"
                            ? phpDetailTabs.length - 1
                            : (currentIndex +
                                (event.key === "ArrowRight" ? 1 : -1) +
                                phpDetailTabs.length) %
                              phpDetailTabs.length;
                      const nextTab = phpDetailTabs[nextIndex].value as PhpTab;
                      setActiveTab(nextTab);
                      requestAnimationFrame(() =>
                        document.getElementById(`php-tab-${nextTab}`)?.focus(),
                      );
                    }}
                    role="tab"
                    tabIndex={activeTab === tab.value ? 0 : -1}
                    aria-selected={activeTab === tab.value}
                    aria-controls={`php-panel-${tab.value}`}
                    type="button"
                  >
                    {tab.label}
                    {tab.value === "extensions" ? <span>4</span> : null}
                  </button>
                ))}
              </div>
              {phpDetailTabs
                .filter((tab) => tab.value !== activeTab)
                .map((tab) => (
                  <div
                    aria-labelledby={`php-tab-${tab.value}`}
                    hidden
                    id={`php-panel-${tab.value}`}
                    key={tab.value}
                    role="tabpanel"
                  />
                ))}
              <div
                className="php-detail-panel"
                id={`php-panel-${activeTab}`}
                role="tabpanel"
                aria-labelledby={`php-tab-${activeTab}`}
                tabIndex={0}
              >
                {activeTab === "logs" ? <PhpLogs /> : null}
                {activeTab === "configuration" ? <PhpConfiguration /> : null}
                {activeTab === "ports" ? <PhpPorts /> : null}
                {activeTab === "extensions" ? <PhpExtensions /> : null}
              </div>
            </div>
          </ScrollArea>
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}

function PhpLogs() {
  return (
    <section className="php-tab-card" aria-labelledby="php-logs-title">
      <PhpPanelHeading
        kicker="Runtime output"
        title="PHP-FPM logs"
        icon={Server}
        actions={
          <>
            <Button variant="outline" size="sm" disabled>
              Follow logs
            </Button>
            <Button variant="outline" size="sm" disabled>
              <Copy aria-hidden="true" /> Copy
            </Button>
          </>
        }
      />
      <div className="php-log-viewer" id="php-logs-title">
        {phpLogFixtures.map((line) => (
          <div key={line.time}>
            <time>{line.time}</time>
            <span className={line.level === "WARN" ? "warn" : "info"}>
              {line.level}
            </span>
            <code>{line.message}</code>
          </div>
        ))}
      </div>
      <p className="php-static-note">
        Static sample output. Follow-to-bottom and live streaming are planned.
      </p>
    </section>
  );
}

function PhpConfiguration() {
  return (
    <section className="php-tab-card" aria-labelledby="php-configuration-title">
      <PhpPanelHeading
        kicker="INI"
        title="Configuration values"
        icon={Settings2}
        actions={
          <>
            <Button variant="outline" size="sm" disabled>
              Edit configuration
            </Button>
            <Button variant="ghost" size="sm" disabled>
              Restore defaults
            </Button>
          </>
        }
      />
      <div className="php-ini-list" id="php-configuration-title">
        {phpIniFixtures.map((item) => (
          <div key={item.key}>
            <code>{item.key}</code>
            <strong>{item.value}</strong>
            <Badge variant="outline">{item.source}</Badge>
          </div>
        ))}
      </div>
      <p className="php-static-note">
        Values are fixture data. Editing and restore confirmation remain
        disabled.
      </p>
    </section>
  );
}

function PhpPorts() {
  return (
    <section className="php-tab-card" aria-labelledby="php-ports-title">
      <PhpPanelHeading
        kicker="Network"
        title="PHP runtime ports"
        icon={Network}
        actions={
          <Button variant="outline" size="sm" disabled>
            Expose port
          </Button>
        }
      />
      <div className="php-port-list" id="php-ports-title">
        {phpPortFixtures.map((port) => (
          <div key={port.port}>
            <Network aria-hidden="true" />
            <span>
              <strong>
                {port.port} · {port.protocol}
              </strong>
              <small>{port.detail}</small>
            </span>
            <Badge
              variant="outline"
              className={
                port.state === "Ready" ? "php-good-badge" : "php-planned-badge"
              }
            >
              {port.state}
            </Badge>
          </div>
        ))}
      </div>
      <p className="php-static-note">
        Port conflict resolution and LAN exposure are planned system operations.
      </p>
    </section>
  );
}

function PhpExtensions() {
  return (
    <section className="php-tab-card" aria-labelledby="php-extensions-title">
      <PhpPanelHeading
        kicker="Modules"
        title="PHP extensions"
        icon={FileCode2}
        actions={
          <>
            <Button variant="outline" size="sm" disabled>
              <Download aria-hidden="true" /> Install extension
            </Button>
            <Button variant="ghost" size="sm" disabled>
              Check updates
            </Button>
          </>
        }
      />
      <div className="php-extension-list" id="php-extensions-title">
        {phpExtensionFixtures.map((extension) => (
          <div key={extension.name}>
            <span className="php-extension-icon">
              {extension.state === "Installed" ? (
                <Check aria-hidden="true" />
              ) : extension.state === "Available" ? (
                <CircleAlert aria-hidden="true" />
              ) : (
                <FileCode2 aria-hidden="true" />
              )}
            </span>
            <span>
              <strong>{extension.name}</strong>
              <small>
                {extension.detail} · {extension.version}
              </small>
            </span>
            <Badge
              variant="outline"
              className={
                extension.state === "Installed"
                  ? "php-good-badge"
                  : extension.state === "Available"
                    ? "php-warning-badge"
                    : "php-planned-badge"
              }
            >
              {extension.state}
            </Badge>
            <Button
              variant="ghost"
              size="icon-sm"
              disabled
              aria-label={`Manage ${extension.name}, planned`}
            >
              <Settings2 aria-hidden="true" />
            </Button>
          </div>
        ))}
      </div>
    </section>
  );
}

function PhpPanelHeading({
  kicker,
  title,
  icon: Icon,
  actions,
}: {
  kicker: string;
  title: string;
  icon: typeof Server;
  actions: React.ReactNode;
}) {
  return (
    <div className="php-panel-heading">
      <div>
        <p className="section-kicker">{kicker}</p>
        <h3>{title}</h3>
      </div>
      <div className="php-panel-actions">
        {actions}
        <Icon className="heading-icon" aria-hidden="true" />
      </div>
    </div>
  );
}
