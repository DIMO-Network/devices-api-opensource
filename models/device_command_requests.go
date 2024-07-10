// Code generated by SQLBoiler 4.16.2 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// DeviceCommandRequest is an object representing the database table.
type DeviceCommandRequest struct {
	ID            string    `boil:"id" json:"id" toml:"id" yaml:"id"`
	UserDeviceID  string    `boil:"user_device_id" json:"user_device_id" toml:"user_device_id" yaml:"user_device_id"`
	IntegrationID string    `boil:"integration_id" json:"integration_id" toml:"integration_id" yaml:"integration_id"`
	Command       string    `boil:"command" json:"command" toml:"command" yaml:"command"`
	Status        string    `boil:"status" json:"status" toml:"status" yaml:"status"`
	CreatedAt     time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt     time.Time `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`

	R *deviceCommandRequestR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L deviceCommandRequestL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var DeviceCommandRequestColumns = struct {
	ID            string
	UserDeviceID  string
	IntegrationID string
	Command       string
	Status        string
	CreatedAt     string
	UpdatedAt     string
}{
	ID:            "id",
	UserDeviceID:  "user_device_id",
	IntegrationID: "integration_id",
	Command:       "command",
	Status:        "status",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

var DeviceCommandRequestTableColumns = struct {
	ID            string
	UserDeviceID  string
	IntegrationID string
	Command       string
	Status        string
	CreatedAt     string
	UpdatedAt     string
}{
	ID:            "device_command_requests.id",
	UserDeviceID:  "device_command_requests.user_device_id",
	IntegrationID: "device_command_requests.integration_id",
	Command:       "device_command_requests.command",
	Status:        "device_command_requests.status",
	CreatedAt:     "device_command_requests.created_at",
	UpdatedAt:     "device_command_requests.updated_at",
}

// Generated where

var DeviceCommandRequestWhere = struct {
	ID            whereHelperstring
	UserDeviceID  whereHelperstring
	IntegrationID whereHelperstring
	Command       whereHelperstring
	Status        whereHelperstring
	CreatedAt     whereHelpertime_Time
	UpdatedAt     whereHelpertime_Time
}{
	ID:            whereHelperstring{field: "\"devices_api\".\"device_command_requests\".\"id\""},
	UserDeviceID:  whereHelperstring{field: "\"devices_api\".\"device_command_requests\".\"user_device_id\""},
	IntegrationID: whereHelperstring{field: "\"devices_api\".\"device_command_requests\".\"integration_id\""},
	Command:       whereHelperstring{field: "\"devices_api\".\"device_command_requests\".\"command\""},
	Status:        whereHelperstring{field: "\"devices_api\".\"device_command_requests\".\"status\""},
	CreatedAt:     whereHelpertime_Time{field: "\"devices_api\".\"device_command_requests\".\"created_at\""},
	UpdatedAt:     whereHelpertime_Time{field: "\"devices_api\".\"device_command_requests\".\"updated_at\""},
}

// DeviceCommandRequestRels is where relationship names are stored.
var DeviceCommandRequestRels = struct {
	UserDevice string
}{
	UserDevice: "UserDevice",
}

// deviceCommandRequestR is where relationships are stored.
type deviceCommandRequestR struct {
	UserDevice *UserDevice `boil:"UserDevice" json:"UserDevice" toml:"UserDevice" yaml:"UserDevice"`
}

// NewStruct creates a new relationship struct
func (*deviceCommandRequestR) NewStruct() *deviceCommandRequestR {
	return &deviceCommandRequestR{}
}

func (r *deviceCommandRequestR) GetUserDevice() *UserDevice {
	if r == nil {
		return nil
	}
	return r.UserDevice
}

// deviceCommandRequestL is where Load methods for each relationship are stored.
type deviceCommandRequestL struct{}

var (
	deviceCommandRequestAllColumns            = []string{"id", "user_device_id", "integration_id", "command", "status", "created_at", "updated_at"}
	deviceCommandRequestColumnsWithoutDefault = []string{"id", "user_device_id", "integration_id", "command", "status"}
	deviceCommandRequestColumnsWithDefault    = []string{"created_at", "updated_at"}
	deviceCommandRequestPrimaryKeyColumns     = []string{"id"}
	deviceCommandRequestGeneratedColumns      = []string{}
)

type (
	// DeviceCommandRequestSlice is an alias for a slice of pointers to DeviceCommandRequest.
	// This should almost always be used instead of []DeviceCommandRequest.
	DeviceCommandRequestSlice []*DeviceCommandRequest
	// DeviceCommandRequestHook is the signature for custom DeviceCommandRequest hook methods
	DeviceCommandRequestHook func(context.Context, boil.ContextExecutor, *DeviceCommandRequest) error

	deviceCommandRequestQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	deviceCommandRequestType                 = reflect.TypeOf(&DeviceCommandRequest{})
	deviceCommandRequestMapping              = queries.MakeStructMapping(deviceCommandRequestType)
	deviceCommandRequestPrimaryKeyMapping, _ = queries.BindMapping(deviceCommandRequestType, deviceCommandRequestMapping, deviceCommandRequestPrimaryKeyColumns)
	deviceCommandRequestInsertCacheMut       sync.RWMutex
	deviceCommandRequestInsertCache          = make(map[string]insertCache)
	deviceCommandRequestUpdateCacheMut       sync.RWMutex
	deviceCommandRequestUpdateCache          = make(map[string]updateCache)
	deviceCommandRequestUpsertCacheMut       sync.RWMutex
	deviceCommandRequestUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var deviceCommandRequestAfterSelectMu sync.Mutex
var deviceCommandRequestAfterSelectHooks []DeviceCommandRequestHook

var deviceCommandRequestBeforeInsertMu sync.Mutex
var deviceCommandRequestBeforeInsertHooks []DeviceCommandRequestHook
var deviceCommandRequestAfterInsertMu sync.Mutex
var deviceCommandRequestAfterInsertHooks []DeviceCommandRequestHook

var deviceCommandRequestBeforeUpdateMu sync.Mutex
var deviceCommandRequestBeforeUpdateHooks []DeviceCommandRequestHook
var deviceCommandRequestAfterUpdateMu sync.Mutex
var deviceCommandRequestAfterUpdateHooks []DeviceCommandRequestHook

var deviceCommandRequestBeforeDeleteMu sync.Mutex
var deviceCommandRequestBeforeDeleteHooks []DeviceCommandRequestHook
var deviceCommandRequestAfterDeleteMu sync.Mutex
var deviceCommandRequestAfterDeleteHooks []DeviceCommandRequestHook

var deviceCommandRequestBeforeUpsertMu sync.Mutex
var deviceCommandRequestBeforeUpsertHooks []DeviceCommandRequestHook
var deviceCommandRequestAfterUpsertMu sync.Mutex
var deviceCommandRequestAfterUpsertHooks []DeviceCommandRequestHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *DeviceCommandRequest) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceCommandRequestAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *DeviceCommandRequest) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceCommandRequestBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *DeviceCommandRequest) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceCommandRequestAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *DeviceCommandRequest) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceCommandRequestBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *DeviceCommandRequest) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceCommandRequestAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *DeviceCommandRequest) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceCommandRequestBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *DeviceCommandRequest) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceCommandRequestAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *DeviceCommandRequest) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceCommandRequestBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *DeviceCommandRequest) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceCommandRequestAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddDeviceCommandRequestHook registers your hook function for all future operations.
func AddDeviceCommandRequestHook(hookPoint boil.HookPoint, deviceCommandRequestHook DeviceCommandRequestHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		deviceCommandRequestAfterSelectMu.Lock()
		deviceCommandRequestAfterSelectHooks = append(deviceCommandRequestAfterSelectHooks, deviceCommandRequestHook)
		deviceCommandRequestAfterSelectMu.Unlock()
	case boil.BeforeInsertHook:
		deviceCommandRequestBeforeInsertMu.Lock()
		deviceCommandRequestBeforeInsertHooks = append(deviceCommandRequestBeforeInsertHooks, deviceCommandRequestHook)
		deviceCommandRequestBeforeInsertMu.Unlock()
	case boil.AfterInsertHook:
		deviceCommandRequestAfterInsertMu.Lock()
		deviceCommandRequestAfterInsertHooks = append(deviceCommandRequestAfterInsertHooks, deviceCommandRequestHook)
		deviceCommandRequestAfterInsertMu.Unlock()
	case boil.BeforeUpdateHook:
		deviceCommandRequestBeforeUpdateMu.Lock()
		deviceCommandRequestBeforeUpdateHooks = append(deviceCommandRequestBeforeUpdateHooks, deviceCommandRequestHook)
		deviceCommandRequestBeforeUpdateMu.Unlock()
	case boil.AfterUpdateHook:
		deviceCommandRequestAfterUpdateMu.Lock()
		deviceCommandRequestAfterUpdateHooks = append(deviceCommandRequestAfterUpdateHooks, deviceCommandRequestHook)
		deviceCommandRequestAfterUpdateMu.Unlock()
	case boil.BeforeDeleteHook:
		deviceCommandRequestBeforeDeleteMu.Lock()
		deviceCommandRequestBeforeDeleteHooks = append(deviceCommandRequestBeforeDeleteHooks, deviceCommandRequestHook)
		deviceCommandRequestBeforeDeleteMu.Unlock()
	case boil.AfterDeleteHook:
		deviceCommandRequestAfterDeleteMu.Lock()
		deviceCommandRequestAfterDeleteHooks = append(deviceCommandRequestAfterDeleteHooks, deviceCommandRequestHook)
		deviceCommandRequestAfterDeleteMu.Unlock()
	case boil.BeforeUpsertHook:
		deviceCommandRequestBeforeUpsertMu.Lock()
		deviceCommandRequestBeforeUpsertHooks = append(deviceCommandRequestBeforeUpsertHooks, deviceCommandRequestHook)
		deviceCommandRequestBeforeUpsertMu.Unlock()
	case boil.AfterUpsertHook:
		deviceCommandRequestAfterUpsertMu.Lock()
		deviceCommandRequestAfterUpsertHooks = append(deviceCommandRequestAfterUpsertHooks, deviceCommandRequestHook)
		deviceCommandRequestAfterUpsertMu.Unlock()
	}
}

