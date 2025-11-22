# Example: AWS Production Infrastructure

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

# Production VPC
resource "aws_vpc" "production" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "production-vpc"
    Environment = "production"
  }
}

# Database instance
resource "aws_db_instance" "production_db" {
  identifier        = "production-postgres"
  engine            = "postgres"
  engine_version    = "15.3"
  instance_class    = "db.t3.medium"
  allocated_storage = 100

  db_name  = "myapp"
  username = "dbadmin"
  password = "change-me-please" # This should be in a secret manager

  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "mon:04:00-mon:05:00"

  storage_encrypted = true
  multi_az         = true

  tags = {
    Name        = "production-db"
    Environment = "production"
  }
}

# S3 bucket for application data
resource "aws_s3_bucket" "app_data" {
  bucket = "myapp-production-data-2024"

  tags = {
    Name        = "app-data"
    Environment = "production"
  }
}

resource "aws_s3_bucket_versioning" "app_data" {
  bucket = aws_s3_bucket.app_data.id

  versioning_configuration {
    status = "Enabled"
  }
}

# EC2 instances
resource "aws_instance" "web_server" {
  count = 2

  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t3.medium"

  tags = {
    Name        = "web-server-${count.index + 1}"
    Environment = "production"
  }
}
