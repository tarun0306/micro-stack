import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { Activity, ArrowUpRight, Boxes, CirclePlus, Package, RefreshCw, ShoppingBag, Users } from 'lucide-react'
import { accountsQuery, isAuthenticationError, productsQuery, request, viewerQuery, type Account, type Product } from './api'

type View = 'overview' | 'orders' | 'catalog' | 'accounts'

function Landing() {
  return <main className="landing-shell">
    <nav className="landing-nav">
      <a className="brand" href="/"><span className="brand-mark">M</span><span>micro<span>stack</span></span></a>
      <div className="landing-nav-links"><a href="#how-it-works">How it works</a><a href="#workspaces">Workspaces</a><a className="landing-sign-in" href="/login">Sign in <ArrowUpRight size={15} /></a></div>
    </nav>
    <section className="landing-hero">
      <div className="landing-hero-copy"><span className="tag">COMMERCE, WITHOUT THE CLUTTER</span><h1>Useful things, thoughtfully assembled.</h1><p>Discover a focused collection of products, build your cart, and place your order in a calm storefront designed around your experience.</p><div className="landing-actions"><a className="primary" href="/shop">Browse the store <ArrowUpRight size={16} /></a><a className="landing-secondary" href="/login">Sign in</a></div><div className="landing-proof"><span><strong>01</strong> Simple product discovery</span><span><strong>02</strong> Secure checkout</span></div></div>
      <div className="landing-visual"><div className="landing-glow" /><div className="landing-window"><div className="landing-window-top"><span /><span /><span /><small>microstack / store</small></div><div className="landing-window-body"><p className="eyebrow">Your selection</p><strong>Everything in view.</strong><div className="landing-metrics"><div><small>Products</small><b>24</b><em>Curated for you</em></div><div><small>Cart total</small><b>$84</b><em>Ready when you are</em></div></div><div className="landing-chart"><i /><i /><i /><i /><i /><i /><i /></div><div className="landing-activity"><span /><div><b>Ready to order</b><small>Secure checkout · Your account</small></div><strong>→</strong></div></div></div></div>
    </section>
    <section className="landing-section" id="workspaces"><div className="landing-section-heading"><span className="tag">A BETTER WAY TO SHOP</span><h2>Everything you need, without the noise.</h2></div><div className="landing-cards"><article><div className="landing-card-number">01</div><h3>Explore</h3><p>Browse a curated catalog with clear descriptions and straightforward pricing.</p><a href="/shop">View the catalog <ArrowUpRight size={15} /></a></article><article><div className="landing-card-number">02</div><h3>Choose</h3><p>Add products to your cart and adjust quantities until your order feels right.</p><a href="/shop">Build your cart <ArrowUpRight size={15} /></a></article><article><div className="landing-card-number">03</div><h3>Order securely</h3><p>Sign in only when you are ready to place an order and keep your purchases connected to your account.</p><a href="/login?returnTo=/shop">Create an account <ArrowUpRight size={15} /></a></article></div></section>
    <section className="landing-process" id="how-it-works"><div><span className="tag">A SIMPLE FLOW</span><h2>From first click to confirmed order.</h2></div><div className="landing-process-steps"><div><b>01</b><span>Explore</span><p>Find products that fit your needs.</p></div><div><b>02</b><span>Choose</span><p>Build your cart at your own pace.</p></div><div><b>03</b><span>Order</span><p>Sign in and check out securely.</p></div></div></section>
    <footer className="landing-footer"><a className="brand" href="/"><span className="brand-mark">M</span><span>micro<span>stack</span></span></a><span>Commerce, made clear.</span><a href="/shop">Start shopping <ArrowUpRight size={15} /></a></footer>
  </main>
}

