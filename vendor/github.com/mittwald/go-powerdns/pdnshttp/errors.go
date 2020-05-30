package pdnshttp

import "fmt"

type ErrNotFound struct {
	URL string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("not found: %s", e.URL)
}

type ErrUnexpectedStatus struct {
	URL        string
	StatusCode int
}

func (e ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status code %d: %s", e.StatusCode, e.URL)
}

func IsNotFound(err error) bool {
	switch err.(type) {
	case ErrNotFound:
		return true
	case *ErrNotFound:
		return true
	}

	return false
}
