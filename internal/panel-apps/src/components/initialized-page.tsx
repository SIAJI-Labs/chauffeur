import * as React from "react";
import type { LucideIcon } from "lucide-react";
import type { BreadcrumbEntry } from "@/components/app-navbar";
import { AppShell } from "@/components/app-shell";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export type InitializedPageFixture = {
  eyebrow: string;
  title: string;
  description: string;
  status: string;
  icon: LucideIcon;
  metrics: ReadonlyArray<{ label: string; value: string; detail: string }>;
  sections: ReadonlyArray<{ title: string; detail: string }>;
  rows?: ReadonlyArray<{
    label: string;
    value: string;
    detail: string;
    status?: string;
  }>;
};

export function InitializedPage({
  fixture,
  breadcrumbs,
  actions = [],
  selector,
}: {
  fixture: InitializedPageFixture;
  breadcrumbs: Array<BreadcrumbEntry>;
  actions?: Array<{ label: string; status: string }>;
  selector?: { label: string; options: ReadonlyArray<string> };
}) {
  const Icon = fixture.icon;
  const [selected, setSelected] = React.useState(selector?.options[0] ?? "");
  return (
    <AppShell
      title={fixture.title}
      breadcrumbs={breadcrumbs}
      contentId="initialized-page-content"
      skipLabel={`Skip to ${fixture.title}`}
    >
      <div className="initialized-page" id="initialized-page-content">
        <section
          className="initialized-page-hero"
          aria-labelledby="initialized-page-title"
        >
          <div>
            <p className="section-kicker">{fixture.eyebrow}</p>
            <div className="initialized-page-title">
              <span className="initialized-page-icon">
                <Icon aria-hidden="true" />
              </span>
              <h2 id="initialized-page-title">{fixture.title}</h2>
            </div>
            <p>{fixture.description}</p>
          </div>
          <Badge variant="outline">{fixture.status}</Badge>
        </section>

        {selector ? (
          <section
            className="initialized-selector"
            aria-label={`${selector.label} selector`}
          >
            <label htmlFor="initialized-page-selector">{selector.label}</label>
            <select
              id="initialized-page-selector"
              value={selected}
              onChange={(event) => setSelected(event.target.value)}
            >
              {selector.options.map((option) => (
                <option key={option}>{option}</option>
              ))}
            </select>
            <Badge variant="outline">Static selection</Badge>
          </section>
        ) : null}

        <section
          className="initialized-metric-grid"
          aria-label={`${fixture.title} summary`}
        >
          {fixture.metrics.map((metric) => (
            <Card key={metric.label}>
              <CardHeader>
                <p className="initialized-metric-label">{metric.label}</p>
                <CardTitle>{metric.value}</CardTitle>
              </CardHeader>
              <CardContent>
                <p>{metric.detail}</p>
              </CardContent>
            </Card>
          ))}
        </section>

        <section
          className="initialized-section-grid"
          aria-label={`${fixture.title} sections`}
        >
          {fixture.sections.map((section) => (
            <Card key={section.title}>
              <CardHeader>
                <CardTitle>{section.title}</CardTitle>
              </CardHeader>
              <CardContent>
                <p>{section.detail}</p>
              </CardContent>
            </Card>
          ))}
        </section>

        {fixture.rows?.length ? (
          <section
            className="initialized-data-panel"
            aria-labelledby="initialized-data-title"
          >
            <div className="initialized-section-heading">
              <div>
                <p className="section-kicker">Fixture records</p>
                <h3 id="initialized-data-title">Representative data</h3>
              </div>
              <Badge variant="outline">Static data</Badge>
            </div>
            <div className="initialized-data-list">
              {fixture.rows.map((row) => (
                <div
                  className="initialized-data-row"
                  key={`${row.label}-${row.value}`}
                >
                  <div>
                    <strong>{row.label}</strong>
                    <small>{row.detail}</small>
                  </div>
                  <span>{row.value}</span>
                  {row.status ? (
                    <Badge variant="outline">{row.status}</Badge>
                  ) : null}
                </div>
              ))}
            </div>
          </section>
        ) : null}

        <section
          className="initialized-actions"
          aria-labelledby="initialized-actions-title"
        >
          <div>
            <p className="section-kicker">Controls</p>
            <h3 id="initialized-actions-title">
              Available in the final page layout
            </h3>
            <p>
              These controls are intentionally unavailable until the supporting
              runtime API exists.
            </p>
          </div>
          <div className="initialized-action-list">
            {(actions.length
              ? actions
              : [{ label: "Manage component", status: "Requires backend" }]
            ).map((action) => (
              <Button
                key={action.label}
                variant="outline"
                disabled
                aria-label={`${action.label}, ${action.status}`}
              >
                {action.label}
                <Badge variant="outline">{action.status}</Badge>
              </Button>
            ))}
          </div>
        </section>
      </div>
    </AppShell>
  );
}
