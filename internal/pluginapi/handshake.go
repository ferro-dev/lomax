// Package pluginapi is lomax's plugin protocol: the versioned handshake and
// gRPC service contract subprocess plugins implement, hosted over
// HashiCorp go-plugin (see docs/music-cli-plan.md section 9). Both the host
// (internal/pluginhost) and plugin binaries (plugins/<name>) depend on this
// package so they always agree on the wire contract.
package pluginapi

import (
	hcplugin "github.com/hashicorp/go-plugin"
)

// Handshake is shared by every lomax plugin and the host. ProtocolVersion
// must bump on any breaking change to the RPC contract in proto/plugin.proto
// — go-plugin refuses to connect a client and server whose versions differ,
// which is exactly the "core guarantees backward compatibility within a
// major version" promise from section 9.
var Handshake = hcplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "LOMAX_PLUGIN",
	MagicCookieValue: "lomax-plugin-v1",
}

// MetadataResolverPluginName is the key both ServeConfig.Plugins (on the
// plugin side) and ClientConfig.Plugins (on the host side) register the
// MetadataResolver service under.
const MetadataResolverPluginName = "metadata_resolver"
