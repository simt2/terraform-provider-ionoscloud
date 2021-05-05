package ionoscloud

import (
	"context"
	"fmt"
	ionoscloud "github.com/ionos-cloud/sdk-go/v5"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccNatGateway_Basic(t *testing.T) {
	var natGateway ionoscloud.NatGateway
	natGatewayName := "natGateway"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNatGatewayDestroyCheck,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckNatGatewayConfig_basic, natGatewayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGatewayExists("ionoscloud_natgateway.natgateway", &natGateway),
					resource.TestCheckResourceAttr("ionoscloud_natgateway.natgateway", "name", natGatewayName),
				),
			},
			{
				Config: testAccCheckNatGatewayConfig_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ionoscloud_natgateway.natgateway", "name", "updated"),
				),
			},
		},
	})
}

func testAccCheckNatGatewayDestroyCheck(s *terraform.State) error {
	client := testAccProvider.Meta().(SdkBundle).Client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ionoscloud_datacenter" {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), *resourceDefaultTimeouts.Delete)

		if cancel != nil {
			defer cancel()
		}

		_, apiResponse, err := client.NATGatewaysApi.DatacentersNatgatewaysFindByNatGatewayId(ctx, rs.Primary.Attributes["datacenter_id"], rs.Primary.ID).Execute()

		if _, ok := err.(ionoscloud.GenericOpenAPIError); ok {
			if apiResponse.Response.StatusCode != 404 {
				return fmt.Errorf("Nat gateway still exists %s %s", rs.Primary.ID, string(apiResponse.Payload))
			}
		} else {
			return fmt.Errorf("Unable to fetch nat gateway %s %s", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckNatGatewayExists(n string, natGateway *ionoscloud.NatGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(SdkBundle).Client
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("testAccCheckNatGatewayExists: Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		ctx, cancel := context.WithTimeout(context.Background(), *resourceDefaultTimeouts.Delete)

		if cancel != nil {
			defer cancel()
		}

		foundNatGateway, _, err := client.NATGatewaysApi.DatacentersNatgatewaysFindByNatGatewayId(ctx, rs.Primary.Attributes["datacenter_id"], rs.Primary.ID).Execute()

		if err != nil {
			return fmt.Errorf("Error occured while fetching NatGateway: %s", rs.Primary.ID)
		}
		if *foundNatGateway.Id != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		natGateway = &foundNatGateway

		return nil
	}
}

const testAccCheckNatGatewayConfig_basic = `
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
  datacenter_id = ionoscloud_datacenter.natgateway_datacenter.id
  name          = "%s" 
  public_ips    = [ "77.68.66.153" ]
  lans {
     id          = ionoscloud_lan.natgateway_lan.id
  }
}`

const testAccCheckNatGatewayConfig_update = `
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
  datacenter_id = ionoscloud_datacenter.natgateway_datacenter.id
  name          = "updated" 
  public_ips    = [ "77.68.66.153"]
  lans {
     id          = ionoscloud_lan.natgateway_lan.id
  }
}`