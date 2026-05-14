/*
** Zabbix
** Copyright (C) 2001-2026 Zabbix SIA
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
**     http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
**/

package handlers //nolint:revive

import (
	"context"
	"encoding/json"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.zabbix.com/plugin/nvidia/internal/plugin/params"
	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
	"golang.zabbix.com/sdk/errs"
)

var (
	_ HandlerFunc = WithJSONResponse(nil)
	_ HandlerFunc = (*Handler)(nil).GetNVMLVersion
	_ HandlerFunc = (*Handler)(nil).GetBAR1MemoryInfo
	_ HandlerFunc = (*Handler)(nil).GetDecoderUtilization
	_ HandlerFunc = (*Handler)(nil).GetDeviceCount
	_ HandlerFunc = (*Handler)(nil).GetDeviceEnergyConsumption
	_ HandlerFunc = (*Handler)(nil).GetDeviceFanSpeed
	_ HandlerFunc = (*Handler)(nil).GetDevicePerfState
	_ HandlerFunc = (*Handler)(nil).GetDevicePowerLimit
	_ HandlerFunc = (*Handler)(nil).GetDevicePowerUsage
	_ HandlerFunc = (*Handler)(nil).GetDeviceSerial
	_ HandlerFunc = (*Handler)(nil).GetDeviceTemperature
	_ HandlerFunc = (*Handler)(nil).GetDriverVersion
	_ HandlerFunc = (*Handler)(nil).GetEncoderStats
	_ HandlerFunc = (*Handler)(nil).GetEncoderUtilization
	_ HandlerFunc = (*Handler)(nil).GetFBMemoryInfo
	_ HandlerFunc = (*Handler)(nil).GetGraphicsFrequency
	_ HandlerFunc = (*Handler)(nil).GetMemoryErrors
	_ HandlerFunc = (*Handler)(nil).GetMemoryFrequency
	_ HandlerFunc = (*Handler)(nil).GetPCIeThroughput
	_ HandlerFunc = (*Handler)(nil).GetRegisterErrors
	_ HandlerFunc = (*Handler)(nil).GetVideoFrequency
	_ HandlerFunc = (*Handler)(nil).GetSMFrequency
)

// HandlerFunc describes the signature all metric handler functions must have.
type HandlerFunc func(
	ctx context.Context,
	metricParams map[string]string,
	extraParams ...string,
) (any, error)

// Handler hold client and syscall implementation for request functions.
type Handler struct {
	concurrentDeviceDiscoveries int
	nvmlRunner                  nvml.Runner
	deviceCacheMux              *sync.Mutex
	deviceCache                 map[string]nvml.Device
}

// DiscoveryDevice holds discovered device data.
type DiscoveryDevice struct {
	UUID string `json:"device_uuid"`
	Name string `json:"device_name"`
}

// ECCErrors holds data for ECC errors.
type ECCErrors struct {
	Corrected   uint64 `json:"corrected"`
	Uncorrected uint64 `json:"uncorrected"`
}

// EncoderStats holds gpu encoder stats.
type EncoderStats struct {
	SessionCount uint `json:"session_count"`
	FPS          uint `json:"average_fps"`
	Latency      uint `json:"average_latency_ms"`
}

// PCIeUtil holds information about PCIe throughput.
type PCIeUtil struct {
	Transmit uint `json:"tx_rate_kb_s"`
	Receive  uint `json:"rx_rate_kb_s"`
}

// UtilisationRates holds data about GPU and its Memory utilisation.
type UtilisationRates struct {
	Device uint `json:"device"`
	Memory uint `json:"memory"`
}

// ECCMode returns current and pending status of ECC.
type ECCMode struct {
	Current bool `json:"current"`
	Pending bool `json:"pending"`
}

// New creates a new handler with initialized clients for system and tcp calls.
func New(nvmlRunner nvml.Runner) *Handler {
	return &Handler{
		// negative indicates no limit
		concurrentDeviceDiscoveries: -1,
		nvmlRunner:                  nvmlRunner,
		deviceCacheMux:              &sync.Mutex{},
		deviceCache:                 make(map[string]nvml.Device),
	}
}

// GetNVMLVersion returns local NVML version.
func (h *Handler) GetNVMLVersion(_ context.Context, _ map[string]string, _ ...string) (any, error) {
	version, err := h.nvmlRunner.GetNVMLVersion()
	if err != nil {
		return "", errs.Wrap(err, "failed to get NVML version")
	}

	return version, nil
}

// GetDriverVersion returns local graphics driver version.
func (h *Handler) GetDriverVersion(_ context.Context, _ map[string]string, _ ...string) (any, error) {
	version, err := h.nvmlRunner.GetDriverVersion()
	if err != nil {
		return "", errs.Wrap(err, "failed to get driver version")
	}

	return version, nil
}

