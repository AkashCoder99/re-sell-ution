/**
 * Listings API (EPIC-03) — create, my listings, upload images, mark as sold
 * Uses same base URL and mock pattern as auth.ts
 */

import type {
  Listing,
  ListingImage,
  Category,
  CreateListingRequest,
  ListingStatus
} from '../types/listing'

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true'

// ----- Helpers -----

async function request<TResponse>(
  path: string,
  options: RequestInit & { token?: string } = {}
): Promise<TResponse> {
  const { token, ...fetchOptions } = options
  const headers: Record<string, string> = {
    ...((fetchOptions.headers as Record<string, string>) || {})
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }
  if (
    fetchOptions.body &&
    typeof fetchOptions.body === 'string' &&
    !headers['Content-Type']
  ) {
    headers['Content-Type'] = 'application/json'
  }

  if (USE_MOCK) {
    return mockListingsApi<TResponse>(path, { ...fetchOptions, headers })
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...fetchOptions,
    headers
  })
  const data: unknown = await response.json().catch(() => ({}))

  if (!response.ok) {
    const message =
      typeof data === 'object' &&
      data !== null &&
      'error' in data &&
      typeof (data as { error?: unknown }).error === 'string'
        ? (data as { error: string }).error
        : 'Request failed'
    throw new Error(message)
  }

  return data as TResponse
}

// ----- Mock store -----

