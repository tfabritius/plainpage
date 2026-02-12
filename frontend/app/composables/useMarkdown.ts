import type { Tokens } from 'marked'
import dompurify from 'dompurify'
import { marked } from 'marked'
import slugify from 'slugify'

export interface TocItem {
  id: string
  text: string
  level: number
}

export interface MarkdownResult {
  html: string
  toc: TocItem[]
}

interface HeadingRendererOptions {
  collectToc?: boolean
  addAnchorLinks?: boolean
}

/**
 * Creates a markdown renderer with heading anchor support
 */
function createMarkdownRenderer(
  toc: TocItem[],
  slugCounter: Map<string, number>,
  options: HeadingRendererOptions = {},
) {
  const { collectToc = true, addAnchorLinks = false } = options
  const renderer = new marked.Renderer()

  // Custom heading renderer that adds IDs and optionally collects TOC
  renderer.heading = ({ tokens, depth }: Tokens.Heading) => {
    const text = tokens.map(t => ('text' in t ? t.text : '')).join('')
    const tag = `h${depth}`

    // Only process H1-H4 for TOC
    if (depth <= 4) {
      // Generate slug from heading text
      let slug = slugify(text, { lower: true, strict: true })

      // Handle duplicate slugs
      const count = slugCounter.get(slug) || 0
      if (count > 0) {
        slug = `${slug}-${count}`
      }
      slugCounter.set(slug.replace(/-\d+$/, ''), count + 1)

      // Collect TOC item
      if (collectToc) {
        toc.push({ id: slug, text, level: depth })
      }

      // Render heading with or without anchor link
      if (addAnchorLinks) {
        // Anchor link with icon rendered via CSS
        return `<${tag} id="${slug}">${text}<a class="heading-anchor" href="#${slug}" aria-label="Link to this heading"></a></${tag}>`
      }
      return `<${tag} id="${slug}">${text}</${tag}>`
    }

    // H5-H6 without ID
    return `<${tag}>${text}</${tag}>`
  }

  return renderer
}

/**
 * Configures DOMPurify to allow necessary attributes
 */
function configureDomPurify() {
  // Allow id and href attributes for anchors
  dompurify.addHook('uponSanitizeAttribute', (node, data) => {
    if (data.attrName === 'id' || data.attrName === 'href') {
      // Allow these attributes
    }
  })
}

// Configure DOMPurify once on module load
configureDomPurify()

/**
 * Parses markdown and returns HTML with TOC data
 */
export async function parseMarkdown(
  markdown: string,
  options: HeadingRendererOptions = {},
): Promise<MarkdownResult> {
  const toc: TocItem[] = []
  const slugCounter = new Map<string, number>()
  const renderer = createMarkdownRenderer(toc, slugCounter, options)

  const parsedMarkdown = await marked.parse(markdown, {
    gfm: true,
    renderer,
  })

  const html = dompurify.sanitize(parsedMarkdown, {
    ADD_ATTR: ['id'],
  })

  return { html, toc }
}

/**
 * Composable for reactive markdown parsing
 */
export function useMarkdown(
  markdown: Ref<string> | ComputedRef<string>,
  options: HeadingRendererOptions = {},
) {
  const html = ref('')
  const toc = ref<TocItem[]>([])

  watchEffect(async () => {
    const result = await parseMarkdown(markdown.value, options)
    html.value = result.html
    toc.value = result.toc
  })

  return { html, toc }
}
