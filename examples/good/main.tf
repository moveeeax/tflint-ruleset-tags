# Every taggable resource carries the required tag keys, so `tflint` is clean.
resource "aws_instance" "web" {
  ami           = "ami-0123456789abcdef0"
  instance_type = "t3.micro"

  tags = {
    Environment = "prod"
    Owner       = "platform-team"
    CostCenter  = "cc-1001"
  }
}

resource "aws_s3_bucket" "assets" {
  bucket = "example-assets"

  tags = {
    Environment = "prod"
    Owner       = "web-team"
    CostCenter  = "cc-2002"
  }
}
