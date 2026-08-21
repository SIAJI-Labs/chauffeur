import * as React from "react";
import { Link, createFileRoute } from "@tanstack/react-router";
import {
  AlertTriangle,
  ArrowUpRight,
  Boxes,
  CircleCheck,
  Database,
  HardDrive,
  Plus,
  Search,
  Settings2,
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
  serviceCatalogFixtures,
  serviceCategories,
  serviceOverviewFixtures,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/services")({
  component: ServicesPage,
});

function ServicesPage() {
  const [search, setSearch] = React.useState("");
  const [category, setCategory] = React.useState("All services");
  const [isInstallDialogOpen, setIsInstallDialogOpen] = React.useState(false);
  const deferredSearch = React.useDeferredValue(search);
  const visibleServices = serviceOverviewFixtures.filter((service) => {
    const query = deferredSearch.trim().toLowerCase();
    const matchesCategory =
      category === "All services" || service.category === category;
    const matchesSearch =
      !query ||
      [service.name, service.category, service.version]
        .join(" ")
        .toLowerCase()
        .includes(query);
    return matchesCategory && matchesSearch;
  });

  return (
    <SidebarProvider className="dashboard-frame">
      <a className="skip-link" href="#services-content">
        Skip to services overview
      </a>
      <AppSidebar />
      <SidebarInset className="dashboard-shell">
        <AppNavbar
          title="Services"
          breadcrumbs={[{ label: "Workspace", to: "/" }, { label: "Services" }]}
          shortcuts={
            <Button
              variant="outline"
              size="sm"
              className="header-button"
              render={<Link to="/containers" />}
            >
              <HardDrive aria-hidden="true" />
              <span>Containers</span>
            </Button>
          }
        />
        <main
          className="dashboard-content min-h-0 flex-1"
          id="services-content"
        >
          <ScrollArea className="dashboard-scroll-area">
            <div className="services-page">
              <section className="services-intro">
                <div>
                  <p className="section-kicker">Local service layer</p>
                  <h2>Everything your projects depend on, in one inventory.</h2>
                  <p>
                    Inspect installed services, dependencies, ports, and the
                    next presets to make local development feel complete.
                  </p>
                </div>
                <Button onClick={() => setIsInstallDialogOpen(true)}>
                  <Plus aria-hidden="true" /> Add service
                </Button>
              </section>

              <section
                className="service-summary-grid"
                aria-label="Service summary"
              >
                <ServiceSummary
                  label="Installed"
                  value="5"
                  detail="Across 6 categories"
                  icon={Boxes}
                />
                <ServiceSummary
                  label="Healthy"
                  value="2"
                  detail="Ready for projects"
                  icon={CircleCheck}
                />
                <ServiceSummary
                  label="Attention"
                  value="2"
                  detail="1 port conflict"
                  icon={AlertTriangle}
                />
                <ServiceSummary
                  label="Updates"
                  value="2"
                  detail="Available this week"
                  icon={ArrowUpRight}
                />
              </section>

              <section
                className="service-inventory-panel"
                aria-labelledby="services-inventory-title"
              >
                <div className="service-panel-heading">
                  <div>
                    <p className="section-kicker">Installed services</p>
                    <h3 id="services-inventory-title">
                      Project-ready infrastructure
                    </h3>
                  </div>
                  <Badge variant="outline">Static inventory</Badge>
                </div>
                <div className="service-list-controls">
                  <label className="service-search">
                    <span className="sr-only">Search services</span>
                    <Search aria-hidden="true" />
                    <Input
                      value={search}
                      onChange={(event) => setSearch(event.target.value)}
                      placeholder="Search services or categories"
                    />
                  </label>
                  <Select
                    value={category}
                    onValueChange={(value) => value && setCategory(value)}
                  >
                    <SelectTrigger
                      aria-label="Filter service category"
                      className="service-category-select"
                    >
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {serviceCategories.map((option) => (
                        <SelectItem key={option} value={option}>
                          {option}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="sr-only" aria-live="polite">
                  Showing {visibleServices.length} of{" "}
                  {serviceOverviewFixtures.length} services
                </div>
                {visibleServices.length ? (
                  <div className="service-inventory-list">
                    {visibleServices.map((service) => (
                      <ServiceInventoryRow
                        key={service.name}
                        service={service}
                      />
                    ))}
                  </div>
                ) : (
                  <div className="service-empty-state">
                    <Boxes aria-hidden="true" />
                    <strong>No services match</strong>
                    <span>Try another category or search term.</span>
                  </div>
                )}
              </section>

              <section
                className="service-catalog-panel"
                aria-labelledby="service-catalog-title"
              >
                <div className="service-panel-heading">
                  <div>
                    <p className="section-kicker">Preset catalog</p>
                    <h3 id="service-catalog-title">
                      Discover the next service for a project
                    </h3>
                  </div>
                  <Badge variant="outline">Available and planned</Badge>
                </div>
                <div className="service-catalog-grid">
                  {serviceCatalogFixtures.map((service) => {
                    const Icon = service.icon;
                    const available = service.state === "Available";
                    return (
                      <article
                        className="service-catalog-card"
                        key={`${service.category}-${service.name}`}
                      >
                        <div className="service-catalog-icon">
                          <Icon aria-hidden="true" />
                        </div>
                        <div className="service-catalog-copy">
                          <strong>{service.name}</strong>
                          <span>{service.category}</span>
                          <p>{service.detail}</p>
                        </div>
                        <Badge
                          variant="outline"
                          className={
                            available ? "service-available" : "service-planned"
                          }
                        >
                          {service.state}
                        </Badge>
                      </article>
                    );
                  })}
                </div>
              </section>

              <section
                className="service-state-strip"
                aria-label="Service empty and conflict states"
              >
                <div>
                  <Database aria-hidden="true" />
                  <span>
                    <strong>All installed</strong>
                    <small>Every preset is already available.</small>
                  </span>
                  <Badge variant="outline">Available</Badge>
                </div>
                <div>
                  <Boxes aria-hidden="true" />
                  <span>
                    <strong>Empty category</strong>
                    <small>No queue service is installed yet.</small>
                  </span>
                  <Badge variant="outline">Empty</Badge>
                </div>
                <div>
                  <AlertTriangle aria-hidden="true" />
                  <span>
                    <strong>Dependency warning</strong>
                    <small>
                      Meilisearch shares a port with another process.
                    </small>
                  </span>
                  <Badge variant="outline">Attention</Badge>
                </div>
              </section>
            </div>
          </ScrollArea>
        </main>
      </SidebarInset>

      <Dialog open={isInstallDialogOpen} onOpenChange={setIsInstallDialogOpen}>
        <DialogContent className="service-install-dialog">
          <DialogHeader>
            <DialogTitle>Add a service</DialogTitle>
            <DialogDescription>
              Preview of the future install flow. Nothing is installed during
              the static UI phase.
            </DialogDescription>
          </DialogHeader>
          <div className="service-install-preview">
            <div>
              <Settings2 aria-hidden="true" />
              <span>
                <strong>Choose a preset</strong>
                <small>Select category, version, and project scope.</small>
              </span>
              <Badge variant="outline">Planned</Badge>
            </div>
            <div>
              <Database aria-hidden="true" />
              <span>
                <strong>Review dependencies</strong>
                <small>Check ports, volumes, and required tools.</small>
              </span>
              <Badge variant="outline">Planned</Badge>
            </div>
            <div>
              <HardDrive aria-hidden="true" />
              <span>
                <strong>Confirm install</strong>
                <small>Show the operation plan before making changes.</small>
              </span>
              <Badge variant="outline">Planned</Badge>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsInstallDialogOpen(false)}
            >
              Close preview
            </Button>
            <Button disabled>Continue to preset selection</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SidebarProvider>
  );
}

function ServiceSummary({
  label,
  value,
  detail,
  icon: Icon,
}: {
  label: string;
  value: string;
  detail: string;
  icon: typeof Boxes;
}) {
  return (
    <article className="service-summary-card">
      <div>
        <span>{label}</span>
        <Icon aria-hidden="true" />
      </div>
      <strong>{value}</strong>
      <small>{detail}</small>
    </article>
  );
}

function ServiceInventoryRow({
  service,
}: {
  service: (typeof serviceOverviewFixtures)[number];
}) {
  const Icon = service.icon;
  return (
    <article className="service-inventory-row">
      <div className={`service-inventory-icon ${service.state}`}>
        <Icon aria-hidden="true" />
      </div>
      <div className="service-inventory-name">
        <div>
          <strong>{service.name}</strong>
          <Badge variant="outline">{service.category}</Badge>
        </div>
        <small>{service.version}</small>
      </div>
      <div className="service-inventory-detail">
        <span>
          <strong>{service.projects}</strong>
          <small>Usage</small>
        </span>
        <span>
          <strong>{service.ports}</strong>
          <small>Ports</small>
        </span>
        <span>
          <strong>{service.dependencies}</strong>
          <small>Dependencies</small>
        </span>
      </div>
      <Badge
        variant="outline"
        className={`service-update-state ${service.state}`}
      >
        {service.update}
      </Badge>
      <Button
        variant="outline"
        size="sm"
        render={<Link to="/services/$name" params={{ name: service.name }} />}
        aria-label={`Open ${service.name} detail`}
      >
        Details
      </Button>
    </article>
  );
}
