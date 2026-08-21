import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import {
  dedicatedServicePageFixtures,
  serviceOverviewFixtures,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/services/details")({
  component: () => (
    <InitializedPage
      fixture={dedicatedServicePageFixtures.details}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Services", to: "/services" },
        { label: "Service details" },
      ]}
      selector={{
        label: "Service",
        options: serviceOverviewFixtures.map((service) => service.name),
      }}
      actions={[
        { label: "Restart service", status: "Requires backend" },
        { label: "Open admin", status: "Planned" },
      ]}
    />
  ),
});
