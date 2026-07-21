package extractor

import (
	"crypto/sha256"
	"fmt"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/geometry"
)

// ExtractedAnnotation represents an un-parsed @ocq comment directive block extracted from a host Asset.
type ExtractedAnnotation struct {
	ID      string            // Deterministic unique hash ID
	Geo     geometry.Geometry // Physical 2D geometry (StartLine, EndLine, Indent)
	Content string            // Raw directive text stripped of host comment markers
}

// NewExtractedAnnotation creates a new ExtractedAnnotation and generates a deterministic ID.
func NewExtractedAnnotation(filename string, geo geometry.Geometry, content string) ExtractedAnnotation {
	id := generateID(filename, geo, content)
	return ExtractedAnnotation{
		ID:      id,
		Geo:     geo,
		Content: content,
	}
}

func generateID(filename string, geo geometry.Geometry, content string) string {
	h := sha256.New()
	raw := fmt.Sprintf("%s:%d:%d:%d:%s", filename, geo.StartLine, geo.EndLine, geo.Indent, content)
	h.Write([]byte(raw))
	return fmt.Sprintf("%x", h.Sum(nil))[:12]
}
