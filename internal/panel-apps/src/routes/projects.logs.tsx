import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { dedicatedProjectPageFixtures, projects } from "@/data/webui-fixtures";

export const Route = createFileRoute("/projects/logs")({
  component: () => (
    <InitializedPage
      fixture={dedicatedProjectPageFixtures.logs}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Projects", to: "/projects" },
        { label: "Project logs" },
      ]}
      selector={{
        label: "Project",
        options: projects.map((project) => project.name),
      }}
      actions={[
        { label: "Copy logs", status: "Planned" },
        { label: "Reconnect stream", status: "Requires backend" },
      ]}
    />
  ),
});
