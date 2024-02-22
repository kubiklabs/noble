package types

const (
	// ModuleName defines the module name
	ModuleName = "autocctp"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_autocctp"

	// Version defines the current version the IBC module supports
	Version = "autocctp-1"

	// PortID is the default port id that module binds to
	PortID = "autocctp"
)

var (
	// PortKey defines the key to store the port ID in store
	PortKey = KeyPrefix("autocctp-port-")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
