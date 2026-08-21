import * as React from "react";
import { Link, createFileRoute } from "@tanstack/react-router";
import {
  ArrowLeft,
  Check,
  Database,
  ExternalLink,
  FileCode2,
  HardDrive,
  MoreHorizontal,
  Network,
  Play,
  Server,
  Settings2,
  TerminalSquare,
  Wrench,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { AppNavbar } from "@/components/app-navbar";
import { AppSidebar } from "@/components/app-sidebar";
import { ResourceNotFound } from "@/components/resource-not-found";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { useHashTab } from "@/hooks/use-hash-tab";
import {
  serviceDatabaseFixtures,
  serviceDetailActions,
  serviceDetailTabs,
  serviceEntityFixtures,
  serviceImportStateFixtures,
  serviceOverviewFixtures,
  serviceSnapshotFixtures,
} from "@/data/webui-fixtures";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

export const Route = createFileRoute("/services/$name")({
  component: ServiceDetailPage,
});

type ServiceTab =
  | "admin"
  | "databases"
  | "entities"
  | "logs"
  | "environment"
  | "configuration"
  | "tools"
  | "ports";

const serviceTabValues = serviceDetailTabs.map(
  (tab) => tab.value as ServiceTab,
);

type ServicePreview =
  | "service-action"
  | "create-database"
  | "drop-database"
  | "export-database"
  | "import-database"
  | "snapshots"
  | "restore-snapshot"
  | "testing-pairing"
  | "create-entity"
  | "delete-entity";

function ServiceDetailPage() {
  const { name } = Route.useParams();
  const service = serviceOverviewFixtures.find((item) => item.name === name);
  const [activeTab, setActiveTab] = useHashTab(serviceTabValues, "admin");
  const [preview, setPreview] = React.useState<ServicePreview | null>(null);

  if (!service) return <ResourceNotFound kind="service" name={name} />;
  const isHealthy = service.state === "healthy";

  return (
    <SidebarProvider className="dashboard-frame">
      <a className="skip-link" href="#service-detail-content">
        Skip to service detail
      </a>
      <AppSidebar />
      <SidebarInset className="dashboard-shell">
        <AppNavbar
          title={service.name}
          breadcrumbs={[
            { label: "Workspace", to: "/" },
            { label: "Services", to: "/services" },
            { label: service.name },
          ]}
          shortcuts={
            <Button
              variant="outline"
              size="sm"
              className="header-button"
              render={<Link to="/services" />}
            >
              <ArrowLeft aria-hidden="true" />
              <span>Services</span>
            </Button>
          }
        />
        <main
          className="dashboard-content min-h-0 flex-1"
          id="service-detail-content"
        >
          <ScrollArea className="dashboard-scroll-area">
            <div className="service-detail-page">
              <section
                className="service-detail-header"
                aria-labelledby="service-detail-title"
              >
                <div className="service-detail-heading">
                  <div>
                    <div className="service-detail-kicker">
                      <span
                        className={`service-detail-status ${service.state}`}
                      >
                        <span aria-hidden="true" />{" "}
                        {isHealthy ? "Healthy" : "Needs attention"}
                      </span>
                      <Badge variant="outline">{service.category}</Badge>
                      <Badge variant="outline">v{service.version}</Badge>
                    </div>
                    <h2 id="service-detail-title">{service.name}</h2>
                    <p>
                      {service.projects} using this service · local workspace
                      preset
                    </p>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    disabled
                    aria-label="More service actions, planned"
                  >
                    <MoreHorizontal aria-hidden="true" />
                  </Button>
                </div>
                <div className="service-detail-meta">
                  <span>
                    <Server aria-hidden="true" /> Version {service.version}
                  </span>
                  <span>
                    <Network aria-hidden="true" /> Ports {service.ports}
                  </span>
                  <span>
                    <Check aria-hidden="true" /> {service.dependencies}
                  </span>
                  <span>
                    <Database aria-hidden="true" /> {service.projects}
                  </span>
                </div>
              </section>

              <section
                className="service-action-bar"
                aria-label="Service actions"
              >
                {serviceDetailActions.map((action) => (
                  <Button
                    key={action}
                    variant="ghost"
                    className="service-action"
                    onClick={() => setPreview("service-action")}
                  >
                    {action}
                    <Badge variant="outline">Planned</Badge>
                  </Button>
                ))}
              </section>

              <div
                className="service-detail-tabs"
                role="tablist"
                aria-label="Service detail sections"
                aria-orientation="horizontal"
              >
                {serviceDetailTabs.map((tab) => (
                  <button
                    className="service-detail-tab"
                    data-active={activeTab === tab.value}
                    id={`service-tab-${tab.value}`}
                    key={tab.value}
                    onClick={() => setActiveTab(tab.value as ServiceTab)}
                    onKeyDown={(event) => {
                      if (
                        !["ArrowLeft", "ArrowRight", "Home", "End"].includes(
                          event.key,
                        )
                      ) {
                        return;
                      }
                      event.preventDefault();
                      const currentIndex = serviceDetailTabs.findIndex(
                        (item) => item.value === tab.value,
                      );
                      const nextIndex =
                        event.key === "Home"
                          ? 0
                          : event.key === "End"
                            ? serviceDetailTabs.length - 1
                            : (currentIndex +
                                (event.key === "ArrowRight" ? 1 : -1) +
                                serviceDetailTabs.length) %
                              serviceDetailTabs.length;
                      const nextTab = serviceDetailTabs[nextIndex]
                        .value as ServiceTab;
                      setActiveTab(nextTab);
                      requestAnimationFrame(() =>
                        document
                          .getElementById(`service-tab-${nextTab}`)
                          ?.focus(),
                      );
                    }}
                    role="tab"
                    tabIndex={activeTab === tab.value ? 0 : -1}
                    aria-selected={activeTab === tab.value}
                    aria-controls={`service-panel-${tab.value}`}
                    type="button"
                  >
                    {tab.label}
                  </button>
                ))}
              </div>
              {serviceDetailTabs
                .filter((tab) => tab.value !== activeTab)
                .map((tab) => (
                  <div
                    aria-labelledby={`service-tab-${tab.value}`}
                    hidden
                    id={`service-panel-${tab.value}`}
                    key={tab.value}
                    role="tabpanel"
                  />
                ))}
              <div
                className="service-detail-tab-panel"
                id={`service-panel-${activeTab}`}
                role="tabpanel"
                aria-labelledby={`service-tab-${activeTab}`}
                tabIndex={0}
              >
                {activeTab === "admin" ? (
                  <ServiceAdmin service={service} />
                ) : null}
                {activeTab === "databases" ? (
                  <ServiceDatabases onPreview={setPreview} />
                ) : null}
                {activeTab === "entities" ? (
                  <ServiceEntities onPreview={setPreview} />
                ) : null}
                {activeTab === "logs" ? <ServiceLogs /> : null}
                {activeTab === "environment" ? <ServiceEnvironment /> : null}
                {activeTab === "configuration" ? (
                  <ServiceConfiguration />
                ) : null}
                {activeTab === "tools" ? <ServiceTools /> : null}
                {activeTab === "ports" ? <ServicePorts /> : null}
              </div>
            </div>
          </ScrollArea>
        </main>
      </SidebarInset>
      <ServicePreviewDialog
        preview={preview}
        onClose={() => setPreview(null)}
      />
    </SidebarProvider>
  );
}

function ServiceAdmin({
  service,
}: {
  service: (typeof serviceOverviewFixtures)[number];
}) {
  return (
    <div className="service-admin-grid">
      <section className="service-detail-card service-admin-hero">
        <DetailHeading
          kicker="Admin dashboard"
          title="A safe launch point for service tools"
          icon={ExternalLink}
        />
        <p>
          Open the service admin panel or copy a connection URL when the local
          integration is ready.
        </p>
        <div className="service-admin-links">
          <Button variant="outline" size="sm" disabled>
            <ExternalLink aria-hidden="true" /> Open admin dashboard
          </Button>
          <code>
            {service.name.toLowerCase()}://localhost:{service.ports}
          </code>
        </div>
        <small className="service-static-note">
          Admin links are previews until service orchestration is available.
        </small>
      </section>
      <section className="service-detail-card">
        <DetailHeading
          kicker="Project usage"
          title="Connected projects"
          icon={Database}
        />
        <div className="service-usage-list">
          <div>
            <span>
              <strong>commerce-api</strong>
              <small>Primary database · read/write</small>
            </span>
            <Badge variant="outline" className="service-good-badge">
              Connected
            </Badge>
          </div>
          <div>
            <span>
              <strong>client-portal</strong>
              <small>Preview environment · read/write</small>
            </span>
            <Badge variant="outline" className="service-good-badge">
              Connected
            </Badge>
          </div>
        </div>
      </section>
      <section className="service-detail-card service-dependency-card">
        <DetailHeading
          kicker="Dependencies"
          title="Runtime connection map"
          icon={Network}
        />
        <div
          className="service-dependency-map"
          aria-label="Static dependency visualization"
        >
          <span className="service-dependency-node main">{service.name}</span>
          <i className="service-dependency-line one" aria-hidden="true" />
          <i className="service-dependency-line two" aria-hidden="true" />
          <span className="service-dependency-node">commerce-api</span>
          <span className="service-dependency-node">client-portal</span>
        </div>
        <Badge variant="outline">Preview only</Badge>
      </section>
      <section className="service-detail-card">
        <DetailHeading
          kicker="Health"
          title="Checks and capacity"
          icon={Check}
        />
        <div className="service-health-grid">
          <span>
            <strong>98%</strong>
            <small>Availability</small>
          </span>
          <span>
            <strong>84 MB</strong>
            <small>Storage used</small>
          </span>
          <span>
            <strong>12 ms</strong>
            <small>Connection</small>
          </span>
        </div>
      </section>
    </div>
  );
}

function ServiceDatabases({
  onPreview,
}: {
  onPreview: (preview: ServicePreview) => void;
}) {
  return (
    <section
      className="service-tab-card"
      aria-labelledby="service-databases-title"
    >
      <TabHeading
        kicker="Data"
        title="Databases"
        icon={Database}
        action="Create database"
      />
      <div className="service-database-toolbar">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("create-database")}
        >
          Create preview
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("import-database")}
        >
          Import preview
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("export-database")}
        >
          Export preview
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onPreview("snapshots")}
        >
          Snapshots
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onPreview("testing-pairing")}
        >
          Pair testing database
        </Button>
      </div>
      <div className="service-database-list" id="service-databases-title">
        {serviceDatabaseFixtures.map((database) => (
          <div className="service-database-row" key={database.name}>
            <Database aria-hidden="true" />
            <span>
              <strong>{database.name}</strong>
              <small>
                {database.size} · owned by {database.owner}
              </small>
            </span>
            <Badge
              variant="outline"
              className={
                database.status === "Testing"
                  ? "service-testing-badge"
                  : "service-good-badge"
              }
            >
              {database.status}
            </Badge>
            <Button variant="outline" size="sm" disabled>
              Copy DSN
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPreview("drop-database")}
              aria-label={`Preview dropping ${database.name}`}
            >
              Drop
            </Button>
          </div>
        ))}
      </div>
      <div
        className="service-import-state-grid"
        aria-label="Import state examples"
      >
        {serviceImportStateFixtures.map((state) => (
          <div className={state.tone} key={state.label}>
            <span>
              <strong>{state.label}</strong>
              <small>{state.detail}</small>
            </span>
            <Badge variant="outline">Preview</Badge>
          </div>
        ))}
      </div>
      <div className="service-snapshot-list">
        <div className="service-snapshot-heading">
          <span>
            <strong>Snapshot list</strong>
            <small>Restore points for this service</small>
          </span>
          <Badge variant="outline">2 snapshots</Badge>
        </div>
        {serviceSnapshotFixtures.map((snapshot) => (
          <div className="service-snapshot-row" key={snapshot.name}>
            <HardDrive aria-hidden="true" />
            <span>
              <strong>{snapshot.name}</strong>
              <small>
                {snapshot.size} · {snapshot.age}
              </small>
            </span>
            <Badge variant="outline" className="service-good-badge">
              {snapshot.state}
            </Badge>
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPreview("restore-snapshot")}
            >
              Restore preview
            </Button>
          </div>
        ))}
      </div>
      <p className="service-static-note">
        Import, export, restore, and destructive actions stay in preview mode
        and do not change data.
      </p>
    </section>
  );
}

