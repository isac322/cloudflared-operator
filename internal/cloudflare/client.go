package cloudflare

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/goccy/go-json"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
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
	ValidateTunnelCredential(ctx context.Context, credential TunnelCredential) (bool, error)
	GetOrCreateTunnel(ctx context.Context, accountID, name string) (TunnelCredential, error)
	CreateDNSRecordForTunnel(ctx context.Context, accountID, tunnelID, domain string, ttl time.Duration) error
}

type client struct {
	*cloudflare.API
	zoneCache *sync.Map
}

func NewClient(token string) (Client, error) {
	cli, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		return nil, err
	}
	return client{cli, &sync.Map{}}, nil
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

func (c client) getZoneIDFromName(ctx context.Context, accountID, zoneName string) (zoneID string, err error) {
	zoneName = normalizeZoneName(zoneName)

	cacheKey := accountID + "-" + zoneName
	if zoneID, ok := c.zoneCache.Load(cacheKey); ok {
		return zoneID.(string), nil
	}

	res, err := c.API.ListZonesContext(ctx, cloudflare.WithZoneFilters(zoneName, accountID, ""))
	if err != nil {
		return "", fmt.Errorf("ListZonesContext command failed: %w", err)
	}

	switch len(res.Result) {
	case 0:
		return "", errors.New("zone could not be found")
	case 1:
		zoneID = res.Result[0].ID
	default:
		return "", errors.New("ambiguous zone name; an account ID might help")
	}

	c.zoneCache.Store(cacheKey, zoneID)
	return zoneID, nil
}

func normalizeZoneName(name string) string {
	if n, err := idna.ToUnicode(name); err == nil {
		return n
	}
	return name
}

func (c client) CreateDNSRecordForTunnel(
	ctx context.Context,
	accountID, tunnelID, domain string,
	ttl time.Duration,
) error {
	domain = normalizeZoneName(domain)
	zoneName, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return err
	}

	zoneID, err := c.getZoneIDFromName(ctx, accountID, zoneName)
	if err != nil {
		return err
	}

	_, err = c.API.CreateDNSRecord(
		ctx,
		&cloudflare.ResourceContainer{
			Identifier: zoneID,
			Type:       cloudflare.ZoneType,
		},
		cloudflare.CreateDNSRecordParams{
			Type:    "CNAME",
			Name:    domain,
			Content: tunnelID,
			TTL:     int(ttl.Seconds()),
			Proxied: ptr.To(true),
		},
	)
	return err
}
