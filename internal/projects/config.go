package projects

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProjectType represents the detected project type.
type ProjectType string

const (
	TypeLaravel      ProjectType = "laravel"
	TypeWordPress    ProjectType = "wordpress"
	TypePHP          ProjectType = "php"
	TypeReverseProxy ProjectType = "reverse-proxy"
	TypeUnknown      ProjectType = ""

	// TypeGeneric is retained for project configs written before explicit PHP
	// project types were introduced.
	TypeGeneric ProjectType = "generic"
)

// FPMConfig describes the PHP-FPM pool configuration for a project.
type FPMConfig struct {
	Dedicated bool
	Socket    string
}

// Project represents a registered Chauffeur project.
type Project struct {
	Slug              string
	Path              string
	Domain            string
	Aliases           []string
	PHPVersion        string
	SSL               bool
	FPM               FPMConfig
	ProjectType       ProjectType
	NodeMode          string
	NodeVersion       string
	ProxyPort         int
	Database          string
	DatabaseContainer string
	Services          []string
	CreatedAt         string
	UpdatedAt         string
}

// AllDomains returns the primary domain followed by all aliases.
func (p *Project) AllDomains() []string {
	all := []string{p.Domain}
	return append(all, p.Aliases...)
}

// ConfigDir returns the directory where this project's config is stored.
func (p *Project) ConfigDir(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, "projects", p.Slug)
}

// ConfigPath returns the full path to the project's config.yaml.
func (p *Project) ConfigPath(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, "projects", p.Slug, "config.yaml")
}

// FPMSocketPath returns the unix socket path for this project's PHP-FPM pool.
func (p *Project) FPMSocketPath(workspaceRoot string) string {
	if p.FPM.Dedicated && p.FPM.Socket != "" {
		return p.FPM.Socket
	}
	return filepath.Join(workspaceRoot, "php", p.PHPVersion, "runtime", "php-fpm", "php-fpm.sock")
}

// Save writes the project config to disk, updating UpdatedAt.
func Save(p *Project, workspaceRoot string) error {
	dir := p.ConfigDir(workspaceRoot)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create project dir: %w", err)
	}
	p.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return os.WriteFile(p.ConfigPath(workspaceRoot), []byte(marshalProject(p)), 0644)
}

// Load reads a project config from disk.
func Load(workspaceRoot, slug string) (*Project, error) {
	path := filepath.Join(workspaceRoot, "projects", slug, "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read project config %q: %w", slug, err)
	}
	return unmarshalProject(string(data))
}

// Delete removes a project's config directory.
func Delete(workspaceRoot, slug string) error {
	return os.RemoveAll(filepath.Join(workspaceRoot, "projects", slug))
}

// ListSlugs returns all project slugs found in the workspace.
func ListSlugs(workspaceRoot string) ([]string, error) {
	projectsDir := filepath.Join(workspaceRoot, "projects")
	entries, err := os.ReadDir(projectsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var slugs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		cfgPath := filepath.Join(projectsDir, e.Name(), "config.yaml")
		if _, err := os.Stat(cfgPath); err == nil {
			slugs = append(slugs, e.Name())
		}
	}
	return slugs, nil
}

// ListAll loads every project in the workspace. Errors on individual projects
// are skipped (corrupt config) so the command still shows remaining projects.
func ListAll(workspaceRoot string) ([]*Project, error) {
	slugs, err := ListSlugs(workspaceRoot)
	if err != nil {
		return nil, err
	}
	var projects []*Project
	for _, slug := range slugs {
		p, err := Load(workspaceRoot, slug)
		if err != nil {
			continue
		}
		projects = append(projects, p)
	}
	return projects, nil
}

// FindByPath returns the project whose Path matches dirPath, or nil.
func FindByPath(workspaceRoot, dirPath string) (*Project, error) {
	projects, err := ListAll(workspaceRoot)
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		if p.Path == dirPath {
			return p, nil
		}
	}
	return nil, nil
}

// IsDomainInUse checks whether domain is already registered to any project.
// Returns the slug of the conflicting project, or "" if not in use.
// skipSlug allows a project to check against its own domains (during update).
func IsDomainInUse(workspaceRoot, domain, skipSlug string) (string, error) {
	projects, err := ListAll(workspaceRoot)
	if err != nil {
		return "", err
	}
	for _, p := range projects {
		if p.Slug == skipSlug {
			continue
		}
		for _, d := range p.AllDomains() {
			if d == domain {
				return p.Slug, nil
			}
		}
	}
	return "", nil
}

