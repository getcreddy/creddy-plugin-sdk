package sdk

import (
	"context"
	"time"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	pb "github.com/getcreddy/creddy-plugin-sdk/proto"
)

// CredentialGRPCPlugin is the go-plugin implementation for gRPC
type CredentialGRPCPlugin struct {
	plugin.Plugin
	Impl Plugin
}

func (p *CredentialGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterPluginServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *CredentialGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: pb.NewPluginClient(c)}, nil
}

// GRPCServer is the gRPC server that wraps a Plugin implementation
type GRPCServer struct {
	pb.UnimplementedPluginServer
	Impl Plugin
}

func (s *GRPCServer) Info(ctx context.Context, req *pb.InfoRequest) (*pb.InfoResponse, error) {
	info, err := s.Impl.Info(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.InfoResponse{
		Name:             info.Name,
		Version:          info.Version,
		Description:      info.Description,
		MinCreddyVersion: info.MinCreddyVersion,
	}, nil
}

func (s *GRPCServer) Scopes(ctx context.Context, req *pb.ScopesRequest) (*pb.ScopesResponse, error) {
	scopes, err := s.Impl.Scopes(ctx)
	if err != nil {
		return nil, err
	}
	pbScopes := make([]*pb.ScopeSpec, len(scopes))
	for i, scope := range scopes {
		pbScopes[i] = &pb.ScopeSpec{
			Pattern:     scope.Pattern,
			Description: scope.Description,
			Examples:    scope.Examples,
		}
	}
	return &pb.ScopesResponse{Scopes: pbScopes}, nil
}

func (s *GRPCServer) ConfigSchema(ctx context.Context, req *pb.ConfigSchemaRequest) (*pb.ConfigSchemaResponse, error) {
	fields, err := s.Impl.ConfigSchema(ctx)
	if err != nil {
		return nil, err
	}
	pbFields := make([]*pb.ConfigField, len(fields))
	for i, f := range fields {
		pbFields[i] = &pb.ConfigField{
			Name:        f.Name,
			Type:        f.Type,
			Description: f.Description,
			Required:    f.Required,
			Default:     f.Default,
		}
	}
	return &pb.ConfigSchemaResponse{Fields: pbFields}, nil
}

func (s *GRPCServer) Constraints(ctx context.Context, req *pb.ConstraintsRequest) (*pb.ConstraintsResponse, error) {
	constraints, err := s.Impl.Constraints(ctx)
	if err != nil {
		return nil, err
	}
	if constraints == nil {
		return &pb.ConstraintsResponse{HasConstraints: false}, nil
	}
	return &pb.ConstraintsResponse{
		HasConstraints: true,
		MaxTtlSeconds:  int64(constraints.MaxTTL.Seconds()),
		MinTtlSeconds:  int64(constraints.MinTTL.Seconds()),
		Description:    constraints.Description,
	}, nil
}

func (s *GRPCServer) Configure(ctx context.Context, req *pb.ConfigureRequest) (*pb.ConfigureResponse, error) {
	err := s.Impl.Configure(ctx, req.ConfigJson)
	resp := &pb.ConfigureResponse{}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp, nil
}

func (s *GRPCServer) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	err := s.Impl.Validate(ctx)
	resp := &pb.ValidateResponse{Valid: err == nil}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp, nil
}

func (s *GRPCServer) GetCredential(ctx context.Context, req *pb.GetCredentialRequest) (*pb.GetCredentialResponse, error) {
	credReq := &CredentialRequest{
		Agent: Agent{
			ID:     req.Agent.Id,
			Name:   req.Agent.Name,
			Scopes: req.Agent.Scopes,
		},
		Scope:      req.Scope,
		TTL:        time.Duration(req.TtlSeconds) * time.Second,
		Parameters: req.Parameters,
	}

	cred, err := s.Impl.GetCredential(ctx, credReq)
	if err != nil {
		return &pb.GetCredentialResponse{Error: err.Error()}, nil
	}

	return &pb.GetCredentialResponse{
		Value:         cred.Value,
		ExpiresAtUnix: cred.ExpiresAt.Unix(),
		ExternalId:    cred.Credential, // proto field is external_id for backwards compat
		Metadata:      cred.Metadata,
	}, nil
}

func (s *GRPCServer) RevokeCredential(ctx context.Context, req *pb.RevokeCredentialRequest) (*pb.RevokeCredentialResponse, error) {
	err := s.Impl.RevokeCredential(ctx, req.ExternalId)
	resp := &pb.RevokeCredentialResponse{Revoked: err == nil}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp, nil
}

