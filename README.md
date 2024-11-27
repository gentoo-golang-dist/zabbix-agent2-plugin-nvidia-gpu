# NVIDIA GPU Plugin for Zabbix Agent 2

## Plugin Information
This plugin provides a native Zabbix solution for monitoring a broad range of NVIDIA GPU metrics with minimal configuration effort.

For information retrieval, the plugin uses NVIDIA's NVML dynamic library. By default, the NVML library is installed on your host along with NVIDIA drivers.

## Requirements
- Installed NVIDIA driver.

### Notes
- The plugin was developed for NVML API version 12. Older NVML versions may not support some metrics.
- Metrics may report errors or be unsupported if your device cannot provide the required information.

## Plugin Setup
The `Plugins.NVIDIA.System.Path` variable must be set in the Zabbix Agent 2 configuration file, specifying the path to the NVIDIA GPU plugin executable. By default, this variable is set in the **plugin** configuration file `nvidia.conf`, which is then included in the **agent** configuration file `zabbix_agent2.conf`.

### Example Setup:
1. Add the following option to the **plugin** configuration file:
   ```text
   Plugins.NVIDIA.System.Path=/path/to/executable/nvidia
   ```
2. Include the plugin configuration file in the main Zabbix Agent 2 configuration file using the `Include` directive:
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
  1. **`device_uuid`**: Unique identifier for the device.  
  2. **`device_name`**: Name of the device.

- **`nvml.device.count`**  
  Returns a single value: (unsigned int) number of devices.

## General Device Metrics
- **`nvml.device.temperature[<device_uuid>]`**  
  Returns a single value: (unsigned int) temperature of the device in Celsius.

- **`nvml.device.serial[<device_uuid>]`**  
  Returns a single value: (unsigned int) number of devices.

- **`nvml.device.fan.speed.avg[<device_uuid>]`**  
  Returns a single value: (unsigned int) average fan speed as a percentage of maximum speed.

- **`nvml.device.performance.state[<device_uuid>]`**  
  Returns a single value: (unsigned int) performance state of the device (0 = max, 15 = min).

- **`nvml.device.energy.consumption[<device_uuid>]`**  
  Returns a single value: (unsigned int) total energy consumption in millijoules (mJ) since the driver was last reloaded.

- **`nvml.device.power.limit[<device_uuid>]`**  
  Returns a single value: (unsigned int) power limit in milliwatts.

- **`nvml.device.power.usage[<device_uuid>]`**  
  Returns a single value: (unsigned int) current power usage in milliwatts.

## Device Memory Metrics
- **`nvml.device.memory.bar1.get[<device_uuid>]`**  
  Returns a JSON structure with the following fields (in bytes):  
  1. **`total_memory_bytes`**: Total BAR1 memory available on the GPU.  
  2. **`used_memory_bytes`**: BAR1 memory currently in use.  
  3. **`free_memory_bytes`**: Available BAR1 memory.

- **`nvml.device.memory.fb.get[<device_uuid>]`**  
  Returns a JSON structure with the following fields (in bytes):  
  1. **`total_memory_bytes`**: Total framebuffer memory of the GPU.  
  2. **`reserved_memory_bytes`**: Memory reserved for internal GPU operations.  
  3. **`free_memory_bytes`**: Available framebuffer memory.  
  4. **`used_memory_bytes`**: Memory currently in use (includes reserved memory).

### Notes
- Reserved memory is included in the used memory.

## Device ECC Mode
- **`nvml.device.ecc.mode[<device_uuid>]`**  
  Returns a JSON structure with the following fields:  
  1. **`current`**: The current ECC mode (bool).  
  2. **`pending`**: The pending ECC mode (bool) to be applied after reboot.

## Device ECC Error Metrics
- **`nvml.device.errors.memory[<device_uuid>]`**  
  Returns a JSON structure with the following fields:  
  1. **`corrected`**: Count of ECC errors that were corrected in memory.  
  2. **`uncorrected`**: Count of ECC errors that could not be corrected in memory.

- **`nvml.device.errors.registry[<device_uuid>]`**  
  Returns a JSON structure with the following fields:  
  1. **`corrected`**: Count of ECC errors that were corrected in registry file.  
  2. **`uncorrected`**: Count of ECC errors that could not be corrected in registry file.

## Device PCI Metrics
- **`nvml.device.pci.utilization[<device_uuid>]`**  
  Returns a JSON structure with the following fields:  
  1. **`tx_rate_kb_s`**: PCI transmit throughput in KB/s.  
  2. **`rx_rate_kb_s`**: PCI receive throughput in KB/s.

## Device Encoder/Decoder Metrics
- **`nvml.device.encoder.stats.get[<device_uuid>]`**  
  Returns a JSON structure with the following fields:  
  1. **`session_count`**: Count of active encoder sessions.  
  2. **`average_fps`**: Average FPS of all active sessions.  
  3. **`average_latency_ms`**: Encode latency in microseconds.

- **`nvml.device.encoder.utilization[<device_uuid>]`**  
  Returns a single value: (unsigned int) encoder utilization as a percentage.

- **`nvml.device.decoder.utilization[<device_uuid>]`**  
  Returns a single value: (unsigned int) decoder utilization as a percentage.

## Device Frequency Metrics
- **`nvml.device.video.frequency[<device_uuid>]`**  
  Returns a single value: (unsigned int) video clock speed in MHz.

- **`nvml.device.graphics.frequency[<device_uuid>]`**  
  Returns a single value: (unsigned int) graphics clock speed in MHz.

- **`nvml.device.sm.frequency[<device_uuid>]`**  
  Returns a single value: (unsigned int) streaming multiprocessor (SM) clock speed in MHz.

- **`nvml.device.memory.frequency[<device_uuid>]`**  
  Returns a single value: (unsigned int) memory clock speed in MHz.

## Device Utilization Metrics
- **`nvml.device.utilization[<device_uuid>]`**  
  Returns a JSON structure with the following fields:  
  1. **`device`**: GPU utilization as a percentage.  
  2. **`memory`**: Memory utilization as a percentage.

