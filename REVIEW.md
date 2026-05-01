# PR #3: Add phone layout: vertical split for narrow terminals

## Round 1 — 3 comments
| # | File | Issue | Status | Commit |
|---|------|-------|--------|--------|
| 1 | tui_test.go:594 | TestAppModelStickyNoDoublePrefix tests ChatModel, rename for consistency | ✅ fixed | ac0cfc0 |
| 2 | main.go:122 | --phone flag declared but never wired to AppModel | ✅ fixed | ac0cfc0 |
| 3 | chat.go:241 | /layout command sets working=true, leaves UI stuck in "Thinking" | ✅ fixed | ac0cfc0 |

## Round 2 — 2 issues (user-reported)
| # | File | Issue | Status | Commit |
|---|------|-------|--------|--------|
| 4 | app.go | Session list flickers: SetSessions called every 500ms tick even when unchanged | ✅ fixed | TBD |
| 5 | app.go | WindowSizeMsg forwarded to child models, overwrites recalcSizes per-pane dimensions | ✅ fixed | TBD |

## Round 3 — 3 issues (user-reported)
| # | File | Issue | Status | Commit |
|---|------|-------|--------|--------|
| 6 | app.go | combinedTmuxMsg calls viewer.Update(SessionListMsg) on every output change | ✅ fixed | TBD |
| 7 | app.go | Keyboard close on mobile: no resize, blank screen. Added /resize cmd + Ctrl+R | ✅ fixed | TBD |
| 8 | chat.go | No touch/scroll support in chat. Added MouseModeAllMotion + scrollOffset | ✅ fixed | TBD |
