# Web design guide

- Use the dark, neutral theme from `internal/web/styles/app.css`. Keep contrast clear and use the primary color for the main action and progress.
- Use the shared `page-container` (up to 100rem) on standard pages. Keep individual forms at a readable width inside it.
- Give pages room to breathe with consistent container padding. Avoid cards around whole forms or setup steps.
- Use shadcn-templ components from `internal/web/components` for controls and feedback. Keep labels visible, focus states clear, and errors next to the task they affect.
- For onboarding, fix a single progress bar across the viewport top and keep Back and Continue actions predictable.
- Let the first setup step span the page: welcome content takes most of the desktop width, with the form on the right; stack them on smaller screens. Show password strength and require confirmation.
- Write direct copy. Mark unavailable integrations as optional and explain that no action is needed.
- Keep browser handlers in `internal/web/handlers`, browser request gates in `internal/web/middleware`, and route composition in `internal/web/routes.go`. Templates belong in `internal/web/pages`.
