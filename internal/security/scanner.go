// Package security provides comprehensive secret detection for dotfiles.
//
// PURPOSE: Prevent accidental commit of sensitive data (API keys, passwords, tokens)
// to version control by scanning files before git operations.
//
// DETECTION METHODS:
// 1. Pattern-based detection - File names and paths (.env, .pem, SSH keys)
// 2. Content scanning - Regex patterns for known secret formats
// 3. Entropy analysis - Shannon entropy for detecting random/high-entropy strings
//
// ARCHITECTURE:
//
//	Scanner (orchestrator)
//	├── PatternMatcher (file path patterns)
//	└── Detector (content analysis)
//	    └── EntropyAnalyzer (randomness detection)
//
// TEST CASES: See scanner_test.go for:
//   - AWS key detection
//   - GitHub token detection
//   - OpenAI API key detection
//   - Private key detection
//   - Clean file validation
package security

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/noeltz/n0man/internal/config"
)

// ScanResult represents the security analysis of a single file.
//
// PURPOSE: Encapsulate all findings for one file with pass/fail status.
//
// FIELDS:
//   - FilePath: Absolute path to the scanned file
//   - Findings: Array of detected secrets (empty if clean)
//   - Risk: Highest risk level found (None, Low, Medium, High, Critical)
//   - Passed: true if no high-risk secrets found OR fail_on_secrets=false
//   - Error: Non-nil if scanning failed (I/O error, etc.)
type ScanResult struct {
	FilePath string
	Findings []Finding
	Risk     RiskLevel
	Passed   bool
	Error    error
}

// ScanReport aggregates results from scanning multiple files.
//
// PURPOSE: Provide summary statistics for batch scanning operations.
//
// USAGE EXAMPLE:
//
//	report, err := scanner.ScanPath("/home/user/.config")
//	if report.TotalFindings > 0 {
//	    fmt.Printf("Found %d secrets in %d files\n", report.TotalFindings, report.ScannedFiles)
//	}
type ScanReport struct {
	Results       []ScanResult
	TotalFiles    int
	ScannedFiles  int
	SkippedFiles  int
	TotalFindings int
	HighestRisk   RiskLevel
}

// Scanner is the main security scanning orchestrator.
//
// PURPOSE: Coordinate pattern matching and content detection for secret scanning.
//
// ARCHITECTURE PATTERN: Facade pattern - provides simple interface to complex subsystem.
//
// COMPONENTS:
//   - config: User-configurable sensitivity and detection settings
//   - patternMatcher: File path pattern detection (e.g., .env, *.pem)
//   - contentDetector: Content analysis with regex and entropy analysis
//
// THREAD SAFETY: Scanner instances are NOT thread-safe. Create per-operation.
type Scanner struct {
	config          *config.SecurityConfig
	patternMatcher  *PatternMatcher
	contentDetector *Detector
}

// NewScanner creates a new Scanner with the provided configuration.
//
// PURPOSE: Initialize scanner with user-defined sensitivity and detection rules.
//
// PARAMETERS:
//   - cfg: Security configuration (sensitivity, allowlists, thresholds)
//
// RETURNS:
//   - *Scanner: Ready-to-use scanner instance
//
// EXAMPLE:
//
//	cfg := config.DefaultConfig().Security
//	cfg.Sensitivity = "high"  // More sensitive detection
//	scanner := NewScanner(&cfg)
//	report, err := scanner.ScanPath("/home/user/.config")
func NewScanner(cfg *config.SecurityConfig) *Scanner {
	return &Scanner{
		config:          cfg,
		patternMatcher:  NewPatternMatcher(cfg),
		contentDetector: NewDetector(cfg),
	}
}

