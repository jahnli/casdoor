# Code style

- Do not add Go tests (`_test.go`) unless explicitly asked.
- Keep comments sparse: only for genuinely non-obvious mechanics, not one per block.

# Frontend (`web`)

`web` is the only frontend, built with shadcn/Tailwind. The legacy Ant Design
frontend (`web-old`) has been removed; do not restore it.

Preserve the migrated validations, field formats, and conditions unless a
behavior change is explicitly requested. New features and fixes go into `web`.

Not migrated on purpose:

- **AI assistant.** The `aiAssistantUrl` config, its header button, the iframe
  drawer and the `ai-assistant` entry of `WidgetItemTree` are dropped. Do not
  restore them.
- **Web3 providers.** The `Web3` category (MetaMask, Web3-Onboard) was removed
  from the backend and `web`. Do not restore the legacy wallet sign-in code.
