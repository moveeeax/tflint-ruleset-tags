# aws_resource_missing_tags

Requires that every taggable AWS resource defines a configured set of tag keys,
and optionally that each value matches a regular expression.

## Why

Untagged infrastructure is invisible to cost allocation, ownership, and
lifecycle automation. Catching it in `tflint` — before `terraform plan` — keeps
the policy where the change is, not in a nightly report nobody reads.

## Configuration

```hcl
rule "aws_resource_missing_tags" {
  enabled  = true
  required = ["Environment", "Owner", "CostCenter"]

  # Optional: value must match the given regular expression (Go/RE2 syntax).
  patterns = {
    Environment = "^(prod|staging|dev)$"
    CostCenter  = "^cc-[0-9]+$"
  }

  # Optional: add resource types beyond the built-in taggable set.
  resource_types = ["aws_custom_widget"]

  # Optional: skip resource types you don't want checked.
  exclude = ["aws_cloudtrail"]
}
```

| Name             | Type                | Default | Description                                             |
| ---------------- | ------------------- | ------- | ------------------------------------------------------- |
| `required`       | `list(string)`      | `[]`    | Tag keys every taggable resource must define.           |
| `patterns`       | `map(string)`       | `{}`    | Per-key RE2 pattern the tag value must match.           |
| `resource_types` | `list(string)`      | `[]`    | Extra resource types to treat as taggable.              |
| `exclude`        | `list(string)`      | `[]`    | Resource types to remove from the checked set.          |

With no `required` keys the rule is a no-op.

## Behaviour

- A resource with **no** `tags` argument fails with every required key listed.
- A resource missing **some** keys fails, listing only the absent ones.
- `patterns` are only evaluated when the value resolves to a static string;
  values that depend on unresolved variables are treated as present but
  unvalidated.
- Only resource types in the effective taggable set are checked, so
  non-taggable resources (e.g. `aws_iam_role_policy_attachment`) are ignored.

## Example diagnostic

```
Error: aws_s3_bucket.logs is missing required tags: [Environment, Owner] (aws_resource_missing_tags)
```
