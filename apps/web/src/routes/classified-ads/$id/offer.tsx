import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { useAppForm } from '#/hooks/form'
import { makeOffer } from '#/features/classified-ad/api'
import { makeOfferSchema } from '#/features/classified-ad/schemas'

export const Route = createFileRoute('/classified-ads/$id/offer')({
  component: OfferForm,
})

function OfferForm() {
  const { id } = Route.useParams()
  const navigate = useNavigate()
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [submitted, setSubmitted] = useState(false)

  const mutation = useMutation({
    mutationFn: (body: Parameters<typeof makeOffer>[1]) =>
      makeOffer(id, body),
    onSuccess: () => setSubmitted(true),
    onError: (error) =>
      setSubmitError(error instanceof Error ? error.message : 'Offer failed'),
  })

  const form = useAppForm({
    defaultValues: {
      buyerEmail: '',
      buyerPseudo: '',
      amountInCents: 0,
      message: '',
    },
    onSubmit: async ({ value }) => {
      setSubmitError(null)
      const parsed = makeOfferSchema.safeParse(value)
      if (!parsed.success) {
        setSubmitError(parsed.error.issues[0]?.message ?? 'Invalid form')
        return
      }
      await mutation.mutateAsync(parsed.data)
    },
  })

  if (submitted) {
    return (
      <main className="demo-page demo-center">
        <section className="demo-panel w-full max-w-2xl text-center">
          <h1 className="demo-title mb-3">Offer sent!</h1>
          <p className="demo-muted mb-6">
            The seller has been notified of your offer.
          </p>
          <button
            className="demo-button"
            onClick={() =>
              navigate({ to: '/classified-ads/$id', params: { id } })
            }
          >
            Back to the Ad
          </button>
        </section>
      </main>
    )
  }

  return (
    <main className="demo-page demo-center">
      <section className="demo-panel w-full max-w-2xl">
        <div className="mb-6">
          <p className="island-kicker mb-2">Make an Offer</p>
          <h1 className="demo-title">Propose your price</h1>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            e.stopPropagation()
            form.handleSubmit()
          }}
          className="space-y-6"
        >
          <form.AppField name="amountInCents">
            {(field) => {
              const value = field.state.value
              return (
                <div>
                  <label
                    htmlFor="amountInCents"
                    className="mb-2 block text-sm font-semibold text-[var(--sea-ink)]"
                  >
                    Your offer (EUR)
                    <input
                      id="amountInCents"
                      type="number"
                      min={0}
                      step="0.01"
                      value={value === 0 ? '' : value / 100}
                      onBlur={field.handleBlur}
                      onChange={(e) =>
                        field.handleChange(
                          Math.round(Number(e.target.value || 0) * 100),
                        )
                      }
                      className="demo-input mt-2"
                    />
                  </label>
                </div>
              )
            }}
          </form.AppField>

          <form.AppField name="message">
            {(field) => <field.TextArea label="Message to the seller" />}
          </form.AppField>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <form.AppField name="buyerPseudo">
              {(field) => <field.TextField label="Your Pseudo" />}
            </form.AppField>
            <form.AppField name="buyerEmail">
              {(field) => <field.TextField label="Your Email" />}
            </form.AppField>
          </div>

          {submitError && (
            <div className="demo-alert demo-alert-danger">{submitError}</div>
          )}

          <div className="flex justify-end">
            <form.AppForm>
              <form.SubscribeButton label="Send Offer" />
            </form.AppForm>
          </div>
        </form>
      </section>
    </main>
  )
}
