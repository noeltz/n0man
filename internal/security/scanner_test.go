package security

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/noeltz/n0man/internal/config"
)

func TestScanner(t *testing.T) {
	tests := []struct {
		name           string
		fileName       string
		fileContent    string
		expectFindings int
		expectRisk     RiskLevel
		expectPass     bool
	}{
		{
			name:           "AWS key in env file",
			fileName:       "config.env",
			fileContent:    "AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI7MDENG/bPxRfiCYEXAMPLEKEY",
			expectFindings: 1,
			expectRisk:     RiskLevelCritical,
			expectPass:     false,
		},
		{
			name:           "GitHub token",
			fileName:       ".github_token",
			fileContent:    "GITHUB_TOKEN=ghp_1234567890abcdefABCDEFghijklmnopqrstuv",
			expectFindings: 2, // May detect both generic_api_key and github_token patterns
			expectRisk:     RiskLevelCritical,
			expectPass:     false,
		},
		{
			name:           "OpenAI API key",
			fileName:       "api_key.txt",
			fileContent:    "API_KEY=sk-1234567890abcdefghijklmnopqrstuvwxyz",
			expectFindings: 1,
			expectRisk:     RiskLevelHigh,
			expectPass:     false,
		},
		{
			name:           "Private key file",
			fileName:       "id_rsa",
			fileContent:    "-----BEGIN RSA PRIVATE KEY-----\nABC\n-----END RSA PRIVATE KEY-----",
			expectFindings: 1,
			expectRisk:     RiskLevelCritical,
			expectPass:     false,
		},
		{
			name:           "Clean config file",
			fileName:       "config.toml",
			fileContent:    "debug=true\nlog_level=info\nport=8080",
			expectFindings: 0,
			expectRisk:     RiskLevelNone,
			expectPass:     true,
		},
		{
			name:           "Normal text file",
			fileName:       "readme.txt",
			fileContent:    "just normal config\nnothing secret here\n",
			expectFindings: 0,
			expectRisk:     RiskLevelNone,
			expectPass:     true,
		},
	}

	cfg := config.DefaultConfig()
	scanner := NewScanner(&cfg.Security)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, tt.fileName)

			if err := os.WriteFile(filePath, []byte(tt.fileContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			report, err := scanner.ScanPath(tempDir)
			if err != nil {
				t.Fatalf("ScanPath failed: %v", err)
			}

			if report.TotalFindings != tt.expectFindings {
				t.Errorf("Expected %d findings, got %d", tt.expectFindings, report.TotalFindings)
			}

			if report.HighestRisk != tt.expectRisk {
				t.Errorf("Expected risk %v, got %v", tt.expectRisk, report.HighestRisk)
			}
		})
	}
}
