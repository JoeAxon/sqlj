package sqlj

type PageOptions struct {
	pageNumber uint
	pageSize   uint
	order      []OrderBy
}

type OrderBy struct {
	expression string
	direction  string
}
