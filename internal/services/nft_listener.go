package services

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type NFTListener struct {
	db  func() *database.DBReaderWriter
	log *zerolog.Logger
}

type MintSuccessData struct {
	RequestID string   `json:"requestId"`
	TxHash    *string  `json:"txHash"`
	TokenID   *big.Int `json:"tokenId"`
	Status    string   `json:"status"`
}

func NewNFTListener(db func() *database.DBReaderWriter, log *zerolog.Logger) *NFTListener {
	return &NFTListener{db: db, log: log}
}

func (i *NFTListener) ProcessMintStatus(messages <-chan *message.Message) {
	for msg := range messages {
		err := i.processMessage(msg)
		if err != nil {
			i.log.Err(err).Msg("error processing NFT mint status message")
		}
	}
}

func (i *NFTListener) processMessage(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

	i.log.Info().RawJSON("mintStatus", msg.Payload).Msg("Got mint.")

	event := new(shared.CloudEvent[MintSuccessData])
	if err := json.Unmarshal(msg.Payload, event); err != nil {
		return errors.Wrap(err, "error parsing mint status")
	}

	return i.processEvent(event)
}

func (i *NFTListener) processEvent(event *shared.CloudEvent[MintSuccessData]) error {
	var (
		ctx = context.Background()
	)

	mr, err := models.FindMintRequest(ctx, i.db().Writer, event.Data.RequestID)
	if err != nil {
		return err
	}

	switch event.Data.Status {
	case models.TxstateSubmitted:
		mr.TXState = models.TxstateSubmitted
		if event.Data.TxHash != nil {
			mr.TXHash = null.BytesFrom(common.FromHex(*event.Data.TxHash))
		}

		if _, err := mr.Update(ctx, i.db().Writer, boil.Infer()); err != nil {
			return err
		}
	case models.TxstateConfirmed:
		n := new(decimal.Big)
		n.SetBigMantScale(event.Data.TokenID, 0)

		mr.TXState = models.TxstateConfirmed
		mr.TokenID = types.NewNullDecimal(n)
		if event.Data.TxHash != nil {
			// This should always be here, for now.
			mr.TXHash = null.BytesFrom(common.FromHex(*event.Data.TxHash))
		}

		if _, err := mr.Update(ctx, i.db().Writer, boil.Infer()); err != nil {
			return err
		}
	}

	return nil
}
