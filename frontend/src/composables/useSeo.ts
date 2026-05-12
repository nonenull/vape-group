import { computed, onBeforeUnmount, watchEffect, type ComputedRef, type Ref } from 'vue'

type MaybeRef<T> = T | Ref<T> | ComputedRef<T>

export interface SeoConfig {
  title: string
  description: string
  image?: string
  type?: 'website' | 'product' | 'article'
  robots?: string
  canonicalPath?: string
  jsonLd?: Record<string, unknown> | null
}

function resolveValue<T>(input: MaybeRef<T>): T {
  if (typeof input === 'object' && input !== null && 'value' in input) {
    return input.value as T
  }
  return input as T
}

function upsertMeta(selector: string, attributes: Record<string, string>) {
  let element = document.head.querySelector<HTMLMetaElement>(selector)
  if (!element) {
    element = document.createElement('meta')
    document.head.appendChild(element)
  }

  Object.entries(attributes).forEach(([key, value]) => {
    element!.setAttribute(key, value)
  })
}

function upsertLink(selector: string, attributes: Record<string, string>) {
  let element = document.head.querySelector<HTMLLinkElement>(selector)
  if (!element) {
    element = document.createElement('link')
    document.head.appendChild(element)
  }

  Object.entries(attributes).forEach(([key, value]) => {
    element!.setAttribute(key, value)
  })
}

function upsertJsonLd(schema: Record<string, unknown> | null) {
  const selector = 'script[data-seo-jsonld="true"]'
  const existing = document.head.querySelector<HTMLScriptElement>(selector)

  if (!schema) {
    existing?.remove()
    return
  }

  const script = existing ?? document.createElement('script')
  script.type = 'application/ld+json'
  script.setAttribute('data-seo-jsonld', 'true')
  script.textContent = JSON.stringify(schema)

  if (!existing) {
    document.head.appendChild(script)
  }
}

export function useSeo(config: MaybeRef<SeoConfig>) {
  const defaultTitle = document.title

  watchEffect(() => {
    const value = resolveValue(config)
    const title = value.title.trim()
    const description = value.description.trim()
    const pageUrl = new URL(value.canonicalPath ?? window.location.pathname, window.location.origin).toString()
    const imageUrl = value.image
      ? new URL(value.image, window.location.origin).toString()
      : `${window.location.origin}/favicon.ico`
    const robots = value.robots ?? 'index,follow'
    const type = value.type ?? 'website'

    document.title = title || defaultTitle

    upsertMeta('meta[name="description"]', { name: 'description', content: description })
    upsertMeta('meta[name="robots"]', { name: 'robots', content: robots })
    upsertMeta('meta[property="og:title"]', { property: 'og:title', content: title })
    upsertMeta('meta[property="og:description"]', { property: 'og:description', content: description })
    upsertMeta('meta[property="og:type"]', { property: 'og:type', content: type })
    upsertMeta('meta[property="og:url"]', { property: 'og:url', content: pageUrl })
    upsertMeta('meta[property="og:image"]', { property: 'og:image', content: imageUrl })
    upsertMeta('meta[name="twitter:card"]', { name: 'twitter:card', content: 'summary_large_image' })
    upsertMeta('meta[name="twitter:title"]', { name: 'twitter:title', content: title })
    upsertMeta('meta[name="twitter:description"]', { name: 'twitter:description', content: description })
    upsertMeta('meta[name="twitter:image"]', { name: 'twitter:image', content: imageUrl })
    upsertLink('link[rel="canonical"]', { rel: 'canonical', href: pageUrl })
    upsertJsonLd(value.jsonLd ?? null)
  })

  onBeforeUnmount(() => {
    upsertJsonLd(null)
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
    url: input.url ?? window.location.origin,
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
    image: input.image,
    sku: input.sku,
    category: input.category,
    url: input.url ?? window.location.href,
    offers: {
      '@type': 'Offer',
      priceCurrency: 'TWD',
      price: input.price.toFixed(2),
      availability: input.availability ? 'https://schema.org/InStock' : 'https://schema.org/OutOfStock',
      url: input.url ?? window.location.href,
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
