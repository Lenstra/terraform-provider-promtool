package provider

import (
	"fmt"
	"testing"
)

func TestCheckConfig(t *testing.T) {
	tests := []PromtoolTestCase{
		{
			TestFile: "./testdata/config_valid.yml",
			Expected: true,
		},
		{
			TestFile: "./testdata/config_invalid.yml",
			Expected: false,
		},
	}

	for _, tt := range tests {
		tt.Run(t, testAccCheckConfig_basic)
	}
}

func testAccCheckConfig_basic(config string) string {
	return fmt.Sprintf(`
locals {
	config = <<EOT
%s
EOT
}
output "test" {
	value = provider::promtool::check_config(local.config)
}
`, config)
}
