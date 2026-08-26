// Command lomax-plugin-discogs is a first-party lomax plugin: a
// MetadataResolver backed by the Discogs API, for releases MusicBrainz
// doesn't have — particularly vinyl rips (see docs/music-cli-plan.md
// section 3). Configure via the LOMAX_DISCOGS_TOKEN environment variable
// (https://www.discogs.com/settings/developers); lomax's plugin host
// inherits its own environment into the plugin subprocess, so the host
// process just needs the variable set.
package main

import (
	"context"
	"os"

	hcplugin "github.com/hashicorp/go-plugin"

	"github.com/ferro-dev/lomax/internal/pluginapi"
)

// resolver adapts *Client to pluginapi.MetadataResolver.
type resolver struct {
	client *Client
}

func (r *resolver) ResolveMetadata(ctx context.Context, artist, album, title string) (*pluginapi.Match, error) {
	return r.client.Search(ctx, artist, album, title)
}

func main() {
	impl := &resolver{client: NewClient(os.Getenv("LOMAX_DISCOGS_TOKEN"))}

	hcplugin.Serve(&hcplugin.ServeConfig{
		HandshakeConfig: pluginapi.Handshake,
		Plugins: map[string]hcplugin.Plugin{
			pluginapi.MetadataResolverPluginName: &pluginapi.MetadataResolverPlugin{Impl: impl},
		},
		GRPCServer: hcplugin.DefaultGRPCServer,
	})
}
