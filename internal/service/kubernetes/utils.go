package kubernetes

import (
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func StorageEncryptionAttribute(description string, computed bool) *schema.StringAttribute {
	return &schema.StringAttribute{
		Description: description,
		Computed:    computed,
		Optional:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Validators: []validator.String{
			stringvalidator.OneOf(string(upcloud.StorageEncryptionDataAtRest)),
		},
	}
}
