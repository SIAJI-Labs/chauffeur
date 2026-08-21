import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import {
  dedicatedServicePageFixtures,
  serviceOverviewFixtures,
} from "@/data/webui-fixtures";

export const Route = createFileRoute("/services/logs")({
  component: () => (
    <InitializedPage
      fixture={dedicatedServicePageFixtures.logs}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Services", to: "/services" },
        { label: "Service logs" },
      ]}
      selector={{
        label: "Service",
        options: serviceOverviewFixtures.map((service) => service.name),
      }}
      actions={[
        { label: "Copy logs", status: "Planned" },
        { label: "Reconnect stream", status: "Requires backend" },
      ]}
    />
  ),
});