function Login({ onLogin }: { onLogin: (destination: string) => void }) {
  const [registering, setRegistering] = useState(false)
  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function submit(event: React.FormEvent) {
    event.preventDefault()
    setLoading(true)
    setError('')
    try {
      if (registering && password !== confirmPassword) {
        throw new Error('Passwords do not match')
      }
      if (registering) {
        await request<{ createAccount: { id: string } }>(
          `mutation CreateAccount($account: AccountInput!) { createAccount(account: $account) { id } }`,
          { account: { name, password } },
        )
      }
      const data = await request<{ login: { token: string } }>(
        `mutation Login($credentials: LoginInput!) { login(credentials: $credentials) { token } }`,
        { credentials: { name, password } },
      )
      localStorage.setItem('microstack_token', data.login.token)
      const returnTo = new URLSearchParams(window.location.search).get('returnTo')
      if (returnTo?.startsWith('/') && !returnTo.startsWith('//')) {
        window.location.replace(returnTo)
        return
      }
      const { viewer } = await request<{ viewer: { role: Account['role'] } }>(viewerQuery)
      onLogin(viewer.role === 'OPERATOR' ? '/console' : '/shop')
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Could not sign in')
    } finally {
      setLoading(false)
    }
  }

  return <main className="login-shell"><form className="panel login-panel" onSubmit={submit}><div className="brand"><span className="brand-mark">M</span><span>micro<span>stack</span></span></div><p className="eyebrow">Customer workspace</p><h1>{registering ? 'Create your account' : 'Sign in'}</h1><p>{registering ? 'Register to browse the catalog and place orders.' : 'Use your account name and password to continue.'}</p><label>Account name<input value={name} onChange={(event) => setName(event.target.value)} autoComplete="username" required /></label><label>Password<input type="password" value={password} onChange={(event) => setPassword(event.target.value)} autoComplete={registering ? 'new-password' : 'current-password'} required /></label>{registering && <label>Confirm password<input type="password" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} autoComplete="new-password" required /></label>}{error && <div className="alert">{error}</div>}<button className="primary full" disabled={loading}>{loading ? (registering ? 'Creating account…' : 'Signing in…') : (registering ? 'Create account' : 'Sign in')}</button><button type="button" className="text-button auth-switch" onClick={() => { setRegistering((current) => !current); setError(''); setConfirmPassword('') }}>{registering ? 'Already have an account? Sign in' : 'New customer? Create an account'}</button></form></main>
}

