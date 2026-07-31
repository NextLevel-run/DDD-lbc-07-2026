import type {
  ClassifiedAdView,
  DeleteClassifiedAdRequest,
  MakeOfferRequest,
  SearchClassifiedAdsParams,
  SearchClassifiedAdsResponse,
  SubmitClassifiedAdRequest,
  SubmitClassifiedAdResponse,
} from './types'

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function parseErrorMessage(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { error?: string }
    return body.error ?? response.statusText
  } catch {
    return response.statusText
  }
}

// The Vite dev proxy (see vite.config.ts) only intercepts requests made by
// the browser. Server-side rendering runs in Node and has no proxy, so it
// must call the Go API directly.
const API_BASE_URL =
  typeof window === 'undefined' ? 'http://localhost:8080' : '/api'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  })

  if (!response.ok) {
    throw new ApiError(await parseErrorMessage(response), response.status)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}

export function submitClassifiedAd(body: SubmitClassifiedAdRequest) {
  return request<SubmitClassifiedAdResponse>('/classified-ads', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function searchClassifiedAds(params: SearchClassifiedAdsParams) {
  const query = new URLSearchParams()
  if (params.category) query.set('category', params.category)
  if (params.zip) query.set('zip', params.zip)
  if (params.city) query.set('city', params.city)
  if (params.q) query.set('q', params.q)
  if (params.minPrice !== undefined)
    query.set('minPrice', String(params.minPrice))
  if (params.maxPrice !== undefined)
    query.set('maxPrice', String(params.maxPrice))
  if (params.sortBy) query.set('sortBy', params.sortBy)
  if (params.limit !== undefined) query.set('limit', String(params.limit))
  if (params.offset !== undefined) query.set('offset', String(params.offset))

  const queryString = query.toString()
  return request<SearchClassifiedAdsResponse>(
    `/classified-ads${queryString ? `?${queryString}` : ''}`,
  )
}

export function getClassifiedAd(id: string) {
  return request<ClassifiedAdView>(`/classified-ads/${id}`)
}

export function makeOffer(id: string, body: MakeOfferRequest) {
  return request<void>(`/classified-ads/${id}/offers`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function deleteClassifiedAd(
  id: string,
  body: DeleteClassifiedAdRequest,
) {
  return request<void>(`/classified-ads/${id}`, {
    method: 'DELETE',
    body: JSON.stringify(body),
  })
}
