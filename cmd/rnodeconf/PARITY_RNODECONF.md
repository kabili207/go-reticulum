# rnodeconf parity TODO

Only items still outstanding vs `python/RNS/Utilities/rnodeconf.py` (everything else is already ported and/or covered by sim/integration tests).

## TODO (remaining parity gaps)

- Make autoinstall “guide” UX match Python (greeting text, prompts, flow when `--port` omitted).
- Match Python port listing metadata in prompts (pyserial shows product + serial_number); Go currently only lists device paths.
- Match manual flashing prompts/flow (BOOT/PRG + RESET instructions, when to ask for Enter, per-platform timing).
- Match long-form guidance blocks and rare message text 1:1 (we aligned key ones, but not guaranteed everywhere).
- Verify behaviour on real hardware for ESP32/AVR/NRF52 (timing-sensitive branches, bootloader quirks, actual KISS responses).

