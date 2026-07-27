# This module trips the rule three ways:
#   - aws_instance.web is missing the Owner tag
#   - aws_s3_bucket.logs has no tags block at all
#   - aws_sqs_queue.jobs has an Environment value that violates the pattern
resource "aws_instance" "web" {
  ami           = "ami-0123456789abcdef0"
  instance_type = "t3.micro"

  tags = {
    Environment = "prod"
  }
}

resource "aws_s3_bucket" "logs" {
  bucket = "example-logs"
}

resource "aws_sqs_queue" "jobs" {
  name = "jobs"

  tags = {
    Environment = "production"
    Owner       = "data-team"
  }
}
