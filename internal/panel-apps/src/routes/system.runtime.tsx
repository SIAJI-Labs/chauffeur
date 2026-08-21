import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { initializedPageFixtures } from "@/data/webui-fixtures";

export const Route = createFileRoute("/system/runtime")({
  component: () => (
    <InitializedPage
      fixture={initializedPageFixtures.runtime}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "System", to: "/system" },
        { label: "Chauffeur runtime" },
      ]}
      actions={[
        { label: "Restart runtime", status: "Requires backend" },
        { label: "Update runtime", status: "Planned" },
      ]}
    />
  ),
});
