import { useRequestHeaders } from '#app'
import { computed } from 'vue'
import type { TenantInfo } from '~/types/store'

const DEFAULT_META_DESCRIPTION_LENGTH = 160
type OpenGraphType = 'article' | 'website' | 'product'

function absoluteUrl(path: string) {
  if (import.meta.client) {
    return new URL(path, window.location.origin).toString()
  }

  const headers = useRequestHeaders(['x-forwarded-proto', 'host'])
  const protocol = headers['x-forwarded-proto'] || 'http'
  const host = headers.host || 'localhost:3000'
  return new URL(path, `${protocol}://${host}`).toString()
}

export function useStoreSeo(input: {
  title: string
  description: string
  image?: string
  type?: OpenGraphType
  robots?: string
  canonicalPath?: string
  siteName?: string
  locale?: string
  lang?: string
  jsonLd?: Record<string, unknown> | Array<Record<string, unknown>> | null
}) {
  const pageUrl = computed(() => absoluteUrl(input.canonicalPath ?? useRoute().path))
  const imageUrl = computed(() => {
    if (!input.image) {
      return absoluteUrl('/favicon.ico')
    }
    if (/^(https?:)?\/\//.test(input.image)) {
      return input.image
    }
    return absoluteUrl(input.image)
  })

  useHead({
    htmlAttrs: {
      lang: input.lang ?? 'zh-Hant',
    },
    link: [
      { rel: 'canonical', href: pageUrl.value },
    ],
    script: input.jsonLd
      ? [{
          type: 'application/ld+json',
          innerHTML: JSON.stringify(input.jsonLd),
        }]
      : [],
  })

  useSeoMeta({
    title: input.title,
    description: input.description,
    robots: input.robots ?? 'index,follow',
    ogTitle: input.title,
    ogDescription: input.description,
    ogType: (input.type ?? 'website') as 'article' | 'website',
    ogSiteName: input.siteName ?? 'Vape Group 商城',
    ogLocale: input.locale ?? 'zh_TW',
    ogUrl: pageUrl.value,
    ogImage: imageUrl.value,
    twitterCard: 'summary_large_image',
    twitterTitle: input.title,
    twitterDescription: input.description,
    twitterImage: imageUrl.value,
  })
}

function decodeHtmlEntities(input: string) {
  const namedEntities: Record<string, string> = {
    amp: '&',
    lt: '<',
    gt: '>',
    quot: '"',
    apos: '\'',
    nbsp: ' ',
    mdash: '-',
    ndash: '-',
    hellip: '...',
  }

  return input.replace(/&(#x?[0-9a-fA-F]+|[a-zA-Z]+);/g, (_, entity: string) => {
    if (entity.startsWith('#x') || entity.startsWith('#X')) {
      const codePoint = Number.parseInt(entity.slice(2), 16)
      return Number.isFinite(codePoint) ? String.fromCodePoint(codePoint) : ''
    }

    if (entity.startsWith('#')) {
      const codePoint = Number.parseInt(entity.slice(1), 10)
      return Number.isFinite(codePoint) ? String.fromCodePoint(codePoint) : ''
    }

    return namedEntities[entity] ?? ''
  })
}

function stripHtml(input: string) {
  return input
    .replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, ' ')
    .replace(/<style\b[^>]*>[\s\S]*?<\/style>/gi, ' ')
    .replace(/<br\s*\/?>/gi, ' ')
    .replace(/<\/(p|div|li|ul|ol|h[1-6]|table|tr|td|section|article)>/gi, ' ')
    .replace(/<[^>]+>/g, ' ')
}

function truncateText(input: string, maxLength: number) {
  if (input.length <= maxLength) {
    return input
  }

  return `${input.slice(0, Math.max(0, maxLength - 1)).trim()}…`
}

export function sanitizeMetaDescription(input: string | null | undefined, maxLength = DEFAULT_META_DESCRIPTION_LENGTH) {
  const plainText = decodeHtmlEntities(stripHtml(String(input ?? '')))
    .replace(/\s+/g, ' ')
    .trim()

  return truncateText(plainText, maxLength)
}

