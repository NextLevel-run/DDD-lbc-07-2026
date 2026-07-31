import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { renderRoute } from '#/test/renderRoute'

describe('Delete confirmation form', () => {
  it('renders only credentials and reason fields, no ad id input', async () => {
    await renderRoute('/classified-ads/ad-1/delete')

    expect(
      screen.getByRole('heading', { name: /confirm deletion/i }),
    ).toBeInTheDocument()
    expect(screen.getByLabelText(/your email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/your password/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/reason/i)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: /delete ad/i }),
    ).toBeInTheDocument()
  })
})
