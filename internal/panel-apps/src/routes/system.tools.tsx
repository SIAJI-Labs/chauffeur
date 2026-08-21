import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { initializedPageFixtures } from "@/data/webui-fixtures";

export const Route = createFileRoute("/system/tools")({
  component: () => (
    <InitializedPage
      fixture={initializedPageFixtures.tools}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "System", to: "/system" },
        { label: "Tools and watchers" },
      ]}
      actions={[
        { label: "Check for updates", status: "Requires backend" },
        { label: "Start watcher", status: "Planned" },
      ]}
    />
  ),
});
