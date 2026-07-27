package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_AwsResourceMissingTagsRule(t *testing.T) {
	rule := NewAwsResourceMissingTagsRule()

	cases := []struct {
		name     string
		files    map[string]string
		expected helper.Issues
	}{
		{
			name: "all required tags present",
			files: map[string]string{
				".tflint.hcl": config(`required = ["Environment", "Owner"]`),
				"main.tf": `
resource "aws_instance" "web" {
  tags = {
    Environment = "prod"
    Owner       = "platform"
  }
}`,
			},
			expected: helper.Issues{},
		},
		{
			name: "missing one required tag",
			files: map[string]string{
				".tflint.hcl": config(`required = ["Environment", "Owner"]`),
				"main.tf": `
resource "aws_instance" "web" {
  tags = {
    Environment = "prod"
  }
}`,
			},
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: "aws_instance.web is missing required tags: [Owner]",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 10},
						End:      hcl.Pos{Line: 5, Column: 4},
					},
				},
			},
		},
		{
			name: "tags attribute entirely absent",
			files: map[string]string{
				".tflint.hcl": config(`required = ["Environment"]`),
				"main.tf": `
resource "aws_s3_bucket" "logs" {
  bucket = "my-logs"
}`,
			},
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: "aws_s3_bucket.logs is missing required tags: [Environment]",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 32},
					},
				},
			},
		},
		{
			name: "non-taggable resource is ignored",
			files: map[string]string{
				".tflint.hcl": config(`required = ["Environment"]`),
				"main.tf": `
resource "aws_iam_role_policy_attachment" "a" {
  role       = "r"
  policy_arn = "arn"
}`,
			},
			expected: helper.Issues{},
		},
		{
			name: "value pattern violated",
			files: map[string]string{
				".tflint.hcl": config(`
    required = ["Environment"]
    patterns = {
      Environment = "^(prod|staging|dev)$"
    }`),
				"main.tf": `
resource "aws_sqs_queue" "q" {
  tags = {
    Environment = "production"
  }
}`,
			},
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `aws_sqs_queue.q: tag "Environment" value "production" does not match required pattern "^(prod|staging|dev)$"`,
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 10},
						End:      hcl.Pos{Line: 5, Column: 4},
					},
				},
			},
		},
		{
			name: "value pattern satisfied",
			files: map[string]string{
				".tflint.hcl": config(`
    required = ["Environment"]
    patterns = {
      Environment = "^(prod|staging|dev)$"
    }`),
				"main.tf": `
resource "aws_sqs_queue" "q" {
  tags = {
    Environment = "prod"
  }
}`,
			},
			expected: helper.Issues{},
		},
		{
			name: "extra resource_types are enforced",
			files: map[string]string{
				".tflint.hcl": config(`
    required       = ["Environment"]
    resource_types = ["aws_custom_widget"]`),
				"main.tf": `
resource "aws_custom_widget" "w" {
  tags = {}
}`,
			},
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: "aws_custom_widget.w is missing required tags: [Environment]",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 10},
						End:      hcl.Pos{Line: 3, Column: 12},
					},
				},
			},
		},
		{
			name: "excluded resource type is skipped",
			files: map[string]string{
				".tflint.hcl": config(`
    required = ["Environment"]
    exclude  = ["aws_instance"]`),
				"main.tf": `
resource "aws_instance" "web" {
  ami = "ami-123"
}`,
			},
			expected: helper.Issues{},
		},
		{
			name: "no required keys is a no-op",
			files: map[string]string{
				".tflint.hcl": config(``),
				"main.tf": `
resource "aws_instance" "web" {
  ami = "ami-123"
}`,
			},
			expected: helper.Issues{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, tc.files)
			if err := rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			helper.AssertIssues(t, tc.expected, runner.Issues)
		})
	}
}

// config wraps a rule body in the pseudo `.tflint.hcl` rule block the helper
// runner expects.
func config(body string) string {
	return `
rule "aws_resource_missing_tags" {
  enabled = true
` + body + `
}`
}