// One returns a single deviceCommandRequest record from the query.
func (q deviceCommandRequestQuery) One(ctx context.Context, exec boil.ContextExecutor) (*DeviceCommandRequest, error) {
	o := &DeviceCommandRequest{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for device_command_requests")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all DeviceCommandRequest records from the query.
func (q deviceCommandRequestQuery) All(ctx context.Context, exec boil.ContextExecutor) (DeviceCommandRequestSlice, error) {
	var o []*DeviceCommandRequest

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to DeviceCommandRequest slice")
	}

	if len(deviceCommandRequestAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all DeviceCommandRequest records in the query.
func (q deviceCommandRequestQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count device_command_requests rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q deviceCommandRequestQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if device_command_requests exists")
	}

	return count > 0, nil
}

// UserDevice pointed to by the foreign key.
func (o *DeviceCommandRequest) UserDevice(mods ...qm.QueryMod) userDeviceQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.UserDeviceID),
	}

	queryMods = append(queryMods, mods...)

	return UserDevices(queryMods...)
}

// LoadUserDevice allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (deviceCommandRequestL) LoadUserDevice(ctx context.Context, e boil.ContextExecutor, singular bool, maybeDeviceCommandRequest interface{}, mods queries.Applicator) error {
	var slice []*DeviceCommandRequest
	var object *DeviceCommandRequest

	if singular {
		var ok bool
		object, ok = maybeDeviceCommandRequest.(*DeviceCommandRequest)
		if !ok {
			object = new(DeviceCommandRequest)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeDeviceCommandRequest)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeDeviceCommandRequest))
			}
		}
	} else {
		s, ok := maybeDeviceCommandRequest.(*[]*DeviceCommandRequest)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeDeviceCommandRequest)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeDeviceCommandRequest))
			}
		}
	}

	args := make(map[interface{}]struct{})
	if singular {
		if object.R == nil {
			object.R = &deviceCommandRequestR{}
		}
		args[object.UserDeviceID] = struct{}{}

	} else {
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &deviceCommandRequestR{}
			}

			args[obj.UserDeviceID] = struct{}{}

		}
	}

	if len(args) == 0 {
		return nil
	}

	argsSlice := make([]interface{}, len(args))
	i := 0
	for arg := range args {
		argsSlice[i] = arg
		i++
	}

	query := NewQuery(
		qm.From(`devices_api.user_devices`),
		qm.WhereIn(`devices_api.user_devices.id in ?`, argsSlice...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load UserDevice")
	}

	var resultSlice []*UserDevice
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice UserDevice")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for user_devices")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for user_devices")
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
		object.R.UserDevice = foreign
		if foreign.R == nil {
			foreign.R = &userDeviceR{}
		}
		foreign.R.DeviceCommandRequests = append(foreign.R.DeviceCommandRequests, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UserDeviceID == foreign.ID {
				local.R.UserDevice = foreign
				if foreign.R == nil {
					foreign.R = &userDeviceR{}
				}
				foreign.R.DeviceCommandRequests = append(foreign.R.DeviceCommandRequests, local)
				break
			}
		}
	}

	return nil
}

