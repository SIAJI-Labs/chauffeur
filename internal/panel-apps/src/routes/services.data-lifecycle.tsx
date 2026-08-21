import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import {
  dedicatedServicePageFixtures,
  serviceOverviewFixtures,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/services/data-lifecycle")({
  component: () => (
    <InitializedPage
      fixture={dedicatedServicePageFixtures.lifecycle}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Services", to: "/services" },
        { label: "Import, export, snapshots" },
      ]}
      selector={{
        label: "Service",
        options: serviceOverviewFixtures.map((service) => service.name),
      }}
      actions={[
        { label: "Import snapshot", status: "Requires backend" },
        { label: "Create export", status: "Planned" },
      ]}
    />
  ),
});
