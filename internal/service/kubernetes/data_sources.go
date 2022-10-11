package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

type kubeconfig struct {
	Clusters       []kubeconfigCluster `yaml:"clusters"`
	Contexts       []kubeconfigContext `yaml:"contexts"`
	CurrentContext string              `yaml:"current-context"`
	Users          []kubeconfigUser    `yaml:"users"`
}

type kubeconfigCluster struct {
	Cluster kubeconfigClusterData `yaml:"cluster"`
	Name    string                `yaml:"name"`
}

type kubeconfigClusterData struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}

type kubeconfigContext struct {
	Context kubeconfigContextData `yaml:"context"`
	Name    string                `yaml:"name"`
}

type kubeconfigContextData struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type kubeconfigUser struct {
	User kubeconfigUserData `yaml:"user"`
	Name string             `yaml:"name"`
}

type kubeconfigUserData struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}

func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Kubernetes cluster details. Please refer to https://www.terraform.io/language/state/sensitive-data to keep the credential data as safe as possible. NOTE: this is an experimental feature in an alpha phase, the resource definition will change in the future.",
		ReadContext: dataSourceClusterRead,
		Schema: map[string]*schema.Schema{
			"client_certificate": {
				Description: clientCertificateDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"client_key": {
				Description: clientKeyDescription,
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"cluster_ca_certificate": {
				Description: clusterCACertificateDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"id": {
				Description: idDescription,
				Type:        schema.TypeString,
				Required:    true,
			},
			"host": {
				Description: hostDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"kubeconfig": {
				Description: kubeconfigDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: nameDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client := meta.(*service.ServiceContext)
	clusterID := d.Get("id").(string)

	s, err := client.GetKubernetesKubeconfig(ctx, &request.GetKubernetesKubeconfigRequest{
		UUID: clusterID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if s == "" {
		return diag.FromErr(fmt.Errorf("kubeconfig is empty: %w", err))
	}

	d.SetId(clusterID)

	err = d.Set("kubeconfig", s)
	if err != nil {
		return diag.FromErr(err)
	}

	k := kubeconfig{}

	err = yaml.Unmarshal([]byte(s), &k)
	if err != nil {
		return diag.FromErr(fmt.Errorf("kubeconfig yaml unmarshal failed: %w", err))
	}

	currentContext := strings.Split(k.CurrentContext, "@")
	err = d.Set("name", currentContext[1])
	if err != nil {
		return diag.FromErr(err)
	}

	for _, v := range k.Clusters {
		if v.Name == currentContext[1] {
			err = d.Set("cluster_ca_certificate", v.Cluster.CertificateAuthorityData)
			if err != nil {
				return diag.FromErr(err)
			}

			err = d.Set("host", v.Cluster.Server)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	for _, v := range k.Users {
		if v.Name == currentContext[0] {
			err = d.Set("client_certificate", v.User.ClientCertificateData)
			if err != nil {
				return diag.FromErr(err)
			}

			err = d.Set("client_key", v.User.ClientKeyData)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func DataSourcePlan() *schema.Resource {
	return &schema.Resource{
		Description: "Pricing plans for node groups. NOTE: this is an experimental feature in an alpha phase, the resource definition will change in the future.",
		ReadContext: dataSourcePlanRead,
		Schema: map[string]*schema.Schema{
			"description": {
				Description: planDescriptionDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: planNameDescription,
				Type:        schema.TypeString,
				Required:    true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{
							"small",
							"medium",
							"large",
						},
						false,
					)),
			},
		},
	}
}

func dataSourcePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client := meta.(*service.ServiceContext)

	plans, err := client.GetKubernetesPlans(ctx, &request.GetKubernetesPlansRequest{})
	if err != nil {
		return diag.FromErr(err)
	}

	if len(plans) == 0 {
		return diag.FromErr(fmt.Errorf("no plans available: %w", err))
	}

	name := d.Get("name").(string)
	err = fmt.Errorf("plan not available: %s", name)

	for _, v := range plans {
		if v.Name == name {
			d.SetId(v.Name)
			err = d.Set("description", v.Description)

			break
		}
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
