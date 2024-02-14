package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v7/upcloud/service"
)

func getTagsAsMap(ctx context.Context, service *service.Service) (map[string]upcloud.Tag, error) {
	currTags := make(map[string]upcloud.Tag)

	resp, err := service.GetTags(ctx)
	if err != nil {
		return nil, err
	}

	for _, tag := range resp.Tags {
		// Convert key to lowercase as tag names are case-insensitive
		currTags[strings.ToLower(tag.Name)] = tag
	}

	return currTags, nil
}

// removeTags removes tags from server and
// deletes the tag if it is not used by any other server.
func removeTags(ctx context.Context, service *service.Service, serverUUID string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	if _, err := service.UntagServer(ctx, &request.UntagServerRequest{
		UUID: serverUUID,
		Tags: tags,
	}); err != nil {
		return err
	}

	currTags, err := getTagsAsMap(ctx, service)
	if err != nil {
		return err
	}

	for _, tagName := range tags {
		// Find tag to be removed
		if tag, ok := currTags[strings.ToLower(tagName)]; ok {
			// Delete tag if it is not used by any servers
			if len(tag.Servers) == 0 {
				if err := service.DeleteTag(ctx, &request.DeleteTagRequest{
					Name: tag.Name,
				}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type tagsExistsWarning struct {
	tags map[string]string
}

func (w tagsExistsWarning) Error() string {
	var duplicates []string
	for k, v := range w.tags {
		duplicates = append(duplicates, fmt.Sprintf("%s = %s", k, v))
	}
	return fmt.Sprintf("some tags already exist with different letter casing, existing names will be used (%s)", strings.Join(duplicates, ", "))
}

// AddTags creates tags that do not yet exist and tags server with given tags.
func addTags(ctx context.Context, service *service.Service, serverUUID string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	// Store current tags in map to easily check if a tag exists
	currTags, err := getTagsAsMap(ctx, service)
	if err != nil {
		return err
	}

	duplicates := make(map[string]string)

	// Create tags that do not yet exist
	for _, tag := range tags {
		existingTag, ok := currTags[strings.ToLower(tag)]
		if !ok {
			if _, err := service.CreateTag(ctx, &request.CreateTagRequest{
				Tag: upcloud.Tag{
					Name: tag,
				},
			}); err != nil {
				return err
			}
		} else if tag != existingTag.Name {
			duplicates[tag] = existingTag.Name
		}
	}

	if _, err := service.TagServer(ctx, &request.TagServerRequest{
		UUID: serverUUID,
		Tags: tags,
	}); err != nil {
		return err
	}

	if len(duplicates) != 0 {
		return tagsExistsWarning{duplicates}
	}

	return nil
}

func updateTags(ctx context.Context, service *service.Service, serverUUID string, oldTags, newTags []string) error {
	tagsToAdd, tagsToDelete := getTagChange(oldTags, newTags)

	if err := removeTags(ctx, service, serverUUID, tagsToDelete); err != nil {
		return err
	}

	return addTags(ctx, service, serverUUID, tagsToAdd)
}

// getTagChange determines tags to add and delete based on given old and new tags.
// Order of the tags is ignored.
func getTagChange(oldTags, newTags []string) (tagsToAdd, tagsToDelete []string) {
	olgTagsMap := sliceToMap(oldTags)
	newTagsMap := sliceToMap(newTags)

	for tag := range newTagsMap {
		if _, ok := olgTagsMap[tag]; !ok {
			tagsToAdd = append(tagsToAdd, tag)
		}
	}

	for tag := range olgTagsMap {
		if _, ok := newTagsMap[tag]; !ok {
			tagsToDelete = append(tagsToDelete, tag)
		}
	}

	return
}

func tagsHasChange(old, new interface{}) bool {
	// Check how tags would change
	toAdd, toDelete := getTagChange(utils.ExpandStrings(old), utils.ExpandStrings(new))

	// If no tags would be added or deleted, no change will be made
	if len(toAdd) == 0 && len(toDelete) == 0 {
		return false
	}
	return true
}
