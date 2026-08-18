### BLOCKED: upstream-channel-pricing-s232

Upstream `8f6f45983` is not an independent local slice. Its channel pricing
sync changes assume the local presence of the upstream model-sync and time-
pricing owners (`92ad68a31` and `9f24a5530`), but this checkout has neither
`SyncPricingModels` nor the associated time-pricing product surface.

No product files were changed, no cherry-pick was retained, and no provider,
database, deployment, or remote operation was performed. Reopen only with a
separate approved contract for those prerequisites.
