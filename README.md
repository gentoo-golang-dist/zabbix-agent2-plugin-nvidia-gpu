# NVIDIA GPU plugin for Zabbix agent 2

## Plugin information
This plugin provides native Zabbix solution for monithoring broad range of NVIDIA GPU metrics with minimal configuration efforts.

For information retrival plugin uses NVIDIA's NVML dynamic library. By default NVML library gets installed on your host with NVIDIA drivers.

## Requirements
Installed NVIDIA driver.

### Notes:
- Plugin was developed for NVML API version 12. With older NVML versions some metrics may not be supported.
- Metric may report error and be unsupported if your device can not provide such information.

## Plugin setup
*Plugins.NVIDIA.System.Path* variable needs to be set in Zabbix agent 2 configuration file with the path to the NVIDIA GPU
plugin executable. By default the variable is set in **plugin** configuration file *nvidia.conf* and then included in
the **agent** configuration file *zabbix_agent2.conf*.

For example: 
You should add the following option to the **plugin** configuration file:

    Plugins.NVIDIA.System.Path=/path/to/executable/nvidia

Then, the configuration file needs to be included in the main Zabbix agent 2 configuration file via the
*Include* command.

For example: 
You should add the following option to the **plugin** configuration file:

    Include=/path/to/config/nvidia.conf

## Configuration
To configure plugins, use Zabbix agent configuration file.

Plugins.NVIDIA.Timeout — the maximum amount of time to wait for a server to respond when first connecting and on follow up operations in the session. 
Global item-type timeout (or individual item timeout) will override this value if it is greater. Default value: equals the global Timeout configuration parameter defined in Zabbix agent 2 configuration file. Limits: 1-30

# Metric Keys

## `nvml.version`
- Returns a single value: (string) version of the NVML library.

## `nvml.system.driver.version`
- Returns a single value: (string) version of the installed NVIDIA driver.

## `nvml.device.get`

Returns a JSON array where each element represents a device in the system, consisting of the following fields:

1. **`device_uuid`** - (string) unique identifier for the device.
2. **`device_name`** - (string) name of the device.

## `nvml.device.count`
- Returns a single value: (unsigned int) number of devices.

## `nvml.device.temperature[<device_uuid>]`
- Returns a single value: (unsigned int) temperature of the device in Celsius.

## `nvml.device.serial[<device_uuid>]`
- Returns a single value: (unsigned int) number of devices.

## `nvml.device.fan.speed.avg[<device_uuid>]`
- Returns a single value: (unsigned int) average fan speed as a percentage (%) of maximum speed.

## `nvml.device.performance.state[<device_uuid>]`
- Returns a single value: (unsigned int) performance state of the device, ranging from 0 (max performance) to 15 (min performance).

## `nvml.device.energy.consumption[<device_uuid>]`
- Returns a single value: (unsigned int) total energy consumption in millijoules (mJ) since the driver was last reloaded.

## `nvml.device.power.limit[<device_uuid>]`
- Returns a single value: (unsigned int) power limit of the device in milliwatts.

## `nvml.device.power.usage[<device_uuid>]`
- Returns a single value: (unsigned int) current power usage of the device in milliwatts.

## `nvml.device.memory.bar1.get[<device_uuid>]`

Returns a JSON structure consisting of the following fields (values are in bytes):

1. **`total_memory_bytes`** - (unsigned int) total BAR1 memory available on the GPU.
2. **`used_memory_bytes`** - (unsigned int) BAR1 memory currently in use.
3. **`free_memory_bytes`** - (unsigned int) available BAR1 memory.

## `nvml.device.memory.fb.get[<device_uuid>]`

Returns a JSON structure consisting of the following fields (values are in bytes):

1. **`total_memory_bytes`** - (unsigned int) total framebuffer memory of the GPU.
2. **`reserved_memory_bytes`** - (unsigned int) memory reserved for internal GPU operations.
3. **`free_memory_bytes`** - (unsigned int) available framebuffer memory.
4. **`used_memory_bytes`** - (unsigned int) memory currently in use (includes reserved memory).

### Notes
- Reserved memory is also included in the used memory.

## `nvml.device.errors.memory[<device_uuid>]`

Returns a JSON structure consisting of the following fields:

1. **`corrected`** - (unsigned int) count of memory errors that were corrected.
2. **`uncorrected`** - (unsigned int) count of memory errors that could not be corrected.

## `nvml.device.errors.registry[<device_uuid>]`

Returns a JSON structure consisting of the following fields:

1. **`corrected`** - (unsigned int) count of registry errors that were corrected.
2. **`uncorrected`** - (unsigned int) count of registry errors that could not be corrected.

## `nvml.device.pci.utilization[<device_uuid>]`

Returns a JSON structure consisting of the following fields:

1. **`tx_rate_kb_s`** - (unsigned int) PCI transmit throughput in KB/s.
2. **`rx_rate_kb_s`** - (unsigned int) PCI receive throughput in KB/s.

## `nvml.device.encoder.stats.get[<device_uuid>]`

Returns a JSON structure consisting of the following fields:

1. **`session_count`** - (unsigned int) count of active encoder sessions.
2. **`average_fps`** - (unsigned int) average FPS of all active sessions.
3. **`average_latency_ms`** - (unsigned int) encode latency in microseconds.

## `nvml.device.video.frequency[<device_uuid>]`
- Returns a single value: (unsigned int) video clock speed in MHz.

## `nvml.device.graphics.frequency[<device_uuid>]`
- Returns a single value: (unsigned int) graphics clock speed in MHz.

## `nvml.device.sm.frequency[<device_uuid>]`
- Returns a single value: (unsigned int) streaming multiprocessor (SM) clock speed in MHz.

## `nvml.device.memory.frequency[<device_uuid>]`
- Returns a single value: (unsigned int) memory clock speed in MHz.

## `nvml.device.encoder.utilization[<device_uuid>]`
- Returns a single value: (unsigned int) encoder utilization as a percentage (%).

## `nvml.device.decoder.utilization[<device_uuid>]`
- Returns a single value: (unsigned int) decoder utilization as a percentage (%).

## `nvml.device.utilization[<device_uuid>]`

Returns a JSON structure consisting of the following fields:

1. **`device`** - (unsigned int) GPU utilization as a percentage (%).
2. **`memory`** - (unsigned int) memory utilization as a percentage (%).

##  `nvml.device.ecc.mode[<device_uuid>]`

Returns a JSON structure consisting of the following fields:

1. **`current`** - (bool) the current ECC (Error-Correcting Code) mode of the device.
2. **`pending`** - (bool) the pending ECC mode that will take effect after the next reboot.