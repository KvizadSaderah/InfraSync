package analyzer

import (
	"strings"

	"github.com/kvizadsaderah/infrasync/pkg/parser"
)

// Risk levels for changes
type RiskLevel string

const (
	RiskCritical RiskLevel = "critical"
	RiskHigh     RiskLevel = "high"
	RiskMedium   RiskLevel = "medium"
	RiskLow      RiskLevel = "low"
)

// Warning represents a detected risky change
type Warning struct {
	Level       RiskLevel
	Resource    string
	Type        string
	Message     string
	Explanation string
}

// AnalyzeChanges detects risky operations in the plan
func AnalyzeChanges(summary *parser.PlanSummary) []Warning {
	warnings := make([]Warning, 0)

	for _, change := range summary.Changes {
		// Check for destructive operations
		if change.IsDelete {
			warnings = append(warnings, analyzeDelete(change)...)
		}

		if change.IsReplace {
			warnings = append(warnings, analyzeReplace(change)...)
		}

		// Check for specific risky updates
		if change.IsUpdate {
			warnings = append(warnings, analyzeUpdate(change)...)
		}
	}

	return warnings
}

// analyzeDelete checks for critical deletions
func analyzeDelete(change parser.ResourceChange) []Warning {
	warnings := make([]Warning, 0)

	// Production resources
	if strings.Contains(strings.ToLower(change.Address), "prod") ||
		strings.Contains(strings.ToLower(change.Address), "production") {
		warnings = append(warnings, Warning{
			Level:       RiskCritical,
			Resource:    change.Address,
			Type:        change.Type,
			Message:     "Production resource will be DESTROYED",
			Explanation: "Deleting production resources can cause service outages",
		})
	}

	// Database deletions
	if isDatabase(change.Type) {
		warnings = append(warnings, Warning{
			Level:       RiskCritical,
			Resource:    change.Address,
			Type:        change.Type,
			Message:     "Database will be DESTROYED - data loss risk!",
			Explanation: "Ensure backups are in place before deleting databases",
		})
	}

	// Storage deletions
	if isStorage(change.Type) {
		warnings = append(warnings, Warning{
			Level:       RiskHigh,
			Resource:    change.Address,
			Type:        change.Type,
			Message:     "Storage resource will be DESTROYED - potential data loss",
			Explanation: "Verify data is backed up or migrated before deletion",
		})
	}

	// Network resources
	if isNetwork(change.Type) {
		warnings = append(warnings, Warning{
			Level:       RiskHigh,
			Resource:    change.Address,
			Type:        change.Type,
			Message:     "Network resource will be DESTROYED - connectivity impact",
			Explanation: "This may affect connectivity for other resources",
		})
	}

	return warnings
}

// analyzeReplace checks for risky replacements
func analyzeReplace(change parser.ResourceChange) []Warning {
	warnings := make([]Warning, 0)

	// Database replacements
	if isDatabase(change.Type) {
		warnings = append(warnings, Warning{
			Level:       RiskCritical,
			Resource:    change.Address,
			Type:        change.Type,
			Message:     "Database will be REPLACED - downtime expected!",
			Explanation: "Database recreation causes downtime and potential data loss",
		})
	}

	// Compute instance replacements
	if isCompute(change.Type) {
		warnings = append(warnings, Warning{
			Level:       RiskHigh,
			Resource:    change.Address,
			Type:        change.Type,
			Message:     "Compute instance will be REPLACED - downtime expected",
			Explanation: "Instance recreation causes service interruption",
		})
	}

	// Load balancer replacements
	if isLoadBalancer(change.Type) {
		warnings = append(warnings, Warning{
			Level:       RiskHigh,
			Resource:    change.Address,
			Type:        change.Type,
			Message:     "Load balancer will be REPLACED - traffic interruption",
			Explanation: "This may cause temporary service unavailability",
		})
	}

	return warnings
}

