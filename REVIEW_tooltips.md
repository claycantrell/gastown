# Code Review: Tooltips for Polecat and Convoy Status

## Review Date: 2026-01-23
## Reviewer: Claude (polecat/slit)

## Summary
✅ **APPROVED** - Implementation matches design with minor notes for future enhancement.

## Does it match the design?

### ✅ Implemented as Designed:
1. **Inline expansion pattern** - Uses ▼/▶ indicators like the convoy TUI
2. **Cursor tracking** - Added `treeCursor` and `convoyCursor` for navigation
3. **Expansion state** - Uses `expandedAgents` and `expandedConvoys` maps
4. **Keyboard bindings** - Enter/Space toggle expansion, j/k for navigation
5. **Detail caching** - Caches fetched details in `agentDetailsCache` and `convoyDetailsCache`
6. **Selection highlighting** - Uses `SelectedStyle` when focused
7. **Truncation** - Long titles are truncated with ellipsis

### ⚠️ Minor Differences (Acceptable):
1. **Key bindings**: Design mentions `i` key, implementation uses `o`/`l` (from existing Expand binding)
   - **Impact**: Low - Still discoverable via help text
   - **Recommendation**: Keep as-is, consistent with existing patterns

2. **Data fetching**: Currently uses cached Agent/Convoy data instead of querying beads
   - **Impact**: Medium - Shows basic info but not all detailed fields (branch, cleanup status, mail)
   - **Recommendation**: Enhance in follow-up work to query agent beads for full runtime details

3. **Tree connectors**: Simplified compared to design diagram
   - **Impact**: Low - Visual clarity is still good
   - **Recommendation**: Can enhance if users request it

## Are there obvious bugs?

### ✅ No Critical Bugs Found

**Checked for:**
- ✅ Nil pointer dereferences - Safe: `getAgentAtCursor()` and `getConvoyAtCursor()` checked before use
- ✅ Off-by-one errors - `maxTreeCursor()` and `maxConvoyCursor()` return `count - 1` correctly
- ✅ Map access safety - Uses two-value assignment (`ok` pattern) where appropriate
- ✅ Bounds checking - Cursor movement bounded by max values

**Minor Issues (Non-blocking):**
1. **Cursor persistence across panel switches** - Not reset when switching panels
   - **Impact**: Low - Cursors stay in position, which could be a feature
   - **Fix**: Add `m.treeCursor = 0` and `m.convoyCursor = 0` when switching panels if desired

2. **Position tracking in renderTree** - `pos` variable tracks agent position
   - **Impact**: None - Works correctly
   - **Note**: Position calculation handles expanded agents correctly

## Is it readable and maintainable?

### ✅ Code Quality is Good

**Strengths:**
1. **Clear function names**: `toggleTreeExpansion()`, `getAgentAtCursor()`, etc.
2. **Logical structure**: Follows existing patterns from convoy TUI
3. **Type safety**: Uses typed structs (`AgentDetails`, `ConvoyDetails`)
4. **Separation of concerns**: Rendering, state management, and data fetching are separate
5. **Comments**: Key sections have explanatory comments

**Suggestions for Future:**
1. **Extract common patterns**: `toggleTreeExpansion()` and `toggleConvoyExpansion()` have similar logic
   - Could create a generic `toggleExpansion(id string, cache map[string]bool, fetchFn func() interface{})` helper
2. **Constants for magic numbers**: `titleMaxLen = 20` could be a const
3. **Documentation**: Add package-level doc comment explaining the expansion feature

## Are there security concerns?

### ✅ No Security Issues

**Checked for:**
- ✅ SQL injection - None (no raw SQL queries)
- ✅ Command injection - None (no external command execution in feed code)
- ✅ XSS/injection in rendering - None (Lipgloss handles escaping)
- ✅ Unbounded memory growth - Caches use maps which could grow, but bounded by agent/convoy count
- ✅ Race conditions - None detected (single-threaded Bubble Tea model)

**Notes:**
- Convoy details fetching (in `convoy.go`) already validates convoy IDs with regex pattern
- Agent/convoy caches are scoped to Model lifetime (GC'd when TUI exits)

## Additional Observations

### ✅ Positive Aspects:
1. **Minimal invasiveness**: Changes are localized to feed TUI, no breaking changes
2. **Progressive enhancement**: Existing functionality works unchanged
3. **Consistent with codebase**: Uses Lipgloss styles, follows Bubble Tea patterns
4. **Testable**: Functions are pure and testable (could add unit tests in follow-up)

### 📝 Future Enhancements (Not Blockers):
1. **Fetch full agent runtime data**: Query beads to get:
   - Branch name
   - Git cleanup status
   - Actual mail count and subjects
   - Hook bead details
2. **Fetch full convoy tracked issues**: Query beads to get:
   - Issue assignees
   - Worker names and ages
   - Last activity timestamps
3. **Add cache expiration**: Invalidate cache after N seconds
4. **Add visual feedback**: Show "loading..." when fetching details
5. **Add tests**: Unit tests for cursor logic, expansion toggling

## Compilation

✅ **Code compiles successfully**
```bash
$ go build ./...
# (no errors)
```

## Verdict

**✅ APPROVED FOR TESTING**

The implementation is solid, follows the design, and has no critical bugs or security issues. The code is readable, maintainable, and integrates well with the existing codebase.

**Recommendation**: Proceed to testing phase. Consider the "Future Enhancements" for a follow-up PR to add full bead querying.

---

## Review Checklist
- [x] Matches design specification
- [x] No obvious bugs
- [x] Code is readable and maintainable
- [x] No security concerns
- [x] Compiles without errors
- [x] Follows existing code patterns
- [x] No breaking changes
