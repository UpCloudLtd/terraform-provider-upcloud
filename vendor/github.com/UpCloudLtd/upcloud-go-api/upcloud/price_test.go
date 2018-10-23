package upcloud

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUmarshalPriceZones tests that PrizeZones, PriceZone and Price are unmarshaled correctly
func TestUmarshalPriceZones(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="utf-8"?>
<prices>
    <zone>
        <firewall>
            <amount>1</amount>
            <price>0.56</price>
        </firewall>
        <io_request_backup>
            <amount>1000000</amount>
            <price>0</price>
        </io_request_backup>
        <io_request_maxiops>
            <amount>1000000</amount>
            <price>0</price>
        </io_request_maxiops>
        <ipv4_address>
            <amount>1</amount>
            <price>0.336</price>
        </ipv4_address>
        <ipv6_address>
            <amount>1</amount>
            <price>0</price>
        </ipv6_address>
        <name>de-fra1</name>
        <public_ipv4_bandwidth_in>
            <amount>1</amount>
            <price>0</price>
        </public_ipv4_bandwidth_in>
        <public_ipv4_bandwidth_out>
            <amount>1</amount>
            <price>5.6</price>
        </public_ipv4_bandwidth_out>
        <public_ipv6_bandwidth_in>
            <amount>1</amount>
            <price>0</price>
        </public_ipv6_bandwidth_in>
        <public_ipv6_bandwidth_out>
            <amount>1</amount>
            <price>5.6</price>
        </public_ipv6_bandwidth_out>
        <server_core>
            <amount>1</amount>
            <price>1.12</price>
        </server_core>
        <server_memory>
            <amount>256</amount>
            <price>0.14</price>
        </server_memory>
        <server_plan_1xCPU-1GB>
            <amount>1</amount>
            <price>1.488</price>
        </server_plan_1xCPU-1GB>
        <server_plan_2xCPU-2GB>
            <amount>1</amount>
            <price>2.976</price>
        </server_plan_2xCPU-2GB>
        <server_plan_4xCPU-4GB>
            <amount>1</amount>
            <price>5.952</price>
        </server_plan_4xCPU-4GB>
        <server_plan_6xCPU-8GB>
            <amount>1</amount>
            <price>11.905</price>
        </server_plan_6xCPU-8GB>
        <storage_backup>
            <amount>1</amount>
            <price>0.0078</price>
        </storage_backup>
        <storage_maxiops>
            <amount>1</amount>
            <price>0.031</price>
        </storage_maxiops>
        <storage_template>
            <amount>1</amount>
            <price>0.031</price>
        </storage_template>
    </zone>
    <zone>
        <firewall>
            <amount>1</amount>
            <price>0.56</price>
        </firewall>
        <io_request_backup>
            <amount>1000000</amount>
            <price>0</price>
        </io_request_backup>
        <io_request_hdd>
            <amount>1000000</amount>
            <price>0</price>
        </io_request_hdd>
        <io_request_maxiops>
            <amount>1000000</amount>
            <price>0</price>
        </io_request_maxiops>
        <io_request_ssd>
            <amount>1000000</amount>
            <price>0</price>
        </io_request_ssd>
        <ipv4_address>
            <amount>1</amount>
            <price>0.336</price>
        </ipv4_address>
        <ipv6_address>
            <amount>1</amount>
            <price>0</price>
        </ipv6_address>
        <name>fi-hel1</name>
        <public_ipv4_bandwidth_in>
            <amount>1</amount>
            <price>0</price>
        </public_ipv4_bandwidth_in>
        <public_ipv4_bandwidth_out>
            <amount>1</amount>
            <price>5.6</price>
        </public_ipv4_bandwidth_out>
        <public_ipv6_bandwidth_in>
            <amount>1</amount>
            <price>0</price>
        </public_ipv6_bandwidth_in>
        <public_ipv6_bandwidth_out>
            <amount>1</amount>
            <price>5.6</price>
        </public_ipv6_bandwidth_out>
        <server_core>
            <amount>1</amount>
            <price>1.24</price>
        </server_core>
        <server_memory>
            <amount>256</amount>
            <price>0.34</price>
        </server_memory>
        <server_plan_1xCPU-1GB>
            <amount>1</amount>
            <price>2.232</price>
        </server_plan_1xCPU-1GB>
        <server_plan_2xCPU-2GB>
            <amount>1</amount>
            <price>4.464</price>
        </server_plan_2xCPU-2GB>
        <server_plan_4xCPU-4GB>
            <amount>1</amount>
            <price>8.928</price>
        </server_plan_4xCPU-4GB>
        <server_plan_6xCPU-8GB>
            <amount>1</amount>
            <price>17.857</price>
        </server_plan_6xCPU-8GB>
        <storage_backup>
            <amount>1</amount>
            <price>0.0078</price>
        </storage_backup>
        <storage_hdd>
            <amount>1</amount>
            <price>0.0145</price>
        </storage_hdd>
        <storage_maxiops>
            <amount>1</amount>
            <price>0.031</price>
        </storage_maxiops>
        <storage_ssd>
            <amount>1</amount>
            <price>0.056</price>
        </storage_ssd>
        <storage_template>
            <amount>1</amount>
            <price>0.031</price>
        </storage_template>
    </zone>
</prices>`

	prizeZones := PrizeZones{}
	err := xml.Unmarshal([]byte(originalXML), &prizeZones)
	assert.Nil(t, err)
	assert.Len(t, prizeZones.PrizeZones, 2)

	zone := prizeZones.PrizeZones[0]
	assert.Equal(t, 1, zone.Firewall.Amount)
	assert.Equal(t, 0.56, zone.Firewall.Price)
	assert.Equal(t, 1000000, zone.IORequestBackup.Amount)
	assert.Equal(t, 0.0, zone.IORequestBackup.Price)

	// TODO: Test the remaining fields
}
