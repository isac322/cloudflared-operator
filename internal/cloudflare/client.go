package cloudflare

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/cloudflare/cloudflare-go"
	"github.com/goccy/go-json"
	"k8s.io/utils/ptr"
)

const (
	secretLen = 64
)

type TunnelCredential struct {
	AccountTag   string `json:"AccountTag"`
	TunnelID     string `json:"TunnelID"`
	TunnelSecret string `json:"TunnelSecret"`
}

type Client interface {
	CreateTunnel(ctx context.Context, accountID, name string) (TunnelCredential, error)
	ValidateTunnelCredential(ctx context.Context, credential TunnelCredential) (bool, error)
	GetOrCreateTunnel(ctx context.Context, accountID, name string) (TunnelCredential, error)
}

type client struct {
	*cloudflare.API
}

func NewClient(token string) (Client, error) {
	cli, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		return nil, err
	}
	return client{cli}, nil
}

func (c client) CreateTunnel(ctx context.Context, accountID, name string) (TunnelCredential, error) {
	secret := make([]byte, secretLen)
	if _, err := rand.Read(secret); err != nil {
		return TunnelCredential{}, err
	}

	encodedTunnelSecret := base64.StdEncoding.EncodeToString(secret)
	tunnel, err := c.API.CreateTunnel(
		ctx,
		&cloudflare.ResourceContainer{
			Identifier: accountID,
			Type:       cloudflare.AccountType,
		},
		cloudflare.TunnelCreateParams{
			Name:      name,
			Secret:    encodedTunnelSecret,
			ConfigSrc: "local",
		},
	)
	if err != nil {
		return TunnelCredential{}, err
	}

	return TunnelCredential{
		AccountTag:   accountID,
		TunnelID:     tunnel.ID,
		TunnelSecret: encodedTunnelSecret,
	}, nil
}

func (c client) ValidateTunnelCredential(ctx context.Context, credential TunnelCredential) (bool, error) {
	origin, err := c.getTunnelCredential(ctx, credential.AccountTag, credential.TunnelID)
	if err != nil {
		return false, err
	}

	return credential == origin, nil
}

func (c client) getTunnelCredential(ctx context.Context, accountID, tunnelID string) (TunnelCredential, error) {
	token, err := c.API.GetTunnelToken(ctx, &cloudflare.ResourceContainer{
		Identifier: accountID,
		Type:       cloudflare.AccountType,
	}, tunnelID)
	if err != nil {
		return TunnelCredential{}, err
	}

	jsonToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return TunnelCredential{}, err
	}

	var tmp struct {
		Token string `json:"s"`
	}
	if err := json.UnmarshalNoEscape(jsonToken, &tmp); err != nil {
		return TunnelCredential{}, err
	}

	return TunnelCredential{
		AccountTag:   accountID,
		TunnelID:     tunnelID,
		TunnelSecret: tmp.Token,
	}, nil
}

func (c client) GetOrCreateTunnel(ctx context.Context, accountID, name string) (TunnelCredential, error) {
	tunnels, _, err := c.API.ListTunnels(
		ctx,
		&cloudflare.ResourceContainer{
			Identifier: accountID,
			Type:       cloudflare.AccountType,
		},
		cloudflare.TunnelListParams{
			Name:      name,
			IsDeleted: ptr.To(false),
		},
	)
	if err != nil {
		return TunnelCredential{}, err
	}
	if len(tunnels) == 0 {
		return c.CreateTunnel(ctx, accountID, name)
	}

	return c.getTunnelCredential(ctx, accountID, tunnels[0].ID)
}
