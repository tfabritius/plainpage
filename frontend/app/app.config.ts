export default defineAppConfig({
  ui: {
    colors: {
      // eslint-disable-next-line ts/no-explicit-any
      primary: 'plainpage' as any,
      warning: 'orange',
    },

    icons: {
      close: 'tabler:x',
      check: 'tabler:check',
      chevronRight: 'tabler:chevron-right',
      loading: 'tabler:loader-2',
    },

    button: {
      defaultVariants: {
        // eslint-disable-next-line ts/no-explicit-any
        color: 'neutral' as any,
        // eslint-disable-next-line ts/no-explicit-any
        variant: 'outline' as any,
      },
      slots: {
        base: 'cursor-pointer justify-center font-normal text-sm',
      },
      compoundVariants: [
        {
          color: 'neutral',
          variant: 'link',
          class: 'hover:text-[var(--ui-primary)]',
        },
        {
          color: 'neutral',
          variant: 'outline',
          class: 'hover:bg-[var(--ui-primary)]/10 hover:text-[var(--ui-primary)] hover:ring-[var(--ui-primary)]/50',
        },
        {
          size: 'md',
          // eslint-disable-next-line ts/no-explicit-any
          variant: 'link' as any,
          class: 'px-1.5',
        },
      ],
    },

    dropdownMenu: {
      variants: {
        active: {
          false: {
            item: 'data-highlighted:text-[var(--ui-primary)] data-highlighted:before:bg-[var(--ui-primary)]/10',
            itemLeadingIcon: 'group-data-highlighted:text-[var(--ui-primary)]',
          },
        },
      },
      slots: {
        item: 'cursor-pointer before:rounded-none before:inset-0',
        group: 'px-0',
        separator: 'mx-0',
      },
    },

    input: {
      compoundVariants: [
        {
          color: 'primary',
          // eslint-disable-next-line ts/no-explicit-any
          variant: ['outline', 'subtle'] as any,
          class: 'focus-visible:ring-1',
        },
      ],
    },

    link: {
      variants: {
        active: {
          true: 'hover:text-[var(--ui-primary)]',
          false: 'hover:text-[var(--ui-primary)]',
        },
      },
    },

    modal: {
      slots: {
        footer: 'justify-end',
      },
    },

    table: {
      slots: {
        tr: 'hover:bg-[var(--ui-bg-elevated)]/50',
      },
    },
  },
})
