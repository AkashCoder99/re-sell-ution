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

const mockListings: Listing[] = []
const mockListingImages: ListingImage[] = []
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
  const body = options.body
    ? JSON.parse(options.body as string)
    : null
  const authHeader = (options.headers as Record<string, string>)?.Authorization
  const token = authHeader?.replace('Bearer ', '')
  if (!token && path.startsWith('/api/v1/listings')) {
    throw new Error('Unauthorized')
  }

  // GET /api/v1/categories
  if (path === '/api/v1/categories' && method === 'GET') {
    return { categories: mockCategories } as TResponse
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
      status: 'active',
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
      status === 'draft'
        ? [] // mock has no drafts
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

  // POST /api/v1/listings/:id/images — upload image (mock: accept URL or base64)
  if (path.match(/^\/api\/v1\/listings\/[^/]+\/images$/) && method === 'POST') {
    const id = path.split('/')[4]
    const listing = mockListings.find((l) => l.id === id)
    if (!listing) throw new Error('Listing not found')
    const payload = body as { image_url?: string; position?: number }
    const imageUrl = payload?.image_url || 'https://placehold.co/400x300?text=Photo'
    const pos =
      payload?.position ??
      mockListingImages.filter((img) => img.listing_id === id).length
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

export function getCategories(token: string): Promise<GetCategoriesResponse> {
  return request<GetCategoriesResponse>('/api/v1/categories', { token })
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

export function getMyListings(
  token: string,
  params: { status?: 'active' | 'sold' | 'draft'; page?: number; limit?: number } = {}
): Promise<MyListingsResponse> {
  const sp = new URLSearchParams()
  if (params.status) sp.set('status', params.status)
  if (params.page) sp.set('page', String(params.page))
  if (params.limit) sp.set('limit', String(params.limit))
  const qs = sp.toString()
  return request<MyListingsResponse>(`/api/v1/listings/me${qs ? '?' + qs : ''}`, {
    token
  })
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
