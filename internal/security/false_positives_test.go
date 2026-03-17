package security

import (
	"testing"
)

// Test False Positive Prevention
// Reference: docs/design/security-scanner-improvements-design.md
// AC: FP-001 through FP-003

func TestStructuralPatternDetection_UI(t *testing.T) {
	analyzer := NewEntropyAnalyzer(4.5)

	// Test: UI mnemonic with alphabetic characters after bracket
	uiString := "[x] close file"
	isStructured, confidence := analyzer.analyzeStructuralPattern(uiString)

	if !isStructured || confidence != 0.1 {
		t.Errorf("Expected UI mnemonic to be structured with confidence 0.1, got isStructured=%v, confidence=%v", isStructured, confidence)
	}

	// Test: UI mnemonic with space and 2-char mnemonic
	uiString2 := "[ab] another option"
	isStructured2, confidence2 := analyzer.analyzeStructuralPattern(uiString2)

	if !isStructured2 || confidence2 != 0.1 {
		t.Errorf("Expected UI mnemonic with space to be structured with confidence 0.1, got isStructured=%v, confidence=%v", isStructured2, confidence2)
	}

	// Test: Non-UI string should not match
	nonUIString := "random-high-entropy-string-without-brackets"
	isStructured3, _ := analyzer.analyzeStructuralPattern(nonUIString)

	if isStructured3 {
		t.Errorf("Expected non-UI string to not be structured, got isStructured=%v", isStructured3)
	}
}

func TestStructuralPatternDetection_VimCommand(t *testing.T) {
	analyzer := NewEntropyAnalyzer(4.5)

	// Test: Vim command with optional key notation
	vimString := ":noremap <C-w> <C-q>"
	isStructured, confidence := analyzer.analyzeStructuralPattern(vimString)

	if !isStructured || confidence != 0.1 {
		t.Errorf("Expected vim command to be structured with confidence 0.1, got isStructured=%v, confidence=%v", isStructured, confidence)
	}

	// Test: Simple vim command
	vimString2 := ":w"
	isStructured2, confidence2 := analyzer.analyzeStructuralPattern(vimString2)

	if !isStructured2 || confidence2 != 0.1 {
		t.Errorf("Expected simple vim command to be structured with confidence 0.1, got isStructured=%v, confidence=%v", isStructured2, confidence2)
	}

	// Test: Non-vim string should not match
	nonVimString := "not a vim command"
	isStructured3, _ := analyzer.analyzeStructuralPattern(nonVimString)

	if isStructured3 {
		t.Errorf("Expected non-vim string to not be structured, got isStructured=%v", isStructured3)
	}
}

func TestStructuralPatternDetection_Checkbox(t *testing.T) {
	analyzer := NewEntropyAnalyzer(4.5)

	// Test: Checked checkbox
	checkboxString := "[x] Task completed"
	isStructured, confidence := analyzer.analyzeStructuralPattern(checkboxString)

	if !isStructured || confidence != 0.1 {
		t.Errorf("Expected checkbox to be structured with confidence 0.1, got isStructured=%v, confidence=%v", isStructured, confidence)
	}

	// Test: Unchecked checkbox
	checkboxString2 := "[ ] Task pending"
	isStructured2, confidence2 := analyzer.analyzeStructuralPattern(checkboxString2)

	if !isStructured2 || confidence2 != 0.1 {
		t.Errorf("Expected unchecked checkbox to be structured with confidence 0.1, got isStructured=%v, confidence=%v", isStructured2, confidence2)
	}

	// Test: Checkbox with space and text
	checkboxString3 := "[ ] Task with description here"
	isStructured3, confidence3 := analyzer.analyzeStructuralPattern(checkboxString3)

	if !isStructured3 || confidence3 != 0.1 {
		t.Errorf("Expected checkbox with text to be structured with confidence 0.1, got isStructured=%v, confidence=%v", isStructured3, confidence3)
	}

	// Test: Non-checkbox string should not match
	nonCheckboxString := "not a checkbox pattern"
	isStructured4, _ := analyzer.analyzeStructuralPattern(nonCheckboxString)

	if isStructured4 {
		t.Errorf("Expected non-checkbox string to not be structured, got isStructured=%v", isStructured4)
	}
}

func TestCodePatternDetection_FunctionCall(t *testing.T) {
	analyzer := NewEntropyAnalyzer(4.5)

	// Test: JavaScript function call
	jsString := "function fetchData() { return data; }"
	isStructured, confidence := analyzer.analyzeStructuralPattern(jsString)

	if !isStructured || confidence != 1.0 {
		t.Errorf("Expected JavaScript function to be code pattern with confidence 1.0, got isStructured=%v, confidence=%v", isStructured, confidence)
	}

	// Test: Python function call
	pyString := "def process_data():\n    return data"
	isStructured2, confidence2 := analyzer.analyzeStructuralPattern(pyString)

	if !isStructured2 || confidence2 != 1.0 {
		t.Errorf("Expected Python function to be code pattern with confidence 1.0, got isStructured=%v, confidence=%v", isStructured2, confidence2)
	}

	// Test: Go function call
	goString := "func processData() {\n    return data\n}"
	isStructured3, confidence3 := analyzer.analyzeStructuralPattern(goString)

	if !isStructured3 || confidence3 != 1.0 {
		t.Errorf("Expected Go function to be code pattern with confidence 1.0, got isStructured=%v, confidence=%v", isStructured3, confidence3)
	}

	// Test: High entropy string without code pattern should not be flagged as code
	highEntropyString := "sk-1234567890abcdefghijklmnopqrst"
	isStructured4, _ := analyzer.analyzeStructuralPattern(highEntropyString)

	if isStructured4 {
		t.Errorf("Expected high entropy string to not be code pattern, got isStructured=%v", isStructured4)
	}
}

