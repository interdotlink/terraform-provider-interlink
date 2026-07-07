package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories wires the in-process provider into the
// acceptance-test harness under the name "interlink".
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"interlink": providerserver.NewProtocol6WithError(New()),
}

// testAccPreCheck fails fast if the credentials needed for live acceptance
// tests are absent. The provider reads the key from INTERLINK_API_KEY, so the
// test configs can use an empty provider block.
func testAccPreCheck(t *testing.T) {
	if os.Getenv("INTERLINK_API_KEY") == "" {
		t.Fatal("INTERLINK_API_KEY must be set for acceptance tests")
	}
}
