package upcloud

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceUpCloudStorage() *schema.Resource {
	storageAccessTypes := []string{
		upcloud.StorageAccessPublic,
		upcloud.StorageAccessPrivate,
	}
	storageTypes := []string{
		upcloud.StorageTypeNormal,
		upcloud.StorageTypeBackup,
		upcloud.StorageTypeCDROM,
		upcloud.StorageTypeTemplate,
	}

	return &schema.Resource{
		Description: fmt.Sprintf(`
Returns storage resource information based on defined arguments.  

Data object can be used to map storage to other resource based on the ID or just to read some other storage property like zone information.  
Storage types are: %s`, strings.Join(storageTypes, ", ")),

		ReadContext: dataSourceUpCloudStorageRead,
		Schema: map[string]*schema.Schema{
			"type": {
				Description: fmt.Sprintf("Storage type (%s). Use 'favorite' as type to filter storages on the list of favorites.", strings.Join(storageTypes, ", ")),
				Type:        schema.TypeString,
				Required:    true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					append(storageTypes, "favorite"), false)),
			},
			"access_type": {
				Description:      fmt.Sprintf("Storage access type (%s)", strings.Join(storageAccessTypes, ", ")),
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(storageAccessTypes, false)),
			},
			"name_regex": {
				Description:      "Use regular expression to match storage name",
				Type:             schema.TypeString,
				Optional:         true,
				ExactlyOneOf:     []string{"name", "name_regex"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsValidRegExp),
			},
			"name": {
				Description:      "Exact name of the storage (same as title)",
				Type:             schema.TypeString,
				Optional:         true,
				ExactlyOneOf:     []string{"name", "name_regex"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 64)),
			},
			"zone": {
				Description: "The zone in which the storage resides",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"size": {
				Description: "Size of the storage in gigabytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"state": {
				Description: "Current state of the storage",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tier": {
				Description: "Storage tier in use",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"title": {
				Description: "Title of the storage",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"most_recent": {
				Description: "If more than one result is returned, use the most recent storage. This is only useful with private storages. Public storages might give unpredictable results.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func dataSourceUpCloudStorageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var re *regexp.Regexp

	nameRegex, nameRegexExists := d.GetOk("name_regex")
	if nameRegexExists {
		re = regexp.MustCompile(nameRegex.(string))
	}

	storageType := d.Get("type").(string)
	storages, err := svc.GetStorages(
		&request.GetStoragesRequest{Type: storageType})

	if err != nil {
		return diag.FromErr(err)
	}

	name, nameExists := d.GetOk("name")
	zone, zoneExists := d.GetOk("zone")
	accessType, accessTypeExists := d.GetOk("access_type")
	matches := make([]upcloud.Storage, 0)
	for _, storage := range storages.Storages {
		zoneMatch := (!zoneExists || zone == storage.Zone)
		accessTypeMatch := (!accessTypeExists || accessType == storage.Access)
		if nameExists && name == storage.Title && zoneMatch && accessTypeMatch {
			matches = append(matches, storage)
		} else if nameRegexExists && re != nil && re.MatchString(storage.Title) && zoneMatch && accessTypeMatch {
			matches = append(matches, storage)
		}
	}
	if len(matches) < 1 {
		return diag.Errorf("query returned no results")
	}

	if len(matches) > 1 {
		if !d.Get("most_recent").(bool) {
			return diag.Errorf("query returned more than one result")
		}

		hasUnpredictableResults := false
		if accessTypeExists && accessType == upcloud.StorageAccessPublic {
			// sort storages by UUID because public templates are missing 'created' timestamp
			hasUnpredictableResults = true
			sort.Slice(matches, func(i, j int) bool {
				return matches[i].UUID > matches[j].UUID
			})
		} else {
			// sort storages by created timestamp
			sort.Slice(matches, func(i, j int) bool {
				if !hasUnpredictableResults && (matches[i].Created.IsZero() || matches[j].Created.IsZero()) {
					hasUnpredictableResults = true
				}
				return matches[i].Created.Unix() > matches[j].Created.Unix()
			})
		}
		if hasUnpredictableResults {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "using 'most_recent' attribute with public images might give unpredictable results",
			})
		}
	}

	if err := setStorageResourceData(d, &matches[0]); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func setStorageResourceData(d *schema.ResourceData, storage *upcloud.Storage) (err error) {
	d.SetId(storage.UUID)
	if err = d.Set("access_type", storage.Access); err != nil {
		return err
	}
	if err = d.Set("type", storage.Type); err != nil {
		return err
	}
	if err = d.Set("size", storage.Size); err != nil {
		return err
	}
	if err = d.Set("state", storage.State); err != nil {
		return err
	}
	if err = d.Set("tier", storage.Tier); err != nil {
		return err
	}
	if err = d.Set("name", storage.Title); err != nil {
		return err
	}
	if err = d.Set("title", storage.Title); err != nil {
		return err
	}
	return d.Set("zone", storage.Zone)
}
