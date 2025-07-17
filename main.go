package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/goccy/go-yaml"
)

type Package struct {
	Name         string `yaml:"name"`
	ImportPrefix string `yaml:"import-prefix"`
	Home         string `yaml:"home"`
	VCS          string `yaml:"vcs"`
	RepoRoot     string `yaml:"repo-root"`
}

//go:embed package.template.html
var packageTemplate string

//go:embed index.template.html
var indexTemplate string

func main() {
	// Read packages.yml
	data, err := os.ReadFile("packages.yml")
	if err != nil {
		log.Fatalf("Failed to read packages.yml: %v", err)
	}

	var pkgs []Package
	if err := yaml.Unmarshal(data, &pkgs); err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	// Use embedded package template
	tmplStr := packageTemplate

	// Create dist directory
	if err := os.MkdirAll("dist", 0755); err != nil {
		log.Fatalf("Failed to create dist directory: %v", err)
	}

	for _, pkg := range pkgs {
		out := strings.ReplaceAll(tmplStr, "{{import-prefix}}", pkg.ImportPrefix)
		out = strings.ReplaceAll(out, "{{vcs}}", pkg.VCS)
		out = strings.ReplaceAll(out, "{{repo-root}}", pkg.RepoRoot)

		outPath := filepath.Join("dist", pkg.Name+".html")
		if err := os.WriteFile(outPath, []byte(out), 0644); err != nil {
			log.Printf("Failed to write %s: %v", outPath, err)
		} else {
			fmt.Printf("Generated: %s\n", outPath)
		}
	}

	// Generate index.html with package list using embedded template
	var listHtml strings.Builder
	for _, pkg := range pkgs {
		item := fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", pkg.Home, pkg.ImportPrefix)
		listHtml.WriteString(item)
	}
	indexContent := strings.ReplaceAll(indexTemplate, "{{list}}", listHtml.String())
	indexFile := filepath.Join("dist", "index.html")
	if err := os.WriteFile(indexFile, []byte(indexContent), 0644); err != nil {
		log.Fatalf("Failed to create index.html: %v", err)
	}
	fmt.Println("Generated:", indexFile)

	// Copy files from public/ to dist/ using cp command if public exists
	if stat, err := os.Stat("public"); err == nil && stat.IsDir() {
		cmd := exec.Command("cp", "-va", "public/", "dist")
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to copy public/ to dist/: %v", err)
		}
	}
}
