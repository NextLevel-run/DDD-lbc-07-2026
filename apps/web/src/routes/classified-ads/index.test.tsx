import { screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { renderRoute } from '#/test/renderRoute'
import type { SearchClassifiedAdsResponse } from '#/features/classified-ad/types'

describe('ClassifiedAd list', () => {
  beforeEach(() => {
    const response: SearchClassifiedAdsResponse = {
      items: [
        {
          id: 'ad-1',
          title: 'Vintage Bike',
          priceInCents: 15000,
          category: 'consumer_goods',
          cityName: 'Paris',
          zipCode: '75001',
          firstImageUrl: '',
          submissionDate: '2026-01-01T00:00:00Z',
        },
      ],
    }
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve(response),
      }),
    )
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders search results from the API', async () => {
    await renderRoute('/classified-ads')

    expect(
      screen.getByRole('heading', { name: /browse listings/i }),
    ).toBeInTheDocument()
    expect(await screen.findByText('Vintage Bike')).toBeInTheDocument()
  })
})
