# AX25KISSInterface parity TODO

Only items still outstanding vs `python/RNS/Interfaces/AX25KISSInterface.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Serial library parity: Python uses `pyserial`, Go uses `github.com/tarm/serial`. Verify port option mappings (baud/databits/stopbits/parity/timeouts) match on all platforms.
- Startup timing/flow: Python sleeps and configures KISS parameters; Go mirrors this but should be verified on real hardware (timing-sensitive).
- Flow-control READY handling: Python enables READY notifications; ensure Go’s `FlowControl` behaviour and timeout recovery match Python’s semantics and log messages.
- Frame parsing robustness: verify the read-loop behaviour matches Python for partial frames, timeouts (`FrameTimeout`), and handling of unexpected KISS commands.
- AX.25 header details: confirm address encoding, SSID bits, and UI/PID constants match Python (and real-world AX.25 expectations).

