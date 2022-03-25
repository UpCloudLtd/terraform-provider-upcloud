package upcloud

import (
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccUpcloudLoadBalancer(t *testing.T) {
	testData, err := ioutil.ReadFile("testdata/upcloud_loadbalancer/loadbalancer.tf")
	if err != nil {
		t.Fatal(err)
	}
	var providers []*schema.Provider
	lbName := "upcloud_loadbalancer.lb"
	dnsName := "upcloud_loadbalancer_resolver.lb_dns_1"
	be1Name := "upcloud_loadbalancer_backend.lb_be_1"
	be2Name := "upcloud_loadbalancer_backend.lb_be_2"
	be1sm1Name := "upcloud_loadbalancer_static_backend_member.lb_be_1_sm_1"
	be1dm1Name := "upcloud_loadbalancer_dynamic_backend_member.lb_be_1_dm_1"
	be2sm1Name := "upcloud_loadbalancer_static_backend_member.lb_be_2_sm_1"
	fe1Name := "upcloud_loadbalancer_frontend.lb_fe_1"
	fe1Rule1Name := "upcloud_loadbalancer_frontend_rule.lb_fe_1_r1"
	cb1Name := "upcloud_loadbalancer_dynamic_certificate_bundle.lb-cb-d1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: string(testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(lbName, "plan", "development"),
					resource.TestCheckResourceAttr(lbName, "zone", "fi-hel2"),
					resource.TestCheckResourceAttr(dnsName, "name", "lb-resolver-1-test"),
					resource.TestCheckResourceAttr(dnsName, "nameservers.#", "2"),
					resource.TestCheckResourceAttr(be1Name, "name", "lb-be-1-test"),
					resource.TestCheckResourceAttr(be1Name, "resolver_name", "lb-resolver-1-test"),
					resource.TestCheckResourceAttr(be2Name, "name", "lb-be-2-test"),
					resource.TestCheckResourceAttr(be2Name, "resolver_name", ""),
					resource.TestCheckResourceAttr(be1sm1Name, "backend_name", "lb-be-1-test"),
					resource.TestCheckResourceAttr(be1sm1Name, "name", "lb-be-1-sm-1-test"),
					resource.TestCheckResourceAttr(be1sm1Name, "ip", "10.0.0.10"),
					resource.TestCheckResourceAttr(be1sm1Name, "port", "8000"),
					resource.TestCheckResourceAttr(be1sm1Name, "weight", "100"),
					resource.TestCheckResourceAttr(be1sm1Name, "max_sessions", "1000"),
					resource.TestCheckResourceAttr(be1sm1Name, "enabled", "true"),
					resource.TestCheckResourceAttr(be1dm1Name, "backend_name", "lb-be-1-test"),
					resource.TestCheckResourceAttr(be1dm1Name, "name", "lb-be-1-dm-1-test"),
					resource.TestCheckResourceAttr(be1dm1Name, "weight", "10"),
					resource.TestCheckResourceAttr(be1dm1Name, "max_sessions", "10"),
					resource.TestCheckResourceAttr(be1dm1Name, "enabled", "false"),
					resource.TestCheckResourceAttr(be2sm1Name, "backend_name", "lb-be-2-test"),
					resource.TestCheckResourceAttr(fe1Name, "name", "lb-fe-1-test"),
					resource.TestCheckResourceAttr(fe1Name, "port", "8080"),
					resource.TestCheckResourceAttr(fe1Name, "mode", "http"),
					resource.TestCheckResourceAttr(fe1Name, "default_backend_name", "lb-be-1-test"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "name", "lb-fe-1-r1-test"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "priority", "10"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.src_port.0.method", "equal"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.src_port.0.value", "80"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.src_ip.0.value", "192.168.0.0/24"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.body_size.0.method", "equal"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.body_size.0.value", "8000"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.path.0.value", "/application"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.path.0.method", "starts"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.path.0.ignore_case", "true"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.url.0.value", "/application"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.url.0.method", "starts"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.url.0.ignore_case", "true"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.url_query.0.method", "starts"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.url_query.0.value", "type=app"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.host.0.value", "example.com"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.http_method.0.value", "PATCH"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.cookie.0.value", "123456"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.cookie.0.name", "x-session-id"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.cookie.0.method", "exact"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.header.0.method", "exact"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.header.0.name", "status"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.header.0.value", "active"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.header.0.ignore_case", "true"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.url_param.0.method", "exact"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.url_param.0.name", "status"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.url_param.0.value", "active"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.url_param.0.ignore_case", "true"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.num_members_up.0.backend_name", "lb-be-1-test"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.num_members_up.0.value", "1"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.num_members_up.0.method", "less"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "matchers.0.num_members_up.0.method", "less"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "actions.0.use_backend.0.backend_name", "lb-be-1-test"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "actions.0.http_redirect.0.location", "/app"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "actions.0.http_return.0.status", "404"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "actions.0.http_return.0.content_type", "text/plain"),
					resource.TestCheckResourceAttr(fe1Rule1Name, "actions.0.http_return.0.payload", "UmVzb3VyY2Ugbm90IGZvdW5kIQ=="),
					resource.TestCheckResourceAttr(fe1Rule1Name, "actions.0.tcp_reject.0.active", "true"),
					resource.TestCheckResourceAttr(cb1Name, "name", "lb-cb-d1-test"),
					resource.TestCheckResourceAttr(cb1Name, "key_type", "rsa"),
					resource.TestCheckResourceAttr(cb1Name, "hostnames.0", "example.com"),
				),
			},
		},
	})
}
