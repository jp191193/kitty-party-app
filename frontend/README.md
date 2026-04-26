# Kitty Party — Frontend

A zero-build, single-page web app for the Go backend in this repo. Just static HTML, CSS,
and ES modules — open it in a browser (via any static file server) and it talks to the
backend at `http://localhost:8080`.

## Run it

The easiest way is any local static server. Examples:

```bash
# 1) Start the Go API
go run ./cmd/api          # listens on :8080

# 2) From a second terminal, serve the frontend on :5173
cd frontend
python -m http.server 5173
# or:  npx serve -p 5173 .
```

Then open `http://localhost:5173` in your browser.

> The browser must reach the API directly. The backend already enables permissive CORS
> in `internal/middleware/middleware.go`, so cross-port requests work in development.

### Pointing at a different backend

Click the gear icon in the top-right and set the API base URL, or run this once in the
browser console:

```js
localStorage.setItem('kp.apiBase', 'https://your-api.example.com');
location.reload();
```

## Acting as a user (auth)

A few endpoints (creating a profile, adding members to a group) require a JWT.

1. Click **Switch** in the header.
2. Pick the member you want to act as.
3. The frontend calls `POST /api/v1/auth/token` and stores the JWT in `localStorage`.

The chip in the header shows the active user; click **Switch → Clear** to forget.

## Structure

```
frontend/
├── index.html              shell with header + nav + main
├── css/
│   ├── styles.css          design tokens, layout, components
│   └── animations.css      subtle transitions only
└── js/
    ├── config.js           API base URL & storage keys
    ├── store.js            tiny pub/sub store for the active user
    ├── api.js              one-stop client wrapping every endpoint
    ├── ui.js               toast/modal/formatters/form helpers
    ├── router.js           hash-based router with params
    ├── app.js              boot — registers routes, wires header
    └── pages/
        ├── dashboard.js
        ├── members.js      list + detail + profile editor
        ├── groups.js       list + detail + memberships
        ├── contributions.js generate dues, record payments
        ├── cycles.js       schedule a multi-month cycle, start/cancel
        └── payouts.js      schedule, disburse, cancel
```

## Adding a new page

1. Create `js/pages/<thing>.js` and export an `async function render…({ root, params, query })`.
2. Inside, set `root.innerHTML = …` (use the `loadingHtml` / `errorHtml` / `emptyHtml`
   helpers from `ui.js`).
3. Register the route in `js/app.js`:

   ```js
   import { renderThing } from './pages/thing.js';
   registerRoute('/things', renderThing);
   registerRoute('/things/:id', renderThing);
   ```

4. Add a nav link in `index.html` (`<a href="#/things" data-route="/things">Things</a>`).

## Adding a new endpoint to the client

Edit `js/api.js` — the file mirrors the backend domains 1:1. Each method returns the
unwrapped `data` payload from the `{success, data}` envelope and throws an `ApiError`
with the backend's `error.message` on failure.

## Notes & gotchas

- `prefers-reduced-motion` disables all animations.
- Dates use `<input type=datetime-local>`. The form helper converts them to ISO with the
  local timezone. Profile DOB uses `type=date` and is sent as `YYYY-MM-DD`.
- All money is rendered in INR with `Intl.NumberFormat`. Change the locale in
  `js/ui.js` (`fmtMoney`) if you need a different currency.
- The store only persists user identity to `localStorage`. Nothing else is cached.
