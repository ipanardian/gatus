<template>
  <div class="flex items-center justify-between">
    <Button
      variant="outline"
      size="sm"
      :disabled="currentPage <= 1"
      @click="previousPage"
      class="flex items-center gap-1"
    >
      <ChevronLeft class="h-4 w-4" />
      Previous
    </Button>
    
    <span class="text-sm text-muted-foreground">
      Page {{ currentPage }} of {{ maxPages }}
    </span>
    
    <Button
      variant="outline"
      size="sm"
      :disabled="currentPage >= maxPages"
      @click="nextPage"
      class="flex items-center gap-1"
    >
      Next
      <ChevronRight class="h-4 w-4" />
    </Button>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const props = defineProps({
  numberOfResultsPerPage: Number,
  currentPageProp: {
    type: Number,
    default: 1
  },
  totalResults: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['page'])

const currentPage = ref(props.currentPageProp)

const maxPages = computed(() => {
  return Math.max(1, Math.ceil(props.totalResults / props.numberOfResultsPerPage))
})

const nextPage = () => {
  currentPage.value++
  emit('page', currentPage.value)
}

const previousPage = () => {
  currentPage.value--
  emit('page', currentPage.value)
}
</script>
