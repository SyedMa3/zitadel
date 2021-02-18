package command

import (
	"context"
	"time"

	"github.com/caos/zitadel/internal/errors"
	"github.com/caos/zitadel/internal/telemetry/tracing"
	"github.com/caos/zitadel/internal/v2/domain"
	"github.com/caos/zitadel/internal/v2/repository/user"
)

var (
	//most of us won't survive until 12-31-9999 23:59:59, maybe ZITADEL does
	defaultExpDate = time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC)
)

func (r *CommandSide) AddUserMachineKey(ctx context.Context, machineKey *domain.MachineKey, resourceOwner string) (*domain.MachineKey, error) {
	err := r.checkUserExists(ctx, machineKey.AggregateID, resourceOwner)
	if err != nil {
		return nil, err
	}
	keyID, err := r.idGenerator.Next()
	if err != nil {
		return nil, err
	}
	keyWriteModel := NewMachineKeyWriteModel(machineKey.AggregateID, keyID, resourceOwner)
	err = r.eventstore.FilterToQueryReducer(ctx, keyWriteModel)
	if err != nil {
		return nil, err
	}

	if machineKey.ExpirationDate.IsZero() {
		machineKey.ExpirationDate = defaultExpDate
	}
	if machineKey.ExpirationDate.Before(time.Now()) {
		return nil, errors.ThrowInvalidArgument(nil, "COMMAND-38vns", "Errors.MachineKey.ExpireBeforeNow")
	}

	err = machineKey.GenerateNewMachineKeyPair(r.machineKeySize)
	if err != nil {
		return nil, err
	}

	events, err := r.eventstore.PushEvents(ctx,
		user.NewMachineKeyAddedEvent(
			ctx,
			UserAggregateFromWriteModel(&keyWriteModel.WriteModel),
			keyID,
			machineKey.Type,
			machineKey.ExpirationDate,
			machineKey.PublicKey))
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(keyWriteModel, events...)
	if err != nil {
		return nil, err
	}

	key := keyWriteModelToMachineKey(keyWriteModel)
	key.PrivateKey = machineKey.PrivateKey
	return key, nil
}

func (r *CommandSide) RemoveUserMachineKey(ctx context.Context, userID, keyID, resourceOwner string) error {
	keyWriteModel, err := r.machineKeyWriteModelByID(ctx, userID, keyID, resourceOwner)
	if err != nil {
		return err
	}
	if !keyWriteModel.Exists() {
		return errors.ThrowNotFound(nil, "COMMAND-4m77G", "Errors.User.Machine.Key.NotFound")
	}

	_, err = r.eventstore.PushEvents(ctx,
		user.NewMachineKeyRemovedEvent(ctx, UserAggregateFromWriteModel(&keyWriteModel.WriteModel), keyID))
	return err
}

func (r *CommandSide) machineKeyWriteModelByID(ctx context.Context, userID, keyID, resourceOwner string) (writeModel *MachineKeyWriteModel, err error) {
	if userID == "" {
		return nil, errors.ThrowInvalidArgument(nil, "COMMAND-4n8vs", "Errors.User.UserIDMissing")
	}
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	writeModel = NewMachineKeyWriteModel(userID, keyID, resourceOwner)
	err = r.eventstore.FilterToQueryReducer(ctx, writeModel)
	if err != nil {
		return nil, err
	}
	return writeModel, nil
}