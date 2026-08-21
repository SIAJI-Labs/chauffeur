import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import {
  dedicatedServicePageFixtures,
  serviceOverviewFixtures,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/services/configuration")({
  component: () => (
    <InitializedPage
      fixture={dedicatedServicePageFixtures.configuration}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Services", to: "/services" },
        { label: "Service configuration" },
      ]}
      selector={{
        label: "Service",
        options: serviceOverviewFixtures.map((service) => service.name),
      }}
      actions={[
        { label: "Save configuration", status: "Requires backend" },
        { label: "Reset changes", status: "Planned" },
      ]}
    />
  ),
});
