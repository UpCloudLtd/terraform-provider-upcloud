package utils

import (
	"fmt"
	"regexp"

	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func LabelsSchema(resource string) *schema.Schema {
	description := fmt.Sprintf("Key-value pairs to classify the %s.", resource)
	return &schema.Schema{
		Description: description,
		Type:        schema.TypeMap,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional:         true,
		ValidateDiagFunc: ValidateLabelsDiagFunc,
	}
}

func LabelsMapToSlice(m map[string]interface{}) []upcloud.Label {
	var labels []upcloud.Label

	for k, v := range m {
		labels = append(labels, upcloud.Label{
			Key:   k,
			Value: v.(string),
		})
	}

	return labels
}

func LabelsSliceToMap(s []upcloud.Label) map[string]string {
	labels := make(map[string]string)

	for _, l := range s {
		labels[l.Key] = l.Value
	}

	return labels
}

var ValidateLabelsDiagFunc = validation.AllDiag(
	validation.MapKeyLenBetween(2, 32),
	validation.MapKeyMatch(regexp.MustCompile("^([a-zA-Z0-9])+([a-zA-Z0-9_-])*$"), ""),
	validation.MapValueLenBetween(0, 255),
)
