variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-south-1"
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
variable "subnet_ids" {
  type        = list(string)
  description = "List of subnet IDs for the EKS cluster"
}
variable "ecr_registry" {
  description = "ECR registry URL"
  type        = string
  default     = ""
}

variable "vpc_id" {
  description = "VPC ID for resources"
  type        = string
  default     = ""
}
variable "github_token" {
  description = "GitHub personal access token"
  type        = string
  sensitive   = true
}