// SetUserDevice of the deviceCommandRequest to the related item.
// Sets o.R.UserDevice to related.
// Adds o to related.R.DeviceCommandRequests.
func (o *DeviceCommandRequest) SetUserDevice(ctx context.Context, exec boil.ContextExecutor, insert bool, related *UserDevice) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"devices_api\".\"device_command_requests\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"user_device_id"}),
		strmangle.WhereClause("\"", "\"", 2, deviceCommandRequestPrimaryKeyColumns),
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

	o.UserDeviceID = related.ID
	if o.R == nil {
		o.R = &deviceCommandRequestR{
			UserDevice: related,
		}
	} else {
		o.R.UserDevice = related
	}

	if related.R == nil {
		related.R = &userDeviceR{
			DeviceCommandRequests: DeviceCommandRequestSlice{o},
		}
	} else {
		related.R.DeviceCommandRequests = append(related.R.DeviceCommandRequests, o)
	}

	return nil
}

// DeviceCommandRequests retrieves all the records using an executor.
func DeviceCommandRequests(mods ...qm.QueryMod) deviceCommandRequestQuery {
	mods = append(mods, qm.From("\"devices_api\".\"device_command_requests\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"devices_api\".\"device_command_requests\".*"})
	}

	return deviceCommandRequestQuery{q}
}