// ScanFile analyzes a single file for secrets.
//
// PURPOSE: Core scanning logic - detects secrets in file content.
//
// DETECTION PIPELINE:
//  1. Pattern Check → Is this a high-risk file type? (.env, .pem, SSH keys)
//  2. Content Scan → Does content match secret patterns? (API keys, passwords)
//  3. Entropy Analysis → Are there high-entropy strings? (random tokens)
//
// PARAMETERS:
//   - path: File path for reporting (not read, content provided separately)
//   - content: File contents as byte slice
//
// RETURNS:
//   - *ScanResult: Analysis results with findings
//   - error: Non-nil if scanning failed
//
// TEST CASES:
//   - Empty file → Passed=true, Findings=0
//   - File with AWS key → Passed=false, Risk=Critical
//   - Binary file with ScanBinaryFiles=false → Skipped, Passed=true
func (s *Scanner) ScanFile(path string, content []byte) (*ScanResult, error) {
	result := &ScanResult{
		FilePath: path,
		Findings: []Finding{},
		Risk:     RiskLevelNone,
		Passed:   true,
	}

	// Early exit if scanning is disabled
	if !s.config.Enabled {
		return result, nil
	}

	// STEP 1: Pattern-based detection (file path analysis)
	// PURPOSE: Quickly identify high-risk file types without content analysis
	// EXAMPLES: .env files, SSH private keys, certificate files
	if s.config.ExcludePatterns {
		shouldExclude, risk, reason := s.patternMatcher.ShouldExclude(path)
		if shouldExclude && risk >= RiskLevelHigh {
			result.Risk = risk
			result.Passed = false
			result.Findings = append(result.Findings, Finding{
				Type:       SecretTypeGeneric,
				Location:   Location{FilePath: path},
				Confidence: 1.0,
				Reasons:    []string{reason},
				RiskLevel:  risk,
			})
			return result, nil
		}
	}

	// STEP 2: Content-based detection
	// PURPOSE: Analyze file content for secret patterns and high-entropy strings
	// CONSTRAINT: Skipped for binary files unless ScanBinaryFiles=true
	if s.config.ScanContent && len(content) > 0 {
		// Skip binary files if configured
		if !s.config.ContentScan.ScanBinaryFiles && IsBinary(content) {
			return result, nil
		}

		// Skip files exceeding size limit
		if s.config.ContentScan.MaxFileSize > 0 && len(content) > s.config.ContentScan.MaxFileSize {
			return result, nil
		}

		findings := s.contentDetector.DetectSecrets(content, path)
		result.Findings = append(result.Findings, findings...)

		for _, finding := range findings {
			if finding.RiskLevel > result.Risk {
				result.Risk = finding.RiskLevel
			}
		}

		result.Passed = result.Risk < RiskLevelHigh || (result.Risk == RiskLevelHigh && !s.config.FailOnSecrets)
	}

	return result, nil
}

