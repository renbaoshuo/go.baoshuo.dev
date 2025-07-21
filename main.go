package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Package struct {
	Name         string `yaml:"name" json:"name" xml:"name"`
	ImportPrefix string `yaml:"import-prefix" json:"importPrefix" xml:"import-prefix"`
	Home         string `yaml:"home" json:"home" xml:"home"`
	VCS          string `yaml:"vcs" json:"vcs" xml:"vcs"`
	RepoRoot     string `yaml:"repo-root" json:"repoRoot" xml:"repo-root"`
}

type Packages struct {
	XMLName  xml.Name  `xml:"packages"`
	Packages []Package `xml:"package"`
}

const (
	distDir   = "dist"
	publicDir = "public"

	packagesJsonFileName = "packages.json"
	packagesXmlFileName  = "packages.xml"

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

	// Generate packages.xml
	packagesXmlFile := filepath.Join(distDir, packagesXmlFileName)
	xmlData, err := xml.MarshalIndent(Packages{Packages: pkgs}, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal packages to XML: %v", err)
	}
	if err := os.WriteFile(packagesXmlFile, []byte(xml.Header+string(xmlData)), 0644); err != nil {
		log.Fatalf("Failed to write packages.xml: %v", err)
	}
	fmt.Println("Generated:", packagesXmlFile)

	// Copy files from public/ to dist/ using cp command if public exists
	if stat, err := os.Stat(publicDir); err == nil && stat.IsDir() {
		cmd := exec.Command("cp", "-va", publicDir+"/.", distDir+"/")
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to copy %s/ to %s/: %v", publicDir, distDir, err)
		}
	}
}
