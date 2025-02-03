package ipfs

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

const (
	imagePrefix          = "data:image/png;base64,"
	contentTypeHeaderKey = "Content-Type"
	pngContentType       = "image/png"
	scheme               = "ipfs"
)

type IPFS struct {
	url             *url.URL
	client          *http.Client
	signer          []byte
	permissionValue *big.Int
	permissionArr   []string
	grantee         common.Address
	appName         string
	appID           string
}

func URL(cid string) string {
	return fmt.Sprintf("%s://%s", scheme, cid)
}

type ipfsResponse struct {
	Success bool   `json:"success"`
	CID     string `json:"cid"`
}

func NewGateway(settings *config.Settings, logger *zerolog.Logger) (*IPFS, error) {
	url, err := url.ParseRequestURI(settings.IPFSURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing IPFS URL %q: %w", settings.IPFSURL, err)
	}

	pk, err := base64.RawURLEncoding.DecodeString(settings.IssuerPrivateKey)
	if err != nil {
		return nil, err
	}

	return &IPFS{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		url:             url,
		signer:          pk,
		permissionValue: big.NewInt(settings.SACDPermissionValue),
		permissionArr:   settings.SACDPermissionArr,
		grantee:         settings.SACDGrantee,
		appName:         settings.DIMOAppName,
		appID:           settings.DIMOAppID,
	}, nil
}

func (i *IPFS) UploadImage(ctx context.Context, img string) (string, error) {
	imageData := strings.TrimPrefix(img, imagePrefix)
	image, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	if len(image) == 0 {
		return "", errors.New("empty image field")
	}

	reader := bytes.NewReader(image)
	req, err := http.NewRequest(http.MethodPost, i.url.String(), reader)
	if err != nil {
		return "", fmt.Errorf("failed to create image upload req: %w", err)
	}

	req.Header.Set(contentTypeHeaderKey, pngContentType)
	resp, err := i.client.Do(req.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("IPFS post request failed: %w", err)
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return "", fmt.Errorf("status code %d", code)
	}

	var respb ipfsResponse
	if err := json.NewDecoder(resp.Body).Decode(&respb); err != nil {
		return "", fmt.Errorf("failed to decode IPFS response: %w", err)
	}

	if !respb.Success {
		return "", errors.New("failed to upload image to IPFS")
	}

	return respb.CID, nil
}

func (i *IPFS) FetchImage(ctx context.Context, cid string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, i.url.JoinPath(cid).String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create image upload req: %w", err)
	}

	resp, err := i.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("IPFS post request failed: %w", err)
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return nil, fmt.Errorf("status code %d", code)
	}

	bdy, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read IPFS response: %w", err)
	}

	return bdy, nil
}

func (i *IPFS) UploadSACD(ctx context.Context, driverID string) ([]byte, error) {

	sacd := i.defaultSACDPermissionTemplate(driverID)
	signedDoc, err := i.sign(ctx, sacd)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, signedDoc, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create image upload req: %w", err)

	}
	resp, err := i.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("IPFS post request failed: %w", err)
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return nil, fmt.Errorf("status code %d", code)
	}

	bdy, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read IPFS response: %w", err)
	}

	return bdy, nil
}

// var secp256k1Prefix = []byte{0xe7, 0x01}
// var period byte = '.'

func (i *IPFS) sign(ctx context.Context, document string) (string, error) {
	// header := map[string]any{
	// 	"alg":  "ES256K",
	// 	"crit": []any{"b64"},
	// 	"b64":  false,
	// }
	// hb, err := json.Marshal(header)
	// if err != nil {
	// 	return "", err
	// }
	// hb64 := make([]byte, base64.RawURLEncoding.EncodedLen(len(hb)))
	// base64.RawURLEncoding.Encode(hb64, hb)
	// jw2 := append(hb64, period)
	// jw2 = append(jw2, digest...)
	// enddig := sha256.Sum256(jw2)

	// ecdsa.Sign(rand.Reader, i.signer)

	// hashed := sha256.Sum256([]byte(document))
	// rsa.
	// signature, err := rsa.SignPKCS1v15(rand.Reader, i.signer, crypto.SHA256, hashed[:])
	// if err != nil {
	// 	log.Fatalf("Failed to sign message: %v", err)
	// }
	var signedDocument string
	// TODO
	return signedDocument, nil
}

func (i *IPFS) defaultSACDPermissionTemplate(driverID string) string {
	desc := fmt.Sprintf(`By proceeding, 
		you will grant data access and 
		control functions to %s
		effective as of %+v until %+v. 
		Permissions being granted: %s
		Driver ID: %s App ID: %s
		DIMO Platform, 
		version 1.0.`,
		i.appName,
		time.Now().Format(time.RFC3339),
		time.Now().AddDate(10, 0, 0).Format(time.RFC3339), // expire in 10 years
		strings.Join(i.permissionArr, "; "),
		driverID,
		i.appID)

	return desc
}
