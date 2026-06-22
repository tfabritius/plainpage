import type { EditorCustomHandlers, EditorSuggestionMenuItem, EditorToolbarItem } from '@nuxt/ui'
import type { Editor } from '@tiptap/vue-3'

type ExtendedEditorToolbarItem<T extends EditorCustomHandlers>
  = EditorToolbarItem<T>
    | (Omit<Extract<EditorToolbarItem<T>, { kind: 'mark' }>, 'mark'> & {
      mark: 'superscript' | 'subscript'
    })

interface UseEditorToolbarOptions {
  aiLoading?: Ref<boolean | undefined>
}

export function useEditorToolbar<T extends EditorCustomHandlers>(_customHandlers?: T, _options: UseEditorToolbarOptions = {}) {
  const { t } = useI18n()

  const slashmenuItems = computed(() => [
    [
      {
        type: 'label',
        label: t('editor.text'),
      },
      {
        kind: 'paragraph',
        label: t('editor.paragraph'),
        icon: 'tabler:pilcrow',
      },
      {
        kind: 'heading',
        level: 1,
        label: t('editor.heading-1'),
        icon: 'tabler:h-1',
      },
      {
        kind: 'heading',
        level: 2,
        label: t('editor.heading-2'),
        icon: 'tabler:h-2',
      },
      {
        kind: 'heading',
        level: 3,
        label: t('editor.heading-3'),
        icon: 'tabler:h-3',
      },
      {
        kind: 'heading',
        level: 4,
        icon: 'tabler:h-4',
        label: t('editor.heading-4'),
      },
      {
        kind: 'heading',
        level: 5,
        icon: 'tabler:h-5',
        label: t('editor.heading-5'),
      },
      {
        kind: 'heading',
        level: 6,
        icon: 'tabler:h-6',
        label: t('editor.heading-6'),
      },
    ],
    [
      {
        type: 'label',
        label: t('editor.lists'),
      },
      {
        kind: 'bulletList',
        label: t('editor.list-unordered'),
        icon: 'tabler:list',
      },
      {
        kind: 'orderedList',
        label: t('editor.list-ordered'),
        icon: 'tabler:list-numbers',
      },
      {
        kind: 'taskList',
        label: t('editor.checkbox'),
        icon: 'tabler:checkbox',
      },
    ],
    [
      {
        type: 'label',
        label: t('editor.insert'),
      },
      {
        kind: 'blockquote',
        label: t('editor.quote'),
        icon: 'tabler:quote',
      },
      {
        kind: 'codeBlock',
        label: t('editor.code-block'),
        icon: 'tabler:braces',
      },
      {
        kind: 'horizontalRule',
        label: t('editor.divider'),
        icon: 'tabler:separator-horizontal',
      },
      {
        kind: 'table',
        label: t('editor.table'),
        icon: 'tabler:table',
      },
    ],
  ] satisfies EditorSuggestionMenuItem<T>[][])

  const bubbleToolbarItems = computed(() => [
    [{
      label: t('editor.turn-into'),
      trailingIcon: 'tabler:chevron-down',
      activeColor: 'neutral',
      activeVariant: 'ghost',
      tooltip: { text: t('editor.turn-into') },
      content: {
        align: 'start',
      },
      ui: {
        label: 'text-xs',
      },
      items: [
        {
          type: 'label',
          label: t('editor.turn-into'),
        },
        {
          kind: 'paragraph',
          label: t('editor.paragraph'),
          icon: 'tabler:pilcrow',
        },
        {
          kind: 'heading',
          level: 1,
          icon: 'tabler:h-1',
          label: t('editor.heading-1'),
        },
        {
          kind: 'heading',
          level: 2,
          icon: 'tabler:h-2',
          label: t('editor.heading-2'),
        },

        {
          kind: 'heading',
          level: 3,
          icon: 'tabler:h-3',
          label: t('editor.heading-3'),
        },
        {
          kind: 'heading',
          level: 4,
          icon: 'tabler:h-4',
          label: t('editor.heading-4'),
        },
        {
          kind: 'heading',
          level: 5,
          icon: 'tabler:h-5',
          label: t('editor.heading-5'),
        },
        {
          kind: 'heading',
          level: 6,
          icon: 'tabler:h-6',
          label: t('editor.heading-6'),
        },
        {
          kind: 'bulletList',
          label: t('editor.list-unordered'),
          icon: 'tabler:list',
        },
        {
          kind: 'orderedList',
          label: t('editor.list-ordered'),
          icon: 'tabler:list-numbers',
        },
        {
          kind: 'taskList',
          label: t('editor.checkbox'),
          icon: 'tabler:checkbox',
        },
        {
          kind: 'blockquote',
          label: t('editor.quote'),
          icon: 'tabler:quote',
        },
        {
          kind: 'codeBlock',
          label: t('editor.code-block'),
          icon: 'tabler:braces',
        },
      ],
    }],
    [
      {
        kind: 'mark',
        mark: 'bold',
        icon: 'tabler:bold',
        tooltip: { text: t('editor.bold') },
      },
      {
        kind: 'mark',
        mark: 'italic',
        icon: 'tabler:italic',
        tooltip: { text: t('editor.italic') },
      },
      {
        kind: 'mark',
        mark: 'underline',
        icon: 'tabler:underline',
        tooltip: { text: t('editor.underline') },
      },
      {
        kind: 'mark',
        mark: 'strike',
        icon: 'tabler:strikethrough',
        tooltip: { text: t('editor.strikethrough') },
      },
      {
        kind: 'mark',
        mark: 'code',
        icon: 'tabler:code',
        tooltip: { text: t('editor.code-inline') },
      },
    ],
    [
      {
        kind: 'mark',
        mark: 'superscript',
        icon: 'tabler:superscript',
        tooltip: { text: t('editor.superscript') },
      },
      {
        kind: 'mark',
        mark: 'subscript',
        icon: 'tabler:subscript',
        tooltip: { text: t('editor.subscript') },
      },
    ],
    [
      {
        slot: 'link' as const,
        icon: 'tabler:link',
        tooltip: { text: t('editor.link') },
      },
    ],
  ] satisfies ExtendedEditorToolbarItem<T>[][])

  const getTableToolbarItems = (editor: Editor): EditorToolbarItem<T>[][] => {
    return [[{
      icon: 'tabler:row-insert-top',
      tooltip: { text: t('editor.table-add-row-above') },
      onClick: () => {
        editor.chain().focus().addRowBefore().run()
      },
    }, {
      icon: 'tabler:row-insert-bottom',
      tooltip: { text: t('editor.table-add-row-below') },
      onClick: () => {
        editor.chain().focus().addRowAfter().run()
      },
    }, {
      icon: 'tabler:column-insert-left',
      tooltip: { text: t('editor.table-add-column-before') },
      onClick: () => {
        editor.chain().focus().addColumnBefore().run()
      },
    }, {
      icon: 'tabler:column-insert-right',
      tooltip: { text: t('editor.table-add-column-after') },
      onClick: () => {
        editor.chain().focus().addColumnAfter().run()
      },
    }], [{
      icon: 'tabler:row-remove',
      tooltip: { text: t('editor.table-delete-row') },
      onClick: () => {
        editor.chain().focus().deleteRow().run()
      },
    }, {
      icon: 'tabler:column-remove',
      tooltip: { text: t('editor.table-delete-column') },
      onClick: () => {
        editor.chain().focus().deleteColumn().run()
      },
    }], [{
      icon: 'tabler:trash',
      tooltip: { text: t('editor.table-delete') },
      onClick: () => {
        editor.chain().focus().deleteTable().run()
      },
    }]]
  }

  return {
    slashmenuItems,
    bubbleToolbarItems,
    getTableToolbarItems,
  }
}
