package rsql

// FilterContext contains information needed to build and validate filters
type FilterContext struct {
	ViewerMemberID      *int64
	CanViewSecretContent bool // True if viewer can see all entry content
}
