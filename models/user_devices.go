// Code generated by SQLBoiler 4.8.3 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// UserDevice is an object representing the database table.
type UserDevice struct {
	ID                 string      `boil:"id" json:"id" toml:"id" yaml:"id"`
	UserID             string      `boil:"user_id" json:"user_id" toml:"user_id" yaml:"user_id"`
	DeviceDefinitionID string      `boil:"device_definition_id" json:"device_definition_id" toml:"device_definition_id" yaml:"device_definition_id"`
	VinIdentifier      null.String `boil:"vin_identifier" json:"vin_identifier,omitempty" toml:"vin_identifier" yaml:"vin_identifier,omitempty"`
	Name               null.String `boil:"name" json:"name,omitempty" toml:"name" yaml:"name,omitempty"`
	CustomImageURL     null.String `boil:"custom_image_url" json:"custom_image_url,omitempty" toml:"custom_image_url" yaml:"custom_image_url,omitempty"`
	Region             null.String `boil:"region" json:"region,omitempty" toml:"region" yaml:"region,omitempty"`
	CreatedAt          time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt          time.Time   `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`

	R *userDeviceR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L userDeviceL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var UserDeviceColumns = struct {
	ID                 string
	UserID             string
	DeviceDefinitionID string
	VinIdentifier      string
	Name               string
	CustomImageURL     string
	Region             string
	CreatedAt          string
	UpdatedAt          string
}{
	ID:                 "id",
	UserID:             "user_id",
	DeviceDefinitionID: "device_definition_id",
	VinIdentifier:      "vin_identifier",
	Name:               "name",
	CustomImageURL:     "custom_image_url",
	Region:             "region",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
}

var UserDeviceTableColumns = struct {
	ID                 string
	UserID             string
	DeviceDefinitionID string
	VinIdentifier      string
	Name               string
	CustomImageURL     string
	Region             string
	CreatedAt          string
	UpdatedAt          string
}{
	ID:                 "user_devices.id",
	UserID:             "user_devices.user_id",
	DeviceDefinitionID: "user_devices.device_definition_id",
	VinIdentifier:      "user_devices.vin_identifier",
	Name:               "user_devices.name",
	CustomImageURL:     "user_devices.custom_image_url",
	Region:             "user_devices.region",
	CreatedAt:          "user_devices.created_at",
	UpdatedAt:          "user_devices.updated_at",
}

// Generated where

var UserDeviceWhere = struct {
	ID                 whereHelperstring
	UserID             whereHelperstring
	DeviceDefinitionID whereHelperstring
	VinIdentifier      whereHelpernull_String
	Name               whereHelpernull_String
	CustomImageURL     whereHelpernull_String
	Region             whereHelpernull_String
	CreatedAt          whereHelpertime_Time
	UpdatedAt          whereHelpertime_Time
}{
	ID:                 whereHelperstring{field: "\"devices_api\".\"user_devices\".\"id\""},
	UserID:             whereHelperstring{field: "\"devices_api\".\"user_devices\".\"user_id\""},
	DeviceDefinitionID: whereHelperstring{field: "\"devices_api\".\"user_devices\".\"device_definition_id\""},
	VinIdentifier:      whereHelpernull_String{field: "\"devices_api\".\"user_devices\".\"vin_identifier\""},
	Name:               whereHelpernull_String{field: "\"devices_api\".\"user_devices\".\"name\""},
	CustomImageURL:     whereHelpernull_String{field: "\"devices_api\".\"user_devices\".\"custom_image_url\""},
	Region:             whereHelpernull_String{field: "\"devices_api\".\"user_devices\".\"region\""},
	CreatedAt:          whereHelpertime_Time{field: "\"devices_api\".\"user_devices\".\"created_at\""},
	UpdatedAt:          whereHelpertime_Time{field: "\"devices_api\".\"user_devices\".\"updated_at\""},
}

// UserDeviceRels is where relationship names are stored.
var UserDeviceRels = struct {
	DeviceDefinition string
}{
	DeviceDefinition: "DeviceDefinition",
}

// userDeviceR is where relationships are stored.
type userDeviceR struct {
	DeviceDefinition *DeviceDefinition `boil:"DeviceDefinition" json:"DeviceDefinition" toml:"DeviceDefinition" yaml:"DeviceDefinition"`
}

// NewStruct creates a new relationship struct
func (*userDeviceR) NewStruct() *userDeviceR {
	return &userDeviceR{}
}

// userDeviceL is where Load methods for each relationship are stored.
type userDeviceL struct{}

var (
	userDeviceAllColumns            = []string{"id", "user_id", "device_definition_id", "vin_identifier", "name", "custom_image_url", "region", "created_at", "updated_at"}
	userDeviceColumnsWithoutDefault = []string{"id", "user_id", "device_definition_id", "vin_identifier", "name", "custom_image_url", "region"}
	userDeviceColumnsWithDefault    = []string{"created_at", "updated_at"}
	userDevicePrimaryKeyColumns     = []string{"id"}
)

type (
	// UserDeviceSlice is an alias for a slice of pointers to UserDevice.
	// This should almost always be used instead of []UserDevice.
	UserDeviceSlice []*UserDevice
	// UserDeviceHook is the signature for custom UserDevice hook methods
	UserDeviceHook func(context.Context, boil.ContextExecutor, *UserDevice) error

	userDeviceQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	userDeviceType                 = reflect.TypeOf(&UserDevice{})
	userDeviceMapping              = queries.MakeStructMapping(userDeviceType)
	userDevicePrimaryKeyMapping, _ = queries.BindMapping(userDeviceType, userDeviceMapping, userDevicePrimaryKeyColumns)
	userDeviceInsertCacheMut       sync.RWMutex
	userDeviceInsertCache          = make(map[string]insertCache)
	userDeviceUpdateCacheMut       sync.RWMutex
	userDeviceUpdateCache          = make(map[string]updateCache)
	userDeviceUpsertCacheMut       sync.RWMutex
	userDeviceUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var userDeviceBeforeInsertHooks []UserDeviceHook
var userDeviceBeforeUpdateHooks []UserDeviceHook
var userDeviceBeforeDeleteHooks []UserDeviceHook
var userDeviceBeforeUpsertHooks []UserDeviceHook

var userDeviceAfterInsertHooks []UserDeviceHook
var userDeviceAfterSelectHooks []UserDeviceHook
var userDeviceAfterUpdateHooks []UserDeviceHook
var userDeviceAfterDeleteHooks []UserDeviceHook
var userDeviceAfterUpsertHooks []UserDeviceHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *UserDevice) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *UserDevice) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *UserDevice) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *UserDevice) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *UserDevice) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *UserDevice) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *UserDevice) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *UserDevice) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *UserDevice) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddUserDeviceHook registers your hook function for all future operations.
func AddUserDeviceHook(hookPoint boil.HookPoint, userDeviceHook UserDeviceHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		userDeviceBeforeInsertHooks = append(userDeviceBeforeInsertHooks, userDeviceHook)
	case boil.BeforeUpdateHook:
		userDeviceBeforeUpdateHooks = append(userDeviceBeforeUpdateHooks, userDeviceHook)
	case boil.BeforeDeleteHook:
		userDeviceBeforeDeleteHooks = append(userDeviceBeforeDeleteHooks, userDeviceHook)
	case boil.BeforeUpsertHook:
		userDeviceBeforeUpsertHooks = append(userDeviceBeforeUpsertHooks, userDeviceHook)
	case boil.AfterInsertHook:
		userDeviceAfterInsertHooks = append(userDeviceAfterInsertHooks, userDeviceHook)
	case boil.AfterSelectHook:
		userDeviceAfterSelectHooks = append(userDeviceAfterSelectHooks, userDeviceHook)
	case boil.AfterUpdateHook:
		userDeviceAfterUpdateHooks = append(userDeviceAfterUpdateHooks, userDeviceHook)
	case boil.AfterDeleteHook:
		userDeviceAfterDeleteHooks = append(userDeviceAfterDeleteHooks, userDeviceHook)
	case boil.AfterUpsertHook:
		userDeviceAfterUpsertHooks = append(userDeviceAfterUpsertHooks, userDeviceHook)
	}
}

