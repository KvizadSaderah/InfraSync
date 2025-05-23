terraform {
  required_providers {
    time = {
      source = "hashicorp/time"
      version = "~> 0.7"
    }
  }
}

resource "time_static" "example" {
  rfc3339 = "2024-01-02T12:30:00Z"
}

// Для демонстрации различных действий
resource "null_resource" "to_be_created" {} 