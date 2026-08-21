import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { dedicatedProjectPageFixtures, projects } from "@/data/webui-fixtures";

export const Route = createFileRoute("/projects/overview")({
  component: () => (
    <InitializedPage
      fixture={dedicatedProjectPageFixtures.overview}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Projects", to: "/projects" },
        { label: "Project overview" },
      ]}
      selector={{
        label: "Project",
        options: projects.map((project) => project.name),
      }}
      actions={[
        { label: "Manage runtime", status: "Requires backend" },
        { label: "Review workers", status: "Planned" },
      ]}
    />
  ),
});
