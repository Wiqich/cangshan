package orm

import (
	"bytes"
	"container/list"
	"fmt"
)

type Condition struct {
	buffer   *bytes.Buffer
	args     *list.List
	closed   bool
	sub      *Condition
	closable bool
}

func newCondition() *Condition {
	return &Condition{
		buffer:   new(bytes.Buffer),
		args:     list.New(),
		closed:   false,
		sub:      nil,
		closable: false,
	}
}

func (condition *Condition) close() {
	if condition.sub != nil {
		condition.close()
		condition.sub = nil
	}
	condition.closed = true
	fmt.Fprintf(condition.buffer, ")")
}

func (condition *Condition) Sub() *Condition {
	if condition.sub == nil && !condition.closed && !condition.closable {
		condition.sub = &Condition{
			buffer:   condition.buffer,
			args:     condition.args,
			closed:   false,
			sub:      nil,
			closable: false,
		}
		condition.closable = true
		return condition.sub
	}
	return condition
}

func (condition *Condition) Equal(name string, value interface{}) *Condition {
	if !condition.closed && !condition.closable {
		fmt.Fprintf(condition.buffer, "%s = ?", name)
		condition.args.PushBack(value)
		condition.closable = true
	}
	return condition
}

func (condition *Condition) NotEqual(name string, value interface{}) *Condition {
	if !condition.closed && !condition.closable {
		fmt.Fprintf(condition.buffer, "%s != ?", name)
		condition.args.PushBack(value)
		condition.closable = true
	}
	return condition
}

func (condition *Condition) Greater(name string, value interface{}) *Condition {
	if !condition.closed && !condition.closable {
		fmt.Fprintf(condition.buffer, "%s > ?", name)
		condition.args.PushBack(value)
		condition.closable = true
	}
	return condition
}

func (condition *Condition) GreaterOrEqual(name string, value interface{}) *Condition {
	if !condition.closed && !condition.closable {
		fmt.Fprintf(condition.buffer, "%s >= ?", name)
		condition.args.PushBack(value)
		condition.closable = true
	}
	return condition
}

func (condition *Condition) Less(name string, value interface{}) *Condition {
	if !condition.closed && !condition.closable {
		fmt.Fprintf(condition.buffer, "%s < ?", name)
		condition.args.PushBack(value)
		condition.closable = true
	}
	return condition
}

func (condition *Condition) LessOrEqual(name string, value interface{}) *Condition {
	if !condition.closed && !condition.closable {
		fmt.Fprintf(condition.buffer, "%s <= ?", name)
		condition.args.PushBack(value)
		condition.closable = true
	}
	return condition
}

func (condition *Condition) Like(name string, value interface{}) *Condition {
	if !condition.closed && !condition.closable {
		fmt.Fprintf(condition.buffer, "%s LIKE ?", name)
		condition.args.PushBack(value)
		condition.closable = true
	}
	return condition
}

func (condition *Condition) NotLike(name string, value interface{}) *Condition {
	if !condition.closed && !condition.closable {
		fmt.Fprintf(condition.buffer, "%s NOT LIKE ?", name)
		condition.args.PushBack(value)
		condition.closable = true
	}
	return condition
}

func (condition *Condition) In(name string, value interface{}) *Condition {
	if !condition.closed && !condition.closable {
		fmt.Fprintf(condition.buffer, "%s IN ?", name)
		condition.args.PushBack(value)
		condition.closable = true
	}
	return condition
}

func (condition *Condition) NotIn(name string, value interface{}) *Condition {
	if !condition.closed && !condition.closable {
		fmt.Fprintf(condition.buffer, "%s NOT IN ?", name)
		condition.args.PushBack(value)
		condition.closable = true
	}
	return condition
}

func (condition *Condition) Not() *Condition {
	if !condition.closed && !condition.closable {
		fmt.Fprintf(condition.buffer, "NOT ")
	}
	return condition
}

func (condition *Condition) And() *Condition {
	if condition.sub != nil {
		condition.sub.close()
		condition.sub = nil
	}
	if !condition.closed && condition.closable {
		fmt.Fprintf(condition.buffer, " AND ")
		condition.closable = false
	}
	return condition
}

func (condition *Condition) Or() *Condition {
	if condition.sub != nil {
		condition.sub.close()
		condition.sub = nil
	}
	if !condition.closed && condition.closable {
		fmt.Fprintf(condition.buffer, " OR ")
		condition.closable = false
	}
	return condition
}
