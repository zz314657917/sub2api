### PASS: upstream-v027-ui

# UI port result

## Per-item mapping

| Upstream | Verdict | Local evidence |
| --- | --- | --- |
| 406be7c518 | implemented | `stores/subscriptions.ts`: `clear()` resets `loading`. |
| b51f0759d4 | implemented | `stores/payment.ts`: callers share `configPromise`. |
| be313735dc | implemented | `BaseDialog.vue`: module-level dialog ID counter. |
| 017e9e98d1 | implemented | `AdminRefundDialog.vue`: warning compares `form.amount`. |
| fe36f4a914 | implemented | `RegisterView.vue`: initial promo visibility reads injected public settings. |
| f8e5a0d94f | implemented | `ModelTagInput.vue`: empty Tab is not prevented. |
| 0838e0e610 | not applicable | HEAD `d6e6c347` has no `components/admin/user/UserPlatformQuotaModal.vue` and no equivalent admin quota editor (`rg` for all three quota limit fields found only type/display consumers). Adding the absent module is outside a behavior port. |
| 98321a054e | implemented | `AmountInput.vue`: invalid text restores accepted `customText`. |
| ee9ac3e45f | implemented | `TotpSetupModal.vue`: digit inputs bind `:value="code[index]"`. |
| 14636d2f0e | implemented | TOTP setup/disable normalize API errors through `extractApiErrorMessage`. |
| 1a32b91eb2 | implemented | `Pagination.vue`: jump input is coerced through `String`. |
| 130ba634ad | implemented | `useClipboard.ts`: fallback exceptions return false. |
| 6a4938bdfb | implemented | `announcements.ts`: only successful bulk reads update state. |
| 0ed735d3ba | implemented | `ProxySelector.vue`: batch calls reuse `handleTestProxy`. |
| 21532add46 | implemented | `UserOrdersView.vue`: status change calls `handlePageChange(1)`. |

## Executed evidence

- `pnpm.cmd exec vue-tsc -b`: exit 0.
- Adjusted 14-file `pnpm.cmd exec vitest run ...`: final rerun exit 0, 14 files and 24 tests passed. Clipboard cases exercise `document.execCommand` throwing with no Clipboard API and after `navigator.clipboard.writeText` rejects; both assert `false`, `copied=false`, the error toast, and textarea cleanup. Independent QA follow-ups also verify status selection from page 3 resets the orders request to page 1, real `extractApiErrorMessage` normalization for `Error` and legacy Axios detail shapes, and refund balance equal/slightly-over requested amount boundaries.
- `git diff HEAD --check`: exit 0; unmerged path query was empty.
- `pnpm.cmd exec vite build --outDir ../outputs/upstream-v027-ui/dist`: exit 1 before application bundling because the read-only shared dependency `@airwallex/components-sdk` cannot resolve `@airwallex/airtracker`. No dependency or lockfile change was made.
- Browser smoke and dependency recovery are being handled by the controller. This developer result neither starts nor cleans browser processes after that handoff.

## Blockers

The approved test list included the nonexistent quota modal spec, so its original 15-file Vitest command cannot run on this base. The controller approved the documented 14-item adjustment. A prior production build attempt remains blocked by the pre-existing read-only dependency resolution failure; no dependency or lockfile change was made.
