package source

import "bytes"

// BytesSource loads resources from an in-memory multi-document YAML stream.
// It is the in-memory counterpart to FileSource, for consumers that already
// hold manifests in memory (for example a controller reading them from a
// ConfigMap) rather than on disk.
type BytesSource struct {
	Data []byte
}

// NewBytesSource creates a BytesSource over the given YAML bytes.
func NewBytesSource(data []byte) *BytesSource {
	return &BytesSource{Data: data}
}

// Load parses the in-memory YAML stream into resources, applying the same
// multi-document parsing and skip rules as FileSource.
func (b *BytesSource) Load() ([]Resource, error) {
	return parseYAML(bytes.NewReader(b.Data))
}
