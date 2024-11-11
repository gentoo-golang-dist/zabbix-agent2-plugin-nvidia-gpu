/*
** Copyright (C) 2001-2024 Zabbix SIA
**
** Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
** documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
** rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
** permit persons to whom the Software is furnished to do so, subject to the following conditions:
**
** The above copyright notice and this permission notice shall be included in all copies or substantial portions
** of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
** WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
** COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
** TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
**/

package handlers

import (
	"context"
	"encoding/json"
	"sync"

	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
	"golang.zabbix.com/plugin/nvidia/plugin/params"
	"golang.zabbix.com/sdk/errs"
)

var (
	_ HandlerFunc = WithJSONResponse(nil)
	_ HandlerFunc = (*Handler)(nil).GetNVMLVersion
	_ HandlerFunc = (*Handler)(nil).GetBAR1MemoryInfo
	_ HandlerFunc = (*Handler)(nil).GetDecoderUtilisation
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
	_ HandlerFunc = (*Handler)(nil).GetEncoderUtilisation
	_ HandlerFunc = (*Handler)(nil).GetFBMemoryInfo
	_ HandlerFunc = (*Handler)(nil).GetGraphicsFrequency
	_ HandlerFunc = (*Handler)(nil).GetMemoryErrors
	_ HandlerFunc = (*Handler)(nil).GetMemoryFrequency
	_ HandlerFunc = (*Handler)(nil).GetPCIeThroughput
	_ HandlerFunc = (*Handler)(nil).GetRegistryErrors
	_ HandlerFunc = (*Handler)(nil).GetVideoFrequency
	_ HandlerFunc = (*Handler)(nil).GetSMFrequency
)

// Handler hold client and syscall implementation for request functions.
type Handler struct {
	nvmlRunner     nvml.Runner
	deviceCacheMux *sync.Mutex
	deviceCache    map[string]*nvml.NVMLDevice
}

// New creates a new handler with initialized clients for system and tcp calls.
func New(nvmlRunner nvml.Runner) *Handler {
	return &Handler{
		nvmlRunner:     nvmlRunner,
		deviceCacheMux: &sync.Mutex{},
		deviceCache:    make(map[string]*nvml.NVMLDevice),
	}
}

// HandlerFunc describes the signature all metric handler functions must have.
type HandlerFunc func(
	ctx context.Context,
	metricParams map[string]string,
	extraParams ...string,
) (any, error)

func (h *Handler) GetDeviceByUUID(uuid string) (nvml.Device, error) {
	h.deviceCacheMux.Lock()
	defer h.deviceCacheMux.Unlock()

	device, ok := h.deviceCache[uuid]
	if ok {
		return device, nil
	}

	device, err := h.nvmlRunner.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	h.deviceCache[uuid] = device

	return device, nil
}

func (h *Handler) GetNVMLVersion(_ context.Context, _ map[string]string, _ ...string) (any, error) {
	version, err := h.nvmlRunner.GetNVMLVersion()
	if err != nil {
		return "", errs.Wrap(err, "failed to get NVML version")
	}

	return version, nil
}

func (h *Handler) GetDriverVersion(_ context.Context, _ map[string]string, _ ...string) (any, error) {
	version, err := h.nvmlRunner.GetDriverVersion()
	if err != nil {
		return "", errs.Wrap(err, "failed to get driver version")
	}

	return version, nil
}

type DiscoveryDevice struct {
	UUID string `json:"device_uuid"`
	Name string `json:"device_name"`
}

func (h *Handler) DeviceDiscovery(_ context.Context, _ map[string]string, _ ...string) (any, error) {
	deviceCount, err := h.nvmlRunner.GetDeviceCountV2()
	if err != nil {
		return nil, err
	}

	var discovered []DiscoveryDevice
	deviceCache := make(map[string]*nvml.NVMLDevice)

	for i := uint(0); i < deviceCount; i++ {
		device, err := h.nvmlRunner.GetDeviceByIndexV2(i)
		if err != nil {
			return nil, errs.Wrap(err, "failed to get device by index")
		}

		uuid, err := device.GetUUID()
		if err != nil {
			return nil, errs.Wrap(err, "failed to get device uuid")
		}

		name, err := device.GetName()
		if err != nil {
			return nil, errs.Wrap(err, "failed to get device name")
		}

		d := DiscoveryDevice{
			UUID: uuid,
			Name: name,
		}

		deviceCache[uuid] = device
		discovered = append(discovered, d)
	}

	h.deviceCacheMux.Lock()
	defer h.deviceCacheMux.Unlock()

	h.deviceCache = deviceCache

	return discovered, nil
}

func (h *Handler) GetDeviceCount(_ context.Context, _ map[string]string, _ ...string) (any, error) {
	deviceCount, err := h.nvmlRunner.GetDeviceCountV2()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device count")
	}

	return deviceCount, nil
}

func (h *Handler) GetDeviceTemperature(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("Could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	temperature, err := device.GetTemperature()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device temperature")
	}

	return temperature, nil
}

