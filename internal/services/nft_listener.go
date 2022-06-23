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
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type NFTListener struct {
	db  func() *database.DBReaderWriter
	log *zerolog.Logger
}

type MintSuccessData struct {
	RequestID string   `json:"requestId"`
	TokenID   *big.Int `json:"tokenId"`
}

func NewNFTListener(db func() *database.DBReaderWriter, log *zerolog.Logger) *NFTListener {
	return &NFTListener{db: db, log: log}
}

func (i *NFTListener) ProcessMintStatus(messages <-chan *message.Message) {
	for msg := range messages {
		err := i.processMessage(msg)
		if err != nil {
			i.log.Err(err).Msg("error processing task status message")
		}
	}
}

func (i *NFTListener) processMessage(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

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

	ud, err := models.FindUserDevice(ctx, i.db().Writer, mr.UserDeviceID)
	if err != nil {
		return err
	}

	n := new(decimal.Big)
	n.SetBigMantScale(event.Data.TokenID, 0)
	ud.TokenID = types.NewNullDecimal(n)
	if _, err := ud.Update(ctx, i.db().Writer, boil.Infer()); err != nil {
		return err
	}

	return nil
}
