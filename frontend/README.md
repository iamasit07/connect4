# Connect 4 — Frontend

The React-based frontend for the Connect 4 web application.

## Tech Stack

- **React 18** with TypeScript
- **Vite** — fast dev server and bundler
- **Tailwind CSS** + **shadcn/ui** — styling and component library
- **Zustand** — lightweight state management
- **Framer Motion** — animations (piece drops, confetti)
- **TanStack React Query** — server state and caching
- **Axios** — HTTP client

## Development

```bash
npm install
npm run dev
```

The dev server starts at `http://localhost:5173` with hot module replacement.

## Build

```bash
npm run build
```

Outputs a production bundle to `dist/`.

## Project Structure

```
src/
├── components/     # Shared UI components (layout, buttons, etc.)
├── features/       # Feature modules (auth, game)
├── hooks/          # Custom hooks and React Query queries
├── lib/            # Utilities (axios instance, config, helpers)
├── pages/          # Route-level page components
└── stores/         # Zustand stores
```
