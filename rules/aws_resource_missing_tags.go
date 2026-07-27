package rules

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// AwsResourceMissingTagsRule checks that every taggable AWS resource carries a
// configured set of required tag keys, optionally validating each value against
// a regular expression.
type AwsResourceMissingTagsRule struct {
	tflint.DefaultRule
}

// awsResourceMissingTagsConfig is the `.tflint.hcl` configuration for the rule.
type awsResourceMissingTagsConfig struct {
	// Required is the set of tag keys every taggable resource must define.
	Required []string `hclext:"required,optional"`
	// Patterns maps a tag key to a regular expression its value must match.
	Patterns map[string]string `hclext:"patterns,optional"`
	// ResourceTypes adds resource types to the built-in taggable set.
	ResourceTypes []string `hclext:"resource_types,optional"`
	// Exclude removes resource types from the checked set.
	Exclude []string `hclext:"exclude,optional"`
}

// NewAwsResourceMissingTagsRule returns a new rule instance.
func NewAwsResourceMissingTagsRule() *AwsResourceMissingTagsRule {
	return &AwsResourceMissingTagsRule{}
}

// Name returns the rule name.
func (r *AwsResourceMissingTagsRule) Name() string {
	return "aws_resource_missing_tags"
}

// Enabled returns whether the rule is enabled by default.
func (r *AwsResourceMissingTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity.
func (r *AwsResourceMissingTagsRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the documentation URL for the rule.
func (r *AwsResourceMissingTagsRule) Link() string {
	return "https://github.com/moveeeax/tflint-ruleset-tags/blob/main/docs/aws_resource_missing_tags.md"
}

// Check runs the rule against the parsed module.
func (r *AwsResourceMissingTagsRule) Check(runner tflint.Runner) error {
	config := awsResourceMissingTagsConfig{}
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

	// With no required keys there is nothing to enforce.
	if len(config.Required) == 0 {
		return nil
	}

	patterns, err := compilePatterns(config.Patterns)
	if err != nil {
		return err
	}

	taggable := taggableSet(config.ResourceTypes, config.Exclude)

	body, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "tags"}},
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, resource := range body.Blocks {
		resourceType := resource.Labels[0]
		if _, ok := taggable[resourceType]; !ok {
			continue
		}
		address := fmt.Sprintf("%s.%s", resourceType, resource.Labels[1])

		attr, defined := resource.Body.Attributes["tags"]
		if !defined {
			if err := r.emitMissing(runner, address, config.Required, resource.DefRange); err != nil {
				return err
			}
			continue
		}

		keys, values := tagKeys(runner, attr.Expr)

		var missing []string
		for _, key := range config.Required {
			if _, present := keys[key]; !present {
				missing = append(missing, key)
			}
		}
		if len(missing) > 0 {
			if err := r.emitMissing(runner, address, missing, attr.Expr.Range()); err != nil {
				return err
			}
		}

		// Validate values against configured patterns when the value is known.
		for _, key := range config.Required {
			re, ok := patterns[key]
			if !ok {
				continue
			}
			value, known := values[key]
			if !known {
				continue
			}
			if !re.MatchString(value) {
				msg := fmt.Sprintf(
					"%s: tag %q value %q does not match required pattern %q",
					address, key, value, re.String(),
				)
				if err := runner.EmitIssue(r, msg, attr.Expr.Range()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *AwsResourceMissingTagsRule) emitMissing(runner tflint.Runner, address string, missing []string, rng hcl.Range) error {
	sorted := append([]string(nil), missing...)
	sort.Strings(sorted)
	msg := fmt.Sprintf(
		"%s is missing required tags: [%s]",
		address, strings.Join(sorted, ", "),
	)
	return runner.EmitIssue(r, msg, rng)
}

// taggableSet builds the effective set of taggable resource types from the
// built-in defaults plus any extra types, minus excluded types.
func taggableSet(extra, exclude []string) map[string]struct{} {
	set := make(map[string]struct{}, len(builtinTaggable)+len(extra))
	for k := range builtinTaggable {
		set[k] = struct{}{}
	}
	for _, k := range extra {
		set[k] = struct{}{}
	}
	for _, k := range exclude {
		delete(set, k)
	}
	return set
}

func compilePatterns(raw map[string]string) (map[string]*regexp.Regexp, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make(map[string]*regexp.Regexp, len(raw))
	for key, pattern := range raw {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern for tag %q: %w", key, err)
		}
		out[key] = re
	}
	return out, nil
}

// tagKeys extracts the tag keys present in a `tags` expression and, where the
// value is a known static string, its value. It first tries full evaluation
// (handles literals and resolvable references); if that fails it falls back to
// syntactic extraction of the object's keys so that dynamic values (e.g.
// referencing a variable) still count as "present".
func tagKeys(runner tflint.Runner, expr hcl.Expression) (map[string]struct{}, map[string]string) {
	keys := map[string]struct{}{}
	values := map[string]string{}

	var evaluated map[string]string
	if err := runner.EvaluateExpr(expr, &evaluated, nil); err == nil {
		for k, v := range evaluated {
			keys[k] = struct{}{}
			values[k] = v
		}
		return keys, values
	}

	pairs, diags := hcl.ExprMap(expr)
	if diags.HasErrors() {
		return keys, values
	}
	for _, pair := range pairs {
		val, kdiags := pair.Key.Value(nil)
		if kdiags.HasErrors() || val.IsNull() || val.Type() != cty.String {
			continue
		}
		key := val.AsString()
		keys[key] = struct{}{}
		// Try to resolve a static value; leave unknown otherwise.
		var s string
		if err := runner.EvaluateExpr(pair.Value, &s, nil); err == nil {
			values[key] = s
		}
	}
	return keys, values
}
