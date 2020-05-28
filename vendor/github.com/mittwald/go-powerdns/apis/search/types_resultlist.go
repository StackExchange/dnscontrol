package search

// ResultList represents a list of search results. The type itself offers some
// advanced filtering functions for convenience.
type ResultList []Result

// FilterByObjectType returns all elements of a result list that are of a given
// object type.
func (l ResultList) FilterByObjectType(t ObjectType) ResultList {
	return l.FilterBy(func(r *Result) bool {
		return r.ObjectType == t
	})
}

// FilterByRecordType returns all elements of a result list that are a resource
// record and have a certain record type.
func (l ResultList) FilterByRecordType(t string) ResultList {
	return l.FilterBy(func(r *Result) bool {
		return r.ObjectType == ObjectTypeRecord && r.Type == t
	})
}

// FilterBy returns all elements of a result list that match a generic matcher
// function. The "matcher" function will be invoked for each element in the
// result list; if it returns true, the respective item will be included in the
// result list.
func (l ResultList) FilterBy(matcher func(*Result) bool) ResultList {
	out := make(ResultList, 0, len(l))

	for i := range l {
		if matcher(&l[i]) {
			out = append(out, l[i])
		}
	}

	return out
}
