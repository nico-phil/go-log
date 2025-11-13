package log

// Sement ties a store and an index together
type Segment struct {
	store      *store
	index      *Index
	baseOffset uint64
	nextOffset uint64
	config     Config
}
