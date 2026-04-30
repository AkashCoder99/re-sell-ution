export interface PublicUser {
  id: string
  email: string
  full_name: string
  city?: string
  bio?: string
  photo_url?: string
  is_admin?: boolean
}

export interface UpdateProfileRequest {
  full_name?: string
  city?: string
  bio?: string
  photo_url?: string
}
