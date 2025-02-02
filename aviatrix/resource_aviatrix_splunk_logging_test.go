package aviatrix

import (
	"fmt"
	"os"
	"testing"

	"github.com/AviatrixSystems/terraform-provider-aviatrix/v3/goaviatrix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAviatrixSplunkLogging_basic(t *testing.T) {
	if os.Getenv("SKIP_SPLUNK_LOGGING") == "yes" {
		t.Skip("Skipping splunk logging test as SKIP_SPLUNK_LOGGING is set")
	}

	resourceName := "aviatrix_splunk_logging.test_splunk_logging"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSplunkLoggingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSplunkLoggingBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSplunkLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server", "1.2.3.4"),
					resource.TestCheckResourceAttr(resourceName, "port", "10"),
					testAccCheckSplunkLoggingExcludedGatewaysMatch([]string{"a", "b"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"custom_output_config_file"},
			},
		},
	})
}

func testAccSplunkLoggingBasic() string {
	return `
resource "aviatrix_splunk_logging" "test_splunk_logging" {
	server            = "1.2.3.4"
	port              = 10
	excluded_gateways = ["a", "b"]
}
`
}

func testAccCheckSplunkLoggingExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("splunk logging not found: %s", resourceName)
		}

		client := testAccProvider.Meta().(*goaviatrix.Client)

		_, err := client.GetSplunkLoggingStatus()
		if err == goaviatrix.ErrNotFound {
			return fmt.Errorf("splunk logging not found: %s", resourceName)
		}

		return nil
	}
}

func testAccCheckSplunkLoggingExcludedGatewaysMatch(input []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*goaviatrix.Client)

		resp, _ := client.GetSplunkLoggingStatus()
		if !goaviatrix.Equivalent(resp.ExcludedGateways, input) {
			return fmt.Errorf("excluded gateways don't match with the input")
		}
		return nil
	}
}

func testAccCheckSplunkLoggingDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*goaviatrix.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aviatrix_splunk_logging" {
			continue
		}

		_, err := client.GetSplunkLoggingStatus()
		if err != goaviatrix.ErrNotFound {
			return fmt.Errorf("splunk_logging still exists")
		}
	}

	return nil
}
