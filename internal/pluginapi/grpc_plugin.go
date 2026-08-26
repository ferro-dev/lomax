package pluginapi

import (
	"context"
	"fmt"

	hcplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	pb "github.com/ferro-dev/lomax/internal/pluginapi/proto"
)

// Match is a plugin's answer to a metadata lookup — the Go-native shape
// both plugin authors and the host code deal in, so neither has to touch
// generated protobuf types directly.
type Match struct {
	Title       string
	Artist      string
	AlbumArtist string
	Album       string
	Year        int
}

// MetadataResolver is what a plugin implements (server side) and what the
// host gets back from dispensing the plugin (client side) — the same
// Go-native interface on both ends of the wire. Resolve returns (nil, nil)
// when the plugin's source has nothing for this track; that's a normal
// outcome, not an error.
type MetadataResolver interface {
	ResolveMetadata(ctx context.Context, artist, album, title string) (*Match, error)
}

// MetadataResolverPlugin adapts a MetadataResolver to go-plugin's
// GRPCPlugin interface. Plugin binaries set Impl and pass this to
// plugin.Serve; the host leaves Impl nil and only uses GRPCClient (dispense
// never calls GRPCServer on the client side).
type MetadataResolverPlugin struct {
	hcplugin.Plugin
	Impl MetadataResolver
}

func (p *MetadataResolverPlugin) GRPCServer(_ *hcplugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterMetadataResolverServer(s, &grpcServer{impl: p.Impl})
	return nil
}

func (p *MetadataResolverPlugin) GRPCClient(_ context.Context, _ *hcplugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return &grpcClient{client: pb.NewMetadataResolverClient(conn)}, nil
}

// grpcServer runs in the plugin process, translating incoming RPCs to the
// plugin author's MetadataResolver implementation.
type grpcServer struct {
	pb.UnimplementedMetadataResolverServer
	impl MetadataResolver
}

func (s *grpcServer) ResolveMetadata(ctx context.Context, req *pb.ResolveMetadataRequest) (*pb.ResolveMetadataResponse, error) {
	match, err := s.impl.ResolveMetadata(ctx, req.GetArtist(), req.GetAlbum(), req.GetTitle())
	if err != nil {
		return nil, err
	}
	if match == nil {
		return &pb.ResolveMetadataResponse{Matched: false}, nil
	}
	return &pb.ResolveMetadataResponse{
		Matched:     true,
		Title:       match.Title,
		Artist:      match.Artist,
		AlbumArtist: match.AlbumArtist,
		Album:       match.Album,
		Year:        int32(match.Year),
	}, nil
}

// grpcClient runs in the host process, implementing MetadataResolver by
// calling out to the plugin subprocess over gRPC.
type grpcClient struct {
	client pb.MetadataResolverClient
}

func (c *grpcClient) ResolveMetadata(ctx context.Context, artist, album, title string) (*Match, error) {
	resp, err := c.client.ResolveMetadata(ctx, &pb.ResolveMetadataRequest{Artist: artist, Album: album, Title: title})
	if err != nil {
		return nil, fmt.Errorf("pluginapi: ResolveMetadata: %w", err)
	}
	if !resp.GetMatched() {
		return nil, nil
	}
	return &Match{
		Title:       resp.GetTitle(),
		Artist:      resp.GetArtist(),
		AlbumArtist: resp.GetAlbumArtist(),
		Album:       resp.GetAlbum(),
		Year:        int(resp.GetYear()),
	}, nil
}
