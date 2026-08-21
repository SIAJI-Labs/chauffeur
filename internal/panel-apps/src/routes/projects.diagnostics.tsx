import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { dedicatedProjectPageFixtures, projects } from "@/data/webui-fixtures";

export const Route = createFileRoute("/projects/diagnostics")({
  component: () => (
    <InitializedPage
      fixture={dedicatedProjectPageFixtures.diagnostics}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Projects", to: "/projects" },
        { label: "Project diagnostics" },
      ]}
      selector={{
        label: "Project",
        options: projects.map((project) => project.name),
      }}
      actions={[
        { label: "Run diagnostics", status: "Requires backend" },
        { label: "Export report", status: "Planned" },
      ]}
    />
  ),
});
