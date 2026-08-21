import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { initializedPageFixtures } from "@/data/webui-fixtures";

export const Route = createFileRoute("/resources/workers")({
  component: () => (
    <InitializedPage
      fixture={initializedPageFixtures.workers}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Resources" },
        { label: "Workers" },
      ]}
      actions={[
        { label: "Start worker", status: "Requires backend" },
        { label: "Execution mode", status: "Planned" },
      ]}
    />
  ),
});
