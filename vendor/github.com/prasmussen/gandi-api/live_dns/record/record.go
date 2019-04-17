package record

import (
	"fmt"
	"strings"

	"github.com/prasmussen/gandi-api/client"
)

// Record holds the zone client structure
type Record struct {
	*client.Client
	Prefix string
}

// Creator is an interface to create new record entries
type Creator interface {
	// Create creates a new record entry
	// possible calls are:
	// Create(recordInfo)
	// Create(recordInfo, "entry")
	// Create(recordInfo, "entry", "type")
	// where "entry" matches entry.example.com
	// and "type" is the record type (A, CNAME, ...)
	Create(recordInfo Info, args ...string) (status *Status, err error)
}

// Updater is an interface to update existing record entries
type Updater interface {
	// Update creates a new record entry
	// possible calls are:
	// Update(recordInfo)
	// Update(recordInfo, "entry")
	// Update(recordInfo, "entry", "type")
	// where "entry" matches entry.example.com
	// and "type" is the record type (A, CNAME, ...)
	Update(recordInfo Info, args ...string) (status *Status, err error)
}

// Lister is an interface to list existing record entries
type Lister interface {
	// List creates a new record entry
	// possible calls are:
	// List(recordInfo)
	// List(recordInfo, "entry")
	// List(recordInfo, "entry", "type")
	// where "entry" matches entry.example.com
	// and "type" is the record type (A, CNAME, ...)
	List(args ...string) (list []*Info, err error)
}

// Deleter is an interface to delete existing record entries
type Deleter interface {
	// Delete creates a new record entry
	// possible calls are:
	// Delete(recordInfo)
	// Delete(recordInfo, "entry")
	// Delete(recordInfo, "entry", "type")
	// where "entry" matches entry.example.com
	// and "type" is the record type (A, CNAME, ...)
	Delete(args ...string) (err error)
}

// Manager is an interface to manage records (for a zone or domain)
type Manager interface {
	Creator
	Updater
	Lister
	Deleter
}

// New instanciates a new instance of a Zone client
func New(c *client.Client, prefix string) *Record {
	return &Record{c, prefix}
}

func (r *Record) uri(pattern string, paths ...string) string {
	args := make([]interface{}, len(paths))
	for i, v := range paths {
		args[i] = v
	}
	return fmt.Sprintf("%s/%s",
		strings.TrimRight(r.Prefix, "/"),
		strings.TrimLeft(fmt.Sprintf(pattern, args...), "/"))
}

func (r *Record) formatCallError(function string, args ...string) error {
	format := "unexpected arguments for function %s." +
		" supported calls are: %s(), %s(<Name>), %s(<Name>, <Type>)" +
		" %s called with"
	a := []interface{}{
		function,
		function, function, function,
		function,
	}
	for _, v := range args {
		format = format + " %s"
		a = append(a, v)
	}
	return fmt.Errorf(format, a...)
}

// Create creates a new record entry
// possible calls are:
// Create(recordInfo)
// Create(recordInfo, "entry")
// Create(recordInfo, "entry", "type")
// where "entry" matches entry.example.com
// and "type" is the record type (A, CNAME, ...)
func (r *Record) Create(recordInfo Info, args ...string) (status *Status, err error) {
	switch len(args) {
	case 0:
		_, err = r.Post(r.uri("/records"), recordInfo, &status)
	case 1:
		_, err = r.Post(r.uri("/records/%s", args...), recordInfo, &status)
	case 2:
		_, err = r.Post(r.uri("/records/%s/%s", args...), recordInfo, &status)
	default:
		err = r.formatCallError("Create", args...)
	}
	return
}

// Update creates a new record entry
// possible calls are:
// Update(recordInfo)
// Update(recordInfo, "entry")
// Update(recordInfo, "entry", "type")
// where "entry" matches entry.example.com
// and "type" is the record type (A, CNAME, ...)
func (r *Record) Update(recordInfo Info, args ...string) (status *Status, err error) {
	switch len(args) {
	case 0:
		_, err = r.Put(r.uri("/records"), recordInfo, &status)
	case 1:
		_, err = r.Put(r.uri("/records/%s", args...), recordInfo, &status)
	case 2:
		_, err = r.Put(r.uri("/records/%s/%s", args...), recordInfo, &status)
	default:
		err = r.formatCallError("Update", args...)
	}
	return
}

// List creates a new record entry
// possible calls are:
// List(recordInfo)
// List(recordInfo, "entry")
// List(recordInfo, "entry", "type")
// where "entry" matches entry.example.com
// and "type" is the record type (A, CNAME, ...)
func (r *Record) List(args ...string) (list []*Info, err error) {
	switch len(args) {
	case 0:
		_, err = r.Get(r.uri("/records"), &list)
	case 1:
		_, err = r.Get(r.uri("/records/%s", args...), &list)
	case 2:
		_, err = r.Get(r.uri("/records/%s/%s", args...), &list)
	default:
		err = r.formatCallError("List", args...)
	}
	return
}

// Delete creates a new record entry
// possible calls are:
// Delete(recordInfo)
// Delete(recordInfo, "entry")
// Delete(recordInfo, "entry", "type")
// where "entry" matches entry.example.com
// and "type" is the record type (A, CNAME, ...)
func (r *Record) Delete(args ...string) (err error) {
	switch len(args) {
	case 0:
		_, err = r.Client.Delete(r.uri("/records"), nil)
	case 1:
		_, err = r.Client.Delete(r.uri("/records/%s", args...), nil)
	case 2:
		_, err = r.Client.Delete(r.uri("/records/%s/%s", args...), nil)
	default:
		err = r.formatCallError("Delete", args...)
	}
	return
}
