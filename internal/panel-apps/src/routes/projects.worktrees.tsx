import { createFileRoute } from "@tanstack/react-router";

import { InitializedPage } from "@/components/initialized-page";
import { dedicatedProjectPageFixtures, projects } from "@/data/webui-fixtures";

export const Route = createFileRoute("/projects/worktrees")({
  component: () => (
    <InitializedPage
      fixture={dedicatedProjectPageFixtures.worktrees}
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Projects", to: "/projects" },
        { label: "Project worktrees" },
      ]}
      selector={{
        label: "Project",
        options: projects.map((project) => project.name),
      }}
      actions={[
        { label: "Add worktree", status: "Requires backend" },
        { label: "Manage isolation", status: "Planned" },
      ]}
    />
  ),
});
