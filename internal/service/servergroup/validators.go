package servergroup

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func validateTrackMembers(ctx context.Context, d *schema.ResourceDiff, _ interface{}) error {
	members, err := utils.SetOfStringsToSlice(ctx, d.Get("members"))
	if err != nil {
		return fmt.Errorf("parsing members in track_members validation failed: %w", err)
	}

	trackMembers := d.Get("track_members").(bool)

	if !trackMembers && len(members) > 0 {
		return fmt.Errorf("track_members can not be set to false when members set is not empty")
	}

	return nil
}
