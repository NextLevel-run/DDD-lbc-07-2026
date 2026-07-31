import { queryOptions } from '@tanstack/react-query'
import { getClassifiedAd, searchClassifiedAds } from './api'
import type { SearchClassifiedAdsParams } from './types'

export function searchClassifiedAdsQueryOptions(
  params: SearchClassifiedAdsParams,
) {
  return queryOptions({
    queryKey: ['classified-ads', 'search', params],
    queryFn: () => searchClassifiedAds(params),
  })
}

export function classifiedAdQueryOptions(id: string) {
  return queryOptions({
    queryKey: ['classified-ads', id],
    queryFn: () => getClassifiedAd(id),
  })
}
