// Code generated by github.com/spachava753/fibergql, DO NOT EDIT.

package followschema

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spachava753/fibergql/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

// region    ************************** generated!.gotpl **************************

// endregion ************************** generated!.gotpl **************************

// region    ***************************** args.gotpl *****************************

// endregion ***************************** args.gotpl *****************************

// region    ************************** directives.gotpl **************************

// endregion ************************** directives.gotpl **************************

// region    **************************** field.gotpl *****************************

func (ec *executionContext) _A_id(ctx context.Context, field graphql.CollectedField, obj *A) (ret graphql.Marshaler) {
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	fc := &graphql.FieldContext{
		Object:     "A",
		Field:      field,
		Args:       nil,
		IsMethod:   false,
		IsResolver: false,
	}

	ctx = graphql.WithFieldContext(ctx, fc)
	resTmp := ec._fieldMiddleware(ctx, obj, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return obj.ID, nil
	})

	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNID2string(ctx, field.Selections, res)
}

func (ec *executionContext) _B_id(ctx context.Context, field graphql.CollectedField, obj *B) (ret graphql.Marshaler) {
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	fc := &graphql.FieldContext{
		Object:     "B",
		Field:      field,
		Args:       nil,
		IsMethod:   false,
		IsResolver: false,
	}

	ctx = graphql.WithFieldContext(ctx, fc)
	resTmp := ec._fieldMiddleware(ctx, obj, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return obj.ID, nil
	})

	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNID2string(ctx, field.Selections, res)
}

// endregion **************************** field.gotpl *****************************

// region    **************************** input.gotpl *****************************

// endregion **************************** input.gotpl *****************************

// region    ************************** interface.gotpl ***************************

func (ec *executionContext) _TestUnion(ctx context.Context, sel ast.SelectionSet, obj TestUnion) graphql.Marshaler {
	switch obj := (obj).(type) {
	case nil:
		return graphql.Null
	case A:
		return ec._A(ctx, sel, &obj)
	case *A:
		if obj == nil {
			return graphql.Null
		}
		return ec._A(ctx, sel, obj)
	case B:
		return ec._B(ctx, sel, &obj)
	case *B:
		if obj == nil {
			return graphql.Null
		}
		return ec._B(ctx, sel, obj)
	default:
		panic(fmt.Errorf("unexpected type %T", obj))
	}
}

// endregion ************************** interface.gotpl ***************************

// region    **************************** object.gotpl ****************************

var aImplementors = []string{"A", "TestUnion"}

func (ec *executionContext) _A(ctx context.Context, sel ast.SelectionSet, obj *A) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, aImplementors)
	out := graphql.NewFieldSet(fields)
	var invalids uint32
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("A")
		case "id":
			innerFunc := func(ctx context.Context) (res graphql.Marshaler) {
				return ec._A_id(ctx, field, obj)
			}

			out.Values[i] = innerFunc(ctx)

			if out.Values[i] == graphql.Null {
				invalids++
			}
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}
	out.Dispatch()
	if invalids > 0 {
		return graphql.Null
	}
	return out
}

var bImplementors = []string{"B", "TestUnion"}

func (ec *executionContext) _B(ctx context.Context, sel ast.SelectionSet, obj *B) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, bImplementors)
	out := graphql.NewFieldSet(fields)
	var invalids uint32
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("B")
		case "id":
			innerFunc := func(ctx context.Context) (res graphql.Marshaler) {
				return ec._B_id(ctx, field, obj)
			}

			out.Values[i] = innerFunc(ctx)

			if out.Values[i] == graphql.Null {
				invalids++
			}
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}
	out.Dispatch()
	if invalids > 0 {
		return graphql.Null
	}
	return out
}

// endregion **************************** object.gotpl ****************************

// region    ***************************** type.gotpl *****************************

func (ec *executionContext) marshalOTestUnion2githubᚗcomᚋ99designsᚋgqlgenᚋcodegenᚋtestserverᚋfollowschemaᚐTestUnion(ctx context.Context, sel ast.SelectionSet, v TestUnion) graphql.Marshaler {
	if v == nil {
		return graphql.Null
	}
	return ec._TestUnion(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