// DeviceDiscovery discovers devices and returns UUIDs and names of devices.
func (h *Handler) DeviceDiscovery(ctx context.Context, _ map[string]string, _ ...string) (any, error) {
	deviceCount, err := h.nvmlRunner.GetDeviceCountV2()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device count")
	}

	var (
		discoveredMux = &sync.Mutex{}
		discovered    = make([]DiscoveryDevice, 0, deviceCount)
		deviceCache   = make(map[string]nvml.Device)
	)

	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(h.concurrentDeviceDiscoveries)

	// Should be done in parallel
	for i := range deviceCount {
		group.Go(func() error {
			select {
			// fails on first discovery error
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			device, err := h.nvmlRunner.GetDeviceByIndexV2(i) //nolint:govet
			if err != nil {
				return errs.Wrap(err, "failed to get device by index")
			}

			uuid, err := device.GetUUID()
			if err != nil {
				return errs.Wrap(err, "failed to get device UUID")
			}

			name, err := device.GetName()
			if err != nil {
				return errs.Wrap(err, "failed to get device name")
			}

			d := DiscoveryDevice{
				UUID: uuid,
				Name: name,
			}

			discoveredMux.Lock()
			defer discoveredMux.Unlock()

			deviceCache[uuid] = device

			discovered = append(discovered, d)

			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		return nil, errs.Wrap(err, "failed discovering devices")
	}

	h.deviceCacheMux.Lock()
	defer h.deviceCacheMux.Unlock()

	h.deviceCache = deviceCache

	return discovered, nil
}

// GetDeviceCount returns count of gpu's.
func (h *Handler) GetDeviceCount(_ context.Context, _ map[string]string, _ ...string) (any, error) {
	deviceCount, err := h.nvmlRunner.GetDeviceCountV2()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device count")
	}

	return deviceCount, nil
}

// GetDeviceTemperature returns remperature of gpu by UUID.
func (h *Handler) GetDeviceTemperature(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	temperature, err := device.GetTemperature()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device temperature")
	}

	return temperature, nil
}

// GetDeviceSerial returns serial number of gpu by UUID.
func (h *Handler) GetDeviceSerial(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	serial, err := device.GetSerial()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device serial")
	}

	return serial, nil
}

// GetDeviceFanSpeed returns gpu fan by UUID.
func (h *Handler) GetDeviceFanSpeed(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	fanSpeed, err := device.GetFanSpeed()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get fan speed")
	}

	return fanSpeed, nil
}

// GetDevicePerfState returns gpu performance state in range (0-15) by UUID.
func (h *Handler) GetDevicePerfState(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	perfState, err := device.GetPerformanceState()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get performance state")
	}

	return perfState, nil
}

// GetDeviceEnergyConsumption retrieves the total energy consumption of the NVIDIA device in millijoules.
func (h *Handler) GetDeviceEnergyConsumption(
	_ context.Context,
	metricParams map[string]string,
	_ ...string,
) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	energyCons, err := device.GetTotalEnergyConsumption()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get total energy consumption")
	}

	return energyCons, nil
}

// GetDevicePowerLimit retrieves the power management limit of the NVIDIA device in milliwatts.
func (h *Handler) GetDevicePowerLimit(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	powerLimit, err := device.GetPowerManagementLimit()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device power limit")
	}

	return powerLimit, nil
}

// GetDevicePowerUsage retrieves the power usage of the NVIDIA device in milliwatts.
func (h *Handler) GetDevicePowerUsage(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	powerUsage, err := device.GetPowerUsage()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device power usage")
	}

	return powerUsage, nil
}

// GetBAR1MemoryInfo retrieves BAR1 memory information for the NVIDIA device.
func (h *Handler) GetBAR1MemoryInfo(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	memoryInfo, err := device.GetBAR1MemoryInfo()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get BAR1 memory info")
	}

	return memoryInfo, nil
}

// GetFBMemoryInfo retrieves detailed memory information for the NVIDIA device using the NVML v2 interface.
func (h *Handler) GetFBMemoryInfo(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	memoryInfoV2, err := device.GetMemoryInfoV2()
	if err == nil {
		return memoryInfoV2, nil
	}

	memortInfoV1, err := device.GetMemoryInfo()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get memory info")
	}

	return memortInfoV1, nil
}

// GetMemoryErrors retrieves the number of corrected and uncorrected ECC errors in memory.
func (h *Handler) GetMemoryErrors(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	corrected, err := device.GetMemoryErrorCounter(
		nvml.MemoryErrorTypeCorrected,
		nvml.MemoryLocationDevice,
		nvml.EccCounterTypeAggregate,
	)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get corrected memory errors")
	}

	uncorrected, err := device.GetMemoryErrorCounter(
		nvml.MemoryErrorTypeUncorrected,
		nvml.MemoryLocationDevice,
		nvml.EccCounterTypeAggregate,
	)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get uncorrected memory errors")
	}

	return ECCErrors{
		Corrected:   corrected,
		Uncorrected: uncorrected,
	}, nil
}