func (h *Handler) GetDeviceSerial(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	serial, err := device.GetSerial()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device serial")
	}

	return serial, nil
}

func (h *Handler) GetDeviceFanSpeed(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	fanSpeed, err := device.GetFanSpeed()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get fan speed")
	}

	return fanSpeed, nil
}

func (h *Handler) GetDevicePerfState(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	perfState, err := device.GetPerformanceState()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get performance state")
	}

	return perfState, nil
}

func (h *Handler) GetDeviceEnergyConsumption(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	energyCons, err := device.GetTotalEnergyConsumption()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get total energy consumption")
	}

	return energyCons, nil
}

func (h *Handler) GetDevicePowerLimit(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	powerLimit, err := device.GetPowerManagementLimit()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device power limit")
	}

	return powerLimit, nil
}

func (h *Handler) GetDevicePowerUsage(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	powerUsage, err := device.GetPowerUsage()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get device power usage")
	}

	return powerUsage, nil
}

func (h *Handler) GetBAR1MemoryInfo(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	memoryInfo, err := device.GetBAR1MemoryInfo()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get BAR1 memory info")
	}

	return memoryInfo, nil
}

func (h *Handler) GetFBMemoryInfo(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	memoryInfo, err := device.GetMemoryInfoV2()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get memory info")
	}

	return memoryInfo, nil
}

type EccErrors struct {
	Corrected   uint64 `json:"corrected"`
	Uncorrected uint64 `json:"uncorrected"`
}

func (h *Handler) GetMemoryErrors(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
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

	ecc := EccErrors{
		Corrected:   corrected,
		Uncorrected: uncorrected,
	}

	return ecc, nil
}

func (h *Handler) GetRegistryErrors(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
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

	ecc := EccErrors{
		Corrected:   corrected,
		Uncorrected: uncorrected,
	}

	return ecc, nil
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

type PcieUtil struct {
	Transmit uint `json:"tx_rate_kb_s"`
	Receive  uint `json:"rx_rate_kb_s"`
}

func (h *Handler) GetPCIeThroughput(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	rx, err := device.GetPCIeThroughput(nvml.RX)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get rx throughput")
	}

	tx, err := device.GetPCIeThroughput(nvml.TX)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get tx throughput")
	}

	util := PcieUtil{
		Receive:  rx,
		Transmit: tx,
	}

	return util, nil
}

type EncoderStats struct {
	SessionCount uint `json:"session_count"`
	FPS          uint `json:"average_fps"`
	Latency      uint `json:"average_latency_ms"`
}

func (h *Handler) GetEncoderStats(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	sessions, fps, latency, err := device.GetEncoderStats()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get encoder stats")
	}

	stats := EncoderStats{
		SessionCount: sessions,
		FPS:          fps,
		Latency:      latency,
	}

	return stats, nil
}

func (h *Handler) GetVideoFrequency(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("Could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	clock, err := device.GetClockInfo(nvml.Video)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get clock info")
	}

	return clock, nil
}

func (h *Handler) GetGraphicsFrequency(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("Could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	clock, err := device.GetClockInfo(nvml.Graphics)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get clock info")
	}

	return clock, nil
}

func (h *Handler) GetSMFrequency(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("Could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	clock, err := device.GetClockInfo(nvml.SM)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get clock info")
	}

	return clock, nil
}

func (h *Handler) GetMemoryFrequency(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("Could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	clock, err := device.GetClockInfo(nvml.Memory)
	if err != nil {
		return nil, errs.Wrap(err, "failed to get clock info")
	}

	return clock, nil
}

func (h *Handler) GetEncoderUtilisation(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("Could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	utilisation, _, err := device.GetEncoderUtilization()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get encoder utilisation")
	}

	return utilisation, nil
}

func (h *Handler) GetDecoderUtilisation(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("Could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	utilisation, _, err := device.GetDecoderUtilization()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get decoder utilisation")
	}

	return utilisation, nil
}

type UtilisationRates struct {
	GPU    uint `json:"device"`
	Memory uint `json:"memory"`
}

func (h *Handler) GetDeviceUtilisation(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("Could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "error getting device by uuid")
	}

	gpu, memory, err := device.GetUtilizationRates()
	if err != nil {
		return nil, errs.Wrap(err, "failed to get utilisation rates")
	}

	util := UtilisationRates{
		GPU:    gpu,
		Memory: memory,
	}

	return util, nil
}

type EccMode struct {
	Currect bool `json:"current"`
	Pending bool `json:"pending"`
}

func (h *Handler) GetECCMode(_ context.Context, metricParams map[string]string, _ ...string) (any, error) {
	uuid, ok := metricParams[params.DeviceUUIDParamName]
	if !ok {
		return nil, errs.New("Could not find param for uuid")
	}

	device, err := h.GetDeviceByUUID(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "failed getting device by uuid")
	}

	current, pending, err := device.GetEccMode()
	if err != nil {
		return nil, errs.Wrap(err, "failed getting ecc mode")
	}

	mode := EccMode{
		Currect: current,
		Pending: pending,
	}

	return mode, nil
}
