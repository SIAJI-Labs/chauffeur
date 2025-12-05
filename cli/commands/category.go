package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/lib"
)

// RunCategory handles `chauf category` command invocations.
func RunCategory(args []string) error {
	logger := lib.NewCommandLogger("category")

	if len(args) == 0 {
		printCategoryUsage()
		return nil
	}

	command := args[0]
	remainingArgs := args[1:]

	switch command {
	case "list", "ls":
		return runCategoryList(logger, remainingArgs)
	case "rename", "mv":
		return runCategoryRename(logger, remainingArgs)
	case "delete", "del", "rm":
		return runCategoryDelete(logger, remainingArgs)
	case "--help", "-h":
		printCategoryUsage()
		return nil
	default:
		return fmt.Errorf("unknown category command: %s", command)
	}
}

// runCategoryList lists all categories and their project counts.
func runCategoryList(logger *lib.Logger, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	projectsDir := cfg.ProjectsDir
	if projectsDir == "" {
		// Fallback to default projects directory
		ws, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("determine home directory: %w", err)
		}
		projectsDir = filepath.Join(ws, ".chauffeur", "projects")
	}

	allProjects, err := listAllProjects(projectsDir)
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	if len(allProjects) == 0 {
		logger.Info("No projects found.")
		return nil
	}

	// Group projects by category
	categories := make(map[string][]linkedProject)
	totalProjects := 0

	for _, project := range allProjects {
		categories[project.Category] = append(categories[project.Category], project)
		totalProjects++
	}

	logger.PrintSection(fmt.Sprintf("Categories (%d projects total)", totalProjects))

	// Sort categories (put "Uncategorized" last, then alphabetically)
	sortedCategories := make([]string, 0, len(categories))
	for category := range categories {
		sortedCategories = append(sortedCategories, category)
	}

	// Simple bubble sort for category ordering
	for i := 0; i < len(sortedCategories); i++ {
		for j := i + 1; j < len(sortedCategories); j++ {
			// Put "Uncategorized" last
			if sortedCategories[i] == "Uncategorized" {
				sortedCategories[i], sortedCategories[j] = sortedCategories[j], sortedCategories[i]
			}
			// Then sort alphabetically (case-insensitive) for non-Uncategorized categories
			if sortedCategories[i] != "Uncategorized" && sortedCategories[j] != "Uncategorized" {
				if strings.ToLower(sortedCategories[i]) > strings.ToLower(sortedCategories[j]) {
					sortedCategories[i], sortedCategories[j] = sortedCategories[j], sortedCategories[i]
				}
			}
		}
	}

	maxCategoryName := 10 // Minimum width for "CATEGORY"
	for _, category := range sortedCategories {
		if len(category) > maxCategoryName {
			maxCategoryName = len(category)
		}
	}

	// Print header
	header := fmt.Sprintf("%-*s  %s", maxCategoryName, "CATEGORY", "PROJECTS")
	divider := fmt.Sprintf("%s  %s", strings.Repeat("-", maxCategoryName), strings.Repeat("-", 8))
	logger.Info(header)
	logger.Info(divider)

	// Print categories
	for _, category := range sortedCategories {
		projects := categories[category]
		count := len(projects)
		categoryDisplay := category

		// Add icons for special categories
		if category == "Uncategorized" {
			categoryDisplay = category + " 📋"
		}

		row := fmt.Sprintf("%-*s  %d", maxCategoryName, categoryDisplay, count)
		logger.Info(row)
	}

	logger.Info("")
	logger.Info(fmt.Sprintf("Total: %d categories, %d projects", len(categories), totalProjects))

	return nil
}

