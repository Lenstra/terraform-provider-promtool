package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/rulefmt"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &CheckRulesFunction{}

type CheckRulesFunction struct {
}

func NewCheckRulesFunction() function.Function {
	return &CheckRulesFunction{}
}

func (f *CheckRulesFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "check_rules"
}

func (f *CheckRulesFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Validate Prometheus configuration",
		Description: "This function validates a Prometheus configuration file.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "config",
				Description: "prometheus-config",
			},
		},
		Return: function.BoolReturn{},
	}
}

func (f *CheckRulesFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var content string
	if resp.Error = req.Arguments.Get(ctx, &content); resp.Error != nil {
		return
	}

	err := CheckRules(content, resp)
	if err {
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, true))
}

func CheckRules(content string, resp *function.RunResponse) bool {
	rgs, errs := rulefmt.Parse([]byte(content))
	for _, e := range errs {
		if e != nil {
			resp.Error = function.ConcatFuncErrors(resp.Error, &function.FuncError{Text: e.Error()})
			return true
		}
	}

	_, errs = checkRuleGroups(rgs, newLintConfig("all", true))
	for _, e := range errs {
		if e != nil {
			resp.Error = function.ConcatFuncErrors(resp.Error, &function.FuncError{Text: e.Error()})
			return true
		}
	}
	return false
}

type lintConfig struct {
	all            bool
	duplicateRules bool
	fatal          bool
}

func newLintConfig(stringVal string, fatal bool) lintConfig {
	items := strings.Split(stringVal, ",")
	ls := lintConfig{
		fatal: fatal,
	}
	for _, setting := range items {
		switch setting {
		case lintOptionAll:
			ls.all = true
		case lintOptionDuplicateRules:
			ls.duplicateRules = true
		case lintOptionNone:
		default:
			fmt.Printf("WARNING: unknown lint option %s\n", setting)
		}
	}
	return ls
}

func (ls lintConfig) lintDuplicateRules() bool {
	return ls.all || ls.duplicateRules
}

func checkRuleGroups(rgs *rulefmt.RuleGroups, lintSettings lintConfig) (int, []error) {
	numRules := 0
	for _, rg := range rgs.Groups {
		numRules += len(rg.Rules)
	}

	if lintSettings.lintDuplicateRules() {
		dRules := checkDuplicates(rgs.Groups)
		if len(dRules) != 0 {
			errMessage := fmt.Sprintf("%d duplicate rule(s) found.\n", len(dRules))
			for _, n := range dRules {
				errMessage += fmt.Sprintf("Metric: %s\nLabel(s):\n", n.metric)
				n.label.Range(func(l labels.Label) {
					errMessage += fmt.Sprintf("\t%s: %s\n", l.Name, l.Value)
				})
			}
			errMessage += "Might cause inconsistency while recording expressions"
			return 0, []error{fmt.Errorf("%w %s", fmt.Errorf("lint error"), errMessage)}
		}
	}

	return numRules, nil
}

type compareRuleType struct {
	metric string
	label  labels.Labels
}

type compareRuleTypes []compareRuleType

func (c compareRuleTypes) Len() int           { return len(c) }
func (c compareRuleTypes) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c compareRuleTypes) Less(i, j int) bool { return compare(c[i], c[j]) < 0 }

func compare(a, b compareRuleType) int {
	if res := strings.Compare(a.metric, b.metric); res != 0 {
		return res
	}

	return labels.Compare(a.label, b.label)
}

func ruleMetric(rule rulefmt.RuleNode) string {
	if rule.Alert.Value != "" {
		return rule.Alert.Value
	}
	return rule.Record.Value
}

func checkDuplicates(groups []rulefmt.RuleGroup) []compareRuleType {
	var duplicates []compareRuleType
	var rules compareRuleTypes

	for _, group := range groups {
		for _, rule := range group.Rules {
			rules = append(rules, compareRuleType{
				metric: ruleMetric(rule),
				label:  labels.FromMap(rule.Labels),
			})
		}
	}
	if len(rules) < 2 {
		return duplicates
	}
	sort.Sort(rules)

	last := rules[0]
	for i := 1; i < len(rules); i++ {
		if compare(last, rules[i]) == 0 {
			// Don't add a duplicated rule multiple times.
			if len(duplicates) == 0 || compare(last, duplicates[len(duplicates)-1]) != 0 {
				duplicates = append(duplicates, rules[i])
			}
		}
		last = rules[i]
	}

	return duplicates
}
