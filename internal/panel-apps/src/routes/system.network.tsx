import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { initializedPageFixtures } from "@/data/webui-fixtures";

export const Route = createFileRoute("/system/network")({
  component: () => (
    <InitializedPage
      fixture={initializedPageFixtures.network}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "System", to: "/system" },
        { label: "DNS and Nginx" },
      ]}
      actions={[
        { label: "Edit configuration", status: "Requires backend" },
        { label: "Reload gateway", status: "Planned" },
      ]}
    />
  ),
});
