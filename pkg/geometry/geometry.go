package geometry

import "fmt"

// Geometry represents 2D physical coordinates within a source Asset file.
type Geometry struct {
	StartLine int // Starting line number (1-indexed)
	EndLine   int // Ending line number (1-indexed)
	Indent    int // Measured indentation offset (number of spaces)
}

// NewGeometry creates a new Geometry coordinate struct.
func NewGeometry(startLine, endLine, indent int) Geometry {
	return Geometry{
		StartLine: startLine,
		EndLine:   endLine,
		Indent:    indent,
	}
}

// String returns a human-readable representation of the physical geometry.
func (g Geometry) String() string {
	if g.StartLine == g.EndLine {
		return fmt.Sprintf("L%d (indent:%d)", g.StartLine, g.Indent)
	}
	return fmt.Sprintf("L%d-L%d (indent:%d)", g.StartLine, g.EndLine, g.Indent)
}
