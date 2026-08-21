import { createFileRoute } from "@tanstack/react-router";
import { Check, FolderGit2, Globe2, Server } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { projectLinkPageFixture } from "@/data/webui-fixtures";

export const Route = createFileRoute("/projects/link")({
  component: ProjectLinkPage,
});

function ProjectLinkPage() {
  return (
    <AppShell
      title="Link a project"
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Projects", to: "/projects" },
        { label: "Link a project" },
      ]}
      contentId="project-link-content"
      skipLabel="Skip to link project"
    >
      <div
        className="initialized-page project-link-page"
        id="project-link-content"
      >
        <section
          className="initialized-page-hero"
          aria-labelledby="project-link-title"
        >
          <div>
            <p className="section-kicker">Projects / onboarding</p>
            <div className="initialized-page-title">
              <span className="initialized-page-icon">
                <FolderGit2 aria-hidden="true" />
              </span>
              <h2 id="project-link-title">Link a project</h2>
            </div>
            <p>
              Review a local folder, runtime profile, and route before adding it
              to the workspace.
            </p>
          </div>
          <Badge variant="outline">Next</Badge>
        </section>
        <div className="project-link-step-grid">
          {projectLinkPageFixture.steps.map((step) => (
            <Card key={step.number}>
              <CardHeader>
                <div className="project-link-step-number">{step.number}</div>
                <CardTitle>{step.title}</CardTitle>
              </CardHeader>
              <CardContent>
                <p>{step.detail}</p>
                <Badge variant="outline">Not connected</Badge>
              </CardContent>
            </Card>
          ))}
        </div>
        <section
          className="project-link-review"
          aria-labelledby="project-link-review-title"
        >
          <div className="initialized-section-heading">
            <div>
              <p className="section-kicker">Sample review</p>
              <h3 id="project-link-review-title">
                Example project configuration
              </h3>
            </div>
            <Badge variant="outline">Fixture data</Badge>
          </div>
          <dl>
            <div>
              <dt>
                <FolderGit2 aria-hidden="true" />
                Local folder
              </dt>
              <dd>{projectLinkPageFixture.samplePath}</dd>
            </div>
            <div>
              <dt>
                <Server aria-hidden="true" />
                Runtime
              </dt>
              <dd>{projectLinkPageFixture.runtime}</dd>
            </div>
            <div>
              <dt>
                <Globe2 aria-hidden="true" />
                Local route
              </dt>
              <dd>{projectLinkPageFixture.route}</dd>
            </div>
            <div>
              <dt>
                <Check aria-hidden="true" />
                HTTPS
              </dt>
              <dd>{projectLinkPageFixture.https}</dd>
            </div>
          </dl>
        </section>
        <section className="initialized-actions">
          <div>
            <p className="section-kicker">Final step</p>
            <h3>Linking is unavailable</h3>
            <p>
              No folder picker, filesystem change, route creation, or submit
              request is performed.
            </p>
          </div>
          <Button disabled>
            Link project <Badge variant="outline">Requires backend</Badge>
          </Button>
        </section>
      </div>
    </AppShell>
  );
}
