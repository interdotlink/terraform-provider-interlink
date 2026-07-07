package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStatusDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "interlink" {}

data "interlink_status" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// status is a computed scalar the API always returns; assert
					// it is set rather than pinning an environment-dependent value.
					resource.TestCheckResourceAttrSet("data.interlink_status.test", "status"),
				),
			},
		},
	})
}
