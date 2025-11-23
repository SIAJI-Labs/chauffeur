package example

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/siaji/chauffeur/cli/internal/config"
	"github.com/siaji/chauffeur/cli/internal/projects"
	"github.com/siaji/chauffeur/cli/internal/templates"
	"github.com/siaji/chauffeur/cli/lib"
)

// ExampleProjectName is the fixed name for the example project
const ExampleProjectName = "example-project"

// ExampleProjectDomain is the domain assigned to the example project
const ExampleProjectDomain = "example-project.test"

// CreateExampleProject creates and links an example project in the workspace
func CreateExampleProject() error {
	logger := lib.NewCommandLogger("example-project")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Create example project directory
	exampleDir := filepath.Join(cfg.ProjectsDir, ExampleProjectName)
	if _, err := os.Stat(exampleDir); err == nil {
		logger.Info(fmt.Sprintf("Example project already exists at: %s", exampleDir))
		return nil
	}

	if err := os.MkdirAll(exampleDir, 0755); err != nil {
		return fmt.Errorf("create example project directory: %w", err)
	}

	// Create index.php with welcome content
	indexContent := generateIndexContent()
	if err := os.WriteFile(filepath.Join(exampleDir, "index.php"), []byte(indexContent), 0644); err != nil {
		return fmt.Errorf("create index.php: %w", err)
	}

	// Create .gitignore
	gitignoreContent := `# Example project .gitignore
# This is an example project managed by Chauffeur
# Feel free to modify or remove this file
`
	if err := os.WriteFile(filepath.Join(exampleDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("create .gitignore: %w", err)
	}

	logger.Success(fmt.Sprintf("Example project created at: %s", exampleDir), "")

	// Check if PHP is installed
	defaultPHP := cfg.PHP.Default
	if defaultPHP == "" {
		defaultPHP = "8.3"
	}

	if !projects.IsPHPVersionInstalled(defaultPHP) {
		logger.Info("PHP not installed yet. Example project will be linked after PHP installation.")
		return nil
	}

	// Link the example project
	return linkExampleProject(cfg, logger, defaultPHP)
}

// LinkExampleProjectIfReady links the example project if both nginx and php are installed
func LinkExampleProjectIfReady() error {
	logger := lib.NewCommandLogger("example-project")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Check if example project already exists
	exampleDir := filepath.Join(cfg.ProjectsDir, ExampleProjectName)
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		// Create the project if it doesn't exist
		if err := CreateExampleProject(); err != nil {
			return err
		}
	}

	// Check if already linked
	projectPath := filepath.Join(cfg.ProjectsDir, ExampleProjectName)
	if _, _, err := projects.FindByPath(cfg.ProjectsDir, projectPath); err == nil {
		// Already linked
		logger.Info(fmt.Sprintf("Example project already linked at: %s", ExampleProjectDomain))
		return nil
	}

	// Check if both nginx and php are installed
	defaultPHP := cfg.PHP.Default
	if defaultPHP == "" {
		defaultPHP = "8.3"
	}

	if !projects.IsPHPVersionInstalled(defaultPHP) {
		logger.Info("PHP not installed yet. Example project will be linked after PHP installation.")
		return nil
	}

	// Check nginx installation
	nginxBinPath := filepath.Join(cfg.WorkspaceDir, "nginx", "bin", "nginx")
	if _, err := os.Stat(nginxBinPath); os.IsNotExist(err) {
		logger.Info("Nginx not installed yet. Example project will be linked after nginx installation.")
		return nil
	}

	// Both services are installed, link the project
	return linkExampleProject(cfg, logger, defaultPHP)
}

