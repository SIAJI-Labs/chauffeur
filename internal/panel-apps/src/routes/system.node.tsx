import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { initializedPageFixtures } from "@/data/webui-fixtures";

export const Route = createFileRoute("/system/node")({
  component: () => (
    <InitializedPage
      fixture={initializedPageFixtures.node}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "System", to: "/system" },
        { label: "Node.js and Bun" },
      ]}
      actions={[
        { label: "Install runtime", status: "Requires backend" },
        { label: "Set default", status: "Planned" },
      ]}
    />
  ),
});
