package extractor

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/ouvrage-systems/ouvrage-calque-go/pkg/geometry"
)

// CommentStyle defines host comment markers for an Asset.
type CommentStyle struct {
	Prefix string // Prefix marker e.g. "#", "//", "::", "<!--"
	Suffix string // Suffix marker e.g. "-->" or empty
}

// KnownCommentStyles maps common file extensions to host comment markers.
var KnownCommentStyles = map[string]CommentStyle{
	".yaml": {Prefix: "#"},
	".yml":  {Prefix: "#"},
	".bat":  {Prefix: "::"},
	".cmd":  {Prefix: "::"},
	".sh":   {Prefix: "#"},
	".bash": {Prefix: "#"},
	".c":    {Prefix: "//"},
	".h":    {Prefix: "//"},
	".cpp":  {Prefix: "//"},
	".go":   {Prefix: "//"},
	".xml":  {Prefix: "<!--", Suffix: "-->"},
	".html": {Prefix: "<!--", Suffix: "-->"},
}

// Extract performs Phase 1 extraction: scanning an Asset and returning all raw @ocq annotations.
func Extract(r io.Reader, filename string) ([]ExtractedAnnotation, error) {
	isNakedDSL := strings.HasSuffix(filename, ".ocq")
	commentStyle := getCommentStyle(filename)

	scanner := bufio.NewScanner(r)
	var annotations []ExtractedAnnotation

	lineNum := 0
	accumulating := false
	accumulatedContent := ""
	accumulatedStartLine := 0
	accumulatedEndLine := 0
	accumulatedIndent := 0
	openParens := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimRight(scanner.Text(), " \t\r")

		if isNakedDSL {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
				continue // Skip comments and empty lines in naked DSL .ocq files
			}

			if accumulating {
				accumulatedContent += "\n" + line
				openParens += strings.Count(trimmed, "(") - strings.Count(trimmed, ")")
				openParens += strings.Count(trimmed, "[") - strings.Count(trimmed, "]")
				accumulatedEndLine = lineNum
				if openParens <= 0 {
					geo := geometry.NewGeometry(accumulatedStartLine, accumulatedEndLine, accumulatedIndent)
					ann := NewExtractedAnnotation(filename, geo, accumulatedContent)
					annotations = append(annotations, ann)
					accumulating = false
					accumulatedContent = ""
				}
				continue
			}

			indent := calculateIndent(line)
			openCount := strings.Count(trimmed, "(") - strings.Count(trimmed, ")") + strings.Count(trimmed, "[") - strings.Count(trimmed, "]")
			if openCount > 0 {
				accumulating = true
				accumulatedStartLine = lineNum
				accumulatedEndLine = lineNum
				accumulatedIndent = indent
				accumulatedContent = line
				openParens = openCount
				continue
			}

			geo := geometry.NewGeometry(lineNum, lineNum, indent)
			ann := NewExtractedAnnotation(filename, geo, trimmed)
			annotations = append(annotations, ann)
		} else {
			// Search for @ocq: annotation marker in host line
			content, indent, ok := extractAnnotationFromHostLine(line, commentStyle)
			if ok {
				geo := geometry.NewGeometry(lineNum, lineNum, indent)
				ann := NewExtractedAnnotation(filename, geo, content)
				annotations = append(annotations, ann)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("extractor error reading %s: %w", filename, err)
	}

	return annotations, nil
}

func getCommentStyle(filename string) CommentStyle {
	for ext, style := range KnownCommentStyles {
		if strings.HasSuffix(filename, ext) {
			return style
		}
	}
	return CommentStyle{Prefix: "#"}
}

func calculateIndent(line string) int {
	indent := 0
	for _, r := range line {
		if r == ' ' {
			indent++
		} else if r == '\t' {
			indent += 4
		} else {
			break
		}
	}
	return indent
}

func extractAnnotationFromHostLine(line string, style CommentStyle) (string, int, bool) {
	trimmed := strings.TrimSpace(line)

	idx := strings.Index(trimmed, "@ocq:")
	if idx == -1 {
		return "", 0, false
	}

	indent := calculateIndent(line)
	content := trimmed[idx+len("@ocq:"):]

	if style.Suffix != "" {
		content = strings.TrimSuffix(content, style.Suffix)
	}
	content = strings.TrimSpace(content)

	if content == "" {
		return "", 0, false
	}

	return content, indent, true
}
