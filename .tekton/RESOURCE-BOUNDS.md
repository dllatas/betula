# Betula PipelineRun resource budgets

Both active Betula PipelineRuns set explicit CPU, memory, and ephemeral-storage requests and limits on their inline Go test and privileged Buildah steps. The Go test step requests 500m CPU, 1Gi memory, and 2Gi ephemeral storage, with limits of 2 CPU, 2Gi memory, and 4Gi ephemeral storage. The Buildah step requests 2 CPU, 4Gi memory, and 10Gi ephemeral storage, with limits of 4 CPU, 8Gi memory, and 20Gi ephemeral storage. These are provisional review budgets: the repository has no retained representative Tekton measurements, and local Go tests do not measure pod or image-build usage. The heavyweight image build has its own substantial explicit budget; it does not rely on a small generic fallback.

Both pipelines also reference the shared `git-clone` Task. Its resources and Tekton helper-container defaults are proposed in netcup-apps PR #153, but are not effective until that PR is merged to `codex/tekton-bootstrap` and ArgoCD reconciles it. These shared limits are separate from, and do not replace, the inline workload budgets in this branch.

Cancellation is managed by separate concurrency changes. Per the required order in netcup-apps PR #151, merge all source annotation PRs including Betula PR #5 and netcup-apps PR #152 first, then merge #151 immediately after that set. This leaves a brief interval with cancellation disabled before positive caps apply; minimize it and avoid build-triggering PR events until #151 is effective. Only merge this resource PR after the Betula cap is effective. This branch carries the #5 annotation as a separate commit, `2681b5586d8c6b3f8bcebd031486bea1180fa474`, sourced from PR #5 commit `d281be8596e3d94e78dd454115089f04f9366f7d`. If #5 merges before this resource PR, rebase this branch and remove the duplicate cherry-pick while retaining the annotation as a separate commit before the resource change. Merge this application-specific budget change before relying on PR #153's generic defaults for these workloads. PR #153's shared clone/helper bounds remain ineffective until merge to `codex/tekton-bootstrap` and ArgoCD reconciliation.

## Checks completed for review

- Ruby YAML parsing passed for both PipelineRuns.
- `kubectl create --dry-run=client --validate=false` rendered both PipelineRuns. This was client-side only; it did not call the Kubernetes API or check server schema/admission.
- `git diff --check` passed.
- `go vet ./...` passed.
- `go test ./...` passed for `cmd/server` and `lib`.
- The image build was unavailable and was not run.

Cluster CI is intentionally deferred for review because Kubernetes API reads are failing. A deferred or skipped CI check is not a passing test. Before merge or rollout acceptance, after API recovery and after shared-task/helper dependencies are effective, exercise each pipeline independently with representative pull-request and master builds. Retain run URLs and pod measurements for CPU throttling, peak memory, ephemeral-storage use, OOMs, and evictions. Revisit these provisional budgets against those observations. No trigger, security, or deployment behavior is changed here.
