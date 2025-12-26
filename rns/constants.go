package rns

const (
	// ReticulumTruncatedHashLength matches Reticulum.TRUNCATED_HASHLENGTH (bits).
	ReticulumTruncatedHashLength = 128
	// IdentityNameHashLength matches Identity.NAME_HASH_LENGTH (bits).
	IdentityNameHashLength = 80
	// IdentityKeySize matches Identity.KEYSIZE (bits).
	IdentityKeySize = 512
	// IdentityHashLength matches Identity.HASHLENGTH (bits).
	IdentityHashLength = 256
	// IdentitySigLength matches Identity.SIGLENGTH (bits), which equals Ed25519 signature size.
	IdentitySigLength = 512
)

// Compatibility aliases (older Go utilities / Python parity helpers)
const (
	DestIn  = DestinationIN
	DestOut = DestinationOUT
)

const (
	DestinationAllowNone = DestinationALLOW_NONE
	DestinationAllowAll  = DestinationALLOW_ALL
	DestinationAllowList = DestinationALLOW_LIST
)

const (
	HashLengthBytes = IdentityHashLength / 8
	SigLengthBytes  = IdentitySigLength / 8
)
