import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { useAppForm } from '#/hooks/form'
import { submitClassifiedAd } from '#/features/classified-ad/api'
import { submitClassifiedAdSchema } from '#/features/classified-ad/schemas'
import { CATEGORIES } from '#/features/classified-ad/types'

export const Route = createFileRoute('/deposit')({ component: Deposit })

function Deposit() {
  const navigate = useNavigate()
  const [submitError, setSubmitError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: submitClassifiedAd,
    onSuccess: (data) => {
      navigate({ to: '/classified-ads/$id', params: { id: data.id } })
    },
    onError: (error) => {
      setSubmitError(error instanceof Error ? error.message : 'Submission failed')
    },
  })

  const form = useAppForm({
    defaultValues: {
      title: '',
      description: '',
      priceInCents: 0,
      sellerEmail: '',
      sellerPseudo: '',
      sellerPassword: '',
      imageUrls: '',
      category: CATEGORIES[0],
      zipCode: '',
      cityName: '',
    },
    onSubmit: async ({ value }) => {
      setSubmitError(null)
      const imageUrls = value.imageUrls
        .split(',')
        .map((url) => url.trim())
        .filter((url) => url.length > 0)

      const parsed = submitClassifiedAdSchema.safeParse({
        ...value,
        imageUrls,
      })

      if (!parsed.success) {
        setSubmitError(parsed.error.issues[0]?.message ?? 'Invalid form')
        return
      }

      await mutation.mutateAsync(parsed.data)
    },
  })

  return (
    <main className="demo-page demo-center">
      <section className="demo-panel w-full max-w-2xl">
        <div className="mb-6">
          <p className="island-kicker mb-2">Sell an Item</p>
          <h1 className="demo-title">Post a Classified Ad</h1>
          <p className="demo-muted mt-2">
            Fill in the details below. Your ad will be published immediately.
          </p>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            e.stopPropagation()
            form.handleSubmit()
          }}
          className="space-y-6"
        >
          <form.AppField name="title">
            {(field) => <field.TextField label="Title" />}
          </form.AppField>

          <form.AppField name="description">
            {(field) => <field.TextArea label="Description" rows={5} />}
          </form.AppField>

          <form.AppField name="priceInCents">
            {(field) => {
              const value = field.state.value
              return (
                <div>
                  <label
                    htmlFor="priceInCents"
                    className="mb-2 block text-sm font-semibold text-[var(--sea-ink)]"
                  >
                    Price (EUR)
                    <input
                      id="priceInCents"
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

          <form.AppField name="category">
            {(field) => (
              <field.Select
                label="Category"
                values={CATEGORIES.map((category) => ({
                  label: category,
                  value: category,
                }))}
              />
            )}
          </form.AppField>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <form.AppField name="zipCode">
              {(field) => <field.TextField label="Zip Code" />}
            </form.AppField>
            <form.AppField name="cityName">
              {(field) => <field.TextField label="City" />}
            </form.AppField>
          </div>

          <form.AppField name="imageUrls">
            {(field) => (
              <field.TextField
                label="Image URLs (comma-separated)"
                placeholder="https://example.com/photo1.jpg, https://example.com/photo2.jpg"
              />
            )}
          </form.AppField>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <form.AppField name="sellerPseudo">
              {(field) => <field.TextField label="Your Pseudo" />}
            </form.AppField>
            <form.AppField name="sellerEmail">
              {(field) => <field.TextField label="Your Email" />}
            </form.AppField>
          </div>

          <form.AppField name="sellerPassword">
            {(field) => (
              <field.TextField
                label="Password (used to delete or edit your ad later)"
                type="password"
              />
            )}
          </form.AppField>

          {submitError && (
            <div className="demo-alert demo-alert-danger">{submitError}</div>
          )}

          <div className="flex justify-end">
            <form.AppForm>
              <form.SubscribeButton label="Publish Ad" />
            </form.AppForm>
          </div>
        </form>
      </section>
    </main>
  )
}