// linkExampleProject links the example project using the link command logic
func linkExampleProject(cfg config.Config, logger *lib.Logger, phpVersion string) error {
	// Check if already linked
	projectPath := filepath.Join(cfg.ProjectsDir, ExampleProjectName)
	if _, _, err := projects.FindByPath(cfg.ProjectsDir, projectPath); err == nil {
		// Already linked
		logger.Info(fmt.Sprintf("Example project already linked at: %s", ExampleProjectDomain))
		return nil
	}

	// Change to example project directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(projectPath); err != nil {
		return fmt.Errorf("change to example project directory: %w", err)
	}

	// Create project layout
	slug := projects.Slugify(ExampleProjectName)
	layout, err := projects.EnsureLayout(cfg.ProjectsDir, slug)
	if err != nil {
		return err
	}

	// Determine socket path (shared FPM for efficiency)
	phpVersionDir := filepath.Join(cfg.WorkspaceDir, "php", phpVersion)
	runtimeDir := filepath.Join(phpVersionDir, "runtime", "php-fpm")
	socketPath := filepath.Join(runtimeDir, "php-fpm.sock")

	// Create FPM configuration (shared by default)
	fpmConfig := &projects.FPM{
		Dedicated: false,
		Socket:    socketPath,
	}

	// Create project configuration
	proj := projects.Config{
		Version: projects.ConfigVersion,
		Path:    projectPath,
		PHP:     phpVersion,
		Runtime: projects.Runtime{
			PHPFPM: socketPath,
			FPM:    fpmConfig,
		},
		CreatedAt: time.Now().UTC(),
	}

	proj.Site = &projects.Site{
		Domain: ExampleProjectDomain,
		SSL:    false,
	}

	// Write project configuration
	if err := projects.WriteConfig(proj, layout.ConfigPath, false); err != nil {
		return fmt.Errorf("write project config: %w", err)
	}

	// Generate nginx configuration
	templateEngine, err := templates.NewTemplateEngine()
	if err != nil {
		return fmt.Errorf("create template engine: %w", err)
	}

	nginxOptions := templates.NginxConfigOptions{
		HTTPPort:  cfg.Nginx.HTTPPort,
		HTTPSPort: cfg.Nginx.HTTPSPort,
	}

	// Generate and write nginx configuration
	if err := templateEngine.WriteNginxConfig(proj, layout, "general", nginxOptions); err != nil {
		return fmt.Errorf("generate nginx config: %w", err)
	}

	// Print success message with access instructions
	logger.Success(fmt.Sprintf("Example project linked successfully at: %s", ExampleProjectDomain), "")
	httpURL := fmt.Sprintf("http://%s", ExampleProjectDomain)
	if cfg.Nginx.HTTPPort != 80 {
		httpURL = fmt.Sprintf("http://%s:%d", ExampleProjectDomain, cfg.Nginx.HTTPPort)
	}
	logger.Info(fmt.Sprintf("You can now access the example project at: %s", httpURL))
	logger.Info("The example project demonstrates:")
	logger.Info("  - PHP functionality with phpinfo()")
	logger.Info("  - Chauffeur's automatic .test domain resolution")
	logger.Info("  - nginx configuration")
	logger.Info("")
	logger.Info(fmt.Sprintf("To remove the example project when ready: cd %s && chauf unlink", projectPath))

	return nil
}

