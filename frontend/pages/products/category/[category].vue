<script setup lang="ts">
import ProductListPage from '~/components/store/ProductListPage.vue'
import { parseCategoryIdFromRouteParam } from '~/composables/useProductSlug'

const route = useRoute()
const categoryId = parseCategoryIdFromRouteParam(route.params.category)

if (!Number.isFinite(categoryId) || categoryId <= 0) {
  throw createError({
    statusCode: 404,
    statusMessage: '分類不存在',
  })
}
</script>

<template>
  <ProductListPage :initial-category-id="categoryId" />
</template>
