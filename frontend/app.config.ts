export default defineAppConfig({
  ui: {
    colors: {
      // eslint-disable-next-line ts/no-explicit-any
      primary: 'plainpage' as any,
    },

    icons: {
      close: 'ci:close-md',
      check: 'ci:check',
      chevronRight: 'ci:chevron-right-md',
      loading: 'ci:arrows-reload-01',
    },

    button: {
      defaultVariants: {
        // eslint-disable-next-line ts/no-explicit-any
        color: 'neutral' as any,
      },
      slots: {
        // eslint-disable-next-line ts/no-explicit-any
        base: 'cursor-pointer justify-center font-normal text-sm' as any,
      },
      variants: {
        size: {
          md: {
            // height of icon only: 14px
            //   + 8 px padding top/bottom = 30px
            // height of icon + text: 20px
            //   + 5 px padding top/bottom = 30px
            base: 'px-4 py-[8px] md:py-[5px]',
          },
        },
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
            // eslint-disable-next-line ts/no-explicit-any
            item: 'data-highlighted:text-[var(--ui-primary)] data-highlighted:before:bg-[var(--ui-primary)]/10' as any,
            // eslint-disable-next-line ts/no-explicit-any
            itemLeadingIcon: 'group-data-highlighted:text-[var(--ui-primary)]' as any,
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
          variant: ['outline', 'subtle'],
          class: 'focus-visible:ring-1',
        },
      ],
    },

    link: {
      variants: {
        active: {
          true: 'hover:text-[var(--ui-primary)]',
          // eslint-disable-next-line ts/no-explicit-any
          false: 'hover:text-[var(--ui-primary)]' as any,
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
