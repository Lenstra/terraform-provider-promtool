// Lenstra - This file contains methods used by the promtool package. - https://github.com/prometheus/prometheus/blob/5a6c8f9c152dfab5f96f5b6f14703b801b014255/cmd/promtool/main.go

// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package promtool

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/rulefmt"
)

func CheckRules(content string, resp *function.RunResponse) bool {
	rgs, errs := rulefmt.Parse([]byte(content), false)
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

func ruleMetric(rule rulefmt.Rule) string {
	if rule.Alert != "" {
		return rule.Alert
	}
	return rule.Record
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
