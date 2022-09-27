package registry

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/segmentio/ksuid"
)

type Client struct {
	Producer     sarama.SyncProducer
	RequestTopic string
	Contract     Contract
}

type Contract struct {
	ChainID *big.Int
	Address common.Address
	Name    string
	Version string
}

type requestData struct {
	ID   string `json:"id"`
	To   string `json:"to"`
	Data string `json:"data"`
}

type MintVehicleSign struct {
	ManufacturerNode *big.Int
	Owner            common.Address
	Attributes       []string
	Infos            []string
}

func anySlice[A any](v []A) []any {
	n := len(v)
	out := make([]any, n)

	for i := 0; i < n; i++ {
		out[i] = v[i]
	}

	return out
}

func (m *MintVehicleSign) Name() string {
	return "MintVehicleSign"
}

func (m *MintVehicleSign) Type() []signer.Type {
	return []signer.Type{
		{Name: "manufacturerNode", Type: "uint256"},
		{Name: "owner", Type: "address"},
		{Name: "attributes", Type: "string[]"},
		{Name: "infos", Type: "string[]"},
	}
}

func (m *MintVehicleSign) Message() signer.TypedDataMessage {
	return signer.TypedDataMessage{
		"manufacturerNode": hexutil.EncodeBig(m.ManufacturerNode),
		"owner":            m.Owner.Hex(),
		"attributes":       anySlice(m.Attributes),
		"infos":            anySlice(m.Infos),
	}
}

type Message interface {
	Name() string
	Type() []signer.Type
	Message() signer.TypedDataMessage
}

// mintVehicleSign(uint256 manufacturerNode, address owner,	string[] calldata attributes, string[] calldata infos, bytes calldata signature)
// MintVehicleSign(uint256 manufacturerNode, address owner, string[] attributes, string[] infos)
func (c *Client) MintVehicleSign(requestID string, manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string, signature []byte) error {
	abi, err := AbiMetaData.GetAbi()
	if err != nil {
		return err
	}

	data, err := abi.Pack("mintVehicleSign", manufacturerNode, owner, attributes, infos, signature)
	if err != nil {
		return err
	}

	return c.sendRequest(requestID, data)
}

func (c *Client) sendRequest(requestID string, data []byte) error {
	event := shared.CloudEvent[requestData]{
		ID:          ksuid.New().String(),
		Source:      "devices-api",
		SpecVersion: "1.0",
		Subject:     requestID,
		Time:        time.Now(),
		Type:        "zone.dimo.transaction.request",
		Data: requestData{
			ID:   requestID,
			To:   hexutil.Encode(c.Contract.Address[:]),
			Data: hexutil.Encode(data),
		},
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, _, err = c.Producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: c.RequestTopic,
			Key:   sarama.StringEncoder(requestID),
			Value: sarama.ByteEncoder(eventBytes),
		},
	)

	return err
}

func (c *Client) GetPayload(msg Message) *signer.TypedData {
	return &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"MintVehicleSign": msg.Type(),
		},
		PrimaryType: msg.Name(),
		Domain: signer.TypedDataDomain{
			Name:              c.Contract.Name,
			Version:           c.Contract.Version,
			ChainId:           (*math.HexOrDecimal256)(c.Contract.ChainID),
			VerifyingContract: c.Contract.Address.Hex(),
		},
		Message: msg.Message(),
	}
}
