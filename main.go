// Command tflint-ruleset-tags is a TFLint plugin that enforces a required set
// of tags on taggable AWS resources.
package main

import (
	"github.com/moveeeax/tflint-ruleset-tags/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "0.2.0"

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &tflint.BuiltinRuleSet{
			Name:    "tags",
			Version: version,
			Rules: []tflint.Rule{
				rules.NewAwsResourceMissingTagsRule(),
			},
		},
	})
}
