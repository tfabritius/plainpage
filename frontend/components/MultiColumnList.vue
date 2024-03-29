<script setup lang="ts" generic="T">
const props = defineProps<{
  items: T[]
  sortAndGroupBy: (e: T) => string
  groupIfMoreThan?: number
}>()

interface ItemGroup {
  header: string
  items: T[]
}

const sortedItems = computed(
  () => props.items
    .map(e => e) // Create a shallow copy to allow in-place sorting
    .sort((a, b) => props.sortAndGroupBy(a).localeCompare(props.sortAndGroupBy(b))),
)

const groupedItems = computed<ItemGroup[]>(() => {
  // Don't group items if there are too few items
  if (sortedItems.value.length <= (props.groupIfMoreThan ?? 0)) {
    return []
  }

  // Group items by first character of string returned by mapper function
  const groups: ItemGroup[] = []

  let currentGroup: ItemGroup = { header: '', items: [] }
  for (const item of sortedItems.value) {
    const firstLetter = props.sortAndGroupBy(item).charAt(0)

    if (firstLetter !== currentGroup.header) {
      currentGroup = { header: firstLetter, items: [] }
      groups.push(currentGroup)
    }

    currentGroup.items.push(item)
  }

  return groups
})
</script>

<template>
  <div v-if="groupedItems.length > 0 " class="columns-1 md:columns-2 lg:columns-3">
    <span v-for="(group, i) of groupedItems" :key="i">
      <span class="font-semibold">
        {{ group.header }}
      </span>
      <ul class="pl-1 mt-0 mb-2 list-none">
        <li v-for="(item, idx) of group.items" :key="idx">
          <slot name="item" :item="item" />
        </li>
      </ul>
    </span>
  </div>
  <div v-else>
    <ul class="pl-0 list-none">
      <li v-for="(item, idx) of sortedItems" :key="idx">
        <slot name="item" :item="item" />
      </li>
    </ul>
  </div>
</template>

<style scoped>

</style>
