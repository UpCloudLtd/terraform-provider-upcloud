package upcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUpcloudServer(t *testing.T) {
	const (
		server1 = "upcloud_server.server1"
		server2 = "upcloud_server.server2"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			testServer1Step1(t, server1),
			testServer2Step1(t, server2),
			testServerStep2(t, server1),
			testServerStep3(t, server1),
		},
	})
}

func testServer1Step1(t *testing.T, name string) resource.TestStep {
	s1, err := renderTerraformConfig("testdata/upcloud_server/s1", map[string]string{"zone": zone})
	if err != nil {
		t.Fatal(err)
	}

	return resource.TestStep{
		ResourceName: name,
		Config:       s1,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(name, "plan", "custom"),
			resource.TestCheckResourceAttr(name, "cpu", "1"),
			resource.TestCheckResourceAttr(name, "mem", "1024"),
			resource.TestCheckResourceAttr(name, "template.0.backup_rule.0.time", "0100"),
			resource.TestCheckResourceAttr(name, "template.0.backup_rule.0.interval", "mon"),
			resource.TestCheckResourceAttr(name, "template.0.backup_rule.0.retention", "2"),
		),
	}
}

func testServer2Step1(t *testing.T, name string) resource.TestStep {
	s1, err := renderTerraformConfig("testdata/upcloud_server/s1", map[string]string{"zone": zone})
	if err != nil {
		t.Fatal(err)
	}

	return resource.TestStep{
		ResourceName: name,
		Config:       s1,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckTypeSetElemNestedAttrs(name, "simple_backup.*", map[string]string{
				"time": "2200",
				"plan": "dailies",
			}),
		),
	}
}

func testServerStep2(t *testing.T, name string) resource.TestStep {
	s2, err := renderTerraformConfig("testdata/upcloud_server/s2", map[string]string{"zone": zone})
	if err != nil {
		t.Fatal(err)
	}

	return resource.TestStep{
		ResourceName: name,
		Config:       s2,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(name, "plan", "custom"),
			resource.TestCheckResourceAttr(name, "cpu", "2"),
			resource.TestCheckResourceAttr(name, "mem", "2048"),
			resource.TestCheckResourceAttr(name, "template.0.backup_rule.0.time", "0200"),
			resource.TestCheckResourceAttr(name, "template.0.backup_rule.0.interval", "fri"),
			resource.TestCheckResourceAttr(name, "template.0.backup_rule.0.retention", "1"),
		),
	}
}

func testServerStep3(t *testing.T, name string) resource.TestStep {
	s2, err := renderTerraformConfig("testdata/upcloud_server/s3", map[string]string{"zone": zone})
	if err != nil {
		t.Fatal(err)
	}

	return resource.TestStep{
		ResourceName: name,
		Config:       s2,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(name, "plan", "1xCPU-1GB"),
		),
	}
}
