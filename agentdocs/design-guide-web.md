# Web design guide

- Use the dark, neutral theme from `internal/web/styles/app.css`. Keep contrast clear and use the primary color for the main action and progress.
- Follow the auth pages: open background, simple headings, short supporting text, and a narrow content column (`max-w-md`) for forms.
- Give pages room to breathe with consistent container padding. Avoid cards around whole forms or setup steps.
- Use shadcn-templ components from `internal/web/components` for controls and feedback. Keep labels visible, focus states clear, and errors next to the task they affect.
- For onboarding, show one progress bar near the top with a plain “Step X of 3” label. Show one step at a time; keep Back and Continue actions predictable.
- Write direct copy. Mark unavailable integrations as optional and explain that no action is needed.
- Keep browser handlers in `internal/web/handlers`, browser request gates in `internal/web/middleware`, and route composition in `internal/web/routes.go`. Templates belong in `internal/web/pages`.
