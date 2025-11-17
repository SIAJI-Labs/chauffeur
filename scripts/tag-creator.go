package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// Global flags
var dryRun = flag.Bool("dry-run", false, "Run in dry-run mode (no actual Git operations)")
var showHelp = flag.Bool("help", false, "Show help message")
var showVersion = flag.Bool("version", false, "Show version information")

// Git operations interface for easier testing
type GitOperations interface {
	FetchTags() error
	GetLatestTag() (string, error)
	GetCurrentBranch() (string, error)
	HasUncommittedChanges() (bool, error)
	CreateTag(tag string) error
	PushTag(tag string) error
	SwitchBranch(branch string) error
	PullCurrentBranch() error
}

// RealGitOps implements actual Git operations
type RealGitOps struct{}

func (g *RealGitOps) FetchTags() error {
	cmd := exec.Command("git", "fetch", "--tags")
	if *dryRun {
		fmt.Printf("🔍 DRY RUN: Would execute: git fetch --tags\n")
		return nil
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *RealGitOps) GetLatestTag() (string, error) {
	cmd := exec.Command("git", "tag", "-l", "--sort=-v:refname")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	var validTags []string
	for _, tag := range tags {
		if tag != "" {
			validTags = append(validTags, tag)
		}
	}

	if len(validTags) == 0 {
		return "v0.0.0", nil
	}

	return validTags[0], nil
}

func (g *RealGitOps) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (g *RealGitOps) HasUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(output)) != "", nil
}

