package server

import "fmt"

type storageDeviceSet struct {
	devices map[string]storageDeviceModel
}

func storageDeviceKey(device storageDeviceModel) string {
	return fmt.Sprintf(
		"%s-%s:%s",
		device.Storage.ValueString(),
		device.Address.ValueString(),
		device.AddressPosition.ValueString(),
	)
}

func newStorageDeviceSet(storageDevices []storageDeviceModel) *storageDeviceSet {
	devices := make(map[string]storageDeviceModel)
	for _, device := range storageDevices {
		devices[storageDeviceKey(device)] = device
	}

	return &storageDeviceSet{
		devices: devices,
	}
}

func (s *storageDeviceSet) includes(device storageDeviceModel) bool {
	_, ok := s.devices[storageDeviceKey(device)]
	return ok
}