// FindDeviceCommandRequest retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindDeviceCommandRequest(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*DeviceCommandRequest, error) {
	deviceCommandRequestObj := &DeviceCommandRequest{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"devices_api\".\"device_command_requests\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, deviceCommandRequestObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from device_command_requests")
	}

	if err = deviceCommandRequestObj.doAfterSelectHooks(ctx, exec); err != nil {
		return deviceCommandRequestObj, err
	}

	return deviceCommandRequestObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *DeviceCommandRequest) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no device_command_requests provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(deviceCommandRequestColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	deviceCommandRequestInsertCacheMut.RLock()
	cache, cached := deviceCommandRequestInsertCache[key]
	deviceCommandRequestInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			deviceCommandRequestAllColumns,
			deviceCommandRequestColumnsWithDefault,
			deviceCommandRequestColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(deviceCommandRequestType, deviceCommandRequestMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(deviceCommandRequestType, deviceCommandRequestMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"devices_api\".\"device_command_requests\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"devices_api\".\"device_command_requests\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into device_command_requests")
	}

	if !cached {
		deviceCommandRequestInsertCacheMut.Lock()
		deviceCommandRequestInsertCache[key] = cache
		deviceCommandRequestInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the DeviceCommandRequest.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *DeviceCommandRequest) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	deviceCommandRequestUpdateCacheMut.RLock()
	cache, cached := deviceCommandRequestUpdateCache[key]
	deviceCommandRequestUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			deviceCommandRequestAllColumns,
			deviceCommandRequestPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update device_command_requests, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"devices_api\".\"device_command_requests\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, deviceCommandRequestPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(deviceCommandRequestType, deviceCommandRequestMapping, append(wl, deviceCommandRequestPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update device_command_requests row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for device_command_requests")
	}

	if !cached {
		deviceCommandRequestUpdateCacheMut.Lock()
		deviceCommandRequestUpdateCache[key] = cache
		deviceCommandRequestUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q deviceCommandRequestQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for device_command_requests")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for device_command_requests")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o DeviceCommandRequestSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceCommandRequestPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"devices_api\".\"device_command_requests\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, deviceCommandRequestPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in deviceCommandRequest slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all deviceCommandRequest")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *DeviceCommandRequest) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no device_command_requests provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(deviceCommandRequestColumnsWithDefault, o)

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

	deviceCommandRequestUpsertCacheMut.RLock()
	cache, cached := deviceCommandRequestUpsertCache[key]
	deviceCommandRequestUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			deviceCommandRequestAllColumns,
			deviceCommandRequestColumnsWithDefault,
			deviceCommandRequestColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			deviceCommandRequestAllColumns,
			deviceCommandRequestPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert device_command_requests, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(deviceCommandRequestPrimaryKeyColumns))
			copy(conflict, deviceCommandRequestPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"devices_api\".\"device_command_requests\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(deviceCommandRequestType, deviceCommandRequestMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(deviceCommandRequestType, deviceCommandRequestMapping, ret)
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
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert device_command_requests")
	}

	if !cached {
		deviceCommandRequestUpsertCacheMut.Lock()
		deviceCommandRequestUpsertCache[key] = cache
		deviceCommandRequestUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single DeviceCommandRequest record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *DeviceCommandRequest) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no DeviceCommandRequest provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), deviceCommandRequestPrimaryKeyMapping)
	sql := "DELETE FROM \"devices_api\".\"device_command_requests\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from device_command_requests")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for device_command_requests")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q deviceCommandRequestQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no deviceCommandRequestQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from device_command_requests")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for device_command_requests")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o DeviceCommandRequestSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(deviceCommandRequestBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceCommandRequestPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"devices_api\".\"device_command_requests\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, deviceCommandRequestPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from deviceCommandRequest slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for device_command_requests")
	}

	if len(deviceCommandRequestAfterDeleteHooks) != 0 {
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
func (o *DeviceCommandRequest) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindDeviceCommandRequest(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *DeviceCommandRequestSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := DeviceCommandRequestSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceCommandRequestPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"devices_api\".\"device_command_requests\".* FROM \"devices_api\".\"device_command_requests\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, deviceCommandRequestPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in DeviceCommandRequestSlice")
	}

	*o = slice

	return nil
}

// DeviceCommandRequestExists checks if the DeviceCommandRequest row exists.
func DeviceCommandRequestExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"devices_api\".\"device_command_requests\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if device_command_requests exists")
	}

	return exists, nil
}

// Exists checks if the DeviceCommandRequest row exists.
func (o *DeviceCommandRequest) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return DeviceCommandRequestExists(ctx, exec, o.ID)
}
