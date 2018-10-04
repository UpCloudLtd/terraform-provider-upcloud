package upcloud

import (
	"fmt"
)

//CPU count has to be positive integer
func validateCPUCount(v interface{}, k string) (ws []string, errors []error) {
	if v.(int) < 1 {
		errors = append(errors, fmt.Errorf(
			"CPU %q must be a positive number", k))
	}
	return
}

func validateMemoryCount(v interface{}, k string) (ws []string, errors []error) {
	if v.(int) < 1 {
		errors = append(errors, fmt.Errorf(
			"Memory %q must be a positive number", k))
	}
	return
}

//CPU count has to be positive integer
func validateSoregeSize(v interface{}, k string) (ws []string, errors []error) {

	if v.(int) < 10 || v.(int) > 1024 {
		errors = append(errors, fmt.Errorf(
			"Storage size %q must be between 10-1024", k))
	}
	return
}
