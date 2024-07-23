package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/lenstra/terraform-provider-promtool/internal/promtool"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &CheckConfigFunction{}

type CheckConfigFunction struct {
}

func NewCheckConfigFunction() function.Function {
	return &CheckConfigFunction{}
}

func (f *CheckConfigFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "check_config"
}

func (f *CheckConfigFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
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

func (f *CheckConfigFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var content string
	if resp.Error = req.Arguments.Get(ctx, &content); resp.Error != nil {
		return
	}

	_, err := promtool.CheckConfig(content, false)

	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, &function.FuncError{Text: err.Error()})
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, true))
}