export function sanitizeProductMetaDescription(input: {
  description: string | null | undefined
  productName?: string | null | undefined
  maxLength?: number
}) {
  const plainText = decodeHtmlEntities(stripHtml(String(input.description ?? '')))
    .replace(/\s+/g, ' ')
    .trim()

  if (!plainText) {
    return ''
  }

  const startCandidates = [
    /產品名稱\s*[:：]/,
    /产品名称\s*[:：]/,
    input.productName ? new RegExp(input.productName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')) : null,
  ].filter(Boolean) as RegExp[]

  const startIndex = startCandidates
    .map((pattern) => plainText.search(pattern))
    .filter((index) => index >= 0)
    .sort((a, b) => a - b)[0]

  const normalizedText = startIndex != null ? plainText.slice(startIndex).trim() : plainText
  return truncateText(normalizedText, input.maxLength ?? DEFAULT_META_DESCRIPTION_LENGTH)
}

export function createStoreJsonLd(input: {
  name: string
  description: string
  url?: string
}) {
  return {
    '@context': 'https://schema.org',
    '@type': 'Store',
    name: input.name,
    description: input.description,
    url: input.url ?? absoluteUrl('/'),
  }
}

export function createBreadcrumbJsonLd(input: {
  items: Array<{
    name: string
    path?: string
    url?: string
  }>
}) {
  return {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: input.items.map((item, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      name: item.name,
      item: item.url ?? absoluteUrl(item.path ?? '/'),
    })),
  }
}

export function createItemListJsonLd(input: {
  name: string
  description?: string
  items: Array<{
    name: string
    path?: string
    url?: string
    image?: string
  }>
}) {
  return {
    '@context': 'https://schema.org',
    '@type': 'ItemList',
    name: input.name,
    ...(input.description ? { description: input.description } : {}),
    itemListElement: input.items.map((item, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      name: item.name,
      url: item.url ?? absoluteUrl(item.path ?? '/'),
      ...(item.image
        ? {
            image: [/^(https?:)?\/\//.test(item.image) ? item.image : absoluteUrl(item.image)],
          }
        : {}),
    })),
  }
}

export function createProductJsonLd(input: {
  name: string
  description: string
  image: string[]
  sku: string
  category: string
  price: number
  availability: boolean
  url?: string
  brand?: string
  mpn?: string
  currency?: string
  rating?: number
  reviews?: number
  itemCondition?: string
  additionalProperty?: Array<{
    name: string
    value: string
  }>
}) {
  return {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: input.name,
    description: input.description,
    image: input.image.map((value) => (/^(https?:)?\/\//.test(value) ? value : absoluteUrl(value))),
    sku: input.sku,
    ...(input.mpn
      ? {
          mpn: input.mpn,
        }
      : {}),
    ...(input.brand
      ? {
          brand: {
            '@type': 'Brand',
            name: input.brand,
          },
        }
      : {}),
    category: input.category,
    url: input.url ?? absoluteUrl(useRoute().path),
    ...(input.additionalProperty?.length
      ? {
          additionalProperty: input.additionalProperty.map((item) => ({
            '@type': 'PropertyValue',
            name: item.name,
            value: item.value,
          })),
        }
      : {}),
    offers: {
      '@type': 'Offer',
      priceCurrency: input.currency ?? 'TWD',
      price: input.price.toFixed(2),
      availability: input.availability ? 'https://schema.org/InStock' : 'https://schema.org/OutOfStock',
      itemCondition: input.itemCondition ?? 'https://schema.org/NewCondition',
      url: input.url ?? absoluteUrl(useRoute().path),
    },
    ...(input.reviews && input.rating
      ? {
          aggregateRating: {
            '@type': 'AggregateRating',
            ratingValue: input.rating.toFixed(1),
            reviewCount: input.reviews,
          },
        }
      : {}),
  }
}

export function resolveTenantSeo(tenant: TenantInfo | null, fallbackTitle: string, fallbackDescription: string) {
  return {
    title: tenant?.seoTitle?.trim() || fallbackTitle,
    description: tenant?.seoDescription?.trim() || fallbackDescription,
  }
}
