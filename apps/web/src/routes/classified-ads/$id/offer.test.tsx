import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { renderRoute } from '#/test/renderRoute'

describe('Offer form', () => {
  it('renders all fields required to make an offer', async () => {
    await renderRoute('/classified-ads/ad-1/offer')

    expect(
      screen.getByRole('heading', { name: /propose your price/i }),
    ).toBeInTheDocument()
    expect(screen.getByLabelText(/your offer/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/message to the seller/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/your pseudo/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/your email/i)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: /send offer/i }),
    ).toBeInTheDocument()
  })
})
