import { z } from 'zod'
import { CATEGORIES, DELETE_REASONS } from './types'

// Mirrors the Go domain's email regex (internal/classified-ad/domain/email.go).
const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const ZIP_CODE_REGEX = /^\d{5}$/

const email = z.string().regex(EMAIL_REGEX, 'Invalid email address')

export const submitClassifiedAdSchema = z.object({
  title: z.string().min(1, 'Title is required').max(100, 'Title must not exceed 100 characters'),
  description: z
    .string()
    .min(1, 'Description is required')
    .max(4000, 'Description must not exceed 4000 characters'),
  priceInCents: z.number().min(0, 'Price must not be negative'),
  sellerEmail: email,
  sellerPseudo: z
    .string()
    .min(1, 'Pseudo is required')
    .max(30, 'Pseudo must not exceed 30 characters'),
  sellerPassword: z.string().min(8, 'Password must be at least 8 characters long'),
  imageUrls: z.array(z.string().min(1, 'Image url must not be empty')).max(10, 'A classified ad cannot have more than 10 images'),
  category: z.enum(CATEGORIES),
  zipCode: z.string().regex(ZIP_CODE_REGEX, 'Zip code must be exactly 5 digits'),
  cityName: z.string().min(1, 'City name is required'),
})

export type SubmitClassifiedAdFormValues = z.infer<
  typeof submitClassifiedAdSchema
>

export const makeOfferSchema = z.object({
  buyerEmail: email,
  buyerPseudo: z.string().min(1, 'Pseudo is required'),
  amountInCents: z.number().min(0, 'Offer amount must not be negative'),
  message: z
    .string()
    .min(1, 'Offer message must not be empty')
    .max(1000, 'Offer message must not exceed 1000 characters'),
})

export type MakeOfferFormValues = z.infer<typeof makeOfferSchema>

export const deleteClassifiedAdSchema = z.object({
  email,
  password: z.string().min(1, 'Password is required'),
  reason: z.enum(DELETE_REASONS),
})

export type DeleteClassifiedAdFormValues = z.infer<
  typeof deleteClassifiedAdSchema
>
