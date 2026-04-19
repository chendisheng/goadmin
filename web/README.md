# GoAdmin Web

Front-end shell for GoAdmin Phase 10.

## Tech Stack

- Vue 3
- TypeScript
- Vite
- Pinia
- Vue Router
- Axios
- Element Plus

## Scripts

```bash
npm install
npm run dev
npm run build
npm run preview
npm run test
```

## Testing Notes

- `npm run test` runs the Vitest regression suite for the multi-tab store and restore flow.
- The tabs coverage focuses on route syncing, cache-name generation, persistence restore, and fixed/public/404 route handling.

## Environment

Copy `.env.example` to `.env` and adjust API settings if needed.
