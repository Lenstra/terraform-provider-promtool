package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

type PromtoolTestCase struct {
	TestFile string
	Expected bool
}

type PromtoolTerraformConfigBuilder func(string) string

func (i *PromtoolTestCase) Run(t *testing.T, makeTerraformConfig PromtoolTerraformConfigBuilder) {
	testFile, err := os.ReadFile(i.TestFile)
	if err != nil {
		t.Fatal(err)
	}

	testStep := []resource.TestStep{
		{
			Config: makeTerraformConfig(string(testFile)),
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
