package kubernetes

import (
	"context"
	"fmt"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	clientCertificateDescription        = "TLS authentication client certificate, encoded (PEM)."
	clientKeyDescription                = "Key to pair with `client_certificate`, encoded (PEM)."
	clusterCACertificateDescription     = "TLS authentication root certificate bundle, encoded (PEM)."
	clusterStorageEncryptionDescription = "Set default storage encryption strategy for all nodes in the cluster."
	controlPlaneIPFilterDescription     = "IP addresses or IP ranges in CIDR format which are allowed to access the cluster control plane. To allow access from any source, use `[\"0.0.0.0/0\"]`. To deny access from all sources, use `[]`. Values set here do not restrict access to node groups or exposed Kubernetes services."
	hostDescription                     = "Hostname of the cluster API. Defined as URI."
	idDescription                       = "UUID of the cluster."
	kubeconfigDescription               = "Kubernetes config file contents for the cluster."
	nameDescription                     = "Cluster name. Needs to be unique within the account."
	networkDescription                  = "Network ID for the cluster to run in."
	networkCIDRDescription              = "Network CIDR for the given network. Computed automatically."
	nodeGroupNamesDescription           = "Names of the node groups configured to cluster"
	planDescription                     = "The pricing plan used for the cluster. Default plan is `development`. You can list available plans with `upctl kubernetes plans`."
	privateNodeGroupsDescription        = "Enable private node groups. Private node groups requires a network that is routed through NAT gateway."
	stateDescription                    = "Operational state of the cluster."
	versionDescription                  = "Kubernetes version ID, e.g. `1.27`. You can list available version IDs with `upctl kubernetes versions`."
	zoneDescription                     = "Zone in which the Kubernetes cluster will be hosted, e.g. `de-fra1`. You can list available zones with `upctl zone list`."

	cleanupWaitTimeSeconds = 240
	resourceNameMaxLength  = 63
	resourceNameRegexpStr  = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
)

var resourceNameRegexp = regexp.MustCompile(resourceNameRegexpStr)

func getClusterDeleted(ctx context.Context, svc *service.Service, id ...string) (map[string]interface{}, error) {
	c, err := svc.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: id[0]})

	return map[string]interface{}{"resource": "cluster", "name": c.Name, "state": c.State}, err
}

func waitForClusterToBeDeleted(ctx context.Context, svc *service.Service, id string) (diags diag.Diagnostics) {
	err := utils.WaitForResourceToBeDeleted(ctx, svc, getClusterDeleted, id)
	if err != nil {
		diags.AddError("Error waiting for cluster to be deleted", err.Error())
	}
	return
}

var validateResourceName = validation.ToDiagFunc(func(i interface{}, _ string) (warns []string, errs []error) {
	val, ok := i.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("provided value is not a string"))
		return
	}

	if len(val) > resourceNameMaxLength {
		errs = append(errs, fmt.Errorf("resource name (%s) too long, max allowed length is %d", val, resourceNameMaxLength))
		return
	}

	nameValid := resourceNameRegexp.MatchString(val)
	if !nameValid {
		errs = append(errs, fmt.Errorf("name (%s) is not valid. Regular expresion used to check validation: %s", val, resourceNameRegexp))
		return
	}

	return
})
