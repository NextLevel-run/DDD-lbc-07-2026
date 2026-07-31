export const CATEGORIES = [
  'immo',
  'auto',
  'consumer_goods',
  'holidays',
] as const

export type Category = (typeof CATEGORIES)[number]

export const DELETE_REASONS = ['sold', 'no_more_to_sell', 'edit'] as const

export type DeleteReason = (typeof DELETE_REASONS)[number]

export const SORT_OPTIONS = [
  'date_desc',
  'date_asc',
  'price_asc',
  'price_desc',
] as const

export type SortBy = (typeof SORT_OPTIONS)[number]

export interface SubmitClassifiedAdRequest {
  title: string
  description: string
  priceInCents: number
  sellerEmail: string
  sellerPseudo: string
  sellerPassword: string
  imageUrls: Array<string>
  category: Category
  zipCode: string
  cityName: string
}

export interface SubmitClassifiedAdResponse {
  id: string
}

export interface MakeOfferRequest {
  buyerEmail: string
  buyerPseudo: string
  amountInCents: number
  message: string
}

export interface DeleteClassifiedAdRequest {
  email: string
  password: string
  reason: DeleteReason
}

export interface ClassifiedAdListItem {
  id: string
  title: string
  priceInCents: number
  category: Category
  cityName: string
  zipCode: string
  firstImageUrl: string
  submissionDate: string
}

export interface SearchClassifiedAdsResponse {
  items: Array<ClassifiedAdListItem>
}

export interface ClassifiedAdView {
  id: string
  title: string
  description: string
  priceInCents: number
  category: Category
  sellerPseudo: string
  imageUrls: Array<string>
  zipCode: string
  cityName: string
  submissionDate: string
}

export interface SearchClassifiedAdsParams {
  category?: Category
  zip?: string
  city?: string
  q?: string
  minPrice?: number
  maxPrice?: number
  sortBy?: SortBy
  limit?: number
  offset?: number
}
