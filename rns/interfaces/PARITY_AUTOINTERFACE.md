# AutoInterface parity TODO

Only items still outstanding vs `python/RNS/Interfaces/AutoInterface.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Interface enumeration parity: Python uses `netinfo` (and has platform-specific ignore lists like `DARWIN_IGNORE_IFS`/`ANDROID_IGNORE_IFS`); Go uses `net.Interfaces()` + `shouldIgnoreAutoInterface()`. Verify ignore/allow semantics match 1:1 per platform.
- Link-local handling: ensure `descopeLinkLocal()` matches Python `descope_linklocal()` for all observed formats (`%ifname` zones, embedded `fe80:<hex>::` patterns).
- Multicast discovery address: confirm `buildMulticastAddress()` reproduces Python’s `mcast_discovery_address` formatting (hash slicing/endian, scope nibble, “ff<type><scope>:…” layout).
- Peering timing: Python sets `reverse_peering_interval = announce_interval*3.25` and uses multiple loops (announce, peer jobs, reverse announce). Go’s loops/timers should be validated against Python timing and backoff behaviour.
- Socket semantics parity: Python sets `SO_REUSEADDR` and optionally `SO_REUSEPORT` and binds unicast + multicast sockets per interface; Go uses `net.ListenConfig` + `joinIPv6Multicast()`. Verify bind addresses, zones, and Windows-specific handling match.
- Peer lifecycle parity: adoption/spawn/cleanup logic (timeouts `PEERING_TIMEOUT`, multicast echo timeouts, multi-interface deque TTL) should match Python behaviour under churn.

