package ionoscloud

import (
	"context"
	"fmt"
	ionoscloud "github.com/ionos-cloud/sdk-go/v5"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccNatGatewayRule_Basic(t *testing.T) {
	var natGatewayRule ionoscloud.NatGatewayRule
	natGatewayRuleName := "natGatewayRule"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNatGatewayRuleDestroyCheck,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckNatGatewayRuleConfig_basic, natGatewayRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGatewayRuleExists("ionoscloud_natgateway_rule.natgateway_rule", &natGatewayRule),
					resource.TestCheckResourceAttr("ionoscloud_natgateway_rule.natgateway_rule", "name", natGatewayRuleName),
				),
			},
			{
				Config: testAccCheckNatGatewayRuleConfig_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ionoscloud_natgateway_rule.natgateway_rule", "name", "updated"),
				),
			},
		},
	})
}

func testAccCheckNatGatewayRuleDestroyCheck(s *terraform.State) error {
	client := testAccProvider.Meta().(SdkBundle).Client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ionoscloud_datacenter" {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), *resourceDefaultTimeouts.Delete)

		if cancel != nil {
			defer cancel()
		}

		_, apiResponse, err := client.NATGatewaysApi.DatacentersNatgatewaysRulesDelete(ctx, rs.Primary.Attributes["datacenter_id"], rs.Primary.Attributes["natgateway_id"], rs.Primary.ID).Execute()

		if _, ok := err.(ionoscloud.GenericOpenAPIError); ok {
			if apiResponse.Response.StatusCode != 404 {
				return fmt.Errorf("Nat gateway rule still exists %s %s", rs.Primary.ID, string(apiResponse.Payload))
			}
		} else {
			return fmt.Errorf("Unable to fetch nat gateway rule %s %s", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckNatGatewayRuleExists(n string, natGateway *ionoscloud.NatGatewayRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(SdkBundle).Client
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("testAccCheckNatGatewayRuleExists: Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		ctx, cancel := context.WithTimeout(context.Background(), *resourceDefaultTimeouts.Delete)

		if cancel != nil {
			defer cancel()
		}

		foundNatGatewayRule, _, err := client.NATGatewaysApi.DatacentersNatgatewaysRulesFindByNatGatewayRuleId(ctx, rs.Primary.Attributes["datacenter_id"], rs.Primary.Attributes["natgateway_id"], rs.Primary.ID).Execute()

		if err != nil {
			return fmt.Errorf("Error occured while fetching NatGatewayRule: %s", rs.Primary.ID)
		}
		if *foundNatGatewayRule.Id != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		natGateway = &foundNatGatewayRule

		return nil
	}
}

const testAccCheckNatGatewayRuleConfig_basic = `
resource "ionoscloud_datacenter" "natgateway_datacenter" {
  name              = "test_natgateway"
  location          = "gb/lhr"
  description       = "datacenter for hosting "
}

resource "ionoscloud_lan" "natgateway_lan" {
  datacenter_id = ionoscloud_datacenter.natgateway_datacenter.id
  public        = false
  name          = "test_natgateway_lan"
}

resource "ionoscloud_natgateway" "natgateway" {
  depends_on    = [ ionoscloud_lan.natgateway_lan ]
  datacenter_id = ionoscloud_datacenter.natgateway_datacenter.id
  name          = "natgateway_test"
  public_ips    = [ "77.68.66.153" ]
  lans {
     id          = ionoscloud_lan.natgateway_lan.id
     gateway_ips = [ "10.12.1.2/24"]
  }
}

resource "ionoscloud_natgateway_rule" "natgateway_rule" {
  datacenter_id = ionoscloud_datacenter.natgateway_datacenter.id
  natgateway_id = ionoscloud_natgateway.natgateway.id
  name          = "%s"
  type          = "SNAT"
  protocol      = "TCP"
  source_subnet = "10.0.1.0/24"
  public_ip     = "77.68.66.153"
  target_subnet = "10.0.1.0/24"
  target_port_range {
      start = 500
      end   = 1000
  }
}
`

const testAccCheckNatGatewayRuleConfig_update = `
resource "ionoscloud_datacenter" "natgateway_datacenter" {
  name              = "test_natgateway"
  location          = "gb/lhr"
  description       = "datacenter for hosting "
}

resource "ionoscloud_lan" "natgateway_lan" {
  datacenter_id = ionoscloud_datacenter.natgateway_datacenter.id
  public        = false
  name          = "test_natgateway_lan"
}

resource "ionoscloud_natgateway" "natgateway" {
  depends_on    = [ ionoscloud_lan.natgateway_lan ]
  datacenter_id = ionoscloud_datacenter.natgateway_datacenter.id
  name          = "natgateway_test"
  public_ips    = [ "77.68.66.153" ]
  lans {
     id          = ionoscloud_lan.natgateway_lan.id
     gateway_ips = [ "10.12.1.2/24"]
  }
}

resource "ionoscloud_natgateway_rule" "natgateway_rule" {
  datacenter_id = ionoscloud_datacenter.natgateway_datacenter.id
  natgateway_id = ionoscloud_natgateway.natgateway.id
  name          = "updated"
  type          = "SNAT"
  protocol      = "TCP"
  source_subnet = "10.0.1.0/24"
  public_ip     = "77.68.66.153"
  target_subnet = "10.0.1.0/24"
  target_port_range {
      start = 500
      end   = 1000
  }
}`