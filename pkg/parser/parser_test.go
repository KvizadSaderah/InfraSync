package parser

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
)

func TestClassifyChange_Create(t *testing.T) {
	rc := &tfjson.ResourceChange{
		Address: "aws_instance.test",
		Type:    "aws_instance",
		Change: &tfjson.Change{
			Actions: []tfjson.Action{tfjson.ActionCreate},
		},
	}

	change := classifyChange(rc)

	if !change.IsCreate {
		t.Error("Expected IsCreate to be true")
	}
	if change.IsUpdate || change.IsDelete || change.IsReplace || change.IsNoOp {
		t.Error("Expected other flags to be false")
	}
}

func TestClassifyChange_Delete(t *testing.T) {
	rc := &tfjson.ResourceChange{
		Address: "aws_instance.test",
		Type:    "aws_instance",
		Change: &tfjson.Change{
			Actions: []tfjson.Action{tfjson.ActionDelete},
		},
	}

	change := classifyChange(rc)

	if !change.IsDelete {
		t.Error("Expected IsDelete to be true")
	}
	if change.IsCreate || change.IsUpdate || change.IsReplace || change.IsNoOp {
		t.Error("Expected other flags to be false")
	}
}

func TestClassifyChange_Update(t *testing.T) {
	rc := &tfjson.ResourceChange{
		Address: "aws_instance.test",
		Type:    "aws_instance",
		Change: &tfjson.Change{
			Actions: []tfjson.Action{tfjson.ActionUpdate},
		},
	}

	change := classifyChange(rc)

	if !change.IsUpdate {
		t.Error("Expected IsUpdate to be true")
	}
	if change.IsCreate || change.IsDelete || change.IsReplace || change.IsNoOp {
		t.Error("Expected other flags to be false")
	}
}

func TestClassifyChange_Replace(t *testing.T) {
	rc := &tfjson.ResourceChange{
		Address: "aws_instance.test",
		Type:    "aws_instance",
		Change: &tfjson.Change{
			Actions: []tfjson.Action{tfjson.ActionDelete, tfjson.ActionCreate},
		},
	}

	change := classifyChange(rc)

	if !change.IsReplace {
		t.Error("Expected IsReplace to be true")
	}
	if change.IsCreate || change.IsUpdate || change.IsDelete || change.IsNoOp {
		t.Error("Expected other flags to be false")
	}
}

func TestClassifyChange_NoOp(t *testing.T) {
	rc := &tfjson.ResourceChange{
		Address: "aws_instance.test",
		Type:    "aws_instance",
		Change: &tfjson.Change{
			Actions: []tfjson.Action{tfjson.ActionNoop},
		},
	}

	change := classifyChange(rc)

	if !change.IsNoOp {
		t.Error("Expected IsNoOp to be true")
	}
	if change.IsCreate || change.IsUpdate || change.IsDelete || change.IsReplace {
		t.Error("Expected other flags to be false")
	}
}

func TestParsePlan_CountsChanges(t *testing.T) {
	plan := &tfjson.Plan{
		TerraformVersion: "1.0.0",
		FormatVersion:    "1.0",
		ResourceChanges: []*tfjson.ResourceChange{
			{
				Address: "aws_instance.create",
				Type:    "aws_instance",
				Change:  &tfjson.Change{Actions: []tfjson.Action{tfjson.ActionCreate}},
			},
			{
				Address: "aws_instance.update",
				Type:    "aws_instance",
				Change:  &tfjson.Change{Actions: []tfjson.Action{tfjson.ActionUpdate}},
			},
			{
				Address: "aws_instance.delete",
				Type:    "aws_instance",
				Change:  &tfjson.Change{Actions: []tfjson.Action{tfjson.ActionDelete}},
			},
			{
				Address: "aws_instance.replace",
				Type:    "aws_instance",
				Change:  &tfjson.Change{Actions: []tfjson.Action{tfjson.ActionDelete, tfjson.ActionCreate}},
			},
			{
				Address: "aws_instance.noop",
				Type:    "aws_instance",
				Change:  &tfjson.Change{Actions: []tfjson.Action{tfjson.ActionNoop}},
			},
		},
	}

	summary, err := ParsePlan(plan)
	if err != nil {
		t.Fatalf("ParsePlan failed: %v", err)
	}

	if summary.ToCreate != 1 {
		t.Errorf("Expected ToCreate=1, got %d", summary.ToCreate)
	}
	if summary.ToUpdate != 1 {
		t.Errorf("Expected ToUpdate=1, got %d", summary.ToUpdate)
	}
	if summary.ToDelete != 1 {
		t.Errorf("Expected ToDelete=1, got %d", summary.ToDelete)
	}
	if summary.ToReplace != 1 {
		t.Errorf("Expected ToReplace=1, got %d", summary.ToReplace)
	}
	if summary.NoChanges != 1 {
		t.Errorf("Expected NoChanges=1, got %d", summary.NoChanges)
	}
}

func TestToMap_NilInput(t *testing.T) {
	result := toMap(nil)
	if result == nil {
		t.Error("Expected non-nil map")
	}
	if len(result) != 0 {
		t.Error("Expected empty map")
	}
}

func TestToMap_ValidMap(t *testing.T) {
	input := map[string]interface{}{
		"key": "value",
	}

	result := toMap(input)
	if result == nil {
		t.Error("Expected non-nil map")
	}
	if result["key"] != "value" {
		t.Error("Expected map to contain key-value pair")
	}
}

func TestToMap_InvalidInput(t *testing.T) {
	result := toMap("not a map")
	if result == nil {
		t.Error("Expected non-nil map")
	}
	if len(result) != 0 {
		t.Error("Expected empty map for invalid input")
	}
}
