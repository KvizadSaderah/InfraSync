package parser

import (
	"encoding/json"
	"fmt"
	"os"

	tfjson "github.com/hashicorp/terraform-json"
)

// ResourceChange represents a simplified resource change with action details
type ResourceChange struct {
	Address    string
	Type       string
	Actions    []string
	Before     map[string]interface{}
	After      map[string]interface{}
	IsCreate   bool
	IsUpdate   bool
	IsDelete   bool
	IsReplace  bool
	IsNoOp     bool
	BeforeSensitive interface{}
	AfterSensitive  interface{}
	AfterUnknown    interface{}
}

// PlanSummary contains the summary of terraform plan
type PlanSummary struct {
	TerraformVersion string
	FormatVersion    string
	Changes          []ResourceChange
	ToCreate         int
	ToUpdate         int
	ToDelete         int
	ToReplace        int
	NoChanges        int
}

// ParsePlanFile reads and parses a terraform plan JSON file
func ParsePlanFile(filename string) (*PlanSummary, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading plan file: %w", err)
	}

	var plan tfjson.Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("error unmarshalling plan JSON: %w", err)
	}

	return ParsePlan(&plan)
}

// ParsePlan processes a terraform plan and creates a summary
func ParsePlan(plan *tfjson.Plan) (*PlanSummary, error) {
	summary := &PlanSummary{
		TerraformVersion: plan.TerraformVersion,
		FormatVersion:    plan.FormatVersion,
		Changes:          make([]ResourceChange, 0),
	}

	if plan.ResourceChanges == nil {
		return summary, nil
	}

	for _, rc := range plan.ResourceChanges {
		change := classifyChange(rc)
		summary.Changes = append(summary.Changes, change)

		// Count changes
		if change.IsCreate {
			summary.ToCreate++
		}
		if change.IsUpdate {
			summary.ToUpdate++
		}
		if change.IsDelete {
			summary.ToDelete++
		}
		if change.IsReplace {
			summary.ToReplace++
		}
		if change.IsNoOp {
			summary.NoChanges++
		}
	}

	return summary, nil
}

// classifyChange determines the type of change for a resource
func classifyChange(rc *tfjson.ResourceChange) ResourceChange {
	change := ResourceChange{
		Address: rc.Address,
		Type:    rc.Type,
		Actions: make([]string, len(rc.Change.Actions)),
	}

	for i, action := range rc.Change.Actions {
		change.Actions[i] = string(action)
	}

	// Handle before/after as maps
	change.Before = toMap(rc.Change.Before)
	change.After = toMap(rc.Change.After)
	change.BeforeSensitive = rc.Change.BeforeSensitive
	change.AfterSensitive = rc.Change.AfterSensitive
	change.AfterUnknown = rc.Change.AfterUnknown

	// Classify the action
	actions := rc.Change.Actions

	// Check for replace (delete + create)
	if len(actions) == 2 {
		hasDelete := false
		hasCreate := false
		for _, a := range actions {
			if a == tfjson.ActionDelete {
				hasDelete = true
			}
			if a == tfjson.ActionCreate {
				hasCreate = true
			}
		}
		if hasDelete && hasCreate {
			change.IsReplace = true
			return change
		}
	}

	// Single action
	if len(actions) == 1 {
		switch actions[0] {
		case tfjson.ActionCreate:
			change.IsCreate = true
		case tfjson.ActionDelete:
			change.IsDelete = true
		case tfjson.ActionUpdate:
			change.IsUpdate = true
		case tfjson.ActionNoop:
			change.IsNoOp = true
		}
	}

	return change
}

// toMap safely converts interface{} to map[string]interface{}
func toMap(v interface{}) map[string]interface{} {
	if v == nil {
		return make(map[string]interface{})
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return make(map[string]interface{})
}
