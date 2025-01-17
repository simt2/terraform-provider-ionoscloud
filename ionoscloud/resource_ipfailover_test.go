package ionoscloud

import (
	"context"
	"fmt"
	ionoscloud "github.com/ionos-cloud/sdk-go/v5"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccLanIPFailover_Basic(t *testing.T) {
	var lan ionoscloud.Lan
	var ipfailover ionoscloud.IPFailover

	testDeleted := func(n string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			_, ok := s.RootModule().Resources[n]
			if ok {
				return fmt.Errorf("Expected Failover group  %s to be deleted.", n)
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckLanIPFailoverDestroyCheck,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testaccchecklanipfailoverconfigBasic),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLanIPFailoverGroupExists("ionoscloud_ipfailover.failover-test", &lan, &ipfailover),
				),
			},
			{
				Config: testaccchecklanipfailoverconfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testDeleted("ionoscloud_ipfailover.failover-test"),
				),
			},
			{
				Config: `/* */`,
			},
		},
	})
}

func testAccCheckLanIPFailoverGroupExists(n string, _ *ionoscloud.Lan, _ *ionoscloud.IPFailover) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ionoscloud.APIClient)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID` is set")
		}

		dcId := rs.Primary.Attributes["datacenter_id"]
		lanId := rs.Primary.Attributes["lan_id"]
		nicUuid := rs.Primary.Attributes["nicuuid"]

		ctx, cancel := context.WithTimeout(context.Background(), *resourceDefaultTimeouts.Default)

		if cancel != nil {
			defer cancel()
		}

		lan, _, err := client.LanApi.DatacentersLansFindById(ctx, dcId, lanId).Execute()

		if err != nil {
			return fmt.Errorf("lan %s not found %s", lanId, err)
		}

		if lan.Properties.IpFailover == nil {
			return fmt.Errorf("lan %s has no failover groups", lanId)
		}
		found := false
		for _, fo := range *lan.Properties.IpFailover {
			if *fo.NicUuid == nicUuid {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("expected NIC %s to be a part of a failover group, but not found in lans %s failover groups", nicUuid, lanId)
		}

		return nil
	}
}

func testAccCheckLanIPFailoverDestroyCheck(s *terraform.State) error {
	client := testAccProvider.Meta().(*ionoscloud.APIClient)

	ctx, cancel := context.WithTimeout(context.Background(), *resourceDefaultTimeouts.Default)

	if cancel != nil {
		defer cancel()
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ionoscloud_ipfailover" {
			continue
		}

		dcId := rs.Primary.Attributes["datacenter_id"]
		lanId := rs.Primary.Attributes["lan_id"]
		nicUuid := rs.Primary.Attributes["nicuuid"]

		lan, apiResponse, err := client.LanApi.DatacentersLansFindById(ctx, dcId, lanId).Execute()

		if err != nil {
			if apiResponse == nil || apiResponse.StatusCode != 404 {
				return fmt.Errorf("an error occured while fetching a Lan ID %s %s", rs.Primary.Attributes["lan_id"], err)
			}
		} else {
			found := false
			if lan.Properties.IpFailover != nil {
				for _, fo := range *lan.Properties.IpFailover {
					if *fo.NicUuid == nicUuid {
						found = true
					}
				}
				if found {
					_, _, err := client.DataCenterApi.DatacentersDelete(ctx, dcId).Execute()
					if err != nil {
						return fmt.Errorf("IP failover group with nicId %s still exists %s %s, removing datacenter", nicUuid, rs.Primary.ID, err)
					}
				}
			}
		}
	}

	return nil
}

const testaccchecklanipfailoverconfigBasic = `
resource "ionoscloud_datacenter" "foobar" {
	name       = "ipfailover-test"
	location = "us/las"
}

resource "ionoscloud_ipblock" "webserver_ip" {
  location = "us/las"
  size = 1
  name = "failover test"
}

resource "ionoscloud_lan" "webserver_lan1" {
  datacenter_id = "${ionoscloud_datacenter.foobar.id}"
  public = true
  name = "terraform test"
}

resource "ionoscloud_server" "webserver" {
  name = "server"
  datacenter_id = "${ionoscloud_datacenter.foobar.id}"
  cores = 1
  ram = 1024
  availability_zone = "ZONE_1"
  cpu_family = "AMD_OPTERON"
	image_name = "centos:latest"
	image_password = "K3tTj8G14a3EgKyNeeiY"
  volume {
    name = "system"
    size = 5
    disk_type = "SSD"
  }
  nic {
    lan = "${ionoscloud_lan.webserver_lan1.id}"
    dhcp = true
    firewall_active = true
     ips = ["${ionoscloud_ipblock.webserver_ip.ips[0]}"]
  }
}
resource "ionoscloud_ipfailover" "failover-test" {
  datacenter_id = "${ionoscloud_datacenter.foobar.id}"
  lan_id="${ionoscloud_lan.webserver_lan1.id}"
  ip ="${ionoscloud_ipblock.webserver_ip.ips[0]}"
  nicuuid= "${ionoscloud_server.webserver.primary_nic}"
}
`

const testaccchecklanipfailoverconfigUpdate = `
resource "ionoscloud_datacenter" "foobar" {
	name       = "ipfailover-test"
	location = "us/las"
}

resource "ionoscloud_ipblock" "webserver_ip" {
  location = "us/las"
  size = 1
  name = "failover test"
}

resource "ionoscloud_lan" "webserver_lan1" {
  datacenter_id = "${ionoscloud_datacenter.foobar.id}"
  public = true
  name = "terraform test"
}

resource "ionoscloud_server" "webserver" {
  name = "server"
  datacenter_id = "${ionoscloud_datacenter.foobar.id}"
  cores = 1
  ram = 1024
  availability_zone = "ZONE_1"
  cpu_family = "AMD_OPTERON"
	image_name = "centos:latest"
	image_password = "K3tTj8G14a3EgKyNeeiY"
  volume {
    name = "system"
    size = 5
    disk_type = "SSD"
  }
  nic {
    lan = "1"
    dhcp = true
    firewall_active = true
     ips =["${ionoscloud_ipblock.webserver_ip.ips[0]}"]
  }
}
`
