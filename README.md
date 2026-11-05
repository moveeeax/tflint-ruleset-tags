# tflint-ruleset-tags

[![ci](https://github.com/moveeeax/tflint-ruleset-tags/actions/workflows/ci.yml/badge.svg)](https://github.com/moveeeax/tflint-ruleset-tags/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.26%2B-00ADD8?logo=go)](https://go.dev/)
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

There is no published release yet, so `tflint --init` has nothing to download.
Build the plugin from source (Go 1.26 or newer) and install it into TFLint's
local plugin directory:

```bash
go install github.com/moveeeax/tflint-ruleset-tags@latest
mkdir -p ~/.tflint.d/plugins
cp "$(go env GOPATH)/bin/tflint-ruleset-tags" ~/.tflint.d/plugins/
```

Or from a clone:

```bash
git clone https://github.com/moveeeax/tflint-ruleset-tags
cd tflint-ruleset-tags
go build -o ~/.tflint.d/plugins/tflint-ruleset-tags .
```

Then enable the plugin in your `.tflint.hcl` **without** `source`/`version`,
so TFLint loads the local binary instead of trying to download a release:

```hcl
plugin "tags" {
  enabled = true
}
```

TFLint also honours `TFLINT_PLUGIN_DIR` if you keep plugins elsewhere.

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

To try a local build, drop the binary in
`~/.tflint.d/plugins/tflint-ruleset-tags` and enable `plugin "tags"` with no
`source` (see [Install](#install)).

## License

[MIT](LICENSE)
