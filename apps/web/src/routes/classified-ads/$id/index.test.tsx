import { screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { renderRoute } from '#/test/renderRoute'
import type { ClassifiedAdView } from '#/features/classified-ad/types'

describe('ClassifiedAd detail', () => {
  beforeEach(() => {
    const view: ClassifiedAdView = {
      id: 'ad-1',
      title: 'Vintage Bike',
      description: 'A great vintage bike.',
      priceInCents: 15000,
      category: 'consumer_goods',
      sellerPseudo: 'seller1',
      imageUrls: [],
      zipCode: '75001',
      cityName: 'Paris',
      submissionDate: '2026-01-01T00:00:00Z',
    }
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve(view),
      }),
    )
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders the ad detail and action links', async () => {
    await renderRoute('/classified-ads/ad-1')

    expect(
      screen.getByRole('heading', { name: 'Vintage Bike' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: /make an offer/i }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: /delete this ad/i }),
    ).toBeInTheDocument()
  })
})
