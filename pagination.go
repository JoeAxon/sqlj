package sqlj

type PageOptions struct {
	PageNumber uint
	PageSize   uint
	Order      []OrderBy
}

type OrderBy struct {
	Expression string
	Direction  string
}