function App() {
  const [authenticated, setAuthenticated] = useState(() => Boolean(localStorage.getItem('microstack_token')))
  const [view, setView] = useState<View>('overview')
  const [accounts, setAccounts] = useState<Account[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [selectedAccount, setSelectedAccount] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [refreshKey, setRefreshKey] = useState(0)
  const [checkingAccess, setCheckingAccess] = useState(() => (
    Boolean(localStorage.getItem('microstack_token')) &&
    (window.location.pathname === '/login' || window.location.pathname === '/console')
  ))

  useEffect(() => {
    const pathname = window.location.pathname
    if (!authenticated || (pathname !== '/console' && pathname !== '/login')) return
    let active = true
    request<{ viewer: Account }>(viewerQuery)
      .then((viewerData) => {
        if (!active) return
        if (pathname === '/login' || viewerData.viewer.role === 'CUSTOMER') {
          window.location.replace(viewerData.viewer.role === 'OPERATOR' ? '/console' : '/shop')
          return
        }
        setCheckingAccess(false)
        setLoading(true)
        return Promise.all([
          request<{ accounts: Account[] }>(accountsQuery),
          request<{ products: Product[] }>(productsQuery),
        ])
      })
      .then((data) => {
        if (!active || !data) return
        const [accountData, productData] = data
        setAccounts(accountData.accounts)
        setProducts(productData.products)
        setSelectedAccount((current) => current || accountData.accounts.find((account) => account.role === 'CUSTOMER')?.id || accountData.accounts[0]?.id || '')
        setError('')
      })
      .catch((reason: Error) => {
        if (!active) return
        if (isAuthenticationError(reason)) {
          localStorage.removeItem('microstack_token')
          setAuthenticated(false)
          window.location.replace(`/login?returnTo=${encodeURIComponent(pathname)}`)
          return
        }
        setCheckingAccess(false)
        setError(reason.message)
      })
      .finally(() => active && setLoading(false))
    return () => {
      active = false
    }

  }, [authenticated, refreshKey])

  const orderCount = useMemo(() => accounts.reduce((sum, account) => sum + account.orders.length, 0), [accounts])
  const revenue = useMemo(
    () => accounts.reduce((sum, account) => sum + account.orders.reduce((orders, order) => orders + order.totalPrice, 0), 0),
    [accounts],
  )

  if (window.location.pathname === '/') return <Landing />
  if (window.location.pathname === '/shop') return <CustomerShop />
  if (checkingAccess) return <div className="loading"><RefreshCw className="spin" size={22} /> Checking access…</div>
  if (!authenticated && window.location.pathname !== '/login') return <Login onLogin={(destination) => window.location.replace(destination)} />
  if (!authenticated) return <Login onLogin={(destination) => window.location.replace(destination)} />

    function CustomerShop() {
      const [accounts, setAccounts] = useState<Account[]>([])
      const [products, setProducts] = useState<Product[]>([])
      const [accountId, setAccountId] = useState('')
      const [cart, setCart] = useState<Record<string, number>>({})
      const [loading, setLoading] = useState(true)
      const [placingOrder, setPlacingOrder] = useState(false)
      const [message, setMessage] = useState('')
      const [error, setError] = useState('')

      useEffect(() => {
        request<{ products: Product[] }>(productsQuery)
          .then((productData) => {
            setProducts(productData.products)
            const token = localStorage.getItem('microstack_token')
            if (!token) return
            return request<{ viewer: Account }>(viewerQuery)
              .then(({ viewer }) => {
                setAccounts([viewer])
                setAccountId(viewer.id)
              })
              .catch((reason: unknown) => {
                if (isAuthenticationError(reason)) {
                  localStorage.removeItem('microstack_token')
                  setAccounts([])
                  setAccountId('')
                } else {
                  setError(reason instanceof Error ? reason.message : 'Could not load your account')
                }
              })
          })
          .catch((reason: Error) => setError(reason.message))
          .finally(() => setLoading(false))
      }, [])

      const cartItems = products.filter((product) => cart[product.id]).map((product) => ({
        product,
        quantity: cart[product.id],
      }))
      const total = cartItems.reduce((sum, item) => sum + item.product.price * item.quantity, 0)

      function addToCart(id: string) {
        setCart((current) => ({ ...current, [id]: (current[id] || 0) + 1 }))
        setMessage('')
      }

      function changeQuantity(id: string, quantity: number) {
        setCart((current) => {
          const next = { ...current }
          if (quantity <= 0) delete next[id]
          else next[id] = quantity
          return next
        })
      }

      async function placeOrder() {
        if (placingOrder) return
        if (!localStorage.getItem('microstack_token')) {
          window.location.href = '/login?returnTo=/shop'
          return
        }
        if (!accountId || !cartItems.length) return
        setPlacingOrder(true)
        try {
          await request(`mutation CreateOrder($order: OrderInput!) { createOrder(order: $order) { id } }`, {
            order: {
              accountId,
              products: cartItems.map(({ product, quantity }) => ({ id: product.id, quantity })),
            },
          })
          setCart({})
          setMessage('Thanks — your order has been placed.')
        } catch (reason) {
          setError(reason instanceof Error ? reason.message : 'Could not place order')
        } finally {
          setPlacingOrder(false)
        }
      }

      if (loading) return <div className="shop-loading"><RefreshCw className="spin" size={22} /> Loading storefront…</div>

      return <div className="shop-shell">
        <header className="shop-header">
          <a className="brand shop-brand" href="/"><span className="brand-mark">M</span><span>micro<span>stack</span></span></a>
          <div className="shop-customer">{accounts[0] ? <span>Welcome, {accounts[0].name}</span> : <a href="/login?returnTo=/shop">Sign in to order</a>}</div>
        </header>
        <main className="shop-content">
          <div className="shop-intro"><div><span className="tag">MICROSTACK STORE</span><h1>Useful things, thoughtfully assembled.</h1><p>Browse the catalog and build your next order.</p></div><div className="shop-badge"><ShoppingBag size={18} /><span>{cartItems.length} item types</span></div></div>
          {error && <div className="alert">{error}</div>}
          {message && <div className="shop-success">{message}</div>}
          <div className="shop-layout">
            <section><div className="shop-section-title"><p className="eyebrow">The catalog</p><h2>Explore products</h2></div><div className="shop-product-grid">{products.map((product) => <article className="shop-product" key={product.id}><div className="shop-product-art"><Package size={25} /></div><div className="shop-product-info"><div><h3>{product.name}</h3><p>{product.description || 'A carefully selected product for your workflow.'}</p></div><strong>${product.price.toFixed(2)}</strong></div><button className="primary full" onClick={() => addToCart(product.id)}>Add to cart</button></article>)}</div></section>
            <aside className="cart-panel"><div className="panel-title"><div><p className="eyebrow">Your selection</p><h3>Cart</h3></div><ShoppingBag size={20} /></div>{cartItems.length ? <><div className="cart-items">{cartItems.map(({ product, quantity }) => <div className="cart-item" key={product.id}><div><strong>{product.name}</strong><span>${product.price.toFixed(2)} each</span></div><div className="quantity"><button onClick={() => changeQuantity(product.id, quantity - 1)} disabled={placingOrder}>−</button><span>{quantity}</span><button onClick={() => changeQuantity(product.id, quantity + 1)} disabled={placingOrder}>+</button></div></div>)}</div><div className="cart-total"><span>Total</span><strong>${total.toFixed(2)}</strong></div><button className="primary full" onClick={placeOrder} disabled={placingOrder}>{placingOrder ? 'Placing order…' : 'Place order'}</button></> : <Empty title="Your cart is empty" text="Add a product to get started." />}</aside>
          </div>
        </main>
      </div>
    }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand"><span className="brand-mark">M</span><span>micro<span>stack</span></span></div>
        <p className="eyebrow">Operations console</p>
        <nav>
          <NavItem icon={<Activity size={17} />} label="Overview" active={view === 'overview'} onClick={() => setView('overview')} />
          <NavItem icon={<ShoppingBag size={17} />} label="Orders" active={view === 'orders'} onClick={() => setView('orders')} count={orderCount} />
          <NavItem icon={<Boxes size={17} />} label="Catalog" active={view === 'catalog'} onClick={() => setView('catalog')} count={products.length} />
          <NavItem icon={<Users size={17} />} label="Accounts" active={view === 'accounts'} onClick={() => setView('accounts')} count={accounts.length} />
        </nav>
        <div className="sidebar-foot"><span className="status-dot" />Live data connected</div>
      </aside>

      <main className="content">
        <header className="topbar">
          <div><p className="eyebrow">{new Intl.DateTimeFormat(undefined, { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' }).format(new Date())}</p><h1>{view === 'overview' ? 'Good morning, operator.' : view[0].toUpperCase() + view.slice(1)}</h1></div>
          <button className="icon-button" onClick={() => setRefreshKey((key) => key + 1)} aria-label="Refresh data"><RefreshCw size={18} /></button>
        </header>
        {error && <div className="alert">{error}</div>}
        {loading ? <div className="loading"><RefreshCw className="spin" size={22} /> Loading service data…</div> : (
          <>
            {view === 'overview' && <Overview accounts={accounts} products={products} orderCount={orderCount} revenue={revenue} onNavigate={setView} />}
            {view === 'orders' && <Orders accounts={accounts} products={products} selectedAccount={selectedAccount} setSelectedAccount={setSelectedAccount} onCreated={() => setRefreshKey((key) => key + 1)} />}
            {view === 'catalog' && <Catalog products={products} onCreated={() => setRefreshKey((key) => key + 1)} />}
            {view === 'accounts' && <Accounts accounts={accounts} />}
          </>
        )}
      </main>
    </div>
  )
}

function NavItem({ icon, label, active, onClick, count }: { icon: ReactNode; label: string; active: boolean; onClick: () => void; count?: number }) {
  return <button className={`nav-item ${active ? 'active' : ''}`} onClick={onClick}>{icon}<span>{label}</span>{count !== undefined && <b>{count}</b>}</button>
}

function Overview({ accounts, products, orderCount, revenue, onNavigate }: { accounts: Account[]; products: Product[]; orderCount: number; revenue: number; onNavigate: (view: View) => void }) {
  const latestOrders = accounts.flatMap((account) => account.orders.map((order) => ({ ...order, accountName: account.name }))).slice(-4).reverse()
  return <section>
    <div className="hero"><div><span className="tag">LIVE GRAPHQL WORKSPACE</span><h2>Your services, in one calm place.</h2><p>Track customers, inventory, and order activity across the microservice stack.</p></div><button className="primary" onClick={() => onNavigate('orders')}><CirclePlus size={17} /> Create an order</button></div>
    <div className="stats">
      <Stat icon={<Users />} label="Accounts" value={accounts.length} detail="registered customers" />
      <Stat icon={<Package />} label="Catalog items" value={products.length} detail="products available" />
      <Stat icon={<ShoppingBag />} label="Orders" value={orderCount} detail="across all accounts" />
      <Stat icon={<ArrowUpRight />} label="Gross value" value={`$${revenue.toFixed(2)}`} detail="order total" />
    </div>
    <div className="section-heading"><div><p className="eyebrow">Recent activity</p><h3>Latest orders</h3></div><button className="text-button" onClick={() => onNavigate('orders')}>View all <ArrowUpRight size={15} /></button></div>
    <div className="order-list">{latestOrders.length ? latestOrders.map((order) => <OrderRow key={order.id} order={order} accountName={order.accountName} />) : <Empty title="No orders yet" text="Create your first order to see activity here." />}</div>
  </section>
}

function Stat({ icon, label, value, detail }: { icon: ReactNode; label: string; value: string | number; detail: string }) {
  return <div className="stat-card"><div className="stat-icon">{icon}</div><p>{label}</p><strong>{value}</strong><small>{detail}</small></div>
}

function Orders({ accounts, products, selectedAccount, setSelectedAccount, onCreated }: { accounts: Account[]; products: Product[]; selectedAccount: string; setSelectedAccount: (id: string) => void; onCreated: () => void }) {
  const [quantity, setQuantity] = useState(1)
  const [productId, setProductId] = useState(products[0]?.id || '')
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')
  const allOrders = accounts.flatMap((account) => account.orders.map((order) => ({ ...order, accountName: account.name }))).reverse()
  const product = products.find((item) => item.id === productId)
  async function createOrder() {
    if (!selectedAccount || !productId) return
    setSaving(true); setMessage('')
    try {
      await request(`mutation CreateOrder($order: OrderInput!) { createOrder(order: $order) { id } }`, { order: { accountId: selectedAccount, products: [{ id: productId, quantity }] } })
      setMessage('Order created successfully'); onCreated()
    } catch (reason) { setMessage(reason instanceof Error ? reason.message : 'Could not create order') } finally { setSaving(false) }
  }
  return <section><div className="section-heading"><div><p className="eyebrow">Commerce flow</p><h2>Orders</h2></div></div>
    <div className="split-layout"><div className="panel"><div className="panel-title"><div><p className="eyebrow">New order</p><h3>Assemble a purchase</h3></div><ShoppingBag size={20} /></div><label>Account<select value={selectedAccount} onChange={(event) => setSelectedAccount(event.target.value)}>{accounts.map((account) => <option key={account.id} value={account.id}>{account.name} · {account.role.toLowerCase()}</option>)}</select></label><label>Product<select value={productId} onChange={(event) => setProductId(event.target.value)}>{products.map((item) => <option key={item.id} value={item.id}>{item.name} · ${item.price.toFixed(2)}</option>)}</select></label><label>Quantity<input type="number" min="1" value={quantity} onChange={(event) => setQuantity(Math.max(1, Number(event.target.value)))} /></label><div className="total-line"><span>Estimated total</span><strong>${((product?.price || 0) * quantity).toFixed(2)}</strong></div><button className="primary full" disabled={saving || !selectedAccount || !productId} onClick={createOrder}>{saving ? 'Creating…' : 'Create order'}</button>{message && <p className="form-message">{message}</p>}</div><div><p className="eyebrow">Order history</p><h3 className="history-title">Recent purchases</h3><div className="order-list">{allOrders.length ? allOrders.map((order) => <OrderRow key={order.id} order={order} accountName={order.accountName} />) : <Empty title="No purchases yet" text="Your order history will appear here." />}</div></div></div>
  </section>
}

function OrderRow({ order, accountName }: { order: { id: string; createdAt: string; totalPrice: number; products: { name: string; quantity: number }[] }; accountName: string }) {
  const createdAt = new Date(order.createdAt)
  const createdLabel = Number.isNaN(createdAt.getTime())
    ? 'Creation time unavailable'
    : `Created ${createdAt.toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })}`
  return <div className="order-row"><div className="order-avatar">{accountName.slice(0, 1).toUpperCase()}</div><div className="order-main"><strong>{accountName}</strong><span>{order.products.map((product) => `${product.name} × ${product.quantity}`).join(', ')}</span><small>{createdLabel}</small></div><div className="order-total">${order.totalPrice.toFixed(2)}</div></div>
}

function Catalog({ products, onCreated }: { products: Product[]; onCreated: () => void }) {
  const [form, setForm] = useState({ name: '', description: '', price: '' }); const [message, setMessage] = useState('')
  async function createProduct() { try { await request(`mutation CreateProduct($product: ProductInput!) { createProduct(product: $product) { id } }`, { product: { ...form, price: Number(form.price) } }); setForm({ name: '', description: '', price: '' }); setMessage('Product added to catalog'); onCreated() } catch (reason) { setMessage(reason instanceof Error ? reason.message : 'Could not create product') } }
  return <section><div className="section-heading"><div><p className="eyebrow">Inventory</p><h2>Catalog</h2></div></div><div className="panel product-form"><div className="panel-title"><div><p className="eyebrow">Add a product</p><h3>Expand your catalog</h3></div><Package size={20} /></div><div className="form-grid"><label>Name<input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="e.g. Starter kit" /></label><label>Price<input type="number" min="0" step="0.01" value={form.price} onChange={(e) => setForm({ ...form, price: e.target.value })} placeholder="0.00" /></label></div><label>Description<textarea value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder="What makes this product useful?" /></label><button className="primary" disabled={!form.name || !form.price} onClick={createProduct}>Add product</button>{message && <p className="form-message">{message}</p>}</div><div className="product-grid">{products.map((product) => <div className="product-card" key={product.id}><div className="product-art"><Package size={22} /></div><div><span className="product-price">${product.price.toFixed(2)}</span><h3>{product.name}</h3><p>{product.description || 'No description provided.'}</p></div></div>)}</div></section>
}

function Accounts({ accounts }: { accounts: Account[] }) {
  return <section><div className="section-heading"><div><p className="eyebrow">Customer base</p><h2>Accounts</h2></div></div><div className="account-grid">{accounts.map((account) => <div className="account-card" key={account.id}><div className="account-avatar">{account.name.slice(0, 1).toUpperCase()}</div><div className="account-card-heading"><h3>{account.name}</h3><span className={`role-badge ${account.role.toLowerCase()}`}>{account.role.toLowerCase()}</span></div><p>{account.orders.length} order{account.orders.length === 1 ? '' : 's'}</p><code>{account.id}</code></div>)}</div></section>
}

function Empty({ title, text }: { title: string; text: string }) { return <div className="empty"><Boxes size={24} /><strong>{title}</strong><span>{text}</span></div> }

export default App
