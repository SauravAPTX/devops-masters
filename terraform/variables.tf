variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "devops-masters-2025"
}

variable "github_repo" {
  description = "GitHub repository in format owner/repo"
  type        = string
}


variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}