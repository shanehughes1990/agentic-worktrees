package taskboard

import (
	"fmt"
	"sort"
	"strings"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

const (
	DefaultBoardQualityThreshold = 85
)

type QualitySeverity string

const (
	QualitySeverityCritical QualitySeverity = "critical"
	QualitySeverityMajor    QualitySeverity = "major"
	QualitySeverityMinor    QualitySeverity = "minor"
)

type BoardQualityFinding struct {
	RuleID         string          `json:"rule_id"`
	Severity       QualitySeverity `json:"severity"`
	Message        string          `json:"message"`
	TaskID         string          `json:"task_id,omitempty"`
	RelatedTaskID  string          `json:"related_task_id,omitempty"`
	Deduction      int             `json:"deduction"`
	Similarity     float64         `json:"similarity,omitempty"`
}

type BoardQualityReport struct {
	Threshold        int                          `json:"threshold"`
	Score            int                          `json:"score"`
	Passed           bool                         `json:"passed"`
	CategoryScores   map[string]int               `json:"category_scores"`
	CriticalFailures []string                     `json:"critical_failures,omitempty"`
	Findings         []BoardQualityFinding        `json:"findings,omitempty"`
}

func EvaluateBoardQuality(board *domaintaskboard.Board, threshold int) BoardQualityReport {
	if threshold <= 0 {
		threshold = DefaultBoardQualityThreshold
	}
	report := BoardQualityReport{
		Threshold:      threshold,
		CategoryScores: map[string]int{"uniqueness": 35, "conflict_risk": 25, "dependencies": 20, "granularity": 20},
	}
	if board == nil {
		report.Score = 0
		report.Passed = false
		report.CriticalFailures = []string{"board is nil"}
		return report
	}

	if err := board.ValidateBasics(); err != nil {
		report.CriticalFailures = append(report.CriticalFailures, fmt.Sprintf("invalid board basics: %v", err))
		report.CategoryScores["dependencies"] = 0
	}

	taskMap := map[string]domaintaskboard.Task{}
	for _, epic := range board.Epics {
		for _, task := range epic.Tasks {
			taskMap[task.ID] = task
		}
	}

	titleIndex := map[string]string{}
	for taskID, task := range taskMap {
		normalizedTitle := normalizeTitle(task.Title)
		if normalizedTitle == "" {
			applyDeduction(&report, "granularity", BoardQualityFinding{RuleID: "title.empty", Severity: QualitySeverityMajor, TaskID: taskID, Message: "task title is empty after normalization", Deduction: 6})
			continue
		}
		if priorTaskID, exists := titleIndex[normalizedTitle]; exists {
			applyDeduction(&report, "uniqueness", BoardQualityFinding{RuleID: "task.duplicate_title", Severity: QualitySeverityCritical, TaskID: taskID, RelatedTaskID: priorTaskID, Message: "duplicate task intent detected by normalized title", Deduction: 20})
			report.CriticalFailures = append(report.CriticalFailures, fmt.Sprintf("duplicate task intent: %s and %s", priorTaskID, taskID))
		} else {
			titleIndex[normalizedTitle] = taskID
		}

		words := strings.Fields(strings.ToLower(task.Title))
		if len(words) < 4 {
			applyDeduction(&report, "granularity", BoardQualityFinding{RuleID: "title.too_short", Severity: QualitySeverityMinor, TaskID: taskID, Message: "task title is too short to express concrete scope + action", Deduction: 3})
		}
		if hasMultiIntentTitle(task.Title) {
			applyDeduction(&report, "granularity", BoardQualityFinding{RuleID: "title.multi_intent", Severity: QualitySeverityMajor, TaskID: taskID, Message: "task title appears to include multiple intents", Deduction: 6})
		}
		if hasBundledTestChecks(task.Title) {
			applyDeduction(&report, "granularity", BoardQualityFinding{RuleID: "test.bundled_checks", Severity: QualitySeverityMinor, TaskID: taskID, Message: "test task appears to bundle multiple independent checks", Deduction: 2})
		}
	}

	allTaskIDs := make([]string, 0, len(taskMap))
	for taskID := range taskMap {
		allTaskIDs = append(allTaskIDs, taskID)
	}
	sort.Strings(allTaskIDs)

	reachability := buildTaskReachability(taskMap)
	for _, fromTaskID := range allTaskIDs {
		for _, dependencyTaskID := range taskMap[fromTaskID].DependsOn {
			cleanDependencyTaskID := strings.TrimSpace(dependencyTaskID)
			if cleanDependencyTaskID == "" {
				continue
			}
			if _, exists := taskMap[cleanDependencyTaskID]; !exists {
				applyDeduction(&report, "dependencies", BoardQualityFinding{RuleID: "dependency.missing", Severity: QualitySeverityCritical, TaskID: fromTaskID, RelatedTaskID: cleanDependencyTaskID, Message: "task depends_on references missing task", Deduction: 20})
				report.CriticalFailures = append(report.CriticalFailures, fmt.Sprintf("missing dependency: %s -> %s", fromTaskID, cleanDependencyTaskID))
			}
		}
		if reachability[fromTaskID][fromTaskID] {
			applyDeduction(&report, "dependencies", BoardQualityFinding{RuleID: "dependency.cycle", Severity: QualitySeverityCritical, TaskID: fromTaskID, Message: "task dependency cycle detected", Deduction: 20})
			report.CriticalFailures = append(report.CriticalFailures, fmt.Sprintf("dependency cycle involving task %s", fromTaskID))
		}
	}

	for leftIndex := 0; leftIndex < len(allTaskIDs); leftIndex++ {
		for rightIndex := leftIndex + 1; rightIndex < len(allTaskIDs); rightIndex++ {
			leftTaskID := allTaskIDs[leftIndex]
			rightTaskID := allTaskIDs[rightIndex]

			if reachability[leftTaskID][rightTaskID] || reachability[rightTaskID][leftTaskID] {
				continue
			}

			similarity := jaccardSimilarity(tokenizeTaskTitle(taskMap[leftTaskID].Title), tokenizeTaskTitle(taskMap[rightTaskID].Title))
			if similarity >= 0.75 {
				applyDeduction(&report, "conflict_risk", BoardQualityFinding{RuleID: "parallel.high_overlap", Severity: QualitySeverityCritical, TaskID: leftTaskID, RelatedTaskID: rightTaskID, Similarity: similarity, Message: "high-overlap parallel tasks may conflict", Deduction: 15})
				report.CriticalFailures = append(report.CriticalFailures, fmt.Sprintf("parallel high-overlap tasks: %s and %s", leftTaskID, rightTaskID))
				continue
			}
			if similarity >= 0.55 {
				applyDeduction(&report, "conflict_risk", BoardQualityFinding{RuleID: "parallel.medium_overlap", Severity: QualitySeverityMajor, TaskID: leftTaskID, RelatedTaskID: rightTaskID, Similarity: similarity, Message: "medium-overlap parallel tasks may need ordering", Deduction: 8})
			}
		}
	}

	report.Score = clampScore(report.CategoryScores["uniqueness"] + report.CategoryScores["conflict_risk"] + report.CategoryScores["dependencies"] + report.CategoryScores["granularity"])
	report.Passed = len(report.CriticalFailures) == 0 && report.Score >= report.Threshold
	return report
}

func applyDeduction(report *BoardQualityReport, category string, finding BoardQualityFinding) {
	if report == nil {
		return
	}
	if _, exists := report.CategoryScores[category]; !exists {
		report.CategoryScores[category] = 0
	}
	report.CategoryScores[category] -= finding.Deduction
	if report.CategoryScores[category] < 0 {
		report.CategoryScores[category] = 0
	}
	report.Findings = append(report.Findings, finding)
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func normalizeTitle(input string) string {
	tokens := tokenizeTaskTitle(input)
	if len(tokens) == 0 {
		return ""
	}
	return strings.Join(tokens, " ")
}

func tokenizeTaskTitle(input string) []string {
	replacer := strings.NewReplacer("/", " ", "-", " ", "_", " ", ",", " ", ":", " ", ";", " ", ".", " ", "(", " ", ")", " ")
	clean := strings.ToLower(strings.TrimSpace(replacer.Replace(input)))
	if clean == "" {
		return nil
	}
	stopWords := map[string]struct{}{
		"the": {}, "a": {}, "an": {}, "to": {}, "for": {}, "and": {}, "or": {}, "of": {}, "in": {}, "with": {}, "as": {}, "by": {}, "on": {}, "this": {}, "that": {},
	}
	parts := strings.Fields(clean)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if _, isStopWord := stopWords[part]; isStopWord {
			continue
		}
		result = append(result, part)
	}
	return result
}

func hasMultiIntentTitle(title string) bool {
	clean := strings.ToLower(strings.TrimSpace(title))
	if !strings.Contains(clean, " and ") {
		return false
	}
	actionVerbs := map[string]struct{}{
		"add": {}, "define": {}, "implement": {}, "map": {}, "wire": {}, "replace": {}, "move": {}, "verify": {}, "validate": {}, "document": {}, "refactor": {}, "preserve": {}, "enforce": {}, "run": {}, "create": {}, "switch": {},
	}
	verbCount := 0
	for _, token := range strings.Fields(clean) {
		if _, exists := actionVerbs[token]; exists {
			verbCount++
		}
	}
	return verbCount >= 2
}

func hasBundledTestChecks(title string) bool {
	clean := strings.ToLower(strings.TrimSpace(title))
	if !strings.Contains(clean, "test") && !strings.Contains(clean, "tests") {
		return false
	}
	if strings.Contains(clean, " plus ") || strings.Contains(clean, " | ") {
		return true
	}
	if strings.Contains(clean, " and ") {
		checkKeywords := []string{"parity", "error", "mapping", "lifecycle", "regression", "schema", "compatibility", "provider", "flow"}
		matchCount := 0
		for _, keyword := range checkKeywords {
			if strings.Contains(clean, keyword) {
				matchCount++
			}
		}
		return matchCount >= 2
	}
	return false
}

func jaccardSimilarity(left []string, right []string) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	leftSet := make(map[string]struct{}, len(left))
	for _, token := range left {
		leftSet[token] = struct{}{}
	}
	rightSet := make(map[string]struct{}, len(right))
	for _, token := range right {
		rightSet[token] = struct{}{}
	}
	intersection := 0
	union := len(leftSet)
	for token := range rightSet {
		if _, exists := leftSet[token]; exists {
			intersection++
		} else {
			union++
		}
	}
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func buildTaskReachability(tasks map[string]domaintaskboard.Task) map[string]map[string]bool {
	reachability := make(map[string]map[string]bool, len(tasks))
	for taskID := range tasks {
		reachability[taskID] = make(map[string]bool, len(tasks))
	}
	for taskID, task := range tasks {
		for _, dependencyTaskID := range task.DependsOn {
			cleanDependencyTaskID := strings.TrimSpace(dependencyTaskID)
			if cleanDependencyTaskID == "" {
				continue
			}
			reachability[taskID][cleanDependencyTaskID] = true
		}
	}

	for pivotTaskID := range tasks {
		for fromTaskID := range tasks {
			if !reachability[fromTaskID][pivotTaskID] {
				continue
			}
			for toTaskID := range tasks {
				if reachability[pivotTaskID][toTaskID] {
					reachability[fromTaskID][toTaskID] = true
				}
			}
		}
	}

	return reachability
}