function ServiceEntities({
  onPreview,
}: {
  onPreview: (preview: ServicePreview) => void;
}) {
  return (
    <section
      className="service-tab-card"
      aria-labelledby="service-entities-title"
    >
      <TabHeading
        kicker="Schema"
        title="Entities"
        icon={FileCode2}
        action="Create entity"
      />
      <div className="service-entity-toolbar">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPreview("create-entity")}
        >
          Create entity preview
        </Button>
        <Badge variant="outline">Grouped by kind</Badge>
      </div>
      <div className="service-entity-grid" id="service-entities-title">
        {serviceEntityFixtures.map((entity) => (
          <div key={entity.kind}>
            <span>
              <strong>{entity.kind}</strong>
              <small>{entity.detail}</small>
            </span>
            <b>{entity.count}</b>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onPreview("delete-entity")}
              aria-label={`Preview deleting ${entity.kind}`}
            >
              <MoreHorizontal aria-hidden="true" /> Delete preview
            </Button>
          </div>
        ))}
      </div>
      <p className="service-static-note">
        Entity creation and deletion previews remain disabled until the service
        API exists.
      </p>
    </section>
  );
}

function ServicePreviewDialog({
  preview,
  onClose,
}: {
  preview: ServicePreview | null;
  onClose: () => void;
}) {
  if (!preview) return null;
  const copy: Record<
    ServicePreview,
    { title: string; description: string; action: string }
  > = {
    "service-action": {
      title: "Service action preview",
      description:
        "Review this service operation before connecting it to a local runtime API.",
      action: "Continue",
    },
    "create-database": {
      title: "Create database preview",
      description:
        "Review the database name, owner project, and testing pairing before creation.",
      action: "Create database",
    },
    "drop-database": {
      title: "Drop database confirmation",
      description:
        "This destructive operation would require an explicit confirmation and a database name check.",
      action: "Confirm drop",
    },
    "export-database": {
      title: "Export database preview",
      description:
        "Choose a format and snapshot metadata before exporting a local database.",
      action: "Export database",
    },
    "import-database": {
      title: "Import database preview",
      description:
        "Review progress, warnings, errors, and manual issues before importing a snapshot.",
      action: "Start import",
    },
    snapshots: {
      title: "Snapshot management preview",
      description:
        "Review snapshot metadata and choose a restore point without changing local data.",
      action: "Manage snapshots",
    },
    "restore-snapshot": {
      title: "Restore snapshot preview",
      description:
        "Confirm the target database, backup age, and overwrite warning before restoring.",
      action: "Restore snapshot",
    },
    "testing-pairing": {
      title: "Testing database pairing preview",
      description:
        "Pair a testing database with a project and preview its isolated connection details.",
      action: "Pair testing database",
    },
    "create-entity": {
      title: "Create entity preview",
      description:
        "Choose an entity kind and preview the schema change before applying it.",
      action: "Create entity",
    },
    "delete-entity": {
      title: "Delete entity confirmation",
      description:
        "Review the entity name and dependency warning before a destructive schema change.",
      action: "Confirm delete",
    },
  };
  const content = copy[preview];
  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="service-preview-dialog">
        <DialogHeader>
          <DialogTitle>{content.title}</DialogTitle>
          <DialogDescription>{content.description}</DialogDescription>
        </DialogHeader>
        <div className="service-preview-dialog-body">
          <Badge variant="outline" className="service-planned-badge">
            Static preview only
          </Badge>
          <div>
            <span>Operation plan</span>
            <strong>{content.action}</strong>
            <small>No API request or data mutation is performed.</small>
          </div>
          {preview === "import-database" ? (
            <div className="service-preview-progress">
              <span>
                <i />
                Import metadata
              </span>
              <span>
                <i />
                Check dependencies
              </span>
              <span>
                <i />
                Review issue states
              </span>
            </div>
          ) : null}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Close preview
          </Button>
          <Button disabled>{content.action}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function ServiceLogs() {
  return (
    <section className="service-tab-card" aria-labelledby="service-logs-title">
      <TabHeading
        kicker="Runtime output"
        title="Service logs"
        icon={HardDrive}
        action="Follow logs"
      />
      <div className="service-log-viewer" id="service-logs-title">
        <div>
          <time>09:43:12</time>
          <span className="info">INFO</span>
          <code>connection pool ready · 12 clients</code>
        </div>
        <div>
          <time>09:42:57</time>
          <span className="info">INFO</span>
          <code>checkpoint complete · 84 MB</code>
        </div>
        <div>
          <time>09:41:04</time>
          <span className="warn">WARN</span>
          <code>client-portal retrying connection</code>
        </div>
      </div>
      <p className="service-static-note">
        Static sample output. Follow, copy, and clear actions are planned.
      </p>
    </section>
  );
}

function ServiceEnvironment() {
  return (
    <section
      className="service-tab-card"
      aria-labelledby="service-environment-title"
    >
      <TabHeading
        kicker="Runtime context"
        title="Environment"
        icon={Settings2}
        action="Edit environment"
      />
      <div className="service-environment-list" id="service-environment-title">
        <div>
          <code>POSTGRES_HOST</code>
          <span>localhost</span>
          <Badge variant="outline">Inherited</Badge>
        </div>
        <div>
          <code>POSTGRES_PORT</code>
          <span>5432</span>
          <Badge variant="outline">Preset</Badge>
        </div>
        <div>
          <code>POSTGRES_PASSWORD</code>
          <span>••••••••••••</span>
          <Badge variant="outline">Secret</Badge>
        </div>
      </div>
      <p className="service-static-note">
        Secrets stay masked. Editing and restore confirmation are planned.
      </p>
    </section>
  );
}

function ServiceConfiguration() {
  return (
    <section
      className="service-tab-card"
      aria-labelledby="service-configuration-title"
    >
      <TabHeading
        kicker="Tuning"
        title="Configuration"
        icon={Settings2}
        action="Save changes"
      />
      <div className="service-config-grid" id="service-configuration-title">
        <div>
          <span>Max connections</span>
          <strong>100</strong>
          <small>Project-safe default</small>
        </div>
        <div>
          <span>Shared buffers</span>
          <strong>256 MB</strong>
          <small>Preset recommendation</small>
        </div>
        <div>
          <span>Log level</span>
          <strong>Notice</strong>
          <small>Warnings remain visible</small>
        </div>
        <div>
          <span>Persistence</span>
          <strong>Enabled</strong>
          <small>Local volume attached</small>
        </div>
      </div>
      <p className="service-static-note">
        Configuration tuning is a preview and does not write service files.
      </p>
    </section>
  );
}

function ServiceTools() {
  return (
    <section className="service-tab-card" aria-labelledby="service-tools-title">
      <TabHeading
        kicker="Client shims"
        title="Tools and clients"
        icon={TerminalSquare}
        action="Check availability"
      />
      <div className="service-tools-grid" id="service-tools-title">
        <div>
          <TerminalSquare aria-hidden="true" />
          <span>
            <strong>psql</strong>
            <small>CLI client · available</small>
          </span>
          <Badge variant="outline" className="service-good-badge">
            Available
          </Badge>
        </div>
        <div>
          <Play aria-hidden="true" />
          <span>
            <strong>Adminer</strong>
            <small>Browser dashboard · planned</small>
          </span>
          <Badge variant="outline">Planned</Badge>
        </div>
        <div>
          <Wrench aria-hidden="true" />
          <span>
            <strong>Schema browser</strong>
            <small>Panel integration · planned</small>
          </span>
          <Badge variant="outline">Planned</Badge>
        </div>
      </div>
    </section>
  );
}

function ServicePorts() {
  return (
    <section className="service-tab-card" aria-labelledby="service-ports-title">
      <TabHeading
        kicker="Network"
        title="Ports"
        icon={Network}
        action="Resolve conflicts"
      />
      <div className="service-ports-list" id="service-ports-title">
        <div>
          <Network aria-hidden="true" />
          <span>
            <strong>5432</strong>
            <small>PostgreSQL TCP · localhost</small>
          </span>
          <Badge variant="outline" className="service-good-badge">
            Ready
          </Badge>
        </div>
        <div>
          <Network aria-hidden="true" />
          <span>
            <strong>5433</strong>
            <small>Testing database · localhost</small>
          </span>
          <Badge variant="outline" className="service-warning-badge">
            Attention
          </Badge>
        </div>
      </div>
      <p className="service-static-note">
        Port changes, conflict resolution, and LAN exposure require backend
        support.
      </p>
    </section>
  );
}

function DetailHeading({
  kicker,
  title,
  icon: Icon,
}: {
  kicker: string;
  title: string;
  icon: LucideIcon;
}) {
  return (
    <div className="service-detail-heading-block">
      <div>
        <p className="section-kicker">{kicker}</p>
        <h3>{title}</h3>
      </div>
      <Icon aria-hidden="true" />
    </div>
  );
}

function TabHeading({
  kicker,
  title,
  icon,
  action,
}: {
  kicker: string;
  title: string;
  icon: LucideIcon;
  action: string;
}) {
  return (
    <div className="service-tab-heading">
      <DetailHeading kicker={kicker} title={title} icon={icon} />
      <Button variant="outline" size="sm" disabled>
        {action}
      </Button>
    </div>
  );
}
