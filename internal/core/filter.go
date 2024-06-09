package core

type Order string

const (
	ASC  Order = "ASC"
	DESC Order = "DESC"
)

func (o Order) String() string {
	return string(o)
}

func ToOrder(s string) Order {
	switch s {
	case "ASC":
		return ASC
	case "DESC":
		return DESC
	default:
		return ASC
	}
}

type Pagination struct {
	Limit  int
	Offset int
}

type Sorting struct {
	OrderBy string
	Order   Order
}
