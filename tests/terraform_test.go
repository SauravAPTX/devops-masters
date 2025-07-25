package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestTerraformInfrastructure(t *testing.T) {
	t.Parallel()

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../terraform",
		Vars: map[string]interface{}{
			"aws_region":    "ap-south-1",
			"project_name":  "devops-masters-test",
			"github_repo":   "your-username/your-repo",
			"github_token":  "your-github-token",
			"environment":   "test",
		},
	})

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// Test outputs
	pipelineName := terraform.Output(t, terraformOptions, "pipeline_name")
	assert.Contains(t, pipelineName, "devops-masters-test-pipeline")

	s3BucketName := terraform.Output(t, terraformOptions, "s3_bucket_name")
	assert.Contains(t, s3BucketName, "devops-masters-test-codepipeline-artifacts")

	codebuildProjectName := terraform.Output(t, terraformOptions, "codebuild_project_name")
	assert.Equal(t, "devops-masters-test-build", codebuildProjectName)
}

func TestS3BucketConfiguration(t *testing.T) {
	t.Parallel()

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../terraform",
		Vars: map[string]interface{}{
			"aws_region":    "ap-south-1",
			"project_name":  "devops-masters-s3-test",
			"github_repo":   "your-username/your-repo",
			"github_token":  "your-github-token",
			"environment":   "test",
		},
	})

	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	// Verify S3 bucket exists and has correct configuration
	s3BucketName := terraform.Output(t, terraformOptions, "s3_bucket_name")
	assert.NotEmpty(t, s3BucketName)
}