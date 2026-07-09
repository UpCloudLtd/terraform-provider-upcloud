package managedobjectstoragetests

import (
	"testing"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const (
	objsto     = "upcloud_managed_object_storage.this"
	bucket     = "upcloud_managed_object_storage_bucket.this"
	staticSite = "upcloud_managed_object_storage_static_site.this"
)

func TestAccUpcloudManagedObjectStorageStaticSite(t *testing.T) {
	testDataS1 := utils.ReadTestDataFile(t, "testdata/managed_object_storage_static_site_s1.tf")
	testDataS2 := utils.ReadTestDataFile(t, "testdata/managed_object_storage_static_site_s2.tf")
	staticSiteDomainName := ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { upcloud.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: upcloud.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataS1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(objsto, "configured_status", "started"),
					resource.TestCheckResourceAttr(bucket, "name", "website"),
					resource.TestCheckResourceAttr(staticSite, "bucket_name", "website"),
					resource.TestCheckResourceAttr(staticSite, "bucket_prefix", ""),
					resource.TestCheckResourceAttr(staticSite, "index_document", "index.html"),
					resource.TestCheckResourceAttr(staticSite, "spa_mode", "false"),
					resource.TestCheckResourceAttr(staticSite, "enabled", "true"),
					resource.TestCheckResourceAttrSet(staticSite, "domain_name"),
					upcloud.CheckStringDoesNotChange(staticSite, "domain_name", &staticSiteDomainName),
				),
			},
			{
				Config:   testDataS1,
				PlanOnly: true,
			},
			{
				Config:            testDataS1,
				ResourceName:      staticSite,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testDataS2,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(staticSite, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(staticSite, "bucket_prefix", "public/"),
					resource.TestCheckResourceAttr(staticSite, "spa_mode", "true"),
					resource.TestCheckResourceAttr(staticSite, "enabled", "true"),
					upcloud.CheckStringDoesNotChange(staticSite, "domain_name", &staticSiteDomainName),
					resource.TestCheckResourceAttr(staticSite, "error_pages.#", "1"),
					resource.TestCheckResourceAttr(staticSite, "error_pages.0.error_document", "errors/404.html"),
					resource.TestCheckResourceAttr(staticSite, "error_pages.0.status_code", "404"),
				),
			},
		},
	})
}
