// Code generated by SQLBoiler 4.14.1 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// NFTPrivilege is an object representing the database table.
type NFTPrivilege struct {
	ContractAddress []byte        `boil:"contract_address" json:"contract_address" toml:"contract_address" yaml:"contract_address"`
	TokenID         types.Decimal `boil:"token_id" json:"token_id" toml:"token_id" yaml:"token_id"`
	Privilege       int64         `boil:"privilege" json:"privilege" toml:"privilege" yaml:"privilege"`
	UserAddress     []byte        `boil:"user_address" json:"user_address" toml:"user_address" yaml:"user_address"`
	Expiry          time.Time     `boil:"expiry" json:"expiry" toml:"expiry" yaml:"expiry"`
	CreatedAt       time.Time     `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt       time.Time     `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`

	R *nftPrivilegeR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L nftPrivilegeL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var NFTPrivilegeColumns = struct {
	ContractAddress string
	TokenID         string
	Privilege       string
	UserAddress     string
	Expiry          string
	CreatedAt       string
	UpdatedAt       string
}{
	ContractAddress: "contract_address",
	TokenID:         "token_id",
	Privilege:       "privilege",
	UserAddress:     "user_address",
	Expiry:          "expiry",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
}

var NFTPrivilegeTableColumns = struct {
	ContractAddress string
	TokenID         string
	Privilege       string
	UserAddress     string
	Expiry          string
	CreatedAt       string
	UpdatedAt       string
}{
	ContractAddress: "nft_privileges.contract_address",
	TokenID:         "nft_privileges.token_id",
	Privilege:       "nft_privileges.privilege",
	UserAddress:     "nft_privileges.user_address",
	Expiry:          "nft_privileges.expiry",
	CreatedAt:       "nft_privileges.created_at",
	UpdatedAt:       "nft_privileges.updated_at",
}

// Generated where

type whereHelpertypes_Decimal struct{ field string }

func (w whereHelpertypes_Decimal) EQ(x types.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertypes_Decimal) NEQ(x types.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertypes_Decimal) LT(x types.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertypes_Decimal) LTE(x types.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertypes_Decimal) GT(x types.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertypes_Decimal) GTE(x types.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

type whereHelperint64 struct{ field string }

func (w whereHelperint64) EQ(x int64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperint64) NEQ(x int64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperint64) LT(x int64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperint64) LTE(x int64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperint64) GT(x int64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperint64) GTE(x int64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperint64) IN(slice []int64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperint64) NIN(slice []int64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

var NFTPrivilegeWhere = struct {
	ContractAddress whereHelper__byte
	TokenID         whereHelpertypes_Decimal
	Privilege       whereHelperint64
	UserAddress     whereHelper__byte
	Expiry          whereHelpertime_Time
	CreatedAt       whereHelpertime_Time
	UpdatedAt       whereHelpertime_Time
}{
	ContractAddress: whereHelper__byte{field: "\"devices_api\".\"nft_privileges\".\"contract_address\""},
	TokenID:         whereHelpertypes_Decimal{field: "\"devices_api\".\"nft_privileges\".\"token_id\""},
	Privilege:       whereHelperint64{field: "\"devices_api\".\"nft_privileges\".\"privilege\""},
	UserAddress:     whereHelper__byte{field: "\"devices_api\".\"nft_privileges\".\"user_address\""},
	Expiry:          whereHelpertime_Time{field: "\"devices_api\".\"nft_privileges\".\"expiry\""},
	CreatedAt:       whereHelpertime_Time{field: "\"devices_api\".\"nft_privileges\".\"created_at\""},
	UpdatedAt:       whereHelpertime_Time{field: "\"devices_api\".\"nft_privileges\".\"updated_at\""},
}

// NFTPrivilegeRels is where relationship names are stored.
var NFTPrivilegeRels = struct {
}{}

// nftPrivilegeR is where relationships are stored.
type nftPrivilegeR struct {
}

// NewStruct creates a new relationship struct
func (*nftPrivilegeR) NewStruct() *nftPrivilegeR {
	return &nftPrivilegeR{}
}

// nftPrivilegeL is where Load methods for each relationship are stored.
type nftPrivilegeL struct{}

var (
	nftPrivilegeAllColumns            = []string{"contract_address", "token_id", "privilege", "user_address", "expiry", "created_at", "updated_at"}
	nftPrivilegeColumnsWithoutDefault = []string{"contract_address", "token_id", "privilege", "user_address", "expiry"}
	nftPrivilegeColumnsWithDefault    = []string{"created_at", "updated_at"}
	nftPrivilegePrimaryKeyColumns     = []string{"contract_address", "token_id", "privilege", "user_address"}
	nftPrivilegeGeneratedColumns      = []string{}
)

type (
	// NFTPrivilegeSlice is an alias for a slice of pointers to NFTPrivilege.
	// This should almost always be used instead of []NFTPrivilege.
	NFTPrivilegeSlice []*NFTPrivilege
	// NFTPrivilegeHook is the signature for custom NFTPrivilege hook methods
	NFTPrivilegeHook func(context.Context, boil.ContextExecutor, *NFTPrivilege) error

	nftPrivilegeQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	nftPrivilegeType                 = reflect.TypeOf(&NFTPrivilege{})
	nftPrivilegeMapping              = queries.MakeStructMapping(nftPrivilegeType)
	nftPrivilegePrimaryKeyMapping, _ = queries.BindMapping(nftPrivilegeType, nftPrivilegeMapping, nftPrivilegePrimaryKeyColumns)
	nftPrivilegeInsertCacheMut       sync.RWMutex
	nftPrivilegeInsertCache          = make(map[string]insertCache)
	nftPrivilegeUpdateCacheMut       sync.RWMutex
	nftPrivilegeUpdateCache          = make(map[string]updateCache)
	nftPrivilegeUpsertCacheMut       sync.RWMutex
	nftPrivilegeUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var nftPrivilegeAfterSelectHooks []NFTPrivilegeHook

var nftPrivilegeBeforeInsertHooks []NFTPrivilegeHook
var nftPrivilegeAfterInsertHooks []NFTPrivilegeHook

var nftPrivilegeBeforeUpdateHooks []NFTPrivilegeHook
var nftPrivilegeAfterUpdateHooks []NFTPrivilegeHook

var nftPrivilegeBeforeDeleteHooks []NFTPrivilegeHook
var nftPrivilegeAfterDeleteHooks []NFTPrivilegeHook

var nftPrivilegeBeforeUpsertHooks []NFTPrivilegeHook
var nftPrivilegeAfterUpsertHooks []NFTPrivilegeHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *NFTPrivilege) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range nftPrivilegeAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *NFTPrivilege) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range nftPrivilegeBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *NFTPrivilege) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range nftPrivilegeAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *NFTPrivilege) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range nftPrivilegeBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *NFTPrivilege) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range nftPrivilegeAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *NFTPrivilege) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range nftPrivilegeBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *NFTPrivilege) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range nftPrivilegeAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *NFTPrivilege) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range nftPrivilegeBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *NFTPrivilege) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range nftPrivilegeAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddNFTPrivilegeHook registers your hook function for all future operations.
func AddNFTPrivilegeHook(hookPoint boil.HookPoint, nftPrivilegeHook NFTPrivilegeHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		nftPrivilegeAfterSelectHooks = append(nftPrivilegeAfterSelectHooks, nftPrivilegeHook)
	case boil.BeforeInsertHook:
		nftPrivilegeBeforeInsertHooks = append(nftPrivilegeBeforeInsertHooks, nftPrivilegeHook)
	case boil.AfterInsertHook:
		nftPrivilegeAfterInsertHooks = append(nftPrivilegeAfterInsertHooks, nftPrivilegeHook)
	case boil.BeforeUpdateHook:
		nftPrivilegeBeforeUpdateHooks = append(nftPrivilegeBeforeUpdateHooks, nftPrivilegeHook)
	case boil.AfterUpdateHook:
		nftPrivilegeAfterUpdateHooks = append(nftPrivilegeAfterUpdateHooks, nftPrivilegeHook)
	case boil.BeforeDeleteHook:
		nftPrivilegeBeforeDeleteHooks = append(nftPrivilegeBeforeDeleteHooks, nftPrivilegeHook)
	case boil.AfterDeleteHook:
		nftPrivilegeAfterDeleteHooks = append(nftPrivilegeAfterDeleteHooks, nftPrivilegeHook)
	case boil.BeforeUpsertHook:
		nftPrivilegeBeforeUpsertHooks = append(nftPrivilegeBeforeUpsertHooks, nftPrivilegeHook)
	case boil.AfterUpsertHook:
		nftPrivilegeAfterUpsertHooks = append(nftPrivilegeAfterUpsertHooks, nftPrivilegeHook)
	}
}

// One returns a single nftPrivilege record from the query.
func (q nftPrivilegeQuery) One(ctx context.Context, exec boil.ContextExecutor) (*NFTPrivilege, error) {
	o := &NFTPrivilege{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for nft_privileges")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all NFTPrivilege records from the query.
func (q nftPrivilegeQuery) All(ctx context.Context, exec boil.ContextExecutor) (NFTPrivilegeSlice, error) {
	var o []*NFTPrivilege

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to NFTPrivilege slice")
	}

	if len(nftPrivilegeAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all NFTPrivilege records in the query.
func (q nftPrivilegeQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count nft_privileges rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q nftPrivilegeQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if nft_privileges exists")
	}

	return count > 0, nil
}

// NFTPrivileges retrieves all the records using an executor.
func NFTPrivileges(mods ...qm.QueryMod) nftPrivilegeQuery {
	mods = append(mods, qm.From("\"devices_api\".\"nft_privileges\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"devices_api\".\"nft_privileges\".*"})
	}

	return nftPrivilegeQuery{q}
}

// FindNFTPrivilege retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindNFTPrivilege(ctx context.Context, exec boil.ContextExecutor, contractAddress []byte, tokenID types.Decimal, privilege int64, userAddress []byte, selectCols ...string) (*NFTPrivilege, error) {
	nftPrivilegeObj := &NFTPrivilege{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"devices_api\".\"nft_privileges\" where \"contract_address\"=$1 AND \"token_id\"=$2 AND \"privilege\"=$3 AND \"user_address\"=$4", sel,
	)

	q := queries.Raw(query, contractAddress, tokenID, privilege, userAddress)

	err := q.Bind(ctx, exec, nftPrivilegeObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from nft_privileges")
	}

	if err = nftPrivilegeObj.doAfterSelectHooks(ctx, exec); err != nil {
		return nftPrivilegeObj, err
	}

	return nftPrivilegeObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *NFTPrivilege) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no nft_privileges provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(nftPrivilegeColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	nftPrivilegeInsertCacheMut.RLock()
	cache, cached := nftPrivilegeInsertCache[key]
	nftPrivilegeInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			nftPrivilegeAllColumns,
			nftPrivilegeColumnsWithDefault,
			nftPrivilegeColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(nftPrivilegeType, nftPrivilegeMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(nftPrivilegeType, nftPrivilegeMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"devices_api\".\"nft_privileges\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"devices_api\".\"nft_privileges\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into nft_privileges")
	}

	if !cached {
		nftPrivilegeInsertCacheMut.Lock()
		nftPrivilegeInsertCache[key] = cache
		nftPrivilegeInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the NFTPrivilege.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *NFTPrivilege) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	nftPrivilegeUpdateCacheMut.RLock()
	cache, cached := nftPrivilegeUpdateCache[key]
	nftPrivilegeUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			nftPrivilegeAllColumns,
			nftPrivilegePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update nft_privileges, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"devices_api\".\"nft_privileges\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, nftPrivilegePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(nftPrivilegeType, nftPrivilegeMapping, append(wl, nftPrivilegePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update nft_privileges row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for nft_privileges")
	}

	if !cached {
		nftPrivilegeUpdateCacheMut.Lock()
		nftPrivilegeUpdateCache[key] = cache
		nftPrivilegeUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q nftPrivilegeQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for nft_privileges")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for nft_privileges")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o NFTPrivilegeSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), nftPrivilegePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"devices_api\".\"nft_privileges\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, nftPrivilegePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in nftPrivilege slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all nftPrivilege")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *NFTPrivilege) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no nft_privileges provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(nftPrivilegeColumnsWithDefault, o)

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

	nftPrivilegeUpsertCacheMut.RLock()
	cache, cached := nftPrivilegeUpsertCache[key]
	nftPrivilegeUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			nftPrivilegeAllColumns,
			nftPrivilegeColumnsWithDefault,
			nftPrivilegeColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			nftPrivilegeAllColumns,
			nftPrivilegePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert nft_privileges, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(nftPrivilegePrimaryKeyColumns))
			copy(conflict, nftPrivilegePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"devices_api\".\"nft_privileges\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(nftPrivilegeType, nftPrivilegeMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(nftPrivilegeType, nftPrivilegeMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert nft_privileges")
	}

	if !cached {
		nftPrivilegeUpsertCacheMut.Lock()
		nftPrivilegeUpsertCache[key] = cache
		nftPrivilegeUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single NFTPrivilege record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *NFTPrivilege) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no NFTPrivilege provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), nftPrivilegePrimaryKeyMapping)
	sql := "DELETE FROM \"devices_api\".\"nft_privileges\" WHERE \"contract_address\"=$1 AND \"token_id\"=$2 AND \"privilege\"=$3 AND \"user_address\"=$4"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from nft_privileges")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for nft_privileges")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q nftPrivilegeQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no nftPrivilegeQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from nft_privileges")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for nft_privileges")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o NFTPrivilegeSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(nftPrivilegeBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), nftPrivilegePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"devices_api\".\"nft_privileges\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, nftPrivilegePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from nftPrivilege slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for nft_privileges")
	}

	if len(nftPrivilegeAfterDeleteHooks) != 0 {
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
func (o *NFTPrivilege) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindNFTPrivilege(ctx, exec, o.ContractAddress, o.TokenID, o.Privilege, o.UserAddress)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *NFTPrivilegeSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := NFTPrivilegeSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), nftPrivilegePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"devices_api\".\"nft_privileges\".* FROM \"devices_api\".\"nft_privileges\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, nftPrivilegePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in NFTPrivilegeSlice")
	}

	*o = slice

	return nil
}

// NFTPrivilegeExists checks if the NFTPrivilege row exists.
func NFTPrivilegeExists(ctx context.Context, exec boil.ContextExecutor, contractAddress []byte, tokenID types.Decimal, privilege int64, userAddress []byte) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"devices_api\".\"nft_privileges\" where \"contract_address\"=$1 AND \"token_id\"=$2 AND \"privilege\"=$3 AND \"user_address\"=$4 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, contractAddress, tokenID, privilege, userAddress)
	}
	row := exec.QueryRowContext(ctx, sql, contractAddress, tokenID, privilege, userAddress)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if nft_privileges exists")
	}

	return exists, nil
}

// Exists checks if the NFTPrivilege row exists.
func (o *NFTPrivilege) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return NFTPrivilegeExists(ctx, exec, o.ContractAddress, o.TokenID, o.Privilege, o.UserAddress)
}
