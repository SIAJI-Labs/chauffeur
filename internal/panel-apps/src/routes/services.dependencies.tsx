import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import {
  dedicatedServicePageFixtures,
  serviceOverviewFixtures,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/services/dependencies")({
  component: () => (
    <InitializedPage
      fixture={dedicatedServicePageFixtures.dependencies}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Services", to: "/services" },
        { label: "Ports and dependencies" },
      ]}
      selector={{
        label: "Service",
        options: serviceOverviewFixtures.map((service) => service.name),
      }}
      actions={[
        { label: "Resolve conflict", status: "Requires backend" },
        { label: "Expose port", status: "Planned" },
      ]}
    />
  ),
});
