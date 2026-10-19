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

export async function request<T>(
  query: string,
  variables?: Record<string, unknown>,
): Promise<T> {
  const token = localStorage.getItem('microstack_token')
  const response = await fetch('/graphql', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ query, variables }),
  })
  const payload = (await response.json()) as GraphQLResponse<T>
  if (!response.ok || payload.errors?.length) {
    throw new Error(payload.errors?.map((error) => error.message).join(', ') ?? 'Request failed')
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