const mockListings: Listing[] = [
  {
    id: 'listing_001',
    seller_id: 'mock_user_id',
    category_id: 'cat_1',
    title: 'Apple MacBook Pro 14" M2 — Excellent Condition',
    description: 'Barely used MacBook Pro 14-inch with M2 chip, 16GB RAM, 512GB SSD. Comes with original charger and box. Selling because I upgraded to M3.',
    condition: 'like_new',
    price: 1299,
    currency: 'USD',
    city: 'Gainesville',
    state: 'Florida',
    status: 'active',
    view_count: 42,
    created_at: new Date(Date.now() - 2 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 2 * 86400000).toISOString(),
    images: [{ id: 'img_001_1', listing_id: 'listing_001', image_url: 'https://images.unsplash.com/photo-1517336714731-489689fd1ca8?w=400&h=300&fit=crop', position: 0 }]
  },
  {
    id: 'listing_002',
    seller_id: 'mock_user_id',
    category_id: 'cat_1',
    title: 'Sony WH-1000XM5 Noise Cancelling Headphones',
    description: 'Sony WH-1000XM5 wireless headphones in black. Used for 3 months, perfect working condition. Includes carry case and USB-C cable.',
    condition: 'good',
    price: 220,
    currency: 'USD',
    city: 'Gainesville',
    state: 'Florida',
    status: 'active',
    view_count: 28,
    created_at: new Date(Date.now() - 5 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 5 * 86400000).toISOString(),
    images: [{ id: 'img_002_1', listing_id: 'listing_002', image_url: 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&h=300&fit=crop', position: 0 }]
  },
  {
    id: 'listing_003',
    seller_id: 'mock_user_id',
    category_id: 'cat_2',
    title: 'IKEA MALM Queen Bed Frame — White',
    description: 'IKEA MALM queen bed frame in white. Disassembled and ready for pickup. Minor scuffs on legs but otherwise great condition. No mattress included.',
    condition: 'good',
    price: 85,
    currency: 'USD',
    city: 'Gainesville',
    state: 'Florida',
    status: 'active',
    view_count: 15,
    created_at: new Date(Date.now() - 7 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 7 * 86400000).toISOString(),
    images: [{ id: 'img_003_1', listing_id: 'listing_003', image_url: 'https://images.unsplash.com/photo-1555041469-a586c61ea9bc?w=400&h=300&fit=crop', position: 0 }]
  },
  {
    id: 'listing_004',
    seller_id: 'mock_user_id',
    category_id: 'cat_3',
    title: 'Nike Air Force 1 Low — Size 10 — Brand New',
    description: 'Brand new Nike Air Force 1 Low in white. Size 10 US. Never worn, still in original box. Got as a gift but wrong size.',
    condition: 'new',
    price: 95,
    currency: 'USD',
    city: 'Gainesville',
    state: 'Florida',
    status: 'active',
    view_count: 61,
    created_at: new Date(Date.now() - 1 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 1 * 86400000).toISOString(),
    images: [{ id: 'img_004_1', listing_id: 'listing_004', image_url: 'https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=400&h=300&fit=crop', position: 0 }]
  },
  {
    id: 'listing_005',
    seller_id: 'mock_user_id',
    category_id: 'cat_4',
    title: 'Textbook Bundle — UF Computer Science (COP3502, COP3503)',
    description: 'Selling my UF CS textbooks: Introduction to Programming with Python (3rd ed) and Data Structures & Algorithms. Both in good condition with some highlighting.',
    condition: 'good',
    price: 45,
    currency: 'USD',
    city: 'Gainesville',
    state: 'Florida',
    status: 'active',
    view_count: 33,
    created_at: new Date(Date.now() - 3 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 3 * 86400000).toISOString(),
    images: [{ id: 'img_005_1', listing_id: 'listing_005', image_url: 'https://images.unsplash.com/photo-1544716278-ca5e3f4abd8c?w=400&h=300&fit=crop', position: 0 }]
  },
  {
    id: 'listing_006',
    seller_id: 'mock_user_id',
    category_id: 'cat_1',
    title: 'iPad Air 5th Gen 64GB WiFi — Space Gray',
    description: 'iPad Air 5th generation, 64GB WiFi, Space Gray. Used for one semester. Screen is perfect, no scratches. Comes with Apple Pencil 2nd gen and Smart Folio case.',
    condition: 'like_new',
    price: 550,
    currency: 'USD',
    city: 'Gainesville',
    state: 'Florida',
    status: 'active',
    view_count: 19,
    created_at: new Date(Date.now() - 4 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 4 * 86400000).toISOString(),
    images: [{ id: 'img_006_1', listing_id: 'listing_006', image_url: 'https://images.unsplash.com/photo-1544244015-0df4b3ffc6b0?w=400&h=300&fit=crop', position: 0 }]
  },
  {
    id: 'listing_007',
    seller_id: 'mock_user_id',
    category_id: 'cat_2',
    title: 'Standing Desk — Adjustable Height — 55"',
    description: 'Electric standing desk, 55 inches wide, adjustable height from 28" to 47". Black frame with walnut top. Works perfectly. Moving out of apartment.',
    condition: 'good',
    price: 180,
    currency: 'USD',
    city: 'Gainesville',
    state: 'Florida',
    status: 'active',
    view_count: 24,
    created_at: new Date(Date.now() - 6 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 6 * 86400000).toISOString(),
    images: [{ id: 'img_007_1', listing_id: 'listing_007', image_url: 'https://images.unsplash.com/photo-1593642632559-0c6d3fc62b89?w=400&h=300&fit=crop', position: 0 }]
  },
  {
    id: 'listing_008',
    seller_id: 'mock_user_id',
    category_id: 'cat_5',
    title: 'Trek FX3 Hybrid Bike — Medium Frame',
    description: 'Trek FX3 hybrid bike, medium frame, matte black. Bought last year, ridden about 200 miles. Comes with front/rear lights and lock. Great for campus commuting.',
    condition: 'like_new',
    price: 420,
    currency: 'USD',
    city: 'Gainesville',
    state: 'Florida',
    status: 'active',
    view_count: 57,
    created_at: new Date(Date.now() - 8 * 86400000).toISOString(),
    updated_at: new Date(Date.now() - 8 * 86400000).toISOString(),
    images: [{ id: 'img_008_1', listing_id: 'listing_008', image_url: 'https://images.unsplash.com/photo-1485965120184-e220f721d03e?w=400&h=300&fit=crop', position: 0 }]
  }
]
const mockListingImages: ListingImage[] = mockListings.flatMap((l) => l.images ?? [])
const mockFavoriteIds = new Set<string>()
const mockCategories: Category[] = [
  { id: 'cat_1', name: 'Electronics', slug: 'electronics', parent_id: null },
  { id: 'cat_2', name: 'Furniture', slug: 'furniture', parent_id: null },
  { id: 'cat_3', name: 'Clothing', slug: 'clothing', parent_id: null },
  { id: 'cat_4', name: 'Books', slug: 'books', parent_id: null },
  { id: 'cat_5', name: 'Other', slug: 'other', parent_id: null }
]

function mockDelay(ms = 300): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

async function mockListingsApi<TResponse>(
  path: string,
  options: RequestInit
): Promise<TResponse> {
  await mockDelay()
  const method = options.method || 'GET'
  let body: unknown = null
  if (options.body instanceof FormData) {
    body = options.body
  } else if (typeof options.body === 'string') {
    try {
      body = JSON.parse(options.body)
    } catch {
      body = null
    }
  }
  const authHeader = (options.headers as Record<string, string>)?.Authorization
  const token = authHeader?.replace('Bearer ', '')
  if (!token && (path.startsWith('/api/v1/listings') || path.startsWith('/api/v1/favorites'))) {
    throw new Error('Unauthorized')
  }

  // GET /api/v1/listings/browse?city=&category_id=&page=1&limit=12
  if (path.startsWith('/api/v1/listings/browse') && method === 'GET') {
    const url = new URL(path, 'http://localhost')
    const city = url.searchParams.get('city') || ''
    const category_id = url.searchParams.get('category_id') || ''
    const page = Math.max(1, parseInt(url.searchParams.get('page') || '1', 10))
    const limit = Math.min(20, Math.max(1, parseInt(url.searchParams.get('limit') || '12', 10)))

    let filtered = mockListings.filter((l) => l.status === 'active')
    if (city) filtered = filtered.filter((l) => l.city.toLowerCase() === city.toLowerCase())
    if (category_id) filtered = filtered.filter((l) => l.category_id === category_id)

    const total = filtered.length
    const start = (page - 1) * limit
    const items = filtered.slice(start, start + limit).map((l) => ({
      ...l,
      images: mockListingImages.filter((img) => img.listing_id === l.id)
    }))
    return {
      listings: items,
      total,
      page,
      limit,
      total_pages: Math.ceil(total / limit) || 1
    } as TResponse
  }

  // GET /api/v1/categories
  if (path === '/api/v1/categories' && method === 'GET') {
    return { categories: mockCategories } as TResponse
  }

  // GET /api/v1/listings/search?q=...&city=...
  if (path.startsWith('/api/v1/listings/search') && method === 'GET') {
    const url = new URL(path, 'http://localhost')
    const q = (url.searchParams.get('q') || '').toLowerCase()
    const city = (url.searchParams.get('city') || '').toLowerCase()
    const filtered = mockListings.filter((l) => {
      if (l.status !== 'active') return false
      if (city && l.city.toLowerCase() !== city) return false
      if (!q) return true
      const haystack = `${l.title} ${l.description}`.toLowerCase()
      return haystack.includes(q)
    })
    const items = filtered.map((l) => ({
      ...l,
      images: mockListingImages.filter((img) => img.listing_id === l.id)
    }))
    return { listings: items, total: items.length } as TResponse
  }

  // POST /api/v1/listings — create listing
  if (path === '/api/v1/listings' && method === 'POST') {
    const payload = body as CreateListingRequest & { image_urls?: string[] }
    const id = 'listing_' + Date.now()
    const now = new Date().toISOString()
    const listing: Listing = {
      id,
      seller_id: 'mock_user_id',
      category_id: payload.category_id ?? null,
      title: payload.title,
      description: payload.description,
      condition: payload.condition,
      price: Number(payload.price),
      currency: payload.currency || 'INR',
      city: payload.city,
      state: payload.state ?? null,
      status: payload.status ?? 'active',
      view_count: 0,
      created_at: now,
      updated_at: now
    }
    mockListings.push(listing)
    const imageUrls = (payload as { image_urls?: string[] }).image_urls || []
    imageUrls.forEach((url, i) => {
      mockListingImages.push({
        id: `img_${id}_${i}`,
        listing_id: id,
        image_url: url,
        position: i
      })
    })
    return { listing } as TResponse
  }

  // GET /api/v1/listings/me?status=active|sold|draft&page=1&limit=10
  if (path.startsWith('/api/v1/listings/me') && method === 'GET') {
    const url = new URL(path, 'http://localhost')
    const status = url.searchParams.get('status') || 'active'
    const page = Math.max(1, parseInt(url.searchParams.get('page') || '1', 10))
    const limit = Math.min(20, Math.max(5, parseInt(url.searchParams.get('limit') || '10', 10)))
    const myListings = mockListings.filter((l) => l.seller_id === 'mock_user_id')
    const byStatus =
      status === 'all'
        ? myListings
        : myListings.filter((l) => l.status === status)
    const total = byStatus.length
    const start = (page - 1) * limit
    const items = byStatus.slice(start, start + limit).map((l) => ({
      ...l,
      images: mockListingImages.filter((img) => img.listing_id === l.id)
    }))
    return {
      listings: items,
      total,
      page,
      limit,
      total_pages: Math.ceil(total / limit) || 1
    } as TResponse
  }

  // PATCH /api/v1/listings/:id/status
  if (path.match(/^\/api\/v1\/listings\/[^/]+\/status$/) && method === 'PATCH') {
    const id = path.split('/')[4]
    const payload = body as { status: ListingStatus; sold_to_user_id?: string }
    const listing = mockListings.find((l) => l.id === id)
    if (!listing) throw new Error('Listing not found')
    listing.status = payload.status
    listing.updated_at = new Date().toISOString()
    if (payload.sold_to_user_id !== undefined) {
      ;(listing as Listing & { sold_to_user_id?: string }).sold_to_user_id =
        payload.sold_to_user_id
    }
    return { listing } as TResponse
  }

  // PATCH /api/v1/listings/:id
  if (path.match(/^\/api\/v1\/listings\/[^/]+$/) && method === 'PATCH') {
    const id = path.split('/')[4]
    const payload = body as Partial<CreateListingRequest>
    const listing = mockListings.find((l) => l.id === id)
    if (!listing) throw new Error('Listing not found')

    if (typeof payload.title === 'string') listing.title = payload.title
    if (typeof payload.description === 'string') listing.description = payload.description
    if (typeof payload.condition === 'string') listing.condition = payload.condition as Listing['condition']
    if (typeof payload.price === 'number') listing.price = payload.price
    if (typeof payload.currency === 'string') listing.currency = payload.currency
    if (typeof payload.city === 'string') listing.city = payload.city
    if (typeof payload.state === 'string') listing.state = payload.state
    if (payload.category_id !== undefined) listing.category_id = payload.category_id ?? null
    listing.updated_at = new Date().toISOString()

    return { listing } as TResponse
  }

  // DELETE /api/v1/listings/:id
  if (path.match(/^\/api\/v1\/listings\/[^/]+$/) && method === 'DELETE') {
    const id = path.split('/')[4]
    const idx = mockListings.findIndex((l) => l.id === id)
    if (idx === -1) throw new Error('Listing not found')
    mockListings.splice(idx, 1)
    const toRemove = mockListingImages.filter((img) => img.listing_id === id)
    toRemove.forEach((img) => {
      const i = mockListingImages.indexOf(img)
      if (i !== -1) mockListingImages.splice(i, 1)
    })
    return { message: 'deleted' } as TResponse
  }

  // GET /api/v1/favorites/:listing_id
  if (path.match(/^\/api\/v1\/favorites\/[^/]+$/) && method === 'GET') {
    const listingId = path.split('/')[4]
    return { favorited: mockFavoriteIds.has(listingId) } as TResponse
  }

  // PUT /api/v1/favorites/:listing_id
  if (path.match(/^\/api\/v1\/favorites\/[^/]+$/) && method === 'PUT') {
    const listingId = path.split('/')[4]
    mockFavoriteIds.add(listingId)
    return { message: 'favorited' } as TResponse
  }

  // DELETE /api/v1/favorites/:listing_id
  if (path.match(/^\/api\/v1\/favorites\/[^/]+$/) && method === 'DELETE') {
    const listingId = path.split('/')[4]
    mockFavoriteIds.delete(listingId)
    return { message: 'removed' } as TResponse
  }

  // POST /api/v1/listings/:id/report
  if (path.match(/^\/api\/v1\/listings\/[^/]+\/report$/) && method === 'POST') {
    return { message: 'listing reported' } as TResponse
  }

  // POST /api/v1/listings/:id/images — upload image (mock: accept URL or base64)
  if (path.match(/^\/api\/v1\/listings\/[^/]+\/images$/) && method === 'POST') {
    const id = path.split('/')[4]
    const listing = mockListings.find((l) => l.id === id)
    if (!listing) throw new Error('Listing not found')
    const defaultPosition = mockListingImages.filter((img) => img.listing_id === id).length

    let imageUrl = 'https://placehold.co/400x300?text=Photo'
    let pos = defaultPosition

    if (body instanceof FormData) {
      const file = body.get('file')
      const position = body.get('position')
      if (typeof position === 'string') {
        const parsed = Number.parseInt(position, 10)
        if (!Number.isNaN(parsed)) {
          pos = parsed
        }
      }
      if (file && typeof file === 'object' && 'name' in file && typeof file.name === 'string') {
        imageUrl = `https://placehold.co/400x300?text=${encodeURIComponent(file.name)}`
      }
    } else {
      const payload = body as { image_url?: string; position?: number } | null
      imageUrl = payload?.image_url || imageUrl
      pos = payload?.position ?? pos
    }

    const img: ListingImage = {
      id: `img_${id}_${Date.now()}`,
      listing_id: id,
      image_url: imageUrl,
      position: pos
    }
    mockListingImages.push(img)
    return { image: img } as TResponse
  }

  throw new Error('Mock listings endpoint not implemented: ' + method + ' ' + path)
}

// ----- Public API -----

export interface GetCategoriesResponse {
  categories: Category[]
}

export async function getCategories(token: string): Promise<GetCategoriesResponse> {
  const res = await request<Partial<GetCategoriesResponse>>('/api/v1/categories', { token })
  return {
    categories: Array.isArray(res?.categories) ? res.categories : []
  }
}

export interface PublicListingsResponse {
  listings: Listing[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export async function getPublicListings(
  token: string,
  params: { city?: string; category_id?: string; page?: number; limit?: number } = {}
): Promise<PublicListingsResponse> {
  const sp = new URLSearchParams()
  if (params.city) sp.set('city', params.city)
  if (params.category_id) sp.set('category_id', params.category_id)
  if (params.page) sp.set('page', String(params.page))
  if (params.limit) sp.set('limit', String(params.limit))
  const qs = sp.toString()
  const res = await request<Partial<PublicListingsResponse>>(`/api/v1/listings/browse${qs ? '?' + qs : ''}`, { token })
  const listings = Array.isArray(res?.listings) ? res.listings : []
  const limit = typeof res?.limit === 'number' && res.limit > 0 ? res.limit : (params.limit ?? 12)
  const page = typeof res?.page === 'number' && res.page > 0 ? res.page : (params.page ?? 1)
  const total = typeof res?.total === 'number' && res.total >= 0 ? res.total : listings.length
  const totalPages =
    typeof res?.total_pages === 'number' && res.total_pages > 0
      ? res.total_pages
      : Math.max(1, Math.ceil(total / limit))

  return {
    listings,
    total,
    page,
    limit,
    total_pages: totalPages
  }
}

export interface SearchListingsResponse {
  listings: Listing[]
  total: number
}

export async function searchListings(
  token: string,
  params: { query: string; city?: string }
): Promise<SearchListingsResponse> {
  const sp = new URLSearchParams()
  sp.set('q', params.query)
  if (params.city) sp.set('city', params.city)
  const qs = sp.toString()
  const res = await request<Partial<SearchListingsResponse>>(`/api/v1/listings/search?${qs}`, { token })
  const listings = Array.isArray(res?.listings) ? res.listings : []
  return {
    listings,
    total: typeof res?.total === 'number' && res.total >= 0 ? res.total : listings.length
  }
}

export interface CreateListingResponse {
  listing: Listing
}

export interface CreateListingPayload extends CreateListingRequest {
  image_urls?: string[]
}

export function createListing(
  token: string,
  payload: CreateListingPayload
): Promise<CreateListingResponse> {
  return request<CreateListingResponse>('/api/v1/listings', {
    method: 'POST',
    token,
    body: JSON.stringify(payload)
  })
}

export interface MyListingsResponse {
  listings: Listing[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export async function getMyListings(
  token: string,
  params: { status?: 'active' | 'reserved' | 'sold' | 'draft' | 'all'; page?: number; limit?: number } = {}
): Promise<MyListingsResponse> {
  const sp = new URLSearchParams()
  if (params.status) sp.set('status', params.status)
  if (params.page) sp.set('page', String(params.page))
  if (params.limit) sp.set('limit', String(params.limit))
  const qs = sp.toString()
  const res = await request<Partial<MyListingsResponse>>(`/api/v1/listings/me${qs ? '?' + qs : ''}`, {
    token
  })
  const listings = Array.isArray(res?.listings) ? res.listings : []
  const limit = typeof res?.limit === 'number' && res.limit > 0 ? res.limit : (params.limit ?? 10)
  const page = typeof res?.page === 'number' && res.page > 0 ? res.page : (params.page ?? 1)
  const total = typeof res?.total === 'number' && res.total >= 0 ? res.total : listings.length
  const totalPages =
    typeof res?.total_pages === 'number' && res.total_pages > 0
      ? res.total_pages
      : Math.max(1, Math.ceil(total / limit))

  return {
    listings,
    total,
    page,
    limit,
    total_pages: totalPages
  }
}

export function updateListingStatus(
  token: string,
  listingId: string,
  payload: { status: ListingStatus; sold_to_user_id?: string }
): Promise<{ listing: Listing }> {
  return request<{ listing: Listing }>(
    `/api/v1/listings/${listingId}/status`,
    {
      method: 'PATCH',
      token,
      body: JSON.stringify(payload)
    }
  )
}

export function deleteListing(token: string, listingId: string): Promise<{ message: string }> {
  return request<{ message: string }>(`/api/v1/listings/${listingId}`, {
    method: 'DELETE',
    token
  })
}

export function updateListing(
  token: string,
  listingId: string,
  payload: CreateListingPayload
): Promise<CreateListingResponse> {
  return request<CreateListingResponse>(`/api/v1/listings/${listingId}`, {
    method: 'PATCH',
    token,
    body: JSON.stringify(payload)
  })
}

export interface UploadImageResponse {
  image: ListingImage
}

export function uploadListingImage(
  token: string,
  listingId: string,
  payload: { image_url: string; position?: number }
): Promise<UploadImageResponse> {
  return request<UploadImageResponse>(`/api/v1/listings/${listingId}/images`, {
    method: 'POST',
    token,
    body: JSON.stringify(payload)
  })
}

export function uploadListingImageFile(
  token: string,
  listingId: string,
  file: File,
  payload: { position?: number } = {}
): Promise<UploadImageResponse> {
  const form = new FormData()
  form.append('file', file)
  if (payload.position !== undefined) {
    form.append('position', String(payload.position))
  }

  return request<UploadImageResponse>(`/api/v1/listings/${listingId}/images`, {
    method: 'POST',
    token,
    body: form
  })
}

export interface FavoriteStatusResponse {
  favorited: boolean
}

export interface FavoriteListingsResponse {
  listings: Listing[]
  total: number
}

export function getFavoriteStatus(
  token: string,
  listingId: string
): Promise<FavoriteStatusResponse> {
  return request<FavoriteStatusResponse>(`/api/v1/favorites/${listingId}`, {
    token
  })
}

export function addFavorite(
  token: string,
  listingId: string
): Promise<{ message: string }> {
  return request<{ message: string }>(`/api/v1/favorites/${listingId}`, {
    method: 'PUT',
    token
  })
}

export function removeFavorite(
  token: string,
  listingId: string
): Promise<{ message: string }> {
  return request<{ message: string }>(`/api/v1/favorites/${listingId}`, {
    method: 'DELETE',
    token
  })
}

export async function getFavoriteListings(token: string): Promise<FavoriteListingsResponse> {
  if (USE_MOCK) {
    const listings = mockListings
      .filter((listing) => mockFavoriteIds.has(listing.id))
      .map((listing) => ({
        ...listing,
        images: mockListingImages.filter((img) => img.listing_id === listing.id)
      }))
    return { listings, total: listings.length }
  }

  const res = await request<Partial<FavoriteListingsResponse>>('/api/v1/favorites', { token })
  const listings = Array.isArray(res?.listings) ? res.listings : []
  return {
    listings,
    total: typeof res?.total === 'number' && res.total >= 0 ? res.total : listings.length
  }
}

export function reportListing(
  token: string,
  listingId: string,
  payload: { reason?: string } = {}
): Promise<{ message: string }> {
  return request<{ message: string }>(`/api/v1/listings/${listingId}/report`, {
    method: 'POST',
    token,
    body: JSON.stringify(payload)
  })
}
