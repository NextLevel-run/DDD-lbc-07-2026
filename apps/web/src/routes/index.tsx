import { Link, createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/')({ component: Homepage })

function Homepage() {
  return (
    <main className="page-wrap px-4 pb-8 pt-14">
      <section className="island-shell rise-in relative overflow-hidden rounded-[2rem] px-6 py-10 sm:px-10 sm:py-14">
        <p className="island-kicker mb-3">Second-Hand Marketplace</p>
        <h1 className="display-title mb-5 max-w-3xl text-4xl leading-[1.02] font-bold tracking-tight text-[var(--sea-ink)] sm:text-6xl">
          Buy and sell second-hand items locally.
        </h1>
        <p className="mb-8 max-w-2xl text-base text-[var(--sea-ink-soft)] sm:text-lg">
          Browse classified ads posted by sellers near you, or list your own
          item in a couple of minutes.
        </p>
        <div className="flex flex-wrap gap-3">
          <Link
            to="/classified-ads"
            className="rounded-full border border-[var(--chip-line)] bg-[var(--chip-bg)] px-5 py-2.5 text-sm font-semibold text-[var(--lagoon-deep)] no-underline transition hover:-translate-y-0.5 hover:bg-[var(--link-bg-hover)]"
          >
            Browse Classified Ads
          </Link>
          <Link
            to="/deposit"
            className="rounded-full border border-[var(--line)] bg-[var(--surface)] px-5 py-2.5 text-sm font-semibold text-[var(--sea-ink)] no-underline transition hover:-translate-y-0.5 hover:bg-[var(--link-bg-hover)]"
          >
            Sell an Item
          </Link>
        </div>
      </section>
    </main>
  )
}
