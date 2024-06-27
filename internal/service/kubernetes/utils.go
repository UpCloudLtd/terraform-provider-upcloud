package kubernetes

import (
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func storageEncryptionSchema(description string, computed bool) *sdkv2_schema.Schema {
	return &sdkv2_schema.Schema{
		Description: description,
		Type:        sdkv2_schema.TypeString,
		Computed:    computed,
		Optional:    true,
		ForceNew:    true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
			string(upcloud.StorageEncryptionDataAtReset),
		}, false)),
	}
}

func StorageEncryptionAttribute(description string, computed bool) *schema.StringAttribute {
	return &schema.StringAttribute{
		Description: description,
		Computed:    computed,
		Optional:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Validators: []validator.String{
			stringvalidator.OneOf(string(upcloud.StorageEncryptionDataAtReset)),
		},
	}
}
