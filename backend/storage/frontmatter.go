package storage

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func parseFrontMatter(file string) (PageMeta, string, error) {
	// Split the file into frontmatter and content
	parts := strings.SplitN(file, "---", 3)
	if len(parts) < 3 {
		return PageMeta{}, file, nil
	}

	// Parse the frontmatter as YAML
	meta := PageMeta{}
	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return PageMeta{}, "", fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Return the frontmatter and markdown content
	return meta, strings.TrimSpace(parts[2]), nil
}

func serializeFrontMatter(meta PageMeta, content string) (string, error) {
	frontMatterBytes, err := yaml.Marshal(&meta)
	if err != nil {
		return "", fmt.Errorf("failed to marshal: %w", err)
	}

	frontMatter := strings.TrimSpace(string(frontMatterBytes))

	var buf strings.Builder
	buf.WriteString("---\n")
	buf.WriteString(frontMatter)
	buf.WriteString("\n---\n")
	buf.WriteString(content)

	return buf.String(), nil
}
