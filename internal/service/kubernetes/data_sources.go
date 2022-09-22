package kubernetes

import (
	"context"
	"fmt"
	"strings"

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
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Name                     string `yaml:"name"`
	Server                   string `yaml:"server"`
}

type kubeconfigContext struct {
	Cluster string `yaml:"cluster"`
	Name    string `yaml:"name"`
	User    string `yaml:"user"`
}

type kubeconfigUser struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
	Name                  string `yaml:"name"`
}

func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Kuberenetes cluster details. Please refer to https://www.terraform.io/language/state/sensitive-data to keep the credential data as safe as possible.",
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

	s, err := client.GetKubernetesKubeconfig(ctx, &request.GetKubernetesKubeconfigRequest{
		UUID: d.Get("id").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if s == "" {
		return diag.FromErr(fmt.Errorf("kubeconfig is empty: %w", err))
	}

	err = d.Set("kubeconfig", s)
	if err != nil {
		return diag.FromErr(err)
	}

	k := kubeconfig{}

	err = yaml.Unmarshal([]byte(s), k)
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
			err = d.Set("cluster_ca_certificate", v.CertificateAuthorityData)
			if err != nil {
				return diag.FromErr(err)
			}

			err = d.Set("host", v.Server)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	for _, v := range k.Users {
		if v.Name == currentContext[0] {
			err = d.Set("client_certificate", v.ClientCertificateData)
			if err != nil {
				return diag.FromErr(err)
			}

			err = d.Set("client_key", v.ClientKeyData)
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

func DataSourcePlans() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePlansRead,
		Schema: map[string]*schema.Schema{
			"plans": {
				Description: plansDescription,
				Type:        schema.TypeMap,
				Elem:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourcePlansRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client := meta.(*service.ServiceContext)

	p, err := client.GetKubernetesPlans(ctx, &request.GetKubernetesPlansRequest{})
	if err != nil {
		return diag.FromErr(err)
	}

	if len(p) == 0 {
		return diag.FromErr(fmt.Errorf("no plans available: %w", err))
	}

	dPlans := make([]map[string]string, len(p))

	for k, v := range p {
		dPlans[k] = map[string]string{
			"description": v.Description,
			"name":        v.Name,
		}
	}

	err = d.Set("plans", dPlans)

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
