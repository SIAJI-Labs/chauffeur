import { Link } from "@tanstack/react-router";
import { CircleAlert } from "lucide-react";

import { AppNavbar } from "@/components/app-navbar";
import { AppSidebar } from "@/components/app-sidebar";
import { Button } from "@/components/ui/button";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";

type ResourceNotFoundProps = {
  kind: "project" | "service";
  name: string;
};

export function ResourceNotFound({ kind, name }: ResourceNotFoundProps) {
  const config =
    kind === "project"
      ? {
          label: "Project",
          listLabel: "Projects",
          listTo: "/projects" as const,
          contentId: "project-not-found-content",
        }
      : {
          label: "Service",
          listLabel: "Services",
          listTo: "/services" as const,
          contentId: "service-not-found-content",
        };

  return (
    <SidebarProvider className="dashboard-frame">
      <a className="skip-link" href={`#${config.contentId}`}>
        Skip to not found message
      </a>
      <AppSidebar />
      <SidebarInset className="dashboard-shell">
        <AppNavbar
          title={`${config.label} not found`}
          breadcrumbs={[
            { label: "Workspace", to: "/" },
            { label: config.listLabel, to: config.listTo },
            { label: name },
          ]}
        />
        <main
          className="dashboard-content min-h-0 flex-1"
          id={config.contentId}
        >
          <div className="resource-not-found">
            <CircleAlert aria-hidden="true" />
            <p className="section-kicker">No match</p>
            <h2>
              {config.label} &ldquo;{name}&rdquo; was not found.
            </h2>
            <p>
              Check the address or return to the{" "}
              {config.listLabel.toLowerCase()} list to choose an available{" "}
              {kind}.
            </p>
            <Button render={<Link to={config.listTo} />}>
              View {config.listLabel.toLowerCase()}
            </Button>
          </div>
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}
