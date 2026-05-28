# Prompt: add-templui-component

Add a templUI component in go-pocket:

1. Create `components/ui/<name>/`.
2. Add `<name>.templ`.
3. Optionally add `<name>.go` props and `<name>.min.js` script.
4. Keep styles as Tailwind utility classes.
5. Run `task templ:generate`.

Do not edit generated `*_templ.go` files by hand.

