import { describe, expect, it } from 'vitest'
import { useMarkdownFormatter } from './useMarkdownFormatter'

describe('useMarkdownFormatter', () => {
  const { formatMarkdown } = useMarkdownFormatter()

  describe('table formatting', () => {
    it('formats a basic table with equal column widths', () => {
      const input = `|Name|Age|
|---|---|
|John|25|
|Jane|30|`

      const expected = `| Name | Age |
|------|-----|
| John | 25  |
| Jane | 30  |`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('formats a table with varying column widths', () => {
      const input = `|Name|Age|City|
|---|---|---|
|John|25|New York|
|Elizabeth|30|Los Angeles|`

      const expected = `| Name      | Age | City        |
|-----------|-----|-------------|
| John      | 25  | New York    |
| Elizabeth | 30  | Los Angeles |`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('preserves left alignment marker', () => {
      const input = `|Name|Age|
|:---|---|
|John|25|`

      const expected = `| Name | Age |
|:-----|-----|
| John | 25  |`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('preserves center alignment marker', () => {
      const input = `|Name|Age|
|:---:|---|
|John|25|`

      const expected = `| Name | Age |
|:----:|-----|
| John | 25  |`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('preserves right alignment marker', () => {
      const input = `|Name|Age|
|---:|---|
|John|25|`

      // Name column is right-aligned (---:), Age column is left-aligned (default)
      const expected = `| Name | Age |
|-----:|-----|
| John | 25  |`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('handles empty cells', () => {
      const input = `|Name|Age|
|---|---|
|John||
||30|`

      const expected = `| Name | Age |
|------|-----|
| John |     |
|      | 30  |`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('handles single column tables', () => {
      const input = `|Name|
|---|
|John|
|Jane|`

      const expected = `| Name |
|------|
| John |
| Jane |`

      expect(formatMarkdown(input)).toBe(expected)
    })
  })

  describe('list formatting', () => {
    it('normalizes unordered list bullets to dashes', () => {
      const input = `* Item 1
* Item 2
* Item 3`

      const expected = `- Item 1
- Item 2
- Item 3`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('formats ordered lists', () => {
      const input = `1. First item
2. Second item
3. Third item`

      const expected = `1. First item
2. Second item
3. Third item`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('formats nested unordered lists with proper indentation', () => {
      const input = `- Item 1
  - Nested item 1
  - Nested item 2
- Item 2`

      const expected = `- Item 1
  - Nested item 1
  - Nested item 2
- Item 2`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('formats deeply nested lists', () => {
      const input = `- Level 1
  - Level 2
    - Level 3
      - Level 4`

      const expected = `- Level 1
  - Level 2
    - Level 3
      - Level 4`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('formats checkbox lists', () => {
      const input = `- [ ] Unchecked item
- [x] Checked item
- [ ] Another unchecked`

      const expected = `- [ ] Unchecked item
- [x] Checked item
- [ ] Another unchecked`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('normalizes checkbox lists with asterisk bullets to dashes', () => {
      const input = `* [ ] Unchecked item
* [x] Checked item
* [ ] Another unchecked`

      const expected = `- [ ] Unchecked item
- [x] Checked item
- [ ] Another unchecked`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('formats nested checkbox lists', () => {
      const input = `- [ ] Parent task
  - [ ] Sub-task 1
  - [x] Sub-task 2 (done)
- [x] Completed parent`

      const expected = `- [ ] Parent task
  - [ ] Sub-task 1
  - [x] Sub-task 2 (done)
- [x] Completed parent`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('handles mixed nested list types', () => {
      const input = `- Unordered item
  1. Ordered nested 1
  2. Ordered nested 2
- Another unordered`

      const expected = `- Unordered item
  1. Ordered nested 1
  2. Ordered nested 2
- Another unordered`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('formats ordered list with nested unordered list using correct indentation', () => {
      // Ordered list items need 3-space indentation for nested content
      // because "1. " is 3 characters
      const input = `1. list
2. list
    - nested point`

      const expected = `1. list
2. list
   - nested point`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('formats ordered list starting at 10+ with correct indentation', () => {
      // "10. " is 4 characters, so nested items need 4-space indentation
      const input = `10. item ten
11. item eleven
    - nested point`

      const expected = `10. item ten
11. item eleven
    - nested point`

      expect(formatMarkdown(input)).toBe(expected)
    })
  })

  describe('document formatting', () => {
    it('formats a document with both tables and lists', () => {
      const input = `# Title

|Name|Age|
|---|---|
|John|25|

* Item 1
* Item 2`

      const expected = `# Title

| Name | Age |
|------|-----|
| John | 25  |

- Item 1
- Item 2`

      expect(formatMarkdown(input)).toBe(expected)
    })

    it('preserves non-table, non-list content unchanged', () => {
      const input = `# Heading

This is a paragraph with **bold** and *italic* text.

Another paragraph here.`

      expect(formatMarkdown(input)).toBe(input)
    })

    it('handles empty input', () => {
      expect(formatMarkdown('')).toBe('')
    })

    it('handles input with only whitespace', () => {
      const input = `

`
      expect(formatMarkdown(input)).toBe(input)
    })

    it('preserves code blocks unchanged', () => {
      const input = `\`\`\`javascript
const x = 1;
const y = 2;
\`\`\``

      expect(formatMarkdown(input)).toBe(input)
    })

    it('preserves inline code unchanged', () => {
      const input = 'Use `const` for constants.'

      expect(formatMarkdown(input)).toBe(input)
    })
  })
})