// ── YAML serialization ────────────────────────────────────────────────────────

func marshalProject(p *Project) string {
	var sb strings.Builder
	line := func(s string) { sb.WriteString(s + "\n") }

	line("slug: " + p.Slug)
	line("path: " + p.Path)
	line("domain: " + p.Domain)
	if len(p.Aliases) > 0 {
		line("aliases:")
		for _, a := range p.Aliases {
			line("  - " + a)
		}
	} else {
		line("aliases: []")
	}
	line(`php_version: "` + p.PHPVersion + `"`)
	if p.SSL {
		line("ssl: true")
	} else {
		line("ssl: false")
	}
	line("fpm:")
	if p.FPM.Dedicated {
		line("  dedicated: true")
	} else {
		line("  dedicated: false")
	}
	line(`  socket: "` + p.FPM.Socket + `"`)
	line("project_type: " + string(p.ProjectType))
	if p.NodeMode != "" {
		line("node:")
		line("  mode: " + p.NodeMode)
		line("  version: \"" + p.NodeVersion + "\"")
	}
	if p.ProxyPort > 0 {
		line(fmt.Sprintf("proxy_port: %d", p.ProxyPort))
	}
	if p.Database != "" {
		line("database: " + p.Database)
	}
	if p.DatabaseContainer != "" {
		line("database_container: " + p.DatabaseContainer)
	}
	if len(p.Services) > 0 {
		line("services:")
		for _, service := range p.Services {
			line("  - " + service)
		}
	}
	line(`created_at: "` + p.CreatedAt + `"`)
	line(`updated_at: "` + p.UpdatedAt + `"`)
	return sb.String()
}

func unmarshalProject(data string) (*Project, error) {
	p := &Project{}
	inFPM := false
	inAliases := false
	inNode := false
	inServices := false

	for _, raw := range strings.Split(data, "\n") {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		isIndented := strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t")

		// Alias list items
		if inAliases && isIndented && strings.HasPrefix(trimmed, "- ") {
			p.Aliases = append(p.Aliases, strings.TrimPrefix(trimmed, "- "))
			continue
		}
		if inServices && isIndented && strings.HasPrefix(trimmed, "- ") {
			p.Services = append(p.Services, strings.TrimPrefix(trimmed, "- "))
			continue
		}

		// FPM sub-keys
		if inFPM && isIndented {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
				switch key {
				case "dedicated":
					p.FPM.Dedicated = val == "true"
				case "socket":
					p.FPM.Socket = val
				}
			}
			continue
		}
		if inNode && isIndented {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
				switch key {
				case "mode":
					p.NodeMode = val
				case "version":
					p.NodeVersion = val
				}
			}
			continue
		}

		// A non-indented line resets section tracking
		if !isIndented {
			inFPM = false
			inAliases = false
			inNode = false
			inServices = false
		}

		// Key-value
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)

		switch key {
		case "slug":
			p.Slug = val
		case "path":
			p.Path = val
		case "domain":
			p.Domain = val
		case "aliases":
			// "aliases: []" means empty; "aliases:" means list follows
			if val == "" {
				inAliases = true
			}
			// val == "[]" → stay with nil Aliases
		case "php_version":
			p.PHPVersion = val
		case "ssl":
			p.SSL = val == "true"
		case "fpm":
			inFPM = true
		case "node":
			inNode = true
		case "proxy_port":
			if _, err := fmt.Sscanf(val, "%d", &p.ProxyPort); err != nil {
				return nil, fmt.Errorf("invalid proxy_port %q", val)
			}
		case "database":
			p.Database = val
		case "database_container":
			p.DatabaseContainer = val
		case "services":
			if val == "" {
				inServices = true
			}
		case "project_type":
			p.ProjectType = ProjectType(val)
		case "created_at":
			p.CreatedAt = val
		case "updated_at":
			p.UpdatedAt = val
		}
	}

	if p.Slug == "" {
		return nil, fmt.Errorf("missing slug in project config")
	}
	return p, nil
}