// runCategoryRename renames a category and moves all projects to the new category.
func runCategoryRename(logger *lib.Logger, args []string) error {
	var fromCategory, toCategory string

	for i, arg := range args {
		switch arg {
		case "--from":
			if i+1 >= len(args) {
				return fmt.Errorf("--from requires a category name")
			}
			fromCategory = args[i+1]
		case "--to":
			if i+1 >= len(args) {
				return fmt.Errorf("--to requires a category name")
			}
			toCategory = args[i+1]
		case "--help", "-h":
			printCategoryRenameUsage()
			return nil
		}
	}

	if fromCategory == "" {
		return fmt.Errorf("--from is required (source category name)")
	}
	if toCategory == "" {
		return fmt.Errorf("--to is required (destination category name)")
	}

	// Validate category names
	tempConfig := &projects.Config{}
	if err := tempConfig.SetCategory(fromCategory); err != nil {
		return fmt.Errorf("invalid source category name: %w", err)
	}
	if err := tempConfig.SetCategory(toCategory); err != nil {
		return fmt.Errorf("invalid destination category name: %w", err)
	}

	// Check if trying to rename a protected category
	if projects.IsProtectedCategory(fromCategory) {
		return fmt.Errorf("cannot rename protected category '%s'", fromCategory)
	}

	if fromCategory == toCategory {
		logger.Info("Source and destination categories are the same, nothing to do.")
		return nil
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	projectsDir := cfg.ProjectsDir
	if projectsDir == "" {
		ws, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("determine home directory: %w", err)
		}
		projectsDir = filepath.Join(ws, ".chauffeur", "projects")
	}

	// Get all projects
	allProjects, err := listAllProjects(projectsDir)
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	// Find projects in the source category
	var projectsToMove []linkedProject
	for _, project := range allProjects {
		if strings.EqualFold(project.Category, fromCategory) {
			projectsToMove = append(projectsToMove, project)
		}
	}

	if len(projectsToMove) == 0 {
		logger.Info(fmt.Sprintf("No projects found in category '%s'.", fromCategory))
		return nil
	}

	logger.Info(fmt.Sprintf("Found %d project(s) in category '%s':", len(projectsToMove), fromCategory))
	for _, project := range projectsToMove {
		logger.Info(fmt.Sprintf("  - %s", project.Slug))
	}

	// Update each project's category
	logger.PrintSection(fmt.Sprintf("Moving projects from '%s' to '%s'", fromCategory, toCategory))

	for _, project := range projectsToMove {
		layout, err := projects.EnsureLayout(projectsDir, project.Slug)
		if err != nil {
			logger.Warn("skip project", fmt.Sprintf("failed to ensure layout for %s: %v", project.Slug, err))
			continue
		}

		projConfig, err := projects.LoadConfig(layout.ConfigPath)
		if err != nil {
			logger.Warn("skip project", fmt.Sprintf("failed to load config for %s: %v", project.Slug, err))
			continue
		}

		// Update the category
		if err := projConfig.SetCategory(toCategory); err != nil {
			logger.Warn("skip project", fmt.Sprintf("failed to set category for %s: %v", project.Slug, err))
			continue
		}

		// Save the updated configuration (force=true since we're updating)
		if err := projects.WriteConfig(projConfig, layout.ConfigPath, true); err != nil {
			logger.Warn("skip project", fmt.Sprintf("failed to save config for %s: %v", project.Slug, err))
			continue
		}

		logger.Success("project moved", fmt.Sprintf("'%s' → '%s'", project.Slug, toCategory))
	}

	logger.Success("category rename complete", fmt.Sprintf("moved %d projects from '%s' to '%s'", len(projectsToMove), fromCategory, toCategory))
	return nil
}

