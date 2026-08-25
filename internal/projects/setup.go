package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// RuntimeChoice describes a runtime option without mutating the workspace.
type RuntimeChoice struct {
	Version     string
	State       string
	Recommended bool
	Evidence    string
}

type ProjectFacts struct {
	Path          string
	Slug          string
	Type          ProjectType
	DocumentRoot  string
	PHPConstraint string
	Evidence      []string
	Existing      *Project
}

type SetupChoices struct {
	PHPVersion string
	Domain     string
	Aliases    []string
	SSL        bool
	Dedicated  bool
}

type SetupWarning struct {
	Code     string
	Message  string
	Evidence string
}

type PlannedChange struct {
	Target string
	Action string
}

type ApplyStatus string

const (
	ApplyLinked   ApplyStatus = "linked"
	ApplyReady    ApplyStatus = "ready"
	ApplyDegraded ApplyStatus = "degraded"
	ApplyFailed   ApplyStatus = "failed"
)

// ApplyResult describes the state reached by the mutation boundary. A linked
// project may still be degraded when services are not running yet.
type ApplyResult struct {
	Status      ApplyStatus
	Evidence    []string
	Remediation string
}

type ApplyDependencies struct {
	Snapshot       func() error
	Save           func() error
	GenerateSSL    func() error
	PrepareRuntime func() error
	GenerateNginx  func() error
	EnableRoute    func() error
	Reload         func() error
	Readiness      func() (string, error)
}

// ApplyProjectSetup is the shared mutation boundary. Detection and rendering
// callers provide intent-specific operations, while this function owns the
// ordering, cancellation checks, and structured outcome.
func ApplyProjectSetup(ctx context.Context, plan SetupPlan, deps ApplyDependencies) (ApplyResult, error) {
	if err := plan.Validate(); err != nil {
		return ApplyResult{Status: ApplyFailed, Remediation: "resolve setup blockers and retry"}, err
	}
	step := func(operation func() error, remediation string) (ApplyResult, error) {
		if err := ctx.Err(); err != nil {
			return ApplyResult{Status: ApplyFailed, Remediation: "retry the setup after cancellation"}, err
		}
		if operation == nil {
			return ApplyResult{}, nil
		}
		if err := operation(); err != nil {
			return ApplyResult{Status: ApplyFailed, Remediation: remediation}, err
		}
		return ApplyResult{}, nil
	}
	if result, err := step(deps.Snapshot, "check the existing project configuration and retry"); err != nil {
		return result, err
	}
	if result, err := step(deps.Save, "check project configuration permissions and retry"); err != nil {
		return result, err
	}
	if result, err := step(deps.GenerateSSL, "fix SSL readiness and retry"); err != nil {
		return result, err
	}
	if result, err := step(deps.PrepareRuntime, "fix the selected runtime and retry"); err != nil {
		return result, err
	}
	if result, err := step(deps.GenerateNginx, "fix nginx configuration generation and retry"); err != nil {
		return result, err
	}
	if result, err := step(deps.EnableRoute, "fix nginx site permissions and retry"); err != nil {
		return result, err
	}
	if result, err := step(deps.Reload, "inspect nginx logs and retry reload"); err != nil {
		return result, err
	}
	result := ApplyResult{Status: ApplyLinked, Evidence: []string{"project intent and route configuration were applied"}}
	if deps.Readiness == nil {
		return result, nil
	}
	evidence, err := deps.Readiness()
	if evidence != "" {
		result.Evidence = append(result.Evidence, evidence)
	}
	if err != nil {
		result.Status = ApplyDegraded
		result.Remediation = "run the readiness check again after fixing the reported service state"
		return result, nil
	}
	result.Status = ApplyReady
	return result, nil
}

// SetupPlan is the shared read-only result consumed by the CLI wizard and
// future web UI. It contains intent, not runtime state or secrets.
type SetupPlan struct {
	Facts      ProjectFacts
	Choices    SetupChoices
	PHPChoices []RuntimeChoice
	Changes    []PlannedChange
	Warnings   []SetupWarning
	Blockers   []SetupWarning
}

type DetectionInput struct {
	Path         string
	TypeOverride ProjectType
	Existing     *Project
	DefaultPHP   string
	ExplicitPHP  string
	InstalledPHP []string
	DomainTLD    string
	SSLReady     bool
}

