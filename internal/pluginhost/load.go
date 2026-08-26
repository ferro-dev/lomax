package pluginhost

import (
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	hcplugin "github.com/hashicorp/go-plugin"

	"github.com/ferro-dev/lomax/internal/pluginapi"
)

// Handle is a running plugin subprocess. Callers must Close it when done.
type Handle struct {
	Name     string
	Resolver pluginapi.MetadataResolver

	client *hcplugin.Client
}

// Load launches d's binary, performs the go-plugin handshake, and dispenses
// the MetadataResolver service. It returns an error if the binary fails to
// start, doesn't speak the expected protocol version, or doesn't implement
// MetadataResolver.
func Load(d Descriptor) (*Handle, error) {
	client := hcplugin.NewClient(&hcplugin.ClientConfig{
		HandshakeConfig:  pluginapi.Handshake,
		Plugins:          map[string]hcplugin.Plugin{pluginapi.MetadataResolverPluginName: &pluginapi.MetadataResolverPlugin{}},
		Cmd:              exec.Command(d.Path),
		AllowedProtocols: []hcplugin.Protocol{hcplugin.ProtocolGRPC},
		// The handshake/negotiation log is an implementation detail, not
		// something lomax's own CLI output should carry.
		Logger: hclog.NewNullLogger(),
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("pluginhost: connect to plugin %q: %w", d.Name, err)
	}

	raw, err := rpcClient.Dispense(pluginapi.MetadataResolverPluginName)
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("pluginhost: dispense plugin %q: %w", d.Name, err)
	}

	resolver, ok := raw.(pluginapi.MetadataResolver)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("pluginhost: plugin %q does not implement MetadataResolver", d.Name)
	}

	return &Handle{Name: d.Name, Resolver: resolver, client: client}, nil
}

// Close terminates the plugin subprocess.
func (h *Handle) Close() {
	h.client.Kill()
}
