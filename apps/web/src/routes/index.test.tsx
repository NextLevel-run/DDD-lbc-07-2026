import { screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { renderRoute } from '#/test/renderRoute'

describe('Homepage', () => {
  it('renders the entry points to browse and sell', async () => {
    await renderRoute('/')

    const main = within(screen.getByRole('main'))
    expect(
      main.getByRole('heading', { name: /buy and sell second-hand items/i }),
    ).toBeInTheDocument()
    expect(
      main.getByRole('link', { name: /browse classified ads/i }),
    ).toBeInTheDocument()
    expect(
      main.getByRole('link', { name: /sell an item/i }),
    ).toBeInTheDocument()
  })
})
