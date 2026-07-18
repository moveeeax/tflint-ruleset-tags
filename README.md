# tflint-ruleset-tags

> Never ship an untagged AWS resource again — enforce your tagging policy in CI.

**Status:** 🚧 In development

## Overview

Custom tflint plugin that enforces a required set of tags on taggable AWS resources.

## Features

- Implement the tflint plugin SDK ruleset interface
- Rule: every taggable resource must carry a configurable set of required tag keys
- Configurable required keys and per-key allowed-value patterns via `.tflint.hcl`
- Emit precise diagnostics with resource address and missing keys
- Ship prebuilt binaries and an install snippet

## Stack

Go 1.22, `terraform-linters/tflint-plugin-sdk`, HCL config.

## Usage

```bash
plugin "tags" {
  enabled = true
  source  = "github.com/cybercapybara/tflint-ruleset-tags"
}
tflint --init && tflint
```

## License

MIT
