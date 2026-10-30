/// <reference types="vite/client" />

export type Product = {
  id: string
  name: string
  description: string
  price: number
}

export type OrderProduct = Product & { quantity: number }

export type Order = {
  id: string
  createdAt: string
  totalPrice: number
  products: OrderProduct[]
}

export type Account = {
  id: string
  name: string
  role: 'CUSTOMER' | 'OPERATOR'
  orders: Order[]
}

type GraphQLResponse<T> = { data?: T; errors?: { message: string }[] }

const graphqlEndpoint = import.meta.env.VITE_GRAPHQL_URL || '/graphql'

export function isAuthenticationError(reason: unknown): boolean {
  if (!(reason instanceof Error)) return false
  return /(?:401|403|unauthenticated|unauthorized|authentication|invalid token|token expired|jwt)/i.test(reason.message)
}

export async function request<T>(
  query: string,
  variables?: Record<string, unknown>,
): Promise<T> {
  const token = localStorage.getItem('microstack_token')
  const response = await fetch(graphqlEndpoint, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ query, variables }),
  })

  const responseBody = await response.text()
  let payload: GraphQLResponse<T>
  try {
    payload = responseBody ? JSON.parse(responseBody) as GraphQLResponse<T> : {}
  } catch {
    if (!response.ok) {
      throw new Error(`The service is unavailable (${response.status}). Please try again.`)
    }
    throw new Error('The service returned an invalid response. Please try again.')
  }

  if (!response.ok || payload.errors?.length) {
    if (response.status === 401 || response.status === 403) {
      throw new Error(`${response.status} Unauthorized`)
    }
    throw new Error(payload.errors?.map((error) => error.message).join(', ') ?? `The service is unavailable (${response.status}). Please try again.`)
  }
  if (!payload.data) throw new Error('The API returned no data')
  return payload.data
}

export const accountsQuery = `query Accounts {
  accounts {
    id name role
    orders {
      id createdAt totalPrice
      products { id name description price quantity }
    }
  }
}`

export const productsQuery = `query Products {
  products { id name description price }
}`

export const viewerQuery = `query Viewer {
  viewer { id name role }
}`
