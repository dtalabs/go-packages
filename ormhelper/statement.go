package ormhelper

type Statement struct {
	Conditions []Condition
	Sort       []Sort
	Limit      int
	Offset     int
}

type Operator string

const (
	Equal              Operator = ""
	NotEqual           Operator = "ne"
	LessThan           Operator = "lt"
	LessThanOrEqual    Operator = "lte"
	GreaterThan        Operator = "gt"
	GreaterThanOrEqual Operator = "gte"
	In                 Operator = "in"
	NotIn              Operator = "not_in"
)

type Condition struct {
	Field    string
	Operator Operator
	Value    any
}

type Sort struct {
	Field      string
	Descending bool
}

func (s Statement) Where(field string, operator Operator, value any) Statement {
	return s.WhereCondition(Condition{Field: field, Operator: operator, Value: value})
}

func (s Statement) WhereEquals(field string, value any) Statement {
	return s.Where(field, Equal, value)
}

func (s Statement) WhereCondition(condition Condition) Statement {
	s.Conditions = append(append([]Condition(nil), s.Conditions...), condition)
	return s
}

func (s Statement) OrderBy(field string, descending bool) Statement {
	s.Sort = append(append([]Sort(nil), s.Sort...), Sort{Field: field, Descending: descending})
	return s
}