func (g *RealGitOps) CreateTag(tag string) error {
	if *dryRun {
		fmt.Printf("🏷️  DRY RUN: Would create tag: %s\n", tag)
		return nil
	}
	cmd := exec.Command("git", "tag", tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *RealGitOps) PushTag(tag string) error {
	if *dryRun {
		fmt.Printf("🚀 DRY RUN: Would push tag to remote: %s\n", tag)
		return nil
	}
	cmd := exec.Command("git", "push", "origin", tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *RealGitOps) SwitchBranch(branch string) error {
	if *dryRun {
		fmt.Printf("🌿 DRY RUN: Would switch to branch: %s\n", branch)
		return nil
	}
	cmd := exec.Command("git", "checkout", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *RealGitOps) PullCurrentBranch() error {
	if *dryRun {
		fmt.Printf("📥 DRY RUN: Would pull current branch\n")
		return nil
	}
	cmd := exec.Command("git", "pull", "origin", "main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Version utilities
func incrementVersion(version string, bumpType string) string {
	// Remove leading 'v' and split
	version = strings.TrimPrefix(version, "v")
	parts := strings.Split(version, ".")

	major, minor, patch := 0, 0, 0
	if len(parts) >= 1 {
		fmt.Sscanf(parts[0], "%d", &major)
	}
	if len(parts) >= 2 {
		fmt.Sscanf(parts[1], "%d", &minor)
	}
	if len(parts) >= 3 {
		fmt.Sscanf(parts[2], "%d", &patch)
	}

	switch bumpType {
	case "major":
		return fmt.Sprintf("v%d.0.0", major+1)
	case "minor":
		return fmt.Sprintf("v%d.%d.0", major, minor+1)
	case "patch":
		return fmt.Sprintf("v%d.%d.%d", major, minor, patch+1)
	default:
		return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
	}
}

func validateVersionTag(val interface{}) error {
	input, ok := val.(string)
	if !ok {
		return fmt.Errorf("input must be a string")
	}
	matched, err := regexp.MatchString(`^v\d+\.\d+\.\d+$`, input)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid format. Use vX.Y.Z (e.g., v1.10.15)")
	}
	return nil
}

// Main function
func createTag() {
	gitOps := &RealGitOps{}

	if *dryRun {
		fmt.Println("🏷️  Chauffeur Tag Creator (DRY RUN)")
		fmt.Println("⚠️  No actual Git operations will be performed")
	} else {
		fmt.Println("🏷️  Chauffeur Tag Creator")
	}
	fmt.Println()

	// Check for uncommitted changes
	hasChanges, err := gitOps.HasUncommittedChanges()
	if err != nil {
		fmt.Printf("❌ Error checking git status: %v\n", err)
		os.Exit(1)
	}

	if hasChanges {
		fmt.Println("❌ You have uncommitted changes. Please commit or stash them before creating a tag.")
		fmt.Println("📝 Current status:")
		cmd := exec.Command("git", "status", "--porcelain")
		cmd.Stdout = os.Stdout
		cmd.Run()
		os.Exit(1)
	}

	// Check current branch
	currentBranch, err := gitOps.GetCurrentBranch()
	if err != nil {
		fmt.Printf("❌ Error getting current branch: %v\n", err)
		os.Exit(1)
	}

	originalBranch := currentBranch

	// Switch to main if not already on main
	if currentBranch != "main" {
		fmt.Printf("Currently on %s, switching to main...\n", currentBranch)
		if err := gitOps.SwitchBranch("main"); err != nil {
			fmt.Printf("❌ Error switching to main: %v\n", err)
			os.Exit(1)
		}
	}

	// Sync with remote
	fmt.Println("Fetching latest changes...")
	if err := gitOps.FetchTags(); err != nil {
		fmt.Printf("❌ Error fetching tags: %v\n", err)
		os.Exit(1)
	}
	if err := gitOps.PullCurrentBranch(); err != nil {
		fmt.Printf("❌ Error pulling latest changes: %v\n", err)
		os.Exit(1)
	}

	// Get latest tag
	latestTag, err := gitOps.GetLatestTag()
	if err != nil {
		fmt.Printf("❌ Error getting latest tag: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Latest tag found: %s\n", latestTag)

	// Generate preview versions
	majorVersion := incrementVersion(latestTag, "major")
	minorVersion := incrementVersion(latestTag, "minor")
	patchVersion := incrementVersion(latestTag, "patch")

	// Interactive selection
	var bumpType string
	prompt := &survey.Select{
		Message: "Select version bump type:",
		Options: []string{
			fmt.Sprintf("major (%s)", majorVersion),
			fmt.Sprintf("minor (%s)", minorVersion),
			fmt.Sprintf("patch (%s)", patchVersion),
			"other (enter manually)",
		},
	}

	if err := survey.AskOne(prompt, &bumpType); err != nil {
		fmt.Printf("❌ Error during prompt: %v\n", err)
		os.Exit(1)
	}

	var newTag string

	// Extract the version part from the selection
	if strings.Contains(bumpType, "major") {
		newTag = majorVersion
	} else if strings.Contains(bumpType, "minor") {
		newTag = minorVersion
	} else if strings.Contains(bumpType, "patch") {
		newTag = patchVersion
	} else {
		// Manual input
		inputPrompt := &survey.Input{
			Message: "Enter custom version tag (e.g., v1.10.15):",
		}
		if err := survey.AskOne(inputPrompt, &newTag, survey.WithValidator(validateVersionTag)); err != nil {
			fmt.Printf("❌ Error during input: %v\n", err)
			os.Exit(1)
		}
	}

	// Create and push tag
	if err := gitOps.CreateTag(newTag); err != nil {
		fmt.Printf("❌ Error creating tag: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Tag created: %s\n", newTag)

	if err := gitOps.PushTag(newTag); err != nil {
		fmt.Printf("❌ Error pushing tag: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("🚀 Tag pushed to remote: %s\n", newTag)

	// Switch back to original branch if needed
	if originalBranch != "main" {
		if err := gitOps.SwitchBranch(originalBranch); err != nil {
			fmt.Printf("❌ Error switching back to %s: %v\n", originalBranch, err)
			os.Exit(1)
		}
		fmt.Printf("Switched back to %s branch.\n", originalBranch)
	}

	fmt.Println()
	if *dryRun {
		fmt.Println("✅ Dry run completed - no actual changes made")
	} else {
		fmt.Println("✅ Tag creation completed!")
	}
}

func printHelp() {
	fmt.Println(`Chauffeur Tag Creator

USAGE:
    tag-creator [FLAGS]

FLAGS:
    --dry-run    Run in dry-run mode (no actual Git operations)
    --help       Show this help message
    --version    Show version information

EXAMPLES:
    tag-creator              # Create a tag interactively
    tag-creator --dry-run    # Preview what would happen without making changes

DESCRIPTION:
    Creates Git tags with interactive version selection. Automatically switches
    to main branch, creates the tag, and pushes to remote repository.`)
}

func printVersion() {
	fmt.Println("tag-creator v1.0.0")
}

func main() {
	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}

	if *showVersion {
		printVersion()
		return
	}

	createTag()
}