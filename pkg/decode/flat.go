package decode

import (
	"log"
	"strings"
)

// Flatten255 takes a list of strings and returns a list of strings where each
// is 255-octets or fewer. Longer strings are split into smaller chunks.
func Flatten255(sl []string) []string {
	var checkIn, checkOut string
	var debug = true

	if debug {
		checkIn = strings.Join(sl, "")
	}

	var result []string
	const max = 255
	for _, s := range sl {
		if len(s) <= max {
			result = append(result, s)
		} else {
			chunks := splitChunks(s, max)
			result = append(result, chunks...)
		}
	}

	if debug {
		checkOut = strings.Join(sl, "")
		if checkIn != checkOut {
			log.Fatalf("assertion failed: in != out (%q) (%q)", checkIn, checkOut)
		}
	}

	return result
}

func splitChunks(buf string, lim int) []string {
	var chunk string
	chunks := make([]string, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}
