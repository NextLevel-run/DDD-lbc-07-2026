import { Link, createFileRoute } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { searchClassifiedAdsQueryOptions } from '#/features/classified-ad/queries'
import { CATEGORIES, SORT_OPTIONS } from '#/features/classified-ad/types'

const searchSchema = z.object({
  category: z.enum(CATEGORIES).optional(),
  zip: z.string().optional(),
  city: z.string().optional(),
  q: z.string().optional(),
  minPrice: z.number().optional(),
  maxPrice: z.number().optional(),
  sortBy: z.enum(SORT_OPTIONS).optional(),
})

export const Route = createFileRoute('/classified-ads/')({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => search,
  loader: ({ context, deps }) =>
    context.queryClient.ensureQueryData(
      searchClassifiedAdsQueryOptions(deps),
    ),
  component: ClassifiedAdList,
})

function formatPrice(priceInCents: number) {
  return (priceInCents / 100).toLocaleString('en-US', {
    style: 'currency',
    currency: 'EUR',
  })
}

function ClassifiedAdList() {
  const search = Route.useSearch()
  const navigate = Route.useNavigate()
  const { data } = useSuspenseQuery(searchClassifiedAdsQueryOptions(search))

  return (
    <main className="demo-page demo-page-wide">
      <div className="mb-6">
        <p className="island-kicker mb-2">Classified Ads</p>
        <h1 className="demo-title">Browse listings</h1>
      </div>

      <form
        className="demo-panel mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4"
        onSubmit={(e) => {
          e.preventDefault()
          const formData = new FormData(e.currentTarget)
          const q = String(formData.get('q') ?? '').trim()
          const category = String(formData.get('category') ?? '')
          const city = String(formData.get('city') ?? '').trim()
          const zip = String(formData.get('zip') ?? '').trim()
          navigate({
            search: {
              q: q || undefined,
              category: category
                ? (category as (typeof CATEGORIES)[number])
                : undefined,
              city: city || undefined,
              zip: zip || undefined,
            },
          })
        }}
      >
        <div>
          <label htmlFor="q" className="mb-2 block text-sm font-semibold">
            Keywords
          </label>
          <input
            id="q"
            name="q"
            defaultValue={search.q}
            className="demo-input"
          />
        </div>
        <div>
          <label
            htmlFor="category"
            className="mb-2 block text-sm font-semibold"
          >
            Category
          </label>
          <select
            id="category"
            name="category"
            defaultValue={search.category ?? ''}
            className="demo-select"
          >
            <option value="">All categories</option>
            {CATEGORIES.map((category) => (
              <option key={category} value={category}>
                {category}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label htmlFor="city" className="mb-2 block text-sm font-semibold">
            City
          </label>
          <input
            id="city"
            name="city"
            defaultValue={search.city}
            className="demo-input"
          />
        </div>
        <div>
          <label htmlFor="zip" className="mb-2 block text-sm font-semibold">
            Zip code
          </label>
          <input
            id="zip"
            name="zip"
            defaultValue={search.zip}
            className="demo-input"
          />
        </div>
        <div className="sm:col-span-2 lg:col-span-4">
          <button type="submit" className="demo-button">
            Search
          </button>
        </div>
      </form>

      {data.items.length === 0 ? (
        <p className="demo-muted">No classified ads match your search.</p>
      ) : (
        <ul className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {data.items.map((item) => (
            <li key={item.id}>
              <Link
                to="/classified-ads/$id"
                params={{ id: item.id }}
                className="demo-card block no-underline"
              >
                {item.firstImageUrl && (
                  <img
                    src={item.firstImageUrl}
                    alt=""
                    className="mb-3 h-40 w-full rounded-lg object-cover"
                  />
                )}
                <h2 className="demo-section-title mb-1">{item.title}</h2>
                <p className="demo-muted mb-1 text-sm">
                  {item.cityName} ({item.zipCode})
                </p>
                <p className="font-bold text-[var(--sea-ink)]">
                  {formatPrice(item.priceInCents)}
                </p>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </main>
  )
}