// runCategoryDelete deletes a category and moves all its projects to Uncategorized (or specified destination).
func runCategoryDelete(logger *lib.Logger, args []string) error {
	var category, moveTo string

	for i, arg := range args {
		switch arg {
		case "--name":
			if i+1 >= len(args) {
				return fmt.Errorf("--name requires a category name")
			}
			category = args[i+1]
		case "--move-to":
			if i+1 >= len(args) {
				return fmt.Errorf("--move-to requires a category name")
			}
			moveTo = args[i+1]
		case "--help", "-h":
			printCategoryDeleteUsage()
			return nil
		}
	}

	if category == "" {
		return fmt.Errorf("--name is required (category name to delete)")
	}

	// Default destination is Uncategorized
	if moveTo == "" {
		moveTo = "Uncategorized"
	}

	// Validate category names
	tempConfig := &projects.Config{}
	if err := tempConfig.SetCategory(category); err != nil {
		return fmt.Errorf("invalid category name: %w", err)
	}
	if err := tempConfig.SetCategory(moveTo); err != nil {
		return fmt.Errorf("invalid destination category name: %w", err)
	}

	// Check if trying to delete a protected category
	if projects.IsProtectedCategory(category) {
		return fmt.Errorf("cannot delete protected category '%s'", category)
	}

	if strings.EqualFold(category, moveTo) {
		return fmt.Errorf("cannot delete category '%s' and move projects to the same category", category)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	projectsDir := cfg.ProjectsDir
	if projectsDir == "" {
		ws, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("determine home directory: %w", err)
		}
		projectsDir = filepath.Join(ws, ".chauffeur", "projects")
	}

	// Get all projects
	allProjects, err := listAllProjects(projectsDir)
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	// Find projects in the category to delete
	var projectsToMove []linkedProject
	for _, project := range allProjects {
		if strings.EqualFold(project.Category, category) {
			projectsToMove = append(projectsToMove, project)
		}
	}

	if len(projectsToMove) == 0 {
		logger.Info(fmt.Sprintf("No projects found in category '%s'.", category))
		return nil
	}

	logger.Info(fmt.Sprintf("Found %d project(s) in category '%s':", len(projectsToMove), category))
	for _, project := range projectsToMove {
		logger.Info(fmt.Sprintf("  - %s", project.Slug))
	}

	// Update each project's category
	logger.PrintSection(fmt.Sprintf("Moving projects from '%s' to '%s'", category, moveTo))

	for _, project := range projectsToMove {
		layout, err := projects.EnsureLayout(projectsDir, project.Slug)
		if err != nil {
			logger.Warn("skip project", fmt.Sprintf("failed to ensure layout for %s: %v", project.Slug, err))
			continue
		}

		projConfig, err := projects.LoadConfig(layout.ConfigPath)
		if err != nil {
			logger.Warn("skip project", fmt.Sprintf("failed to load config for %s: %v", project.Slug, err))
			continue
		}

		// Update the category
		if err := projConfig.SetCategory(moveTo); err != nil {
			logger.Warn("skip project", fmt.Sprintf("failed to set category for %s: %v", project.Slug, err))
			continue
		}

		// Save the updated configuration (force=true since we're updating)
		if err := projects.WriteConfig(projConfig, layout.ConfigPath, true); err != nil {
			logger.Warn("skip project", fmt.Sprintf("failed to save config for %s: %v", project.Slug, err))
			continue
		}

		logger.Success("project moved", fmt.Sprintf("'%s' → '%s'", project.Slug, moveTo))
	}

	logger.Success("category delete complete", fmt.Sprintf("moved %d projects from deleted category '%s' to '%s'", len(projectsToMove), category, moveTo))
	return nil
}

func printCategoryUsage() {
	logger := lib.NewCommandLogger("category")
	logger.PrintBlock(`Chauffeur Category Management

Usage:
  chauf category list                       List all categories with project counts.
  chauf category rename --from <old> --to <new>    Rename a category and move all projects.
  chauf category delete --name <category> [--move-to <dest>]   Delete a category and move projects.

Category Operations:
  • Categories organize your projects in 'chauf links'
  • Protected categories (like 'Uncategorized') cannot be renamed or deleted
  • When deleting categories, projects move to 'Uncategorized' by default

Examples:
  chauf category list
  chauf category rename --from "Work" --to "Client Projects"
  chauf category delete --name "Old Category"
  chauf category delete --name "Temp" --move-to "Archive"

Flags:
  --from string       Source category name (for rename)
  --to string         Destination category name (for rename)
  --name string       Category name to delete (for delete)
  --move-to string    Destination category for moved projects (default: Uncategorized)
  --help, -h          Show this help message.
`)
}

func printCategoryRenameUsage() {
	logger := lib.NewCommandLogger("category rename")
	logger.PrintBlock(`Chauffeur Category Rename

Usage:
  chauf category rename --from <old-category> --to <new-category>

Description:
  Renames a category by moving all projects from the old category to the new one.
  The old category will be empty after the operation.

Flags:
  --from string       Source category name (required)
  --to string         Destination category name (required)
  --help, -h          Show this help message.

Examples:
  chauf category rename --from "Work" --to "Client Projects"
  chauf category rename --from "Personal" --to "Hobby Projects"
`)
}

func printCategoryDeleteUsage() {
	logger := lib.NewCommandLogger("category delete")
	logger.PrintBlock(`Chauffeur Category Delete

Usage:
  chauf category delete --name <category> [--move-to <destination>]

Description:
  Deletes a category by moving all its projects to another category.
  If no destination is specified, projects are moved to 'Uncategorized'.

Flags:
  --name string       Category name to delete (required)
  --move-to string    Destination category for moved projects (default: Uncategorized)
  --help, -h          Show this help message.

Examples:
  chauf category delete --name "Old Category"
  chauf category delete --name "Temp" --move-to "Archive"
  chauf category delete --name "Client: Old" --move-to "Client: Current"
`)
}
