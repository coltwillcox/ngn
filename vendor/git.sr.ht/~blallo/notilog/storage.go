package notilog

import (
	"context"
	"fmt"
	"slices"
	"time"
)

// Storage is append only and exposes a method to access the stored
// notifications, with optional filtering.
type Storage interface {
	// Append is the insertion method.
	Append(context.Context, *Notification) error
	// Query gives access to the saved methods. The list of criteria are
	// meant as and-concatenated.
	Query(context.Context, ...Criterion) ([]*Notification, error)
	// Prune should erase all the notifications from the store.
	Prune(context.Context) error
}

type Criterion struct {
	Field Field
	Op    Operator
	Value any
}

type Field string

const (
	FieldProgram   Field = "program"
	FieldTitle     Field = "title"
	FieldBody      Field = "body"
	FieldSender    Field = "sender"
	FieldSerial    Field = "serial"
	FieldCreatedAt Field = "created_at"
)

type Operator string

const (
	OperatorEq    Operator = "="
	OperatorNe    Operator = "!="
	OperatorIn    Operator = "in"
	OperatorNotIn Operator = "not in"
	OperatorGt    Operator = ">"
	OperatorGe    Operator = ">="
	OperatorLt    Operator = "<"
	OperatorLe    Operator = "<="
)

func (c Criterion) Eval(n *Notification) bool {
	switch c.Field {
	case FieldProgram:
		return c.evalString(n.Program)
	case FieldTitle:
		return c.evalString(n.Title)
	case FieldBody:
		return c.evalString(n.Body)
	case FieldSender:
		return c.evalString(n.Body)
	case FieldSerial:
		return c.evalInt(int(n.Serial))
	case FieldCreatedAt:
		return c.evalDate(n.CreatedAt)
	default:
		panic("invalid field")
	}
}

func (c Criterion) evalString(src string) bool {
	switch c.Op {
	case OperatorEq:
		return src == c.Value
	case OperatorNe:
		return src != c.Value
	case OperatorIn:
		return slices.Contains(c.Value.([]string), src)
	case OperatorNotIn:
		// XXX: this is inefficient
		return !slices.Contains(c.Value.([]string), src)
	default:
		panic("impossible operator for string")
	}
}

func (c Criterion) evalInt(src int) bool {
	switch c.Op {
	case OperatorEq:
		return src == c.Value
	case OperatorNe:
		return src != c.Value
	case OperatorGt:
		return src > c.Value.(int)
	case OperatorGe:
		return src >= c.Value.(int)
	case OperatorLt:
		return src < c.Value.(int)
	case OperatorLe:
		return src <= c.Value.(int)
	case OperatorIn:
		return slices.Contains(c.Value.([]int), src)
	case OperatorNotIn:
		// XXX: this is inefficient
		return !slices.Contains(c.Value.([]int), src)
	default:
		panic("impossible operator for int")
	}
}

func (c Criterion) evalDate(src time.Time) bool {
	switch c.Op {
	case OperatorEq:
		return src.Compare(c.Value.(time.Time)) == 0
	case OperatorNe:
		return src.Compare(c.Value.(time.Time)) != 0
	case OperatorGt:
		return src.Compare(c.Value.(time.Time)) > 0
	case OperatorGe:
		return src.Compare(c.Value.(time.Time)) >= 0
	case OperatorLt:
		return src.Compare(c.Value.(time.Time)) < 0
	case OperatorLe:
		return src.Compare(c.Value.(time.Time)) <= 0
	default:
		panic("impossible operator for date")
	}
}

func ValidateCriterion(c Criterion) error {
	switch c.Field {
	case FieldProgram, FieldTitle, FieldBody, FieldSender:
		return validateString(c)
	case FieldSerial:
		return validateInt(c)
	case FieldCreatedAt:
		return validateDate(c)
	default:
		return fmt.Errorf("non-existing field: %s", c.Field)
	}
}

func validateString(c Criterion) error {
	switch c.Op {
	case OperatorEq, OperatorNe:
		if _, ok := c.Value.(string); !ok {
			return fmt.Errorf("unacceptable value for scalar operator on string")
		}
	case OperatorIn, OperatorNotIn:
		if _, ok := c.Value.([]string); !ok {
			return fmt.Errorf("unacceptable value for list operator on string")
		}
	case OperatorGt, OperatorGe, OperatorLt, OperatorLe:
		return fmt.Errorf("impossible operator on string")
	}

	return nil
}

func validateInt(c Criterion) error {
	switch c.Op {
	case OperatorEq, OperatorNe, OperatorGt, OperatorGe, OperatorLt, OperatorLe:
		if _, ok := c.Value.(int); ok {
			return nil
		}
	case OperatorIn, OperatorNotIn:
		if _, ok := c.Value.([]int); ok {
			return nil
		}
	}

	return fmt.Errorf("unacceptable value for operator on int")
}

func validateDate(c Criterion) error {
	switch c.Op {
	case OperatorIn, OperatorNotIn:
		return fmt.Errorf("impossible operator for date")
	default:
		if _, ok := c.Value.(time.Time); !ok {
			return fmt.Errorf("unacceptable value for date")
		}
	}
	return nil
}
