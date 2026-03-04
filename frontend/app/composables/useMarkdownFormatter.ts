import type { Tokens } from 'marked'
import { marked } from 'marked'

/**
 * Composable for formatting markdown documents
 */
export function useMarkdownFormatter() {
  /**
   * Format the entire markdown document
   */
  function formatMarkdown(text: string): string {
    const tokens = marked.lexer(text)
    const lines = text.split('\n')

    // Track which line ranges have been formatted (to avoid double processing)
    const formattedRanges: Array<{ start: number, end: number, replacement: string[] }> = []

    // Process tokens to find tables and lists
    let currentLine = 0

    for (const token of tokens) {
      const tokenLines = token.raw.split('\n')
      const tokenLineCount = tokenLines.length - (token.raw.endsWith('\n') ? 1 : 0)
      const tokenStartLine = currentLine

      if (token.type === 'table') {
        const tableLines = lines.slice(tokenStartLine, tokenStartLine + tokenLineCount)
        const formattedTable = formatTable(tableLines, token as Tokens.Table)
        formattedRanges.push({
          start: tokenStartLine,
          end: tokenStartLine + tokenLineCount,
          replacement: formattedTable,
        })
      } else if (token.type === 'list') {
        const listLines = lines.slice(tokenStartLine, tokenStartLine + tokenLineCount)
        const formattedList = formatList(listLines, token as Tokens.List)
        formattedRanges.push({
          start: tokenStartLine,
          end: tokenStartLine + tokenLineCount,
          replacement: formattedList,
        })
      }

      currentLine += token.raw.split('\n').length - 1
    }

    // Apply replacements from end to start to preserve line numbers
    const resultLines = [...lines]
    for (let i = formattedRanges.length - 1; i >= 0; i--) {
      const range = formattedRanges[i]!
      resultLines.splice(range.start, range.end - range.start, ...range.replacement)
    }

    return resultLines.join('\n')
  }

  type TableAlignment = 'left' | 'center' | 'right' | null

  /**
   * Format a markdown table
   */
  function formatTable(_lines: string[], token: Tokens.Table): string[] {
    // Check if the original token had a trailing blank line (separator)
    const hasTrailingBlankLine = token.raw.endsWith('\n\n')

    // Parse table data from token
    const headers = token.header.map((h: Tokens.TableCell) => h.text)
    const alignments: TableAlignment[] = token.align
    const rows = token.rows.map((row: Tokens.TableCell[]) => row.map((cell: Tokens.TableCell) => cell.text))

    // Calculate max width for each column
    const columnCount = headers.length
    const columnWidths: number[] = []

    for (let col = 0; col < columnCount; col++) {
      let maxWidth = headers[col]?.length ?? 0
      for (const row of rows) {
        const cellWidth = row[col]?.length ?? 0
        if (cellWidth > maxWidth) {
          maxWidth = cellWidth
        }
      }
      // Minimum width of 3 for separator (---)
      columnWidths.push(Math.max(maxWidth, 3))
    }

    // Build formatted table
    const result: string[] = []

    // Header row
    const headerCells = headers.map((h: string, i: number) => padCell(h, columnWidths[i]!, alignments[i] ?? null))
    result.push(`| ${headerCells.join(' | ')} |`)

    // Separator row (no spaces around dashes, dashes fill full column width + 2 for padding spaces)
    const separatorCells = alignments.map((align: TableAlignment, i: number) => {
      const width = columnWidths[i]! + 2 // Add 2 for the spaces around content in data rows
      if (align === 'center') {
        return `:${'-'.repeat(width - 2)}:`
      } else if (align === 'right') {
        return `${'-'.repeat(width - 1)}:`
      } else if (align === 'left') {
        return `:${'-'.repeat(width - 1)}`
      } else {
        return '-'.repeat(width)
      }
    })
    result.push(`|${separatorCells.join('|')}|`)

    // Data rows
    for (const row of rows) {
      const cells = row.map((cell: string, i: number) => padCell(cell, columnWidths[i]!, alignments[i] ?? null))
      result.push(`| ${cells.join(' | ')} |`)
    }

    // Preserve trailing blank line if it existed
    if (hasTrailingBlankLine) {
      result.push('')
    }

    return result
  }

  /**
   * Pad a cell to the specified width based on alignment
   */
  function padCell(text: string, width: number, align: TableAlignment): string {
    const padding = width - text.length
    if (padding <= 0) {
      return text
    }

    if (align === 'center') {
      const leftPad = Math.floor(padding / 2)
      const rightPad = padding - leftPad
      return ' '.repeat(leftPad) + text + ' '.repeat(rightPad)
    } else if (align === 'right') {
      return ' '.repeat(padding) + text
    } else {
      // Default to left alignment
      return text + ' '.repeat(padding)
    }
  }

  /**
   * Format a markdown list
   */
  function formatList(_lines: string[], token: Tokens.List): string[] {
    const result: string[] = []
    const ordered = token.ordered
    const startNumber = token.start || 1

    // Initial call has no parent indent and default marker width of 2 (for "- ")
    formatListItems(token.items, result, '', ordered, startNumber, 2)

    return result
  }

  /**
   * Recursively format list items
   */
  function formatListItems(
    items: Tokens.ListItem[],
    result: string[],
    parentIndent: string,
    ordered: boolean,
    startNumber: number,
    _parentMarkerWidth: number,
  ): void {
    // For nested items, indent by the parent's marker width to align under parent's text
    const indent = parentIndent

    items.forEach((item, index) => {
      const number = startNumber + index
      const bullet = ordered ? `${number}.` : '-'
      // Calculate this item's marker width for potential nested children
      const markerWidth = bullet.length + 1 // +1 for the space after the bullet
      const checkboxPrefix = item.checked === true
        ? '[x] '
        : item.checked === false
          ? '[ ] '
          : ''

      // Get the text content (first token that's text or has text)
      let textContent = ''
      let hasNestedList = false
      let nestedList: Tokens.List | null = null

      for (const subToken of item.tokens) {
        if (subToken.type === 'text') {
          textContent = (subToken as Tokens.Text).text.split('\n')[0] ?? ''
        } else if (subToken.type === 'list') {
          hasNestedList = true
          nestedList = subToken as Tokens.List
        }
      }

      result.push(`${indent}${bullet} ${checkboxPrefix}${textContent}`)

      // Process nested list
      if (hasNestedList && nestedList) {
        // Nested items should be indented by this item's marker width
        const nestedIndent = indent + ' '.repeat(markerWidth)
        formatListItems(
          nestedList.items,
          result,
          nestedIndent,
          nestedList.ordered,
          nestedList.start || 1,
          markerWidth,
        )
      }
    })
  }

  return {
    formatMarkdown,
  }
}
