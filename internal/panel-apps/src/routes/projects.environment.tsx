import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { dedicatedProjectPageFixtures, projects } from "@/data/webui-fixtures";

export const Route = createFileRoute("/projects/environment")({
  component: () => (
    <InitializedPage
      fixture={dedicatedProjectPageFixtures.environment}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Projects", to: "/projects" },
        { label: "Project environment" },
      ]}
      selector={{
        label: "Project",
        options: projects.map((project) => project.name),
      }}
      actions={[
        { label: "Edit environment", status: "Requires backend" },
        { label: "Restore values", status: "Planned" },
      ]}
    />
  ),
});
