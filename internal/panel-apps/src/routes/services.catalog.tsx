import * as React from "react";
import { createFileRoute } from "@tanstack/react-router";
import { Boxes, Search } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  serviceCatalogFixtures,
  serviceCategories,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/services/catalog")({
  component: ServiceCatalogPage,
});

function ServiceCatalogPage() {
  const [query, setQuery] = React.useState("");
  const filteredServices = serviceCatalogFixtures.filter((service) =>
    `${service.name} ${service.category} ${service.detail}`
      .toLowerCase()
      .includes(query.toLowerCase()),
  );
  return (
    <AppShell
      title="Service catalog"
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Services", to: "/services" },
        { label: "Service catalog" },
      ]}
      contentId="service-catalog-content"
      skipLabel="Skip to service catalog"
    >
      <div
        className="initialized-page service-catalog-page"
        id="service-catalog-content"
      >
        <section
          className="initialized-page-hero"
          aria-labelledby="service-catalog-title"
        >
          <div>
            <p className="section-kicker">Services / discovery</p>
            <div className="initialized-page-title">
              <span className="initialized-page-icon">
                <Boxes aria-hidden="true" />
              </span>
              <h2 id="service-catalog-title">Service catalog</h2>
            </div>
            <p>
              Browse local service presets before choosing a future installation
              or project pairing.
            </p>
          </div>
          <Badge variant="outline">Static catalog</Badge>
        </section>
        <section
          className="catalog-toolbar"
          aria-label="Service catalog filters"
        >
          <div className="catalog-search">
            <Search aria-hidden="true" />
            <Input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search services"
              aria-label="Search services"
            />
          </div>
          <div
            className="catalog-categories"
            role="list"
            aria-label="Service categories"
          >
            {serviceCategories.slice(0, 6).map((category) => (
              <Button
                key={category}
                variant={category === "All services" ? "secondary" : "ghost"}
                size="sm"
                disabled={category !== "All services"}
              >
                {category}
              </Button>
            ))}
          </div>
        </section>
        <section
          className="catalog-grid"
          aria-live="polite"
          aria-label="Service catalog results"
        >
          {filteredServices.map((service) => {
            const Icon = service.icon;
            return (
              <article className="catalog-card" key={service.name}>
                <div className="catalog-card-heading">
                  <span className="initialized-page-icon">
                    <Icon aria-hidden="true" />
                  </span>
                  <Badge variant="outline">{service.state}</Badge>
                </div>
                <h3>{service.name}</h3>
                <p className="catalog-card-category">{service.category}</p>
                <p>{service.detail}</p>
                <Button variant="outline" disabled>
                  {service.state === "Available" ? "Install" : "View plan"}
                  <Badge variant="outline">Requires backend</Badge>
                </Button>
              </article>
            );
          })}
          {!filteredServices.length ? (
            <div className="dashboard-empty-state">
              <h3>No matching services</h3>
              <p>Try another catalog search.</p>
            </div>
          ) : null}
        </section>
      </div>
    </AppShell>
  );
}
