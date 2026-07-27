plugin "tags" {
  enabled = true
  source  = "github.com/moveeeax/tflint-ruleset-tags"
  version = "0.1.0"
}

rule "aws_resource_missing_tags" {
  enabled  = true
  required = ["Environment", "Owner"]

  patterns = {
    Environment = "^(prod|staging|dev)$"
  }
}
