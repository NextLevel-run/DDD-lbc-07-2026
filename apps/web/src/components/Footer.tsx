export default function Footer() {
  const year = new Date().getFullYear()

  return (
    <footer className="mt-20 border-t border-[var(--line)] px-4 pb-10 pt-6 text-[var(--sea-ink-soft)]">
      <div className="page-wrap text-center text-sm">
        <p className="m-0">&copy; {year} Second-Hand Marketplace</p>
      </div>
    </footer>
  )
}
