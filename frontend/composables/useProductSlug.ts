const slugUnsafeChars = /[^a-z0-9]+/g

const pinyinMap: Record<string, string> = {
  煙: 'yan',
  烟: 'yan',
  彈: 'dan',
  弹: 'dan',
  桿: 'gan',
  杆: 'gan',
  拋: 'pao',
  棄: 'qi',
  式: 'shi',
  電: 'dian',
  电: 'dian',
  子: 'zi',
  菸: 'yan',
  油: 'you',
  主: 'zhu',
  機: 'ji',
  机: 'ji',
  設: 'she',
  备: 'bei',
  備: 'bei',
  套: 'tao',
  裝: 'zhuang',
  装: 'zhuang',
  口: 'kou',
  味: 'wei',
  冰: 'bing',
  葡: 'pu',
  萄: 'tao',
  經: 'jing',
  经: 'jing',
  典: 'dian',
  草: 'cao',
}

export function slugifyProductName(value: string) {
  const source = value.trim()
  if (!source) {
    return 'product'
  }

  const segments: string[] = []
  let asciiBuffer = ''

  const flushAscii = () => {
    if (!asciiBuffer) {
      return
    }
    segments.push(asciiBuffer.toLowerCase())
    asciiBuffer = ''
  }

  for (const char of source) {
    if (/[a-z0-9]/i.test(char)) {
      asciiBuffer += char
      continue
    }

    flushAscii()

    const mapped = pinyinMap[char]
    if (mapped) {
      segments.push(mapped)
    }
  }

  flushAscii()

  const result = segments.join('-').replace(slugUnsafeChars, '-').replace(/^-+|-+$/g, '')
  return result || 'product'
}

export function buildProductPath(product: { id: number | string, name: string, slug?: string }) {
  const slug = product.slug?.trim() || slugifyProductName(product.name)
  return `/products/${product.id}-${slug}`
}

export function buildCategoryPath(category: { id: number | string, name: string }) {
  return `/products/category/${category.id}-${slugifyProductName(category.name)}`
}

export function parseProductIdFromRouteParam(value: string | string[] | undefined) {
  if (Array.isArray(value)) {
    value = value[0]
  }
  const raw = String(value ?? '').trim()
  if (!raw) {
    return NaN
  }
  const match = raw.match(/^(\d+)/)
  return match ? Number(match[1]) : NaN
}

export function parseCategoryIdFromRouteParam(value: string | string[] | undefined) {
  if (Array.isArray(value)) {
    value = value[0]
  }
  const raw = String(value ?? '').trim()
  if (!raw) {
    return NaN
  }
  const match = raw.match(/^(\d+)/)
  return match ? Number(match[1]) : NaN
}