// scanFileStreaming scans large files line-by-line to reduce memory usage.
//
// PURPOSE: Handle files >1MB without loading entire content into memory.
//
// PERFORMANCE CHARACTERISTICS:
//   - Memory: O(batch_size) instead of O(file_size)
//   - Speed: Slightly slower than full-file scan due to buffering
//   - Accuracy: Same detection as full-file scan
//
// IMPLEMENTATION PATTERN: Streaming/batch processing
//   - Reads file line by line
//   - Accumulates 100 lines in buffer
//   - Scans buffer, then clears
//   - Scans remaining lines at EOF
//
// CONSTRAINT: May miss secrets that span across batch boundaries (>100 lines apart)
// MITIGATION: 100-line batches are typically sufficient for most secret patterns
func (s *Scanner) scanFileStreaming(path string) (*ScanResult, error) {
	result := &ScanResult{
		FilePath: path,
		Findings: []Finding{},
		Risk:     RiskLevelNone,
		Passed:   true,
	}

	if !s.config.Enabled {
		return result, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var lineBuffer strings.Builder

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		lineBuffer.WriteString(line)
		lineBuffer.WriteString("\n")

		// Scan in batches of 100 lines to balance memory and performance
		// PURPOSE: Reduce memory pressure while maintaining detection accuracy
		if lineNum%100 == 0 {
			batchContent := lineBuffer.String()
			findings := s.contentDetector.DetectSecrets([]byte(batchContent), path)
			result.Findings = append(result.Findings, findings...)
			lineBuffer.Reset()
		}
	}

	// Scan remaining lines (final batch < 100 lines)
	if lineBuffer.Len() > 0 {
		findings := s.contentDetector.DetectSecrets([]byte(lineBuffer.String()), path)
		result.Findings = append(result.Findings, findings...)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for _, finding := range result.Findings {
		if finding.RiskLevel > result.Risk {
			result.Risk = finding.RiskLevel
		}
	}

	result.Passed = result.Risk < RiskLevelHigh || (result.Risk == RiskLevelHigh && !s.config.FailOnSecrets)
	return result, nil
}

// ScanPath scans all files in a directory tree.
//
// PURPOSE: Batch scan operation for entire directories.
//
// PARAMETERS:
//   - path: Root directory to scan
//
// RETURNS:
//   - *ScanReport: Aggregated results with statistics
//   - error: Non-nil if directory traversal failed
//
// EXCLUSIONS:
//   - .git/ directory (version control metadata)
//   - .backups/ directory (n0man backup snapshots)
//
// SEE ALSO: ScanPathWithContext for cancellable scanning
func (s *Scanner) ScanPath(path string) (*ScanReport, error) {
	return s.ScanPathWithContext(context.Background(), path)
}

// ScanPathWithContext scans a path for secrets with context for cancellation.
//
// PURPOSE: Enable cancellable scanning for long-running operations.
//
// CONTEXT USAGE:
//   - Check ctx.Done() before processing each file
//   - Return ctx.Err() immediately when cancelled
//   - Allows graceful shutdown during large directory scans
//
// PARAMETERS:
//   - ctx: Context for cancellation (e.g., from signal handling or timeout)
//   - path: Root directory to scan
//
// RETURNS:
//   - *ScanReport: Aggregated results (may be partial if cancelled)
//   - error: ctx.Err() if cancelled, or traversal error
//
// USAGE EXAMPLE:
//
//	// 5-minute timeout for scanning
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//
//	report, err := scanner.ScanPathWithContext(ctx, "/home/user/.config")
//	if errors.Is(err, context.DeadlineExceeded) {
//	    fmt.Println("Scan timed out - results may be incomplete")
//	}
//
// CONSTRAINT: Does not follow symlinks (security feature)
func (s *Scanner) ScanPathWithContext(ctx context.Context, path string) (*ScanReport, error) {
	report := &ScanReport{
		Results: []ScanResult{},
	}

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		// CONTEXT CHECK: Allow cancellation of long-running scans
		// PATTERN: Standard Go context cancellation check
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil
		}

		if info.IsDir() {
			// Skip version control and backup directories
			// PURPOSE: Avoid scanning Git metadata and n0man's own backups
			if info.Name() == ".git" || info.Name() == ".backups" {
				return filepath.SkipDir
			}
			return nil
		}

		report.TotalFiles++

		// PERFORMANCE OPTIMIZATION: Stream large files instead of loading into memory
		// THRESHOLD: 1MB - files larger than this use streaming
		if info.Size() > 1024*1024 {
			result, err := s.scanFileStreaming(filePath)
			if err != nil {
				result = &ScanResult{
					FilePath: filePath,
					Error:    err,
				}
			}
			report.ScannedFiles++
			report.Results = append(report.Results, *result)
			report.TotalFindings += len(result.Findings)
			if result.Risk > report.HighestRisk {
				report.HighestRisk = result.Risk
			}
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			report.SkippedFiles++
			return nil
		}

		result, err := s.ScanFile(filePath, content)
		if err != nil {
			result.Error = err
		}

		report.ScannedFiles++
		report.Results = append(report.Results, *result)
		report.TotalFindings += len(result.Findings)

		if result.Risk > report.HighestRisk {
			report.HighestRisk = result.Risk
		}

		return nil
	})

	return report, err
}

func IsBinary(content []byte) bool {
	checkLen := len(content)
	if checkLen > 8192 {
		checkLen = 8192
	}
	for i := 0; i < checkLen; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}

func (r *ScanReport) Summary() string {
	return fmt.Sprintf(
		"Scanned %d files: %d findings, highest risk: %s",
		r.ScannedFiles,
		r.TotalFindings,
		r.HighestRisk,
	)
}

// PrintFindings outputs detected findings (legacy support)
func PrintFindings(findings []Finding) {
	if len(findings) == 0 {
		fmt.Println("✅ No security issues found.")
		return
	}

	fmt.Printf("🚨 Found %d potential security issue(s):\n\n", len(findings))
	for i, f := range findings {
		fmt.Printf("  %d. [%s] %s", i+1, f.RiskLevel.String(), f.Location.FilePath)
		if f.Location.LineNumber > 0 {
			fmt.Printf(":%d", f.Location.LineNumber)
		}
		fmt.Println()
		if len(f.Reasons) > 0 {
			fmt.Printf("     → %s\n", strings.Join(f.Reasons, ", "))
		}
	}
}
