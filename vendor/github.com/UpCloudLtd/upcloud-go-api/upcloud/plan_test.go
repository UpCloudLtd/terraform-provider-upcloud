package upcloud

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUnmarshalPlans tests that Plans and Plan objects unmarshal correctly
func TestUnmarshalPlans(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<plans>
    <plan>
        <core_number>1</core_number>
        <memory_amount>1024</memory_amount>
        <name>1xCPU-1GB</name>
        <public_traffic_out>2048</public_traffic_out>
        <storage_size>30</storage_size>
        <storage_tier>maxiops</storage_tier>
    </plan>
    <plan>
        <core_number>2</core_number>
        <memory_amount>2048</memory_amount>
        <name>2xCPU-2GB</name>
        <public_traffic_out>3072</public_traffic_out>
        <storage_size>50</storage_size>
        <storage_tier>maxiops</storage_tier>
    </plan>
    <plan>
        <core_number>4</core_number>
        <memory_amount>4096</memory_amount>
        <name>4xCPU-4GB</name>
        <public_traffic_out>4096</public_traffic_out>
        <storage_size>100</storage_size>
        <storage_tier>maxiops</storage_tier>
    </plan>
    <plan>
        <core_number>6</core_number>
        <memory_amount>8192</memory_amount>
        <name>6xCPU-8GB</name>
        <public_traffic_out>8192</public_traffic_out>
        <storage_size>200</storage_size>
        <storage_tier>maxiops</storage_tier>
    </plan>
</plans>`

	plans := Plans{}
	err := xml.Unmarshal([]byte(originalXML), &plans)
	assert.Nil(t, err)
	assert.Len(t, plans.Plans, 4)

	plan := plans.Plans[0]
	assert.Equal(t, 1, plan.CoreNumber)
	assert.Equal(t, 1024, plan.MemoryAmount)
	assert.Equal(t, "1xCPU-1GB", plan.Name)
	assert.Equal(t, 2048, plan.PublicTrafficOut)
	assert.Equal(t, 30, plan.StorageSize)
	assert.Equal(t, StorageTierMaxIOPS, plan.StorageTier)
}