func (s *GRPCServer) MatchScope(ctx context.Context, req *pb.MatchScopeRequest) (*pb.MatchScopeResponse, error) {
	matches, err := s.Impl.MatchScope(ctx, req.Scope)
	if err != nil {
		return nil, err
	}
	return &pb.MatchScopeResponse{Matches: matches}, nil
}

// GRPCClient is the gRPC client that implements the Plugin interface
type GRPCClient struct {
	client pb.PluginClient
}

func (c *GRPCClient) Info(ctx context.Context) (*PluginInfo, error) {
	resp, err := c.client.Info(ctx, &pb.InfoRequest{})
	if err != nil {
		return nil, err
	}
	return &PluginInfo{
		Name:             resp.Name,
		Version:          resp.Version,
		Description:      resp.Description,
		MinCreddyVersion: resp.MinCreddyVersion,
	}, nil
}

func (c *GRPCClient) Scopes(ctx context.Context) ([]ScopeSpec, error) {
	resp, err := c.client.Scopes(ctx, &pb.ScopesRequest{})
	if err != nil {
		return nil, err
	}
	scopes := make([]ScopeSpec, len(resp.Scopes))
	for i, s := range resp.Scopes {
		scopes[i] = ScopeSpec{
			Pattern:     s.Pattern,
			Description: s.Description,
			Examples:    s.Examples,
		}
	}
	return scopes, nil
}

func (c *GRPCClient) ConfigSchema(ctx context.Context) ([]ConfigField, error) {
	resp, err := c.client.ConfigSchema(ctx, &pb.ConfigSchemaRequest{})
	if err != nil {
		return nil, err
	}
	fields := make([]ConfigField, len(resp.Fields))
	for i, f := range resp.Fields {
		fields[i] = ConfigField{
			Name:        f.Name,
			Type:        f.Type,
			Description: f.Description,
			Required:    f.Required,
			Default:     f.Default,
		}
	}
	return fields, nil
}

func (c *GRPCClient) Constraints(ctx context.Context) (*Constraints, error) {
	resp, err := c.client.Constraints(ctx, &pb.ConstraintsRequest{})
	if err != nil {
		return nil, err
	}
	if !resp.HasConstraints {
		return nil, nil
	}
	return &Constraints{
		MaxTTL:      time.Duration(resp.MaxTtlSeconds) * time.Second,
		MinTTL:      time.Duration(resp.MinTtlSeconds) * time.Second,
		Description: resp.Description,
	}, nil
}

func (c *GRPCClient) Configure(ctx context.Context, config string) error {
	resp, err := c.client.Configure(ctx, &pb.ConfigureRequest{ConfigJson: config})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return &PluginError{Message: resp.Error}
	}
	return nil
}

func (c *GRPCClient) Validate(ctx context.Context) error {
	resp, err := c.client.Validate(ctx, &pb.ValidateRequest{})
	if err != nil {
		return err
	}
	if !resp.Valid {
		return &PluginError{Message: resp.Error}
	}
	return nil
}

func (c *GRPCClient) GetCredential(ctx context.Context, req *CredentialRequest) (*Credential, error) {
	pbReq := &pb.GetCredentialRequest{
		Agent: &pb.Agent{
			Id:     req.Agent.ID,
			Name:   req.Agent.Name,
			Scopes: req.Agent.Scopes,
		},
		Scope:      req.Scope,
		TtlSeconds: int64(req.TTL.Seconds()),
		Parameters: req.Parameters,
	}

	resp, err := c.client.GetCredential(ctx, pbReq)
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, &PluginError{Message: resp.Error}
	}

	return &Credential{
		Value:      resp.Value,
		ExpiresAt:  time.Unix(resp.ExpiresAtUnix, 0),
		Credential: resp.ExternalId, // proto field is external_id for backwards compat
		Metadata:   resp.Metadata,
	}, nil
}

func (c *GRPCClient) RevokeCredential(ctx context.Context, externalID string) error {
	resp, err := c.client.RevokeCredential(ctx, &pb.RevokeCredentialRequest{ExternalId: externalID})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return &PluginError{Message: resp.Error}
	}
	return nil
}

func (c *GRPCClient) MatchScope(ctx context.Context, scope string) (bool, error) {
	resp, err := c.client.MatchScope(ctx, &pb.MatchScopeRequest{Scope: scope})
	if err != nil {
		return false, err
	}
	return resp.Matches, nil
}

// PluginError represents an error from a plugin
type PluginError struct {
	Message string
}

func (e *PluginError) Error() string {
	return e.Message
}