func TestCodePatternDetection_Operators(t *testing.T) {
	analyzer := NewEntropyAnalyzer(4.5)

	// Test: Arrow operator
	arrowString := "result -> nextValue"
	isStructured, confidence := analyzer.analyzeStructuralPattern(arrowString)

	if !isStructured || confidence != 1.0 {
		t.Errorf("Expected arrow operator to be code pattern with confidence 1.0, got isStructured=%v, confidence=%v", isStructured, confidence)
	}

	// Test: Assignment operator
	assignString := "value := otherValue"
	isStructured2, confidence2 := analyzer.analyzeStructuralPattern(assignString)

	if !isStructured2 || confidence2 != 1.0 {
		t.Errorf("Expected assignment operator to be code pattern with confidence 1.0, got isStructured=%v, confidence=%v", isStructured2, confidence2)
	}

	// Test: Comparison operators
	compString := "if a == b && c != d"
	isStructured3, confidence3 := analyzer.analyzeStructuralPattern(compString)

	if !isStructured3 || confidence3 != 1.0 {
		t.Errorf("Expected comparison operators to be code pattern with confidence 1.0, got isStructured=%v, confidence=%v", isStructured3, confidence3)
	}

	// Test: Non-operator string should not be flagged
	nonOpString := "not an operator pattern"
	isStructured4, _ := analyzer.analyzeStructuralPattern(nonOpString)

	if isStructured4 {
		t.Errorf("Expected non-operator string to not be code pattern, got isStructured=%v", isStructured4)
	}
}

func TestCodePatternDetection_ControlFlow(t *testing.T) {
	analyzer := NewEntropyAnalyzer(4.5)

	// Test: If statement
	ifString := "if condition {\n    doSomething()\n}"
	isStructured, confidence := analyzer.analyzeStructuralPattern(ifString)

	if !isStructured || confidence != 1.0 {
		t.Errorf("Expected if statement to be code pattern with confidence 1.0, got isStructured=%v, confidence=%v", isStructured, confidence)
	}

	// Test: For loop
	forString := "for i := 0; i < 10; i++ {\n    process(i)\n}"
	isStructured2, confidence2 := analyzer.analyzeStructuralPattern(forString)

	if !isStructured2 || confidence2 != 1.0 {
		t.Errorf("Expected for loop to be code pattern with confidence 1.0, got isStructured=%v, confidence=%v", isStructured2, confidence2)
	}

	// Test: While loop
	whileString := "while condition {\n    doWork()\n}"
	isStructured3, confidence3 := analyzer.analyzeStructuralPattern(whileString)

	if !isStructured3 || confidence3 != 1.0 {
		t.Errorf("Expected while loop to be code pattern with confidence 1.0, got isStructured=%v, confidence=%v", isStructured3, confidence3)
	}

	// Test: Return statement
	returnString := "return result"
	isStructured4, confidence4 := analyzer.analyzeStructuralPattern(returnString)

	if !isStructured4 || confidence4 != 1.0 {
		t.Errorf("Expected return statement to be code pattern with confidence 1.0, got isStructured=%v, confidence=%v", isStructured4, confidence4)
	}

	// Test: Non-control flow string should not be flagged
	nonFlowString := "not a control flow pattern"
	isStructured5, _ := analyzer.analyzeStructuralPattern(nonFlowString)

	if isStructured5 {
		t.Errorf("Expected non-control flow string to not be code pattern, got isStructured=%v", isStructured5)
	}
}

func TestDictionaryFiltering_CommonWords(t *testing.T) {
	analyzer := NewEntropyAnalyzer(4.5)

	// Test: String with common English words
	commonWordsString := "the and for with from find search"
	score := analyzer.calculateNaturalLanguageScore(commonWordsString)

	if score < 0.6 {
		t.Errorf("Expected common words string to have natural language score >= 0.6, got %v", score)
	}

	// Test: String with fewer common words
	fewerWordsString := "some random words"
	score2 := analyzer.calculateNaturalLanguageScore(fewerWordsString)

	if score2 < 0.3 {
		t.Errorf("Expected fewer common words string to have natural language score < 0.3, got %v", score2)
	}

	// Test: High entropy string with no common words
	highEntropyString := "sk-1234567890abcdefghijklmnopqrst"
	score3 := analyzer.calculateNaturalLanguageScore(highEntropyString)

	if score3 >= 0.3 {
		t.Errorf("Expected high entropy string with no common words to have natural language score < 0.3, got %v", score3)
	}

	// Test: Empty string
	emptyString := ""
	score4 := analyzer.calculateNaturalLanguageScore(emptyString)

	if score4 != 0.0 {
		t.Errorf("Expected empty string to have natural language score 0.0, got %v", score4)
	}
}

func TestDictionaryFiltering_TechnicalTerms(t *testing.T) {
	analyzer := NewEntropyAnalyzer(4.5)

	// Test: String with technical terms
	techString := "config settings options preferences profile theme"
	score := analyzer.calculateNaturalLanguageScore(techString)

	if score < 0.5 {
		t.Errorf("Expected technical terms string to have natural language score >= 0.5, got %v", score)
	}

	// Test: String with fewer technical terms
	fewerTechString := "some configuration"
	score2 := analyzer.calculateNaturalLanguageScore(fewerTechString)

	if score2 < 0.2 {
		t.Errorf("Expected fewer technical terms string to have natural language score < 0.2, got %v", score2)
	}

	// Test: Random string with no technical terms
	randomString := "sk-1234567890abcdefghijklmnopqrst"
	score3 := analyzer.calculateNaturalLanguageScore(randomString)

	if score3 >= 0.2 {
		t.Errorf("Expected random string with no technical terms to have natural language score < 0.2, got %v", score3)
	}
}
