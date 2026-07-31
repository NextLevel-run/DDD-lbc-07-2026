import { RouterProvider, createMemoryHistory } from '@tanstack/react-router'
import { render, waitFor } from '@testing-library/react'
import { expect } from 'vitest'
import { getRouter } from '#/router'

export async function renderRoute(path: string) {
  const router = getRouter()
  router.update({
    history: createMemoryHistory({ initialEntries: [path] }),
    context: router.options.context,
  })

  const result = render(<RouterProvider router={router} />)
  await router.load()
  await waitFor(() => expect(router.state.status).toBe('idle'))

  return result
}
