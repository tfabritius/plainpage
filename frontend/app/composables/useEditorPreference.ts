export type EditorType = 'source' | 'visual'

export function useEditorPreference() {
  const editorType = useState<EditorType>('editor-preference', () => {
    // Default to visual editor
    if (import.meta.client) {
      const stored = localStorage.getItem('editor-preference')
      if (stored === 'source' || stored === 'visual') {
        return stored
      }
    }
    return 'visual'
  })

  // Persist to localStorage when changed
  watch(editorType, (newValue) => {
    if (import.meta.client) {
      localStorage.setItem('editor-preference', newValue)
    }
  })

  const isSourceEditor = computed(() => editorType.value === 'source')
  const isVisualEditor = computed(() => editorType.value === 'visual')

  function toggleEditor() {
    editorType.value = editorType.value === 'source' ? 'visual' : 'source'
  }

  function setEditor(type: EditorType) {
    editorType.value = type
  }

  return {
    editorType,
    isSourceEditor,
    isVisualEditor,
    toggleEditor,
    setEditor,
  }
}
