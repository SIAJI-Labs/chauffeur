import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { initializedPageFixtures } from "@/data/webui-fixtures";

export const Route = createFileRoute("/resources/usage")({
  component: () => (
    <InitializedPage
      fixture={initializedPageFixtures.usage}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Resources" },
        { label: "Resource usage" },
      ]}
    />
  ),
});
