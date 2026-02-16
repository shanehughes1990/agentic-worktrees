package board

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/shanehughes1990/agentic-worktrees/internal/domain"
)

var dependsPattern = regexp.MustCompile(`\[depends:([^\]]+)\]`)

func BuildBoardFromFile(scopePath string) (domain.Board, error) {
	content, err := os.ReadFile(scopePath)
	if err != nil {
		return domain.Board{}, fmt.Errorf("read scope file: %w", err)
	}
	return BuildBoardFromContent(scopePath, string(content)), nil
}

func BuildBoardFromContent(scopePath string, content string) domain.Board {
	lines := strings.Split(content, "\n")
	tasks := make([]domain.Task, 0)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			line = strings.TrimSpace(line[1:])
		}
		if line == "" {
			continue
		}

		deps := parseDependencies(line)
		title := stripDependencyMarker(line)
		taskID := fmt.Sprintf("task-%03d", len(tasks)+1)

		tasks = append(tasks, domain.Task{
			ID:           taskID,
			Title:        title,
			Prompt:       title,
			Dependencies: deps,
			Lane:         laneForIndex(len(tasks)),
			Status:       domain.TaskStatusPending,
		})
	}

	if len(tasks) == 0 {
		fallback := strings.TrimSpace(content)
		if fallback == "" {
			fallback = "Bootstrap initial autonomous task"
		}
		tasks = append(tasks, domain.Task{
			ID:     "task-001",
			Title:  fallback,
			Prompt: fallback,
			Lane:   laneForIndex(0),
			Status: domain.TaskStatusPending,
		})
	}

	return domain.Board{
		SchemaVersion: domain.CurrentSchemaVersion,
		SourceScope:   scopePath,
		GeneratedAt:   time.Now().UTC(),
		Tasks:         tasks,
	}
}

func parseDependencies(line string) []string {
	match := dependsPattern.FindStringSubmatch(line)
	if len(match) < 2 {
		return nil
	}
	parts := strings.Split(match[1], ",")
	deps := make([]string, 0, len(parts))
	for _, part := range parts {
		dep := strings.TrimSpace(part)
		if dep != "" {
			deps = append(deps, dep)
		}
	}
	if len(deps) == 0 {
		return nil
	}
	return deps
}

func stripDependencyMarker(line string) string {
	cleaned := dependsPattern.ReplaceAllString(line, "")
	return strings.TrimSpace(cleaned)
}

func laneForIndex(index int) string {
	lanes := []string{"lane-a", "lane-b", "lane-c"}
	return lanes[index%len(lanes)]
}
