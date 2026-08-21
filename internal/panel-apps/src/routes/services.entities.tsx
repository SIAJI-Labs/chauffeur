import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import {
  dedicatedServicePageFixtures,
  serviceOverviewFixtures,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/services/entities")({
  component: () => (
    <InitializedPage
      fixture={dedicatedServicePageFixtures.entities}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Services", to: "/services" },
        { label: "Entity actions" },
      ]}
      selector={{
        label: "Service",
        options: serviceOverviewFixtures.map((service) => service.name),
      }}
      actions={[
        { label: "Create entity", status: "Requires backend" },
        { label: "Delete entity", status: "Requires backend" },
      ]}
    />
  ),
});
