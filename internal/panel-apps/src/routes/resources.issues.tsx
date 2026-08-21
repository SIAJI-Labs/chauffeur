import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { initializedPageFixtures } from "@/data/webui-fixtures";

export const Route = createFileRoute("/resources/issues")({
  component: () => (
    <InitializedPage
      fixture={initializedPageFixtures.issues}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Resources" },
        { label: "Issues and diagnostics" },
      ]}
      actions={[
        { label: "Run workspace doctor", status: "Requires backend" },
        { label: "Refresh issues", status: "Planned" },
      ]}
    />
  ),
});
