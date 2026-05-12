import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { productAPI } from '@/api/products'
import { resolveAssetURL } from '@/api/client'
import type { Product, ProductOptionGroup, ProductSkuVariant, ProductVariant } from '@/data/mockProducts'

export const useProductStore = defineStore('product', () => {
  const products = ref<Product[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const currentPage = ref(1)
  const pageSize = ref(20)
  const total = ref(0)
  const selectedProduct = ref<Product | null>(null)

  const normalizeProduct = (input: any): Product => {
    const gallery =
      input.custom_images?.length
        ? input.custom_images
        : input.gallery?.length
          ? input.gallery
          : input.base_images?.length
            ? input.base_images
            : input.preview_image
              ? [input.preview_image]
              : ['/src/assets/logo.svg']

    const detailImages =
      input.custom_detail_images?.length
        ? input.custom_detail_images
        : input.detail_images?.length
          ? input.detail_images
          : input.specifications?.detailImages?.length
            ? input.specifications.detailImages
            : []

    const normalizedGallery = gallery.map(resolveAssetURL)
    const normalizedDetailImages = detailImages.map(resolveAssetURL)
    const image = resolveAssetURL(input.preview_image ?? normalizedGallery[0] ?? '/src/assets/logo.svg')

    const variants: ProductVariant[] = Array.isArray(input.variants)
      ? input.variants
        .map((item: any) => ({
          name: String(item?.name ?? '').trim(),
          sku: String(item?.sku ?? '').trim(),
        }))
        .filter((item: ProductVariant) => item.name && item.sku)
      : []

    const optionGroups: ProductOptionGroup[] = Array.isArray(input.option_groups ?? input.optionGroups)
      ? (input.option_groups ?? input.optionGroups)
        .map((item: any) => ({
          name: String(item?.name ?? '').trim(),
          values: Array.isArray(item?.values) ? item.values.map((value: any) => String(value).trim()).filter(Boolean) : [],
        }))
        .filter((item: ProductOptionGroup) => item.name && item.values.length)
      : []

    const skuVariants: ProductSkuVariant[] = Array.isArray(input.sku_variants ?? input.skuVariants)
      ? (input.sku_variants ?? input.skuVariants)
        .map((item: any) => ({
          sku: String(item?.sku ?? '').trim(),
          price: item?.price == null ? null : Number(item.price),
          stock: item?.stock == null ? null : Number(item.stock),
          selections: item?.selections && typeof item.selections === 'object' ? item.selections : {},
        }))
        .filter((item: ProductSkuVariant) => item.sku && Object.keys(item.selections).length)
      : []

    return {
      id: Number(input.id),
      name: input.custom_name ?? input.base_name ?? '未命名商品',
      sku: input.sku ?? 'N/A',
      price: Number(input.custom_price ?? input.base_price ?? 0),
      salePrice: undefined,
      category: input.category ?? input.specifications?.category ?? '未分類',
      rating: Number(input.rating ?? input.specifications?.rating ?? 4.5),
      reviews: Number(input.reviews ?? input.specifications?.reviews ?? 0),
      stock: Number(input.custom_stock_quantity ?? input.base_stock_quantity ?? 0),
      image,
      gallery: normalizedGallery,
      detailImages: normalizedDetailImages,
      badge: input.badge ?? input.specifications?.badge ?? '',
      description: input.custom_description ?? input.description ?? input.specifications?.description ?? '暫無商品描述',
      longDescription:
        input.custom_description ??
        input.long_description ??
        input.specifications?.longDescription ??
        input.specifications?.description ??
        '暫無商品詳情',
      flavors: input.flavors ?? input.specifications?.flavors ?? [],
      variants,
      optionGroups,
      skuVariants,
      specs: [
        { label: 'SKU', value: input.sku ?? 'N/A' },
        { label: '分類', value: input.category ?? input.specifications?.category ?? '未分類' },
        { label: '庫存', value: String(input.custom_stock_quantity ?? input.base_stock_quantity ?? 0) },
      ],
    }
  }

  const fetchProducts = async (page = 1, limit = 20) => {
    loading.value = true
    error.value = null
    try {
      const response = await productAPI.getProducts(page, limit)
      products.value = response.data.map(normalizeProduct)
      total.value = response.total
      currentPage.value = page
      pageSize.value = limit
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch products'
      products.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  const fetchProductDetail = async (id: number) => {
    loading.value = true
    error.value = null
    try {
      const product = normalizeProduct(await productAPI.getProductDetail(id))
      selectedProduct.value = product
      return product
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch product'
      selectedProduct.value = null
      return null
    } finally {
      loading.value = false
    }
  }

  const pagination = computed(() => ({
    current: currentPage.value,
    size: pageSize.value,
    total: total.value,
    pages: Math.ceil(total.value / pageSize.value),
  }))

  return {
    products,
    loading,
    error,
    selectedProduct,
    pagination,
    fetchProducts,
    fetchProductDetail,
  }
})
