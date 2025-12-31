# rnodeconf

RNode configuration and firmware utility for the Go port. This is the most hardware-dependent tool in `cmd/`: it talks to an attached RNode over a serial port.

## Overview

- Prints device info and configuration.
- Performs firmware update / flashing and EEPROM bootstrap on supported devices.
- Manages local signing keys and trusted verification keys.
- Configures WiFi/Bluetooth/display and radio (TNC) parameters.
- Can backup/dump/wipe EEPROM.

## Usage

`rnodeconf [options] [port]`

You can specify the serial port either via `--port` or as the final positional argument.

## Storage directory

`rnodeconf` stores its own state separately from Reticulum:

- If `RNODECONF_DIR` is set, it uses that directory.
- Otherwise it uses `~/.config/rnodeconf`.

## Flags

### General

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-version` | bool | `false` | Print program version and exit. |
| `-help`, `-h` | bool | `false` | Show help and exit. |
| `-port` | string | `""` | Serial port for the RNode (or use positional `port`). |
| `-list-models` | bool | `false` | List known hardware models and exit. |
| `-show-model` | string | `""` | Show detailed information about a model code (example `A1` or `0xA1`). |
| `-C`, `-clear-cache` | bool | `false` | Clear cached firmware files and exit. |
| `-trust-key` | string | `""` | Store a trusted public key for device verification (hex DER). |

### Device info / install / firmware

| Flag(s) | Type | Description |
|---|---:|---|
| `-i`, `-info` | bool | Show device info. |
| `-a`, `-autoinstall` | bool | Automatic installation on supported devices. |
| `-u`, `-update` | bool | Update firmware to the latest version. |
| `-U`, `-force-update` | bool | Update even if version matches or is older. |
| `-f`, `-flash` | bool | Flash firmware and bootstrap EEPROM (offline/local firmware only). |
| `-r`, `-rom` | bool | Bootstrap EEPROM without flashing firmware. |
| `-fw-version` | string | Use a specific firmware version for update/autoinstall. |
| `-fw-url` | string | Use an alternate firmware download URL. |
| `-nocheck` | bool | Do not check for updates online. |
| `-e`, `-extract` | bool | Extract firmware from a connected RNode for later use. |
| `-E`, `-use-extracted` | bool | Use previously extracted firmware for autoinstall/update. |
| `-baud-flash` | string | Set a specific baud rate when flashing (default `921600`). |

### Keys and signing

| Flag(s) | Type | Description |
|---|---:|---|
| `-k`, `-key` | bool | Generate a new signing key and exit. |
| `-S`, `-sign` | bool | Sign attached device (store device signature in EEPROM). |
| `-P`, `-public`, `-show-signing-key` | bool | Display the public part of the local signing key (hex DER) and device signing public key. |
| `-H`, `-firmware-hash` | string | Set installed firmware hash (hex). |

### Device modes

| Flag(s) | Type | Description |
|---|---:|---|
| `-N`, `-normal` | bool | Switch device to normal mode. |
| `-T`, `-tnc` | bool | Switch device to TNC mode. |

### Bluetooth

| Flag(s) | Type | Description |
|---|---:|---|
| `-b`, `-bluetooth-on` | bool | Turn Bluetooth on. |
| `-B`, `-bluetooth-off` | bool | Turn Bluetooth off. |
| `-p`, `-bluetooth-pair` | bool | Enter Bluetooth pairing mode. |

### WiFi

| Flag(s) | Type | Description |
|---|---:|---|
| `-w`, `-wifi` | string | Set WiFi mode: `OFF`, `AP`, or `STATION`. |
| `-channel` | string | Set WiFi channel. |
| `-ssid` | string | Set WiFi SSID (`NONE` to delete). |
| `-psk` | string | Set WiFi PSK (`NONE` to delete). |
| `-show-psk` | bool | Display stored WiFi PSK. |
| `-ip` | string | Set static WiFi IP (`NONE` for DHCP). |
| `-nm` | string | Set static netmask (`NONE` for DHCP). |

### Display and LEDs

| Flag(s) | Type | Description |
|---|---:|---|
| `-D`, `-display` | int | Set display intensity (0–255). |
| `-t`, `-timeout` | int | Set display timeout seconds (`0` disables). |
| `-R`, `-rotation` | int | Set display rotation (0–3). |
| `-display-addr` | string | Set display I2C address as a hex byte (`00`–`FF`). |
| `-recondition-display` | bool | Start display reconditioning. |
| `-np` | int | Set NeoPixel intensity (0–255). |

### TNC parameters

| Flag(s) | Type | Description |
|---|---:|---|
| `-freq` | int | Frequency (Hz). |
| `-bw` | int | Bandwidth (Hz). |
| `-txp` | int | TX power (dBm). |
| `-sf` | int | Spreading factor (7–12). |
| `-cr` | int | Coding rate (5–8). |
| `-x`, `-ia-enable` | bool | Enable interference avoidance. |
| `-X`, `-ia-disable` | bool | Disable interference avoidance. |

### EEPROM

| Flag(s) | Type | Description |
|---|---:|---|
| `-c`, `-config` | bool | Print device configuration. |
| `-eeprom-backup` | bool | Backup EEPROM to file. |
| `-eeprom-dump` | bool | Dump EEPROM to console. |
| `-eeprom-wipe` | bool | Unlock and wipe EEPROM. |

## Examples

Show device info:

`go run ./cmd/rnodeconf --port /dev/tty.usbserial-XXXX --info`

Update firmware (may require network access):

`go run ./cmd/rnodeconf --port /dev/tty.usbserial-XXXX --update`

Switch to TNC mode and set frequency:

`go run ./cmd/rnodeconf --port /dev/tty.usbserial-XXXX --tnc --freq 868100000`

## Exit codes

- `0`: Success.
- `2`: CLI usage / unknown flags.
- `99`: Could not create/access `RNODECONF_DIR` storage.

Many additional non-zero codes are used to match upstream behaviour for specific error cases (firmware verification, flashing failures, missing parameters, etc.).