// generateIndexContent creates the content for index.php
func generateIndexContent() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to Chauffeur - Example Project</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f8f9fa;
        }
        
        .container {
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
        }
        
        header {
            text-align: center;
            margin-bottom: 3rem;
            padding: 2rem;
            background: white;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        
        h1 {
            color: #2c3e50;
            font-size: 2.5rem;
            margin-bottom: 0.5rem;
        }
        
        .subtitle {
            color: #7f8c8d;
            font-size: 1.2rem;
        }
        
        .card {
            background: white;
            border-radius: 10px;
            padding: 2rem;
            margin-bottom: 2rem;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        
        .card h2 {
            color: #2c3e50;
            margin-bottom: 1rem;
        }
        
        .links {
            list-style: none;
        }
        
        .links li {
            margin-bottom: 0.5rem;
        }
        
        .links a {
            color: #3498db;
            text-decoration: none;
            padding: 0.5rem 1rem;
            border: 1px solid #3498db;
            border-radius: 5px;
            display: inline-block;
            transition: all 0.3s ease;
        }
        
        .links a:hover {
            background: #3498db;
            color: white;
        }
        
        .phpinfo-section {
            margin-top: 3rem;
            background: white;
            border-radius: 10px;
            padding: 2rem;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        
        .toggle-btn {
            background: #3498db;
            color: white;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: 5px;
            cursor: pointer;
            font-size: 1rem;
            transition: background 0.3s ease;
        }
        
        .toggle-btn:hover {
            background: #2980b9;
        }
        
        .phpinfo-content {
            margin-top: 1rem;
            max-height: 0;
            overflow: hidden;
            transition: max-height 0.5s ease;
        }
        
        .phpinfo-content.show {
            max-height: 2000px;
        }
        
        .footer {
            text-align: center;
            margin-top: 3rem;
            padding: 2rem;
            background: white;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🚗 Welcome to Chauffeur!</h1>
            <p class="subtitle">Your Linux PHP development environment is ready</p>
        </header>
        
        <section class="card">
            <h2>What is Chauffeur?</h2>
            <p>Chauffeur is a Linux CLI that provides per-project PHP development services (nginx, PHP-FPM, dnsmasq-managed .test domains) without touching system prefixes. Everything lives under <code>~/.chauffeur/</code> so multiple projects can coexist safely.</p>
        </section>
        
        <section class="card">
            <h2>Quick Links</h2>
            <ul class="links">
                <li><a href="https://chauffeur.siaji.com" target="_blank">📖 Documentation</a></li>
                <li><a href="https://github.com/SIAJI-Labs/chauffeur" target="_blank">💻 GitHub Repository</a></li>
                <li><a href="https://github.com/SIAJI-Labs/chauffeur/issues" target="_blank">🐛 Report Issues</a></li>
            </ul>
        </section>
        
        <section class="card">
            <h2>Next Steps</h2>
            <ol>
                <li>Create your own project: <code>cd /path/to/your/project && chauf link</code></li>
                <li>Start services: <code>chauf start</code></li>
                <li>Access your project at <code>your-project-name.test</code></li>
                <li>Unlink this example when ready: <code>cd ~/.chauffeur/projects/example-project && chauf unlink</code></li>
            </ol>
        </section>
        
        <section class="phpinfo-section">
            <h2>PHP Information</h2>
            <p>Click the button below to view your PHP configuration:</p>
            <button class="toggle-btn" onclick="togglePhpInfo()">Show/Hide phpinfo()</button>
            <div id="phpinfo-content" class="phpinfo-content">
                <?php phpinfo(); ?>
            </div>
        </section>
        
        <footer class="footer">
            <p>Generated by Chauffeur on ` + time.Now().Format("2006-01-02 at 15:04:05") + `</p>
            <p>This example project is located at: <code><?php echo __DIR__; ?></code></p>
        </footer>
    </div>
    
    <script>
        function togglePhpInfo() {
            const content = document.getElementById('phpinfo-content');
            content.classList.toggle('show');
        }
    </script>
</body>
</html>
`
}

// RemoveExampleProject removes the example project if it exists
func RemoveExampleProject() error {
	logger := lib.NewCommandLogger("example-project")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Check if example project is linked
	projectPath := filepath.Join(cfg.ProjectsDir, ExampleProjectName)
	_, layout, err := projects.FindByPath(cfg.ProjectsDir, projectPath)
	if err != nil {
		// Project not linked, just remove the directory
		if _, err := os.Stat(projectPath); err == nil {
			if err := os.RemoveAll(projectPath); err != nil {
				return logger.Error("Failed to remove example project directory", err.Error())
			}
			logger.Success("Example project directory removed", projectPath)
		}
		return nil
	}

	// Remove nginx configuration
	templateEngine, err := templates.NewTemplateEngine()
	if err != nil {
		return logger.Error("Failed to create template engine", err.Error())
	}

	projectSlug := projects.Slugify(ExampleProjectName)
	if err := templateEngine.RemoveNginxConfig(projectSlug); err != nil {
		logger.Warn("Failed to remove nginx configuration", err.Error())
	}

	// Remove project configuration
	if err := os.Remove(layout.ConfigPath); err != nil {
		return logger.Error("Failed to remove project configuration", err.Error())
	}

	// Remove project directory
	if err := os.RemoveAll(projectPath); err != nil {
		return logger.Error("Failed to remove example project directory", err.Error())
	}

	logger.Success("Example project removed successfully", ExampleProjectDomain)
	return nil
}

// IsExampleProjectExists checks if the example project exists
func IsExampleProjectExists() bool {
	cfg, err := config.Load()
	if err != nil {
		return false
	}

	exampleDir := filepath.Join(cfg.ProjectsDir, ExampleProjectName)
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		return false
	}

	return true
}

// IsExampleProjectLinked checks if the example project is linked
func IsExampleProjectLinked() bool {
	cfg, err := config.Load()
	if err != nil {
		return false
	}

	projectPath := filepath.Join(cfg.ProjectsDir, ExampleProjectName)
	if _, _, err := projects.FindByPath(cfg.ProjectsDir, projectPath); err != nil {
		return false
	}

	return true
}