// DetectSetup gathers project facts and recommendations without writing files,
// starting processes, or invoking project-provided commands.
func DetectSetup(input DetectionInput) (SetupPlan, error) {
	info, err := os.Stat(input.Path)
	if err != nil {
		return SetupPlan{}, fmt.Errorf("inspect project path: %w", err)
	}
	if !info.IsDir() {
		return SetupPlan{}, fmt.Errorf("project path is not a directory: %s", input.Path)
	}
	typeDetected := Detect(input.Path)
	projectType := typeDetected
	if input.TypeOverride != TypeUnknown {
		projectType = input.TypeOverride
	}
	slug := GenerateSlug(input.Path)
	if slug == "" {
		return SetupPlan{}, fmt.Errorf("could not derive project slug from %s", input.Path)
	}
	domain := DomainFromSlug(slug, input.DomainTLD)
	if input.Existing != nil {
		domain = input.Existing.Domain
	}
	constraint := detectPHPConstraint(input.Path)
	choices := SetupChoices{PHPVersion: input.DefaultPHP, Domain: domain}
	if input.Existing != nil {
		choices.PHPVersion = input.Existing.PHPVersion
		choices.Aliases = append([]string(nil), input.Existing.Aliases...)
		choices.SSL = input.Existing.SSL
		choices.Dedicated = input.Existing.FPM.Dedicated
	}
	plan := BuildSetupPlan(ProjectFacts{
		Path:          input.Path,
		Slug:          slug,
		Type:          projectType,
		DocumentRoot:  DocumentRoot(input.Path, projectType),
		PHPConstraint: constraint,
		Evidence:      projectEvidence(input.Path, projectType),
		Existing:      input.Existing,
	}, choices, nil)
	if input.ExplicitPHP != "" {
		choices.PHPVersion = majorMinor(input.ExplicitPHP)
		plan.Choices.PHPVersion = choices.PHPVersion
	}
	plan.PHPChoices, plan.Choices.PHPVersion = recommendPHP(input, constraint, choices.PHPVersion)
	if plan.Choices.SSL && !input.SSLReady {
		plan.Warnings = append(plan.Warnings, SetupWarning{Code: "ssl-not-ready", Message: "SSL is selected but certificate prerequisites are not ready", Evidence: "mkcert/CA readiness must be checked before apply"})
	}
	if plan.Choices.PHPVersion == "" && projectType != TypeReverseProxy {
		plan.Blockers = append(plan.Blockers, SetupWarning{Code: "php-unavailable", Message: "no available PHP version satisfies the project constraint", Evidence: constraint})
	}
	return plan, nil
}

func recommendPHP(input DetectionInput, constraint, preferred string) ([]RuntimeChoice, string) {
	choices := make([]RuntimeChoice, 0, len(input.InstalledPHP))
	explicit := majorMinor(input.ExplicitPHP)
	valid := func(version string) bool { return constraint == "" || PHPVersionSatisfies(version, constraint) }
	for _, version := range input.InstalledPHP {
		mm := majorMinor(version)
		if mm == "" {
			continue
		}
		choice := RuntimeChoice{Version: mm, State: "installed"}
		if !valid(mm) {
			choice.State = "incompatible"
			choice.Evidence = "does not satisfy composer.json PHP constraint " + constraint
		}
		choices = append(choices, choice)
	}
	if explicit != "" {
		for i := range choices {
			if choices[i].Version == explicit && choices[i].State == "installed" && valid(explicit) {
				choices[i].Recommended = true
				choices[i].Evidence = "explicit --php selection"
				return choices, explicit
			}
		}
		choices = append(choices, RuntimeChoice{Version: explicit, State: "unavailable", Evidence: "explicit --php selection is not available or incompatible"})
		return choices, ""
	}
	if preferred != "" {
		preferred = majorMinor(preferred)
		for i := range choices {
			if choices[i].Version == preferred && choices[i].State == "installed" && valid(preferred) {
				choices[i].Recommended = true
				choices[i].Evidence = "matches the existing project or workspace default and composer.json"
				return choices, preferred
			}
		}
	}
	for i := range choices {
		if choices[i].State == "installed" && valid(choices[i].Version) {
			choices[i].Recommended = true
			choices[i].Evidence = "available and satisfies composer.json constraint " + constraint
			return choices, choices[i].Version
		}
	}
	return choices, ""
}

func detectPHPConstraint(path string) string {
	data, err := os.ReadFile(filepath.Join(path, "composer.json"))
	if err != nil {
		return ""
	}
	var manifest struct {
		Require map[string]string `json:"require"`
	}
	if json.Unmarshal(data, &manifest) != nil {
		return ""
	}
	return strings.TrimSpace(manifest.Require["php"])
}

// PHPConstraint reads only composer.json and returns its declared PHP
// requirement. It never invokes Composer or evaluates project code.
func PHPConstraint(path string) string { return detectPHPConstraint(path) }

func majorMinor(version string) string {
	parts := strings.Split(strings.TrimSpace(version), ".")
	if len(parts) < 2 {
		return ""
	}
	if _, err := strconv.Atoi(parts[0]); err != nil {
		return ""
	}
	if _, err := strconv.Atoi(parts[1]); err != nil {
		return ""
	}
	return parts[0] + "." + parts[1]
}