// One returns a single userDevice record from the query.
func (q userDeviceQuery) One(ctx context.Context, exec boil.ContextExecutor) (*UserDevice, error) {
	o := &UserDevice{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for user_devices")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all UserDevice records from the query.
func (q userDeviceQuery) All(ctx context.Context, exec boil.ContextExecutor) (UserDeviceSlice, error) {
	var o []*UserDevice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to UserDevice slice")
	}

	if len(userDeviceAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all UserDevice records in the query.
func (q userDeviceQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count user_devices rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q userDeviceQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if user_devices exists")
	}

	return count > 0, nil
}

// DeviceDefinition pointed to by the foreign key.
func (o *UserDevice) DeviceDefinition(mods ...qm.QueryMod) deviceDefinitionQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.DeviceDefinitionID),
	}

	queryMods = append(queryMods, mods...)

	query := DeviceDefinitions(queryMods...)
	queries.SetFrom(query.Query, "\"devices_api\".\"device_definitions\"")

	return query
}

// LoadDeviceDefinition allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (userDeviceL) LoadDeviceDefinition(ctx context.Context, e boil.ContextExecutor, singular bool, maybeUserDevice interface{}, mods queries.Applicator) error {
	var slice []*UserDevice
	var object *UserDevice

	if singular {
		object = maybeUserDevice.(*UserDevice)
	} else {
		slice = *maybeUserDevice.(*[]*UserDevice)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &userDeviceR{}
		}
		args = append(args, object.DeviceDefinitionID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &userDeviceR{}
			}

			for _, a := range args {
				if a == obj.DeviceDefinitionID {
					continue Outer
				}
			}

			args = append(args, obj.DeviceDefinitionID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`devices_api.device_definitions`),
		qm.WhereIn(`devices_api.device_definitions.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load DeviceDefinition")
	}

	var resultSlice []*DeviceDefinition
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice DeviceDefinition")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for device_definitions")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for device_definitions")
	}

	if len(userDeviceAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.DeviceDefinition = foreign
		if foreign.R == nil {
			foreign.R = &deviceDefinitionR{}
		}
		foreign.R.UserDevices = append(foreign.R.UserDevices, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.DeviceDefinitionID == foreign.ID {
				local.R.DeviceDefinition = foreign
				if foreign.R == nil {
					foreign.R = &deviceDefinitionR{}
				}
				foreign.R.UserDevices = append(foreign.R.UserDevices, local)
				break
			}
		}
	}

	return nil
}

// SetDeviceDefinition of the userDevice to the related item.
// Sets o.R.DeviceDefinition to related.
// Adds o to related.R.UserDevices.
func (o *UserDevice) SetDeviceDefinition(ctx context.Context, exec boil.ContextExecutor, insert bool, related *DeviceDefinition) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"devices_api\".\"user_devices\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"device_definition_id"}),
		strmangle.WhereClause("\"", "\"", 2, userDevicePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.DeviceDefinitionID = related.ID
	if o.R == nil {
		o.R = &userDeviceR{
			DeviceDefinition: related,
		}
	} else {
		o.R.DeviceDefinition = related
	}

	if related.R == nil {
		related.R = &deviceDefinitionR{
			UserDevices: UserDeviceSlice{o},
		}
	} else {
		related.R.UserDevices = append(related.R.UserDevices, o)
	}

	return nil
}

// UserDevices retrieves all the records using an executor.
func UserDevices(mods ...qm.QueryMod) userDeviceQuery {
	mods = append(mods, qm.From("\"devices_api\".\"user_devices\""))
	return userDeviceQuery{NewQuery(mods...)}
}

// FindUserDevice retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindUserDevice(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*UserDevice, error) {
	userDeviceObj := &UserDevice{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"devices_api\".\"user_devices\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, userDeviceObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from user_devices")
	}

	if err = userDeviceObj.doAfterSelectHooks(ctx, exec); err != nil {
		return userDeviceObj, err
	}

	return userDeviceObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *UserDevice) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no user_devices provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = currTime
		}
	}

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(userDeviceColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	userDeviceInsertCacheMut.RLock()
	cache, cached := userDeviceInsertCache[key]
	userDeviceInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			userDeviceAllColumns,
			userDeviceColumnsWithDefault,
			userDeviceColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(userDeviceType, userDeviceMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(userDeviceType, userDeviceMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"devices_api\".\"user_devices\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"devices_api\".\"user_devices\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into user_devices")
	}

	if !cached {
		userDeviceInsertCacheMut.Lock()
		userDeviceInsertCache[key] = cache
		userDeviceInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the UserDevice.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *UserDevice) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	userDeviceUpdateCacheMut.RLock()
	cache, cached := userDeviceUpdateCache[key]
	userDeviceUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			userDeviceAllColumns,
			userDevicePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update user_devices, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"devices_api\".\"user_devices\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, userDevicePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(userDeviceType, userDeviceMapping, append(wl, userDevicePrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update user_devices row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for user_devices")
	}

	if !cached {
		userDeviceUpdateCacheMut.Lock()
		userDeviceUpdateCache[key] = cache
		userDeviceUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q userDeviceQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for user_devices")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for user_devices")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o UserDeviceSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userDevicePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"devices_api\".\"user_devices\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, userDevicePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in userDevice slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all userDevice")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *UserDevice) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no user_devices provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
		o.UpdatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(userDeviceColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	userDeviceUpsertCacheMut.RLock()
	cache, cached := userDeviceUpsertCache[key]
	userDeviceUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			userDeviceAllColumns,
			userDeviceColumnsWithDefault,
			userDeviceColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			userDeviceAllColumns,
			userDevicePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert user_devices, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(userDevicePrimaryKeyColumns))
			copy(conflict, userDevicePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"devices_api\".\"user_devices\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(userDeviceType, userDeviceMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(userDeviceType, userDeviceMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert user_devices")
	}

	if !cached {
		userDeviceUpsertCacheMut.Lock()
		userDeviceUpsertCache[key] = cache
		userDeviceUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single UserDevice record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *UserDevice) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no UserDevice provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), userDevicePrimaryKeyMapping)
	sql := "DELETE FROM \"devices_api\".\"user_devices\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from user_devices")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for user_devices")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q userDeviceQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no userDeviceQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from user_devices")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for user_devices")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o UserDeviceSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(userDeviceBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userDevicePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"devices_api\".\"user_devices\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, userDevicePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from userDevice slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for user_devices")
	}

	if len(userDeviceAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *UserDevice) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindUserDevice(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *UserDeviceSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := UserDeviceSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userDevicePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"devices_api\".\"user_devices\".* FROM \"devices_api\".\"user_devices\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, userDevicePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in UserDeviceSlice")
	}

	*o = slice

	return nil
}

// UserDeviceExists checks if the UserDevice row exists.
func UserDeviceExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"devices_api\".\"user_devices\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if user_devices exists")
	}

	return exists, nil
}
