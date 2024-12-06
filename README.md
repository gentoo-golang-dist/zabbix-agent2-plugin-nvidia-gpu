# NVIDIA GPU Plugin for Zabbix Agent 2

# Table of Contents

- [Plugin Information](#plugin-information)
- [Requirements](#requirements)
    - [Notes](#notes)
- [Build from Source](#build-from-source)
    - [Prerequisites](#prerequisites)
- [Plugin Setup](#plugin-setup)
    - [Example Setup](#example-setup)
- [Configuration](#configuration)
- [Metric Keys](#metric-keys)
    - [General Information](#general-information)
    - [General Device Metrics](#general-device-metrics)
    - [Device Memory Metrics](#device-memory-metrics)
    - [Device ECC Mode](#device-ecc-mode)
    - [Device ECC Error Metrics](#device-ecc-error-metrics)
    - [Device PCI Metrics](#device-pci-metrics)
    - [Device Encoder/Decoder Metrics](#device-encoderdecoder-metrics)
    - [Device Frequency Metrics](#device-frequency-metrics)
    - [Device Utilization Metrics](#device-utilization-metrics)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

## Plugin Information
This plugin provides a native Zabbix solution for monitoring a broad range of NVIDIA GPU metrics with minimal configuration effort.

For information retrieval, the plugin uses NVIDIA's NVML dynamic library. By default, the NVML library is installed on your host along with NVIDIA drivers.

## Requirements
- Installed NVIDIA driver.

### Notes
- The plugin was developed for NVML API version 12. Older NVML versions may not support some metrics.
- Metrics may report errors or be unsupported if your device cannot provide the required information.
- **If the NVML dynamic library, which is installed by default with the NVIDIA driver, is absent, Zabbix Agent 2 with the NVIDIA GPU plugin will not start.**

## Build from Source

The plugin supports building for both Linux and Windows. To avoid errors during cross-compilation, it is recommended to build the plugin directly on the target operating system.

To build the NVIDIA GPU Plugin for Zabbix Agent 2 from source, ensure you have the following prerequisites.

### Prerequisites
- **Go Programming Language**: Version 1.21 or higher.  
- **CGO Enabled**: The build process requires `CGO_ENABLED=1` for proper compilation.
- **C Compiler**: A C compiler is required for building with `CGO_ENABLED=1`.  

## Plugin Setup
The `Plugins.NVIDIA.System.Path` variable must be set in the Zabbix Agent 2 configuration file, specifying the path to the NVIDIA GPU plugin executable. By default, this variable is set in the **plugin** configuration file `nvidia.conf`, which is then included in the **agent** configuration file `zabbix_agent2.conf`.

### Example Setup:
- Add the following option to the **plugin** configuration file:
   ```text
   Plugins.NVIDIA.System.Path=/path/to/executable/nvidia
   ```
- Include the plugin configuration file in the main Zabbix Agent 2 configuration file using the `Include` directive:
   ```text
   Include=/path/to/config/nvidia.conf
   ```

## Configuration
To configure plugins, use the Zabbix Agent configuration file.

- **`Plugins.NVIDIA.Timeout`**: Specifies the maximum time (in seconds) to wait for a server response during connection attempts and subsequent operations in the session. The global item-type timeout or individual item timeout will override this value if greater.  
  - **Default**: Equal to the global `Timeout` parameter in the Zabbix Agent 2 configuration file.  
  - **Limits**: 1-30 seconds.

# Metric Keys

## General Information
- **`nvml.version`**  
  Returns a single value: (string) version of the NVML library.

- **`nvml.system.driver.version`**  
  Returns a single value: (string) version of the installed NVIDIA driver.

- **`nvml.device.get`**  
  Returns a JSON array, where each element represents a device in the system with the following fields:  
  - **`device_uuid`**: Unique identifier for the device.  
  - **`device_name`**: Name of the device.

- **`nvml.device.count`**  
  Returns a single value: (unsigned int) number of devices.

## General Device Metrics
- **`nvml.device.temperature[<deviceUUID>]`**  
  Returns a single value: (unsigned int) temperature of the device in Celsius.

- **`nvml.device.serial[<deviceUUID>]`**  
  Returns a single value: (unsigned int) number of devices.

- **`nvml.device.fan.speed.avg[<deviceUUID>]`**  
  Returns a single value: (unsigned int) average fan speed as a percentage of maximum speed.

- **`nvml.device.performance.state[<deviceUUID>]`**  
  Returns a single value: (unsigned int) performance state of the device (0 = max, 15 = min).

- **`nvml.device.energy.consumption[<deviceUUID>]`**  
  Returns a single value: (unsigned int) total energy consumption in millijoules (mJ) since the driver was last reloaded.

- **`nvml.device.power.limit[<deviceUUID>]`**  
  Returns a single value: (unsigned int) power limit in milliwatts.

- **`nvml.device.power.usage[<deviceUUID>]`**  
  Returns a single value: (unsigned int) current power usage in milliwatts.

## Device Memory Metrics
- **`nvml.device.memory.bar1.get[<deviceUUID>]`**  
  Returns a JSON structure with the following fields (in bytes):  
  - **`total_memory_bytes`**: Total BAR1 memory available on the GPU.  
  - **`free_memory_bytes`**: Available BAR1 memory.
  - **`used_memory_bytes`**: BAR1 memory currently in use.  

- **`nvml.device.memory.fb.get[<deviceUUID>]`**  
  Returns a JSON structure with the following fields (in bytes):  
  - **`total_memory_bytes`**: Total framebuffer memory of the GPU.  
  - **`reserved_memory_bytes`**: Memory reserved for internal GPU operations.  
  - **`free_memory_bytes`**: Available framebuffer memory.  
  - **`used_memory_bytes`**: Memory currently in use (includes reserved memory).

  ### Notes
  - Reserved memory is included in the used memory.

## Device ECC Mode
- **`nvml.device.ecc.mode[<deviceUUID>]`**  
  Returns a JSON structure with the following fields:  
  - **`current`**: The current ECC mode (bool).  
  - **`pending`**: The pending ECC mode (bool) to be applied after reboot.

## Device ECC Error Metrics
- **`nvml.device.errors.memory[<deviceUUID>]`**  
  Returns a JSON structure with the following fields:  
  - **`corrected`**: Count of ECC errors that were corrected in memory.  
  - **`uncorrected`**: Count of ECC errors that could not be corrected in memory.

- **`nvml.device.errors.register[<deviceUUID>]`**  
  Returns a JSON structure with the following fields:  
  - **`corrected`**: Count of ECC errors that were corrected in register file.  
  - **`uncorrected`**: Count of ECC errors that could not be corrected in register file.

## Device PCI Metrics
- **`nvml.device.pci.utilization[<deviceUUID>]`**  
  Returns a JSON structure with the following fields:  
  - **`tx_rate_kb_s`**: PCI transmit throughput in KB/s.  
  - **`rx_rate_kb_s`**: PCI receive throughput in KB/s.

## Device Encoder/Decoder Metrics
- **`nvml.device.encoder.stats.get[<deviceUUID>]`**  
  Returns a JSON structure with the following fields:  
  - **`session_count`**: Count of active encoder sessions.  
  - **`average_fps`**: Average FPS of all active sessions.  
  - **`average_latency_ms`**: Encode latency in microseconds.

- **`nvml.device.encoder.utilization[<deviceUUID>]`**  
  Returns a single value: (unsigned int) encoder utilization as a percentage.

- **`nvml.device.decoder.utilization[<deviceUUID>]`**  
  Returns a single value: (unsigned int) decoder utilization as a percentage.

## Device Frequency Metrics
- **`nvml.device.video.frequency[<deviceUUID>]`**  
  Returns a single value: (unsigned int) video clock speed in MHz.

- **`nvml.device.graphics.frequency[<deviceUUID>]`**  
  Returns a single value: (unsigned int) graphics clock speed in MHz.

- **`nvml.device.sm.frequency[<deviceUUID>]`**  
  Returns a single value: (unsigned int) streaming multiprocessor (SM) clock speed in MHz.

- **`nvml.device.memory.frequency[<deviceUUID>]`**  
  Returns a single value: (unsigned int) memory clock speed in MHz.

## Device Utilization Metrics
- **`nvml.device.utilization[<deviceUUID>]`**  
  Returns a JSON structure with the following fields:  
  - **`device`**: GPU utilization as a percentage.  
  - **`memory`**: Memory utilization as a percentage.

## Troubleshooting

The plugin forwards all its logs to Zabbix Agent 2, which then logs them according to the log location configured for the agent.

For debugging purposes, you can increase the Zabbix Agent 2 log level by either updating the `DebugLevel` field in the configuration file or using runtime control with the following command:

```sh
zabbix_agent2 -R log_level_increase
```

For more detailed information about Zabbix Agent 2, refer to the [official Zabbix documentation](https://www.zabbix.com/documentation/current/en/manual/concepts/agent2).

## Contributing

Found a bug or have a suggestion for improvement? Feel free to open an issue or submit a feature request through the [Zabbix support system](https://support.zabbix.com/secure/Dashboard.jspa).

Interested in contributing? Pull requests are always welcome!