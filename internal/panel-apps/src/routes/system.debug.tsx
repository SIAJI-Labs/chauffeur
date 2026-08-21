import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { initializedPageFixtures } from "@/data/webui-fixtures";

export const Route = createFileRoute("/system/debug")({
  component: () => (
    <InitializedPage
      fixture={initializedPageFixtures.debug}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "System", to: "/system" },
        { label: "Debug bridge" },
      ]}
      actions={[
        { label: "Enable debug bridge", status: "Requires backend" },
        { label: "Configure lenses", status: "Planned" },
      ]}
    />
  ),
});
