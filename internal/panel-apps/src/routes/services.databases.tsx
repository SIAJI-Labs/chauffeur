import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import {
  dedicatedServicePageFixtures,
  serviceOverviewFixtures,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/services/databases")({
  component: () => (
    <InitializedPage
      fixture={dedicatedServicePageFixtures.databases}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Services", to: "/services" },
        { label: "Databases and entities" },
      ]}
      selector={{
        label: "Service",
        options: serviceOverviewFixtures.map((service) => service.name),
      }}
      actions={[
        { label: "Create database", status: "Requires backend" },
        { label: "Copy DSN", status: "Planned" },
      ]}
    />
  ),
});
