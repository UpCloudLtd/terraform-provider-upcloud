package properties

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPropertiesMap(t *testing.T) {
	dbTypes := getTypes()

	for _, dbType := range dbTypes {
		t.Run(string(dbType), func(t *testing.T) {
			assert.NotPanics(t, func() {
				GetProperties(dbType)
			})
		})
	}
}
