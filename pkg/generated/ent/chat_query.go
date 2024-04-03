/*
Copyright YEAR 1block.ai.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ent

import (
	"context"
	"fmt"
	"math"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/llmos/llmos-dashboard/pkg/generated/ent/chat"
	"github.com/llmos/llmos-dashboard/pkg/generated/ent/predicate"
	"github.com/llmos/llmos-dashboard/pkg/generated/ent/user"
)

// ChatQuery is the builder for querying Chat entities.
type ChatQuery struct {
	config
	ctx        *QueryContext
	order      []chat.OrderOption
	inters     []Interceptor
	predicates []predicate.Chat
	withOwner  *UserQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the ChatQuery builder.
func (cq *ChatQuery) Where(ps ...predicate.Chat) *ChatQuery {
	cq.predicates = append(cq.predicates, ps...)
	return cq
}

// Limit the number of records to be returned by this query.
func (cq *ChatQuery) Limit(limit int) *ChatQuery {
	cq.ctx.Limit = &limit
	return cq
}

// Offset to start from.
func (cq *ChatQuery) Offset(offset int) *ChatQuery {
	cq.ctx.Offset = &offset
	return cq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (cq *ChatQuery) Unique(unique bool) *ChatQuery {
	cq.ctx.Unique = &unique
	return cq
}

// Order specifies how the records should be ordered.
func (cq *ChatQuery) Order(o ...chat.OrderOption) *ChatQuery {
	cq.order = append(cq.order, o...)
	return cq
}

// QueryOwner chains the current query on the "owner" edge.
func (cq *ChatQuery) QueryOwner() *UserQuery {
	query := (&UserClient{config: cq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := cq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := cq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(chat.Table, chat.FieldID, selector),
			sqlgraph.To(user.Table, user.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, chat.OwnerTable, chat.OwnerColumn),
		)
		fromU = sqlgraph.SetNeighbors(cq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Chat entity from the query.
// Returns a *NotFoundError when no Chat was found.
func (cq *ChatQuery) First(ctx context.Context) (*Chat, error) {
	nodes, err := cq.Limit(1).All(setContextOp(ctx, cq.ctx, "First"))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{chat.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (cq *ChatQuery) FirstX(ctx context.Context) *Chat {
	node, err := cq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Chat ID from the query.
// Returns a *NotFoundError when no Chat ID was found.
func (cq *ChatQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = cq.Limit(1).IDs(setContextOp(ctx, cq.ctx, "FirstID")); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{chat.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (cq *ChatQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := cq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Chat entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Chat entity is found.
// Returns a *NotFoundError when no Chat entities are found.
func (cq *ChatQuery) Only(ctx context.Context) (*Chat, error) {
	nodes, err := cq.Limit(2).All(setContextOp(ctx, cq.ctx, "Only"))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{chat.Label}
	default:
		return nil, &NotSingularError{chat.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (cq *ChatQuery) OnlyX(ctx context.Context) *Chat {
	node, err := cq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Chat ID in the query.
// Returns a *NotSingularError when more than one Chat ID is found.
// Returns a *NotFoundError when no entities are found.
func (cq *ChatQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = cq.Limit(2).IDs(setContextOp(ctx, cq.ctx, "OnlyID")); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{chat.Label}
	default:
		err = &NotSingularError{chat.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (cq *ChatQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := cq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Chats.
func (cq *ChatQuery) All(ctx context.Context) ([]*Chat, error) {
	ctx = setContextOp(ctx, cq.ctx, "All")
	if err := cq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Chat, *ChatQuery]()
	return withInterceptors[[]*Chat](ctx, cq, qr, cq.inters)
}

// AllX is like All, but panics if an error occurs.
func (cq *ChatQuery) AllX(ctx context.Context) []*Chat {
	nodes, err := cq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Chat IDs.
func (cq *ChatQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if cq.ctx.Unique == nil && cq.path != nil {
		cq.Unique(true)
	}
	ctx = setContextOp(ctx, cq.ctx, "IDs")
	if err = cq.Select(chat.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (cq *ChatQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := cq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (cq *ChatQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, cq.ctx, "Count")
	if err := cq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, cq, querierCount[*ChatQuery](), cq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (cq *ChatQuery) CountX(ctx context.Context) int {
	count, err := cq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (cq *ChatQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, cq.ctx, "Exist")
	switch _, err := cq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (cq *ChatQuery) ExistX(ctx context.Context) bool {
	exist, err := cq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the ChatQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (cq *ChatQuery) Clone() *ChatQuery {
	if cq == nil {
		return nil
	}
	return &ChatQuery{
		config:     cq.config,
		ctx:        cq.ctx.Clone(),
		order:      append([]chat.OrderOption{}, cq.order...),
		inters:     append([]Interceptor{}, cq.inters...),
		predicates: append([]predicate.Chat{}, cq.predicates...),
		withOwner:  cq.withOwner.Clone(),
		// clone intermediate query.
		sql:  cq.sql.Clone(),
		path: cq.path,
	}
}

// WithOwner tells the query-builder to eager-load the nodes that are connected to
// the "owner" edge. The optional arguments are used to configure the query builder of the edge.
func (cq *ChatQuery) WithOwner(opts ...func(*UserQuery)) *ChatQuery {
	query := (&UserClient{config: cq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	cq.withOwner = query
	return cq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Title string `json:"title,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.Chat.Query().
//		GroupBy(chat.FieldTitle).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (cq *ChatQuery) GroupBy(field string, fields ...string) *ChatGroupBy {
	cq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &ChatGroupBy{build: cq}
	grbuild.flds = &cq.ctx.Fields
	grbuild.label = chat.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Title string `json:"title,omitempty"`
//	}
//
//	client.Chat.Query().
//		Select(chat.FieldTitle).
//		Scan(ctx, &v)
func (cq *ChatQuery) Select(fields ...string) *ChatSelect {
	cq.ctx.Fields = append(cq.ctx.Fields, fields...)
	sbuild := &ChatSelect{ChatQuery: cq}
	sbuild.label = chat.Label
	sbuild.flds, sbuild.scan = &cq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a ChatSelect configured with the given aggregations.
func (cq *ChatQuery) Aggregate(fns ...AggregateFunc) *ChatSelect {
	return cq.Select().Aggregate(fns...)
}

func (cq *ChatQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range cq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, cq); err != nil {
				return err
			}
		}
	}
	for _, f := range cq.ctx.Fields {
		if !chat.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if cq.path != nil {
		prev, err := cq.path(ctx)
		if err != nil {
			return err
		}
		cq.sql = prev
	}
	return nil
}

func (cq *ChatQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Chat, error) {
	var (
		nodes       = []*Chat{}
		_spec       = cq.querySpec()
		loadedTypes = [1]bool{
			cq.withOwner != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Chat).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Chat{config: cq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, cq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := cq.withOwner; query != nil {
		if err := cq.loadOwner(ctx, query, nodes, nil,
			func(n *Chat, e *User) { n.Edges.Owner = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (cq *ChatQuery) loadOwner(ctx context.Context, query *UserQuery, nodes []*Chat, init func(*Chat), assign func(*Chat, *User)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*Chat)
	for i := range nodes {
		fk := nodes[i].UserID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(user.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "user_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}

func (cq *ChatQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := cq.querySpec()
	_spec.Node.Columns = cq.ctx.Fields
	if len(cq.ctx.Fields) > 0 {
		_spec.Unique = cq.ctx.Unique != nil && *cq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, cq.driver, _spec)
}

func (cq *ChatQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(chat.Table, chat.Columns, sqlgraph.NewFieldSpec(chat.FieldID, field.TypeUUID))
	_spec.From = cq.sql
	if unique := cq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if cq.path != nil {
		_spec.Unique = true
	}
	if fields := cq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, chat.FieldID)
		for i := range fields {
			if fields[i] != chat.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if cq.withOwner != nil {
			_spec.Node.AddColumnOnce(chat.FieldUserID)
		}
	}
	if ps := cq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := cq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := cq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := cq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (cq *ChatQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(cq.driver.Dialect())
	t1 := builder.Table(chat.Table)
	columns := cq.ctx.Fields
	if len(columns) == 0 {
		columns = chat.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if cq.sql != nil {
		selector = cq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if cq.ctx.Unique != nil && *cq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range cq.predicates {
		p(selector)
	}
	for _, p := range cq.order {
		p(selector)
	}
	if offset := cq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := cq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// ChatGroupBy is the group-by builder for Chat entities.
type ChatGroupBy struct {
	selector
	build *ChatQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (cgb *ChatGroupBy) Aggregate(fns ...AggregateFunc) *ChatGroupBy {
	cgb.fns = append(cgb.fns, fns...)
	return cgb
}

// Scan applies the selector query and scans the result into the given value.
func (cgb *ChatGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, cgb.build.ctx, "GroupBy")
	if err := cgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ChatQuery, *ChatGroupBy](ctx, cgb.build, cgb, cgb.build.inters, v)
}

func (cgb *ChatGroupBy) sqlScan(ctx context.Context, root *ChatQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(cgb.fns))
	for _, fn := range cgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*cgb.flds)+len(cgb.fns))
		for _, f := range *cgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*cgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := cgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// ChatSelect is the builder for selecting fields of Chat entities.
type ChatSelect struct {
	*ChatQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (cs *ChatSelect) Aggregate(fns ...AggregateFunc) *ChatSelect {
	cs.fns = append(cs.fns, fns...)
	return cs
}

// Scan applies the selector query and scans the result into the given value.
func (cs *ChatSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, cs.ctx, "Select")
	if err := cs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ChatQuery, *ChatSelect](ctx, cs.ChatQuery, cs, cs.inters, v)
}

func (cs *ChatSelect) sqlScan(ctx context.Context, root *ChatQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(cs.fns))
	for _, fn := range cs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*cs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := cs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
