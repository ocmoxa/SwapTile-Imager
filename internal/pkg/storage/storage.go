package storage

import "context"

// CatergoryStorage stores known category names.
type CatergoryStorage interface {
	// List return all known categories.
	List(ctx context.Context) (names []string, err error)
	// Delete deletes a category by the name. If the category doesn't
	// exist, it will return a nil error.
	Delete(ctx context.Context, name string) (err error)
	// Upsert updates or inserts a category.
	Upsert(ctx context.Context, name string) (err error)
}

// ImageIDStorage stores known image id.
type ImageIDStorage interface {
	// List returns a list of image ids for the given category.
	List(ctx context.Context, category string, pagination Pagination) (ids []string, err error)
	// Insert saves new image id to the category. The uniquness of the
	// id is not checked.
	Insert(ctx context.Context, cateogry string, id string) (err error)
	// Delete deletes image by id from the category.
	Delete(ctx context.Context, category string, id string) (err error)
}

// Pagination holds query limits.
type Pagination struct {
	// Limit of rows in the result.
	Limit int
	// Offset of rows in the result.
	Offset int
}
