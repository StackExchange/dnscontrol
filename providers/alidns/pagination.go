package alidns

// paginateAll is a small generic paginator helper. The caller provides a
// fetch function that requests a single page (pageNumber,pageSize) and
// returns the items for that page, the total number of items available,
// and an error if any. paginateAll will iterate pages until it has
// collected all items or an error occurs.
func paginateAll[T any](fetch func(pageNumber, pageSize int) ([]T, int, error), maxPageSize int) ([]T, error) {
	page := 1
	pageSize := maxPageSize
	var out []T

	for {
		items, total, err := fetch(page, pageSize)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)

		// If we've collected all items, or the page returned nothing, stop.
		if len(out) >= total || len(items) == 0 {
			break
		}
		page++
	}
	return out, nil
}
