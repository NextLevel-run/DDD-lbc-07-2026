import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { useAppForm } from '#/hooks/form'
import { deleteClassifiedAd } from '#/features/classified-ad/api'
import { deleteClassifiedAdSchema } from '#/features/classified-ad/schemas'
import { DELETE_REASONS } from '#/features/classified-ad/types'

export const Route = createFileRoute('/classified-ads/$id/delete')({
  component: DeleteClassifiedAd,
})

function DeleteClassifiedAd() {
  const { id } = Route.useParams()
  const navigate = useNavigate()
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [deleted, setDeleted] = useState(false)

  const mutation = useMutation({
    mutationFn: (body: Parameters<typeof deleteClassifiedAd>[1]) =>
      deleteClassifiedAd(id, body),
    onSuccess: () => setDeleted(true),
    onError: (error) =>
      setSubmitError(
        error instanceof Error ? error.message : 'Deletion failed',
      ),
  })

  const form = useAppForm({
    defaultValues: {
      email: '',
      password: '',
      reason: DELETE_REASONS[0],
    },
    onSubmit: async ({ value }) => {
      setSubmitError(null)
      const parsed = deleteClassifiedAdSchema.safeParse(value)
      if (!parsed.success) {
        setSubmitError(parsed.error.issues[0]?.message ?? 'Invalid form')
        return
      }
      await mutation.mutateAsync(parsed.data)
    },
  })

  if (deleted) {
    return (
      <main className="demo-page demo-center">
        <section className="demo-panel w-full max-w-2xl text-center">
          <h1 className="demo-title mb-3">Ad deleted</h1>
          <p className="demo-muted mb-6">
            Your classified ad has been removed and is no longer visible.
          </p>
          <button
            className="demo-button"
            onClick={() => navigate({ to: '/classified-ads' })}
          >
            Back to Listings
          </button>
        </section>
      </main>
    )
  }

  return (
    <main className="demo-page demo-center">
      <section className="demo-panel w-full max-w-2xl">
        <div className="mb-6">
          <p className="island-kicker mb-2">Delete Classified Ad</p>
          <h1 className="demo-title">Confirm deletion</h1>
          <p className="demo-muted mt-2">
            Enter your credentials to confirm you own this ad.
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
          <form.AppField name="email">
            {(field) => <field.TextField label="Your Email" />}
          </form.AppField>

          <form.AppField name="password">
            {(field) => <field.TextField label="Your Password" type="password" />}
          </form.AppField>

          <form.AppField name="reason">
            {(field) => (
              <field.Select
                label="Reason"
                values={DELETE_REASONS.map((reason) => ({
                  label: reason,
                  value: reason,
                }))}
              />
            )}
          </form.AppField>

          {submitError && (
            <div className="demo-alert demo-alert-danger">{submitError}</div>
          )}

          <div className="flex justify-end">
            <form.AppForm>
              <form.SubscribeButton label="Delete Ad" />
            </form.AppForm>
          </div>
        </form>
      </section>
    </main>
  )
}