// PHPVersionSatisfies handles the Composer constraint forms used by project
// manifests without invoking Composer or executing project code.
func PHPVersionSatisfies(version, constraint string) bool {
	v := parsePHPVersion(version)
	for _, alternative := range strings.Split(constraint, "||") {
		ok := true
		for _, token := range strings.Fields(strings.TrimSpace(alternative)) {
			if !phpConstraintTokenMatches(v, token) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

type phpVersion struct{ major, minor, patch int }

func parsePHPVersion(version string) phpVersion {
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(version), "v"), ".")
	result := phpVersion{}
	if len(parts) > 0 {
		result.major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) > 1 {
		result.minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) > 2 {
		result.patch, _ = strconv.Atoi(parts[2])
	}
	return result
}

func phpConstraintTokenMatches(version phpVersion, token string) bool {
	token = strings.TrimSpace(token)
	if token == "" || token == "*" {
		return true
	}
	prefix := "="
	value := token
	for _, candidate := range []string{"^", "~", ">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(token, candidate) {
			prefix, value = candidate, strings.TrimSpace(strings.TrimPrefix(token, candidate))
			break
		}
	}
	trimmedValue := strings.TrimSuffix(value, ".*")
	target := parsePHPVersion(trimmedValue)
	componentCount := len(strings.Split(trimmedValue, "."))
	compare := func(a, b phpVersion) int {
		if a.major != b.major {
			if a.major < b.major {
				return -1
			}
			return 1
		}
		if a.minor != b.minor {
			if a.minor < b.minor {
				return -1
			}
			return 1
		}
		if a.patch < b.patch {
			return -1
		}
		if a.patch > b.patch {
			return 1
		}
		return 0
	}
	switch prefix {
	case ">=":
		return compare(version, target) >= 0
	case "<=":
		return compare(version, target) <= 0
	case ">":
		return compare(version, target) > 0
	case "<":
		return compare(version, target) < 0
	case "^":
		if target.major == 0 {
			return compare(version, target) >= 0 && version.major == 0 && version.minor == target.minor
		}
		return compare(version, target) >= 0 && version.major == target.major
	case "~":
		if componentCount <= 2 {
			return compare(version, target) >= 0 && version.major == target.major
		}
		return compare(version, target) >= 0 && version.major == target.major && version.minor == target.minor
	default:
		if strings.HasSuffix(value, ".*") || len(strings.Split(value, ".")) < 3 {
			return version.major == target.major && version.minor == target.minor
		}
		return compare(version, target) == 0
	}
}

func (p SetupPlan) Validate() error {
	if p.Facts.Path == "" || p.Facts.Slug == "" {
		return fmt.Errorf("setup plan is missing project identity")
	}
	if p.Facts.Type != TypeReverseProxy && p.Choices.PHPVersion == "" {
		return fmt.Errorf("setup plan is missing PHP version")
	}
	if p.Choices.Domain == "" {
		return fmt.Errorf("setup plan is missing primary domain")
	}
	if len(p.Blockers) > 0 {
		return fmt.Errorf("setup plan has unresolved blocker: %s", p.Blockers[0].Message)
	}
	return nil
}

// BuildSetupPlan creates a plan from already detected facts and explicit
// choices. It performs no filesystem or process mutation.
func BuildSetupPlan(facts ProjectFacts, choices SetupChoices, phpChoices []RuntimeChoice) SetupPlan {
	plan := SetupPlan{Facts: facts, Choices: choices, PHPChoices: phpChoices}
	plan.Changes = append(plan.Changes,
		PlannedChange{Target: "project config", Action: "write project intent"},
		PlannedChange{Target: "nginx", Action: "generate and enable site route"},
	)
	if choices.Domain != "" {
		plan.Changes = append(plan.Changes, PlannedChange{Target: "domains", Action: "route primary domain " + choices.Domain})
	}
	if len(choices.Aliases) > 0 {
		plan.Changes = append(plan.Changes, PlannedChange{Target: "domains", Action: "route aliases: " + strings.Join(choices.Aliases, ", ")})
	}
	if facts.Type != TypeReverseProxy && choices.PHPVersion != "" {
		plan.Changes = append(plan.Changes, PlannedChange{Target: "PHP runtime", Action: "prepare PHP " + choices.PHPVersion + " FPM runtime"})
	}
	if choices.SSL {
		plan.Changes = append(plan.Changes, PlannedChange{Target: "SSL", Action: "generate trusted certificate when mkcert is ready"})
	}
	if choices.Dedicated {
		plan.Changes = append(plan.Changes, PlannedChange{Target: "PHP-FPM", Action: "configure dedicated project runtime"})
	}
	return plan
}

func projectEvidence(path string, projectType ProjectType) []string {
	evidence := []string{"directory inspected without executing project code"}
	switch projectType {
	case TypeLaravel:
		evidence = append(evidence, filepath.Join(path, "artisan")+" identifies Laravel")
	case TypeWordPress:
		evidence = append(evidence, "WordPress manifest files identify WordPress")
	case TypeReverseProxy:
		evidence = append(evidence, "package manager or proxy hints identify a reverse-proxy project")
	default:
		evidence = append(evidence, "no framework marker found; using generic PHP detection")
	}
	return evidence
}
