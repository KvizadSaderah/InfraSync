package analyzer

import (
	"testing"

	"github.com/kvizadsaderah/infrasync/pkg/parser"
)

func TestAnalyzeChanges_DatabaseDeletion(t *testing.T) {
	summary := &parser.PlanSummary{
		Changes: []parser.ResourceChange{
			{
				Address:  "aws_db_instance.main",
				Type:     "aws_db_instance",
				IsDelete: true,
			},
		},
	}

	warnings := AnalyzeChanges(summary)

	if len(warnings) == 0 {
		t.Error("Expected warnings for database deletion")
	}

	foundCritical := false
	for _, w := range warnings {
		if w.Level == RiskCritical {
			foundCritical = true
		}
	}

	if !foundCritical {
		t.Error("Expected critical warning for database deletion")
	}
}

func TestAnalyzeChanges_ProductionResourceDeletion(t *testing.T) {
	summary := &parser.PlanSummary{
		Changes: []parser.ResourceChange{
			{
				Address:  "aws_instance.production_web",
				Type:     "aws_instance",
				IsDelete: true,
			},
		},
	}

	warnings := AnalyzeChanges(summary)

	if len(warnings) == 0 {
		t.Error("Expected warnings for production resource deletion")
	}

	foundProdWarning := false
	for _, w := range warnings {
		if w.Resource == "aws_instance.production_web" && w.Level == RiskCritical {
			foundProdWarning = true
		}
	}

	if !foundProdWarning {
		t.Error("Expected critical warning for production resource")
	}
}

func TestAnalyzeChanges_DatabaseReplace(t *testing.T) {
	summary := &parser.PlanSummary{
		Changes: []parser.ResourceChange{
			{
				Address:   "aws_rds_cluster.main",
				Type:      "aws_rds_cluster",
				IsReplace: true,
			},
		},
	}

	warnings := AnalyzeChanges(summary)

	if len(warnings) == 0 {
		t.Error("Expected warnings for database replacement")
	}

	foundCritical := false
	for _, w := range warnings {
		if w.Level == RiskCritical {
			foundCritical = true
		}
	}

	if !foundCritical {
		t.Error("Expected critical warning for database replacement")
	}
}

func TestAnalyzeChanges_NoWarningsForSafeChanges(t *testing.T) {
	summary := &parser.PlanSummary{
		Changes: []parser.ResourceChange{
			{
				Address:  "aws_s3_bucket.logs",
				Type:     "aws_s3_bucket",
				IsCreate: true,
			},
			{
				Address:  "aws_cloudwatch_log_group.app",
				Type:     "aws_cloudwatch_log_group",
				IsUpdate: true,
				Before:   map[string]interface{}{"retention_days": 7},
				After:    map[string]interface{}{"retention_days": 30},
			},
		},
	}

	warnings := AnalyzeChanges(summary)

	// Should have no critical warnings for safe changes
	for _, w := range warnings {
		if w.Level == RiskCritical {
			t.Errorf("Unexpected critical warning for safe change: %s", w.Message)
		}
	}
}

func TestIsDatabase(t *testing.T) {
	tests := []struct {
		resourceType string
		expected     bool
	}{
		{"aws_db_instance", true},
		{"aws_rds_cluster", true},
		{"google_sql_database", true},
		{"azurerm_sql_database", true},
		{"aws_dynamodb_table", true},
		{"aws_instance", false},
		{"aws_s3_bucket", false},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			result := isDatabase(tt.resourceType)
			if result != tt.expected {
				t.Errorf("isDatabase(%s) = %v, want %v", tt.resourceType, result, tt.expected)
			}
		})
	}
}

func TestIsStorage(t *testing.T) {
	tests := []struct {
		resourceType string
		expected     bool
	}{
		{"aws_s3_bucket", true},
		{"google_storage_bucket", true},
		{"azurerm_storage_account", true},
		{"aws_ebs_volume", true},
		{"aws_instance", false},
		{"aws_db_instance", false},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			result := isStorage(tt.resourceType)
			if result != tt.expected {
				t.Errorf("isStorage(%s) = %v, want %v", tt.resourceType, result, tt.expected)
			}
		})
	}
}

func TestHasEncryptionDisabled(t *testing.T) {
	tests := []struct {
		name     string
		before   map[string]interface{}
		after    map[string]interface{}
		expected bool
	}{
		{
			name:     "encryption disabled",
			before:   map[string]interface{}{"encrypted": true},
			after:    map[string]interface{}{"encrypted": false},
			expected: true,
		},
		{
			name:     "encryption enabled",
			before:   map[string]interface{}{"encrypted": false},
			after:    map[string]interface{}{"encrypted": true},
			expected: false,
		},
		{
			name:     "encryption unchanged",
			before:   map[string]interface{}{"encrypted": true},
			after:    map[string]interface{}{"encrypted": true},
			expected: false,
		},
		{
			name:     "no encryption field",
			before:   map[string]interface{}{"name": "test"},
			after:    map[string]interface{}{"name": "test"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasEncryptionDisabled(tt.before, tt.after)
			if result != tt.expected {
				t.Errorf("hasEncryptionDisabled() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasBackupDisabled(t *testing.T) {
	tests := []struct {
		name     string
		before   map[string]interface{}
		after    map[string]interface{}
		expected bool
	}{
		{
			name:     "versioning disabled",
			before:   map[string]interface{}{"versioning": true},
			after:    map[string]interface{}{"versioning": false},
			expected: true,
		},
		{
			name:     "backup retention reduced to zero",
			before:   map[string]interface{}{"backup_retention_days": float64(30)},
			after:    map[string]interface{}{"backup_retention_days": float64(0)},
			expected: true,
		},
		{
			name:     "backup retention increased",
			before:   map[string]interface{}{"backup_retention_days": float64(7)},
			after:    map[string]interface{}{"backup_retention_days": float64(30)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasBackupDisabled(tt.before, tt.after)
			if result != tt.expected {
				t.Errorf("hasBackupDisabled() = %v, want %v", result, tt.expected)
			}
		})
	}
}
