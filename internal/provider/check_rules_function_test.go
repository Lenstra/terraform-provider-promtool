package provider

import (
	"fmt"
	"testing"
)

func TestCheckRules(t *testing.T) {
	tests := []PromtoolTestCase{
		{
			TestFile: "./testdata/rules_valid.yml",
			Expected: true,
		},
		{
			TestFile: "./testdata/rules_invalid_duplicate.yml",
			Expected: false,
		},
	}

	for _, tt := range tests {
		tt.Run(t, testAccCheckRulesConfig_basic)
	}
}

func testAccCheckRulesConfig_basic(config string) string {
	return fmt.Sprintf(`
locals {
	config = <<EOT
%s
EOT
}
output "test" {
	value = provider::promtool::check_rules(local.config)
}
`, config)
}
