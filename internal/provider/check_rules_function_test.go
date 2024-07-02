package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
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
		tt.Run(t)
	}
}

type PromtoolTestCase struct {
	TestFile string
	Expected bool
}

func (i *PromtoolTestCase) Run(t *testing.T) {
	testFile, err := os.ReadFile(i.TestFile)
	if err != nil {
		t.Fatal(err)
	}

	testStep := []resource.TestStep{
		{
			Config: testAccCheckRulesConfig_basic(string(testFile)),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckOutput("test", fmt.Sprintf("%t", i.Expected)),
			),
		},
	}

	if !i.Expected {
		testStep[0].ExpectError = regexp.MustCompile(".*")
	}

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    testStep,
	})
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
