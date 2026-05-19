import { computed } from 'vue'
import type { TenantInfo } from '~/types/store'

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
  type?: 'website' | 'product' | 'article'
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
    ogType: input.type ?? 'website',
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
  rating?: number
  reviews?: number
}) {
  return {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: input.name,
    description: input.description,
    image: input.image.map((value) => (/^(https?:)?\/\//.test(value) ? value : absoluteUrl(value))),
    sku: input.sku,
    category: input.category,
    url: input.url ?? absoluteUrl(useRoute().path),
    offers: {
      '@type': 'Offer',
      priceCurrency: 'TWD',
      price: input.price.toFixed(2),
      availability: input.availability ? 'https://schema.org/InStock' : 'https://schema.org/OutOfStock',
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
