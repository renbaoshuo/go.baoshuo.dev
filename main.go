package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Package struct {
	Name         string `yaml:"name" json:"name"`
	ImportPrefix string `yaml:"import-prefix" json:"importPrefix"`
	Home         string `yaml:"home" json:"home"`
	VCS          string `yaml:"vcs" json:"vcs"`
	RepoRoot     string `yaml:"repo-root" json:"repoRoot"`
}

const (
	distDir   = "dist"
	publicDir = "public"

	packagesJsonFileName = "packages.json"

	packageTemplateFileName = "package.template.html"
	indexTemplateFileName   = "index.template.html"
)

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

	// Parse templates from files
	tmpl, err := template.ParseFiles(packageTemplateFileName, indexTemplateFileName)
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Create dist directory
	if err := os.MkdirAll(distDir, 0755); err != nil {
		log.Fatalf("Failed to create dist directory: %v", err)
	}

	for _, pkg := range pkgs {
		outPath := filepath.Join(distDir, fmt.Sprintf("%s.html", pkg.Name))
		f, err := os.Create(outPath)
		if err != nil {
			log.Printf("Failed to create %s: %v", outPath, err)
			continue
		}
		defer f.Close()

		if err := tmpl.ExecuteTemplate(f, packageTemplateFileName, pkg); err != nil {
			log.Printf("Failed to execute template for %s: %v", pkg.Name, err)
		} else {
			fmt.Printf("Generated: %s\n", outPath)
		}
	}

	// Generate index.html with package list
	indexFile := filepath.Join(distDir, "index.html")
	f, err := os.Create(indexFile)
	if err != nil {
		log.Fatalf("Failed to create index.html: %v", err)
	}
	defer f.Close()

	if err := tmpl.ExecuteTemplate(f, indexTemplateFileName, pkgs); err != nil {
		log.Fatalf("Failed to execute index template: %v", err)
	}
	fmt.Println("Generated:", indexFile)

	// Generate packages.json
	packagesJsonFile := filepath.Join(distDir, packagesJsonFileName)
	jsonData, err := json.MarshalIndent(pkgs, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal packages to JSON: %v", err)
	}
	if err := os.WriteFile(packagesJsonFile, jsonData, 0644); err != nil {
		log.Fatalf("Failed to write packages.json: %v", err)
	}
	fmt.Println("Generated:", packagesJsonFile)

	// Copy files from public/ to dist/ using cp command if public exists
	if stat, err := os.Stat(publicDir); err == nil && stat.IsDir() {
		cmd := exec.Command("cp", "-va", publicDir+"/.", distDir+"/")
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to copy %s/ to %s/: %v", publicDir, distDir, err)
		}
	}
}
