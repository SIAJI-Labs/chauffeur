import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { initializedPageFixtures } from "@/data/webui-fixtures";

export const Route = createFileRoute("/resources/telemetry")({
  component: () => (
    <InitializedPage
      fixture={initializedPageFixtures.telemetry}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Resources" },
        { label: "CPU and memory" },
      ]}
      actions={[{ label: "Connect telemetry", status: "Requires backend" }]}
    />
  ),
});
