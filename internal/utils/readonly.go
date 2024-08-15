package utils

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func planModifiers(attribute schema.Attribute) reflect.Value {
	switch attribute.(type) {
	case schema.BoolAttribute:
		return reflect.ValueOf([]planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		})
	case schema.Int64Attribute:
		return reflect.ValueOf([]planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		})
	case schema.MapAttribute:
		return reflect.ValueOf([]planmodifier.Map{
			mapplanmodifier.UseStateForUnknown(),
		})
	case schema.StringAttribute:
		return reflect.ValueOf([]planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		})
	default:
		panic(fmt.Sprintf("Generating PlanModifiers not implemented for %T", attribute))
	}
}

func asReadonly(attribute schema.Attribute) schema.Attribute {
	output := reflect.New(reflect.TypeOf(attribute))
	output.Elem().Set(reflect.ValueOf(attribute))

	reflect.Indirect(output).FieldByName("Computed").SetBool(true)
	reflect.Indirect(output).FieldByName("Required").SetBool(false)
	reflect.Indirect(output).FieldByName("Optional").SetBool(false)

	reflect.Indirect(output).FieldByName("Default").SetZero()
	reflect.Indirect(output).FieldByName("Validators").SetZero()

	reflect.Indirect(output).FieldByName("PlanModifiers").Set(planModifiers(attribute))

	return output.Interface().(schema.Attribute)
}

func asMap(inputs []string) map[string]bool {
	m := make(map[string]bool)
	for _, i := range inputs {
		m[i] = true
	}
	return m
}

func ReadonlyAttributes(input map[string]schema.Attribute, passthrough ...string) map[string]schema.Attribute {
	output := make(map[string]schema.Attribute)
	passthroughMap := asMap(passthrough)

	for field, attribute := range input {
		if passthroughMap[field] {
			output[field] = attribute
		} else {
			output[field] = asReadonly(attribute)
		}
	}

	return output
}
