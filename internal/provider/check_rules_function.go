package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/lenstra/terraform-provider-promtool/internal/promtool"
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

	err := promtool.CheckRules(content, resp)
	if err {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, false))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, true))
}
