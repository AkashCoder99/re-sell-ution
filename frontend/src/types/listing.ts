/**
 * Types for listings and categories (EPIC-03)
 */

export type ListingCondition = 'new' | 'like_new' | 'good' | 'fair' | 'poor'
export type ListingStatus = 'active' | 'reserved' | 'sold' | 'deleted'

export interface Category {
  id: string
  name: string
  slug: string
  parent_id: string | null
}

export interface ListingImage {
  id: string
  listing_id: string
  image_url: string
  position: number
}

export interface Listing {
  id: string
  seller_id: string
  category_id: string | null
  title: string
  description: string
  condition: ListingCondition
  price: number
  currency: string
  city: string
  state: string | null
  status: ListingStatus
  view_count: number
  created_at: string
  updated_at: string
  images?: ListingImage[]
  sold_to_user_id?: string | null
}

export interface CreateListingRequest {
  title: string
  description: string
  condition: ListingCondition
  price: number
  currency?: string
  city: string
  state?: string
  category_id?: string | null
}

export interface CreateListingDraft {
  title: string
  description: string
  condition: ListingCondition
  price: number
  currency: string
  city: string
  state: string
  category_id: string | null
  image_urls: string[]
}

export const LISTING_CONDITIONS: { value: ListingCondition; label: string }[] = [
  { value: 'new', label: 'New' },
  { value: 'like_new', label: 'Like New' },
  { value: 'good', label: 'Good' },
  { value: 'fair', label: 'Fair' },
  { value: 'poor', label: 'Poor' }
]

export const LISTING_STATUS_LABELS: Record<ListingStatus, string> = {
  active: 'Active',
  reserved: 'Reserved',
  sold: 'Sold',
  deleted: 'Deleted'
}
