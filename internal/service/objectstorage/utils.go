package objectstorage

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getMissing(expected, found []string) []string {
	var missing []string
	for _, expectedName := range expected {
		nameFound := false
		for _, foundName := range found {
			if foundName == expectedName {
				nameFound = true
				break
			}
		}

		if !nameFound {
			if missing == nil {
				missing = make([]string, 0, 1)
			}

			missing = append(missing, expectedName)
		}
	}

	return missing
}

func generateObjectStorageEnvVarKey(prefix, objectStorageName string) string {
	name := strings.ToUpper(strings.Replace(objectStorageName, "-", "_", -1))
	return fmt.Sprintf("%s%s", prefix, name)
}

// Attempts to get access key.
// Second return value is a bool, set to true if key value was retrieved from env variable
func getAccessKey(d *schema.ResourceData) (string, bool, error) {
	configVal := d.Get("access_key").(string)

	// If config value is set to something else then empty string, just use it
	if configVal != "" {
		return configVal, false, nil
	}

	// If config value is empty string, use environment variable
	objectStorageName := d.Get("name").(string)
	envVarKey := generateObjectStorageEnvVarKey(accessKeyEnvVarPrefix, objectStorageName)
	envVarValue, envVarSet := os.LookupEnv(envVarKey)

	if !envVarSet {
		return "", false, fmt.Errorf("access_key config field for object storage %s is set to empty string and environment variable %s is not set", objectStorageName, envVarKey)
	}

	length := len(envVarValue)

	if length < accessKeyMinLength {
		return "", false, fmt.Errorf("access_key set in environment variable %s is too short; minimum length is %d, got %d", envVarKey, accessKeyMinLength, length)
	}

	if length > accessKeyMaxLength {
		return "", false, fmt.Errorf("access_key set in environment variable %s is too long; maximum length is %d, got %d", envVarKey, accessKeyMaxLength, length)
	}

	return envVarValue, true, nil
}

// Attempts to get secret key.
// Second return value is a bool, set to true if key value was revtrived from env variable
func getSecretKey(d *schema.ResourceData) (string, bool, error) {
	configVal := d.Get("secret_key").(string)

	// If config value is set to something else then empty string, just use it
	if configVal != "" {
		return configVal, false, nil
	}

	// If config value is empty string, use environment variable
	objectStorageName := d.Get("name").(string)
	envVarKey := generateObjectStorageEnvVarKey(secretKeyEnvVarPrefix, objectStorageName)
	envVarValue, envVarSet := os.LookupEnv(envVarKey)

	if !envVarSet {
		return "", false, fmt.Errorf("secret_key config field for object storage %s is set to empty string and environment variable %s is not set", objectStorageName, envVarKey)
	}

	length := len(envVarValue)

	if length < secretKeyMinLength {
		return "", false, fmt.Errorf("secret_key set in environment variable %s is too short; minimum length is %d, got %d", envVarKey, secretKeyMinLength, length)
	}

	if length > secretKeyMaxLength {
		return "", false, fmt.Errorf("secret_key set in environment variable %s is too long; maximum length is %d, got %d", envVarKey, secretKeyMaxLength, length)
	}

	return envVarValue, true, nil
}

func getNewAndDeletedBucketNames(d *schema.ResourceData) ([]string, []string) {
	beforeNames := make([]string, 0)
	afterNames := make([]string, 0)

	before, after := d.GetChange(bucketKey)

	for _, item := range before.(*schema.Set).List() {
		valueMap := item.(map[string]interface{})
		beforeNames = append(beforeNames, valueMap["name"].(string))
	}

	for _, item := range after.(*schema.Set).List() {
		valueMap := item.(map[string]interface{})
		afterNames = append(afterNames, valueMap["name"].(string))
	}

	return getMissing(beforeNames, afterNames), getMissing(afterNames, beforeNames)
}
