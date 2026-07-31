import { Link, createFileRoute, notFound } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { ApiError } from '#/features/classified-ad/api'
import { classifiedAdQueryOptions } from '#/features/classified-ad/queries'

export const Route = createFileRoute('/classified-ads/$id/')({
  loader: async ({ context, params }) => {
    try {
      await context.queryClient.ensureQueryData(
        classifiedAdQueryOptions(params.id),
      )
    } catch (error) {
      if (error instanceof ApiError && error.status === 404) {
        throw notFound()
      }
      throw error
    }
  },
  notFoundComponent: ClassifiedAdNotFound,
  component: ClassifiedAdDetail,
})

function ClassifiedAdNotFound() {
  return (
    <main className="demo-page demo-center">
      <section className="demo-panel w-full max-w-2xl text-center">
        <h1 className="demo-title mb-3">Ad not found</h1>
        <p className="demo-muted mb-6">
          This classified ad doesn't exist, or has been deleted or expired.
        </p>
        <Link to="/classified-ads" className="demo-button no-underline">
          Back to Listings
        </Link>
      </section>
    </main>
  )
}

function formatPrice(priceInCents: number) {
  return (priceInCents / 100).toLocaleString('en-US', {
    style: 'currency',
    currency: 'EUR',
  })
}

function ClassifiedAdDetail() {
  const { id } = Route.useParams()
  const { data: ad } = useSuspenseQuery(classifiedAdQueryOptions(id))

  return (
    <main className="demo-page">
      <div className="demo-panel">
        {ad.imageUrls.length > 0 && (
          <div className="mb-6 grid grid-cols-1 gap-3 sm:grid-cols-3">
            {ad.imageUrls.map((url) => (
              <img
                key={url}
                src={url}
                alt=""
                className="h-48 w-full rounded-xl object-cover"
              />
            ))}
          </div>
        )}

        <p className="island-kicker mb-2">{ad.category}</p>
        <h1 className="demo-title mb-3">{ad.title}</h1>
        <p className="mb-4 text-2xl font-bold text-[var(--sea-ink)]">
          {formatPrice(ad.priceInCents)}
        </p>
        <p className="demo-muted mb-6">
          {ad.cityName} ({ad.zipCode}) &middot; sold by {ad.sellerPseudo}
        </p>
        <p className="mb-8 whitespace-pre-wrap text-base text-[var(--sea-ink)]">
          {ad.description}
        </p>

        <div className="flex flex-wrap gap-3">
          <Link
            to="/classified-ads/$id/offer"
            params={{ id: ad.id }}
            className="demo-button no-underline"
          >
            Make an Offer
          </Link>
          <Link
            to="/classified-ads/$id/delete"
            params={{ id: ad.id }}
            className="demo-button demo-button-secondary no-underline"
          >
            Delete This Ad
          </Link>
        </div>
      </div>
    </main>
  )
}