// GetRegisterErrors retrieves the number of corrected and uncorrected ECC errors in register file.
func (h *Handler) GetRegisterErrors(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	corrected, err := device.GetMemoryErrorCounter(
		nvml.MemoryErrorTypeCorrected,
		nvml.MemoryLocationRegisterFile,
		nvml.EccCounterTypeAggregate,
	)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get corrected memory errors")
	}

	uncorrected, err := device.GetMemoryErrorCounter(
		nvml.MemoryErrorTypeUncorrected,
		nvml.MemoryLocationRegisterFile,
		nvml.EccCounterTypeAggregate,
	)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get uncorrected memory errors")
	}

	return ECCErrors{
		Corrected:   corrected,
		Uncorrected: uncorrected,
	}, nil
}

// GetPCIeThroughput retrieves the PCIe receive and transmit throughput for the NVIDIA device in KB/s.
func (h *Handler) GetPCIeThroughput(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	rx, err := device.GetPCIeThroughput(nvml.RX)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get rx throughput")
	}

	tx, err := device.GetPCIeThroughput(nvml.TX)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get tx throughput")
	}

	return PCIeUtil{
		Receive:  rx,
		Transmit: tx,
	}, nil
}

// GetEncoderStats retrieves statistics related to the encoder activity on the device.
// Metrics are: session count, fps, latency.
func (h *Handler) GetEncoderStats(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	sessions, fps, latency, err := device.GetEncoderStats()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get encoder stats")
	}

	return EncoderStats{
		SessionCount: sessions,
		FPS:          fps,
		Latency:      latency,
	}, nil
}

// GetVideoFrequency retrieves the clock rate for video encoder/decoder of the NVIDIA device.
func (h *Handler) GetVideoFrequency(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	clock, err := device.GetClockInfo(nvml.Video)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get clock info")
	}

	return clock, nil
}

// GetGraphicsFrequency retrieves the clock rate for the graphics module of the NVIDIA device.
func (h *Handler) GetGraphicsFrequency(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	clock, err := device.GetClockInfo(nvml.Graphics)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get clock info")
	}

	return clock, nil
}

// GetSMFrequency retrieves the clock rate for the specified clock type of the NVIDIA device.
func (h *Handler) GetSMFrequency(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	clock, err := device.GetClockInfo(nvml.SM)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get clock info")
	}

	return clock, nil
}

// GetMemoryFrequency retrieves the clock rate for memory of the NVIDIA device.
func (h *Handler) GetMemoryFrequency(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	clock, err := device.GetClockInfo(nvml.Memory)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get clock info")
	}

	return clock, nil
}

// GetEncoderUtilization retrieves the encoder utilization statistics for the device.
func (h *Handler) GetEncoderUtilization(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	utilisation, _, err := device.GetEncoderUtilization()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get encoder utilisation")
	}

	return utilisation, nil
}

// GetDecoderUtilization retrieves the decoder utilization statistics for the device.
func (h *Handler) GetDecoderUtilization(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	utilisation, _, err := device.GetDecoderUtilization()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get decoder utilisation")
	}

	return utilisation, nil
}

// GetDeviceUtilisation collects data about device utilisation.
func (h *Handler) GetDeviceUtilisation(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	gpu, memory, err := device.GetUtilizationRates()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get utilisation rates")
	}

	return UtilisationRates{
		Device: gpu,
		Memory: memory,
	}, nil
}

// GetECCMode collects data about gpu ECC mode.
func (h *Handler) GetECCMode(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("failed to find param for UUID")
	}

	device, err := h.getDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed getting device by UUID")
	}

	current, pending, err := device.GetEccMode()
	if err != nil {
		return nil, errs.Wrap(err, "failed getting ecc mode")
	}

	return ECCMode{
		Current: current,
		Pending: pending,
	}, nil
}

// WithJSONResponse wraps a handler function, marshaling its response
// to a JSON object and returning it as string.
func WithJSONResponse(handler HandlerFunc) HandlerFunc {
	return func(
		ctx context.Context, metricParams map[string]string, extraParams ...string,
	) (any, error) {
		res, err := handler(ctx, metricParams, extraParams...)
		if err != nil {
			return nil, errs.Wrap(err, "failed to receive the result")
		}

		jsonRes, err := json.Marshal(res)
		if err != nil {
			return nil, errs.Wrap(err, "failed to marshal result to JSON")
		}

		return string(jsonRes), nil
	}
}

// getDeviceByUUID accesses devices from device cache of runner,
// if device with requested UUID not cached it requests it from NVML.
// In case of success it caches device for future requests, else returns error.
//
//nolint:ireturn,nolintlint
func (h *Handler) getDeviceByUUID(uuid string) (nvml.Device, error) {
	h.deviceCacheMux.Lock()
	defer h.deviceCacheMux.Unlock()

	device, ok := h.deviceCache[uuid]
	if ok {
		return device, nil
	}

	device, err := h.nvmlRunner.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device by UUID")
	}

	h.deviceCache[uuid] = device

	return device, nil
}
