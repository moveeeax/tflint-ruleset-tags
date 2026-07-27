# tflint-ruleset-tags

[![ci](https://github.com/moveeeax/tflint-ruleset-tags/actions/workflows/ci.yml/badge.svg)](https://github.com/moveeeax/tflint-ruleset-tags/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.25-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> Never ship an untagged AWS resource again — enforce your tagging policy in CI.

A [TFLint](https://github.com/terraform-linters/tflint) plugin that fails the
lint when a taggable AWS resource is missing required tag keys — or when a tag
value doesn't match the pattern you expect. Catch it before `terraform plan`,
where the change is, instead of in a cost report nobody reads.

## Rules

| Name                                                          | Description                                       | Severity |
| ------------------------------------------------------------ | ------------------------------------------------- | -------- |
| [aws_resource_missing_tags](docs/aws_resource_missing_tags.md) | Required tag keys (and value patterns) per resource | ERROR    |

## Install

Add the plugin to your `.tflint.hcl` and run `tflint --init`:

```hcl
plugin "tags" {
  enabled = true
  source  = "github.com/moveeeax/tflint-ruleset-tags"
  version = "0.1.0"
}
```

`tflint --init` downloads the released binary for your platform and verifies it.

## Configure

```hcl
rule "aws_resource_missing_tags" {
  enabled  = true
  required = ["Environment", "Owner", "CostCenter"]

  # Optional: the value must match this RE2 regular expression.
  patterns = {
    Environment = "^(prod|staging|dev)$"
    CostCenter  = "^cc-[0-9]+$"
  }

  # Optional: extend / trim the built-in taggable resource set.
  resource_types = ["aws_custom_widget"]
  exclude        = ["aws_cloudtrail"]
}
```

| Option           | Type           | Default | Description                                    |
| ---------------- | -------------- | ------- | ---------------------------------------------- |
| `required`       | `list(string)` | `[]`    | Tag keys every taggable resource must define.  |
| `patterns`       | `map(string)`  | `{}`    | Per-key RE2 pattern the tag value must match.  |
| `resource_types` | `list(string)` | `[]`    | Extra resource types to treat as taggable.     |
| `exclude`        | `list(string)` | `[]`    | Resource types to remove from the check.       |

With no `required` keys the rule does nothing.

## How it works

The plugin implements the
[tflint-plugin-sdk](https://github.com/terraform-linters/tflint-plugin-sdk)
ruleset interface. For each `resource` block it:

1. Skips resource types outside the effective **taggable set** (a curated list of
   ~60 common taggable AWS types, plus your `resource_types`, minus `exclude`) —
   so non-taggable resources like `aws_iam_role_policy_attachment` never produce
   false positives.
2. Reads the `tags` argument. A missing argument means every required key is
   missing. A present one is resolved to its keys — first by full evaluation,
   then, if a value references an unresolved variable, by reading the object's
   keys syntactically so it still counts as present.
3. Emits a precise diagnostic naming the resource address and the missing keys,
   and a separate diagnostic when a known value violates its `patterns` regex.

## Usage

```console
$ tflint
Error: aws_s3_bucket.logs is missing required tags: [Environment, Owner] (aws_resource_missing_tags)

  on main.tf line 14:
  14: resource "aws_s3_bucket" "logs" {

Reference: https://github.com/moveeeax/tflint-ruleset-tags/blob/main/docs/aws_resource_missing_tags.md
```

The [`examples/`](examples/) directory has a `good/` module that passes and a
`bad/` module that trips the rule three different ways.

## Develop

```bash
go test ./...        # unit tests (uses the SDK's mock runner)
go build -o tflint-ruleset-tags .
```

To try the local build without a release, drop the binary in
`~/.tflint.d/plugins/tflint-ruleset-tags` and enable `plugin "tags"` with no
`source`.

## License

[MIT](LICENSE)
