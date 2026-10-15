package providers

// DocumentationNote is a way for providers to give more detail about what features they support.
type DocumentationNote struct {
	HasFeature    bool
	Unimplemented bool
	Comment       string
	Link          string
}

// Can is a small helper for concisely creating Documentation Notes
// comments are variadic for easy omission. First is comment, second is link, the rest are ignored.
func Can(comments ...string) *DocumentationNote {
	n := &DocumentationNote{
		HasFeature: true,
	}
	n.addStrings(comments)
	return n
}

// Cannot is a small helper for concisely creating Documentation Notes
// comments are variadic for easy omission. First is comment, second is link, the rest are ignored.
func Cannot(comments ...string) *DocumentationNote {
	n := &DocumentationNote{
		HasFeature: false,
	}
	n.addStrings(comments)
	return n
}

// Unimplemented is a small helper for concisely creating Documentation Notes
// comments are variadic for easy omission. First is comment, second is link, the rest are ignored.
func Unimplemented(comments ...string) *DocumentationNote {
	n := &DocumentationNote{
		HasFeature:    false,
		Unimplemented: true,
	}
	n.addStrings(comments)
	return n
}

func (n *DocumentationNote) addStrings(comments []string) {
	if len(comments) > 0 {
		n.Comment = comments[0]
	}
	if len(comments) > 1 {
		n.Link = comments[1]
	}
}