// analyzeUpdate checks for risky attribute updates
func analyzeUpdate(change parser.ResourceChange) []Warning {
	warnings := make([]Warning, 0)

	before := change.Before
	after := change.After

	// Check security group changes
	if isSecurityGroup(change.Type) {
		if hasSecurityGroupWeakening(before, after) {
			warnings = append(warnings, Warning{
				Level:       RiskHigh,
				Resource:    change.Address,
				Type:        change.Type,
				Message:     "Security group rules are being relaxed",
				Explanation: "Review that new permissions don't expose services unnecessarily",
			})
		}
	}

	// Check for encryption disabling
	if hasEncryptionDisabled(before, after) {
		warnings = append(warnings, Warning{
			Level:       RiskCritical,
			Resource:    change.Address,
			Type:        change.Type,
			Message:     "Encryption is being DISABLED",
			Explanation: "Disabling encryption is a serious security risk",
		})
	}

	// Check for backup/versioning disabling
	if hasBackupDisabled(before, after) {
		warnings = append(warnings, Warning{
			Level:       RiskHigh,
			Resource:    change.Address,
			Type:        change.Type,
			Message:     "Backup or versioning is being DISABLED",
			Explanation: "This increases risk of data loss",
		})
	}

	return warnings
}

// Resource type helpers
func isDatabase(resourceType string) bool {
	dbTypes := []string{"aws_db_instance", "aws_rds", "google_sql_database",
		"azurerm_sql_database", "aws_dynamodb_table", "postgresql", "mysql"}
	return containsAny(resourceType, dbTypes)
}

func isStorage(resourceType string) bool {
	storageTypes := []string{"aws_s3_bucket", "google_storage_bucket",
		"azurerm_storage_account", "aws_ebs_volume"}
	return containsAny(resourceType, storageTypes)
}

func isNetwork(resourceType string) bool {
	networkTypes := []string{"aws_vpc", "aws_subnet", "google_compute_network",
		"azurerm_virtual_network"}
	return containsAny(resourceType, networkTypes)
}

func isCompute(resourceType string) bool {
	computeTypes := []string{"aws_instance", "google_compute_instance",
		"azurerm_virtual_machine", "aws_ecs_service", "aws_lambda_function"}
	return containsAny(resourceType, computeTypes)
}

func isLoadBalancer(resourceType string) bool {
	lbTypes := []string{"aws_lb", "aws_elb", "aws_alb", "google_compute_forwarding_rule",
		"azurerm_lb"}
	return containsAny(resourceType, lbTypes)
}

func isSecurityGroup(resourceType string) bool {
	sgTypes := []string{"aws_security_group", "google_compute_firewall",
		"azurerm_network_security_group"}
	return containsAny(resourceType, sgTypes)
}

func containsAny(s string, substrs []string) bool {
	lower := strings.ToLower(s)
	for _, substr := range substrs {
		if strings.Contains(lower, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// Attribute change helpers
func hasSecurityGroupWeakening(before, after map[string]interface{}) bool {
	// Simplified check - in real implementation, would check ingress/egress rules
	// Check if cidr_blocks changed to allow more open access
	beforeCidr := getStringSlice(before, "cidr_blocks")
	afterCidr := getStringSlice(after, "cidr_blocks")

	for _, cidr := range afterCidr {
		if cidr == "0.0.0.0/0" || cidr == "::/0" {
			found := false
			for _, bc := range beforeCidr {
				if bc == cidr {
					found = true
					break
				}
			}
			if !found {
				return true // New open access added
			}
		}
	}

	return false
}

func hasEncryptionDisabled(before, after map[string]interface{}) bool {
	encryptionKeys := []string{"encryption", "encrypted", "enable_encryption", "encryption_enabled"}

	for _, key := range encryptionKeys {
		beforeVal := getBool(before, key)
		afterVal := getBool(after, key)

		if beforeVal && !afterVal {
			return true
		}
	}

	return false
}

func hasBackupDisabled(before, after map[string]interface{}) bool {
	backupKeys := []string{"versioning", "backup_enabled", "enable_backup",
		"backup_retention_days", "backup_retention_period"}

	for _, key := range backupKeys {
		beforeVal := getBool(before, key)
		afterVal := getBool(after, key)

		if beforeVal && !afterVal {
			return true
		}

		// Check for retention period reduction
		beforeNum := getNumber(before, key)
		afterNum := getNumber(after, key)
		if beforeNum > 0 && afterNum == 0 {
			return true
		}
	}

	return false
}

// Helper functions
func getBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getNumber(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		if n, ok := val.(float64); ok {
			return n
		}
	}
	return 0
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if val, ok := m[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return []string{}
}
