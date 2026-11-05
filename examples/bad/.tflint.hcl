plugin "tags" {
  # Loaded from ~/.tflint.d/plugins (no release published yet; see README).
  enabled = true
}

rule "aws_resource_missing_tags" {
  enabled  = true
  required = ["Environment", "Owner"]

  patterns = {
    Environment = "^(prod|staging|dev)$"
  }
}
