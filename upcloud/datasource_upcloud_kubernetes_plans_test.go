package upcloud

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDataSourceUpcloudKubernetesPlans_basic(t *testing.T) {
	basic, err := os.ReadFile("testdata/datasource_upcloud_kubernetes_plans/basic.tf")
	if err != nil {
		t.Fatal(fmt.Errorf("could not load testdata: %w", err))
	}

	var providers []*schema.Provider

	resourceName := "data.upcloud_kubernetes_plans.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: string(basic),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "plans"),
				),
				Destroy: false,
			},
		},
	})
}
