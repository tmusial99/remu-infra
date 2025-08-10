package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func cleanHost(host string) string {
	if strings.Contains(host, ":") {
		return strings.Split(host, ":")[0]
	}
	return host
}

func hasExtension(path string, exts []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range exts {
		if ext == strings.ToLower(e) {
			return true
		}
	}
	return false
}

func printConfig(config *ServerConfig) {
	var sb strings.Builder

	sb.WriteString("\n🔧 SERVER CONFIGURATION\n")
	sb.WriteString(fmt.Sprintf("Port: %s\n", config.Port))
	sb.WriteString(fmt.Sprintf("Localhost test domain: %s\n", config.LocalhostTestDomain))

	sb.WriteString("\n📁 DOMAIN MAPPINGS:\n")
	for _, domain := range config.Domains {
		apiStatus := "❌ No API"
		if domain.APIEnabled {
			apiStatus = "✅ API enabled"
		}
		sb.WriteString(fmt.Sprintf("   %s -> %s (%s)\n", domain.Host, domain.StaticPath, apiStatus))
	}

	fmt.Println(sb.String())
}