package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformInfrastructure(t *testing.T) {
	t.Parallel()

	// Generate a unique identifier for this test run
	uniqueID := random.UniqueId()
	projectName := fmt.Sprintf("devops-masters-test-%s", uniqueID)

	// Get required environment variables
	awsRegion := getEnvVar(t, "AWS_DEFAULT_REGION", "ap-south-1")
	githubRepo := getEnvVar(t, "GITHUB_REPO", "")
	githubToken := getEnvVar(t, "GITHUB_TOKEN", "")

	// Skip test if required env vars are not set
	if githubRepo == "" || githubToken == "" {
		t.Skip("Skipping test: GITHUB_REPO and GITHUB_TOKEN environment variables must be set")
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../terraform",
		Vars: map[string]interface{}{
			"aws_region":   awsRegion,
			"project_name": projectName,
			"github_repo":  githubRepo,
			"github_token": githubToken,
			"environment":  "test",
		},
		// Add retry configuration
		RetryableTerraformErrors: map[string]string{
			".*": "Terraform command failed",
		},
		MaxRetries:         3,
		TimeBetweenRetries: 5 * time.Second,
		// Set backend configuration to avoid state conflicts
		BackendConfig: map[string]interface{}{
			"key": fmt.Sprintf("test/terraform-%s.tfstate", uniqueID),
		},
	})

	// Ensure cleanup happens even if test fails
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Test panicked: %v", r)
		}
		// Add retry logic for destroy operation
		retry.DoWithRetry(t, "terraform destroy", 3, 10*time.Second, func() (string, error) {
			terraform.Destroy(t, terraformOptions)
			return "", nil
		})
	}()

	// Initialize and validate Terraform configuration
	terraform.Init(t, terraformOptions)
	terraform.Validate(t, terraformOptions)

	// Plan and apply with proper error handling
	terraform.Plan(t, terraformOptions)
	terraform.Apply(t, terraformOptions)

	// Test outputs with retries to handle eventual consistency
	retry.DoWithRetry(t, "check pipeline output", 3, 5*time.Second, func() (string, error) {
		pipelineName := terraform.Output(t, terraformOptions, "pipeline_name")
		expectedPipelineName := fmt.Sprintf("%s-pipeline", projectName)
		
		if !assert.Contains(t, pipelineName, expectedPipelineName) {
			return "", fmt.Errorf("pipeline name %s does not contain expected %s", pipelineName, expectedPipelineName)
		}
		return "", nil
	})

	retry.DoWithRetry(t, "check s3 bucket output", 3, 5*time.Second, func() (string, error) {
		s3BucketName := terraform.Output(t, terraformOptions, "s3_bucket_name")
		expectedBucketPrefix := fmt.Sprintf("%s-codepipeline-artifacts", projectName)
		
		if !assert.Contains(t, s3BucketName, expectedBucketPrefix) {
			return "", fmt.Errorf("s3 bucket name %s does not contain expected %s", s3BucketName, expectedBucketPrefix)
		}
		return "", nil
	})

	retry.DoWithRetry(t, "check codebuild project output", 3, 5*time.Second, func() (string, error) {
		codebuildProjectName := terraform.Output(t, terraformOptions, "codebuild_project_name")
		expectedProjectName := fmt.Sprintf("%s-build", projectName)
		
		if !assert.Equal(t, expectedProjectName, codebuildProjectName) {
			return "", fmt.Errorf("codebuild project name %s does not match expected %s", codebuildProjectName, expectedProjectName)
		}
		return "", nil
	})
}

func TestS3BucketConfiguration(t *testing.T) {
	t.Parallel()

	// Generate a unique identifier for this test run
	uniqueID := random.UniqueId()
	projectName := fmt.Sprintf("devops-masters-s3-test-%s", uniqueID)

	// Get required environment variables
	awsRegion := getEnvVar(t, "AWS_DEFAULT_REGION", "ap-south-1")
	githubRepo := getEnvVar(t, "GITHUB_REPO", "")
	githubToken := getEnvVar(t, "GITHUB_TOKEN", "")

	// Skip test if required env vars are not set
	if githubRepo == "" || githubToken == "" {
		t.Skip("Skipping test: GITHUB_REPO and GITHUB_TOKEN environment variables must be set")
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../terraform",
		Vars: map[string]interface{}{
			"aws_region":   awsRegion,
			"project_name": projectName,
			"github_repo":  githubRepo,
			"github_token": githubToken,
			"environment":  "test",
		},
		RetryableTerraformErrors: map[string]string{
			".*": "Terraform command failed",
		},
		MaxRetries:         3,
		TimeBetweenRetries: 5 * time.Second,
		BackendConfig: map[string]interface{}{
			"key": fmt.Sprintf("test/terraform-s3-%s.tfstate", uniqueID),
		},
	})

	// Ensure cleanup happens
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Test panicked: %v", r)
		}
		retry.DoWithRetry(t, "terraform destroy", 3, 10*time.Second, func() (string, error) {
			terraform.Destroy(t, terraformOptions)
			return "", nil
		})
	}()

	terraform.Init(t, terraformOptions)
	terraform.Validate(t, terraformOptions)
	terraform.Apply(t, terraformOptions)

	// Verify S3 bucket exists and has correct configuration
	s3BucketName := terraform.Output(t, terraformOptions, "s3_bucket_name")
	require.NotEmpty(t, s3BucketName, "S3 bucket name should not be empty")

	// Additional AWS-specific validations
	retry.DoWithRetry(t, "verify s3 bucket exists", 5, 10*time.Second, func() (string, error) {
		// Verify the bucket actually exists in AWS
		bucketExists := aws.S3BucketExists(t, awsRegion, s3BucketName)
		if !bucketExists {
			return "", fmt.Errorf("S3 bucket %s does not exist in AWS", s3BucketName)
		}
		return "", nil
	})

	// Test bucket versioning (if configured in your Terraform)
	retry.DoWithRetry(t, "check bucket versioning", 3, 5*time.Second, func() (string, error) {
		versioning := aws.GetS3BucketVersioning(t, awsRegion, s3BucketName)
		t.Logf("S3 bucket %s versioning status: %s", s3BucketName, versioning)
		return "", nil
	})
}

// Helper function to get environment variables with fallback
func getEnvVar(t *testing.T, key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if fallback != "" {
		t.Logf("Using fallback value for %s", key)
		return fallback
	}
	return ""
}

// TestTerraformPlan - Test that runs only terraform plan (useful for PR validation)
func TestTerraformPlan(t *testing.T) {
	t.Parallel()

	uniqueID := random.UniqueId()
	projectName := fmt.Sprintf("devops-masters-plan-test-%s", uniqueID)

	awsRegion := getEnvVar(t, "AWS_DEFAULT_REGION", "ap-south-1")
	githubRepo := getEnvVar(t, "GITHUB_REPO", "")
	githubToken := getEnvVar(t, "GITHUB_TOKEN", "")

	if githubRepo == "" || githubToken == "" {
		t.Skip("Skipping test: GITHUB_REPO and GITHUB_TOKEN environment variables must be set")
	}

	terraformOptions := &terraform.Options{
		TerraformDir: "../terraform",
		Vars: map[string]interface{}{
			"aws_region":   awsRegion,
			"project_name": projectName,
			"github_repo":  githubRepo,
			"github_token": githubToken,
			"environment":  "test",
		},
		BackendConfig: map[string]interface{}{
			"key": fmt.Sprintf("test/terraform-plan-%s.tfstate", uniqueID),
		},
	}

	terraform.Init(t, terraformOptions)
	terraform.Validate(t, terraformOptions)
	
	// Only run plan, don't apply
	planOutput := terraform.Plan(t, terraformOptions)
	
	// Verify plan contains expected resources
	assert.Contains(t, planOutput, "aws_codepipeline")
	assert.Contains(t, planOutput, "aws_s3_bucket")
	assert.Contains(t, planOutput, "aws_codebuild_project")
}