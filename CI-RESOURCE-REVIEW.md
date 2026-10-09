# CI resource continuation

2026-10-09: keep existing lower Go budgets and change the provisional 8Gi
Buildah memory limit to 4Gi. General build nodes have approximately 7.75Gi
allocatable RAM; an 8Gi step limit cannot contain one build before node
exhaustion. CPU, memory requests and ephemeral-storage budgets stay unchanged.

These are interim bounds. Successful representative Amauta Go/frontend and
Buildah runs establish class-level evidence, not Betula-specific sizing.
Keep per-application peak/storage and startup measurements pending. The
skip marker defers an automatic publication burst; it is not a passed test.
