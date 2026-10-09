# Betula PipelineRun resource budgets

Both active Betula PipelineRuns set explicit CPU, memory, and ephemeral-storage requests and limits on their inline Go test and privileged Buildah steps. The Go test step requests 500m CPU, 1Gi memory, and 2Gi ephemeral storage, with limits of 2 CPU, 2Gi memory, and 4Gi ephemeral storage. The Buildah step requests 2 CPU, 4Gi memory, and 10Gi ephemeral storage, with limits of 4 CPU, 4Gi memory, and 20Gi ephemeral storage. These are provisional containment budgets: no retained representative Betula Tekton measurements are available, and local Go tests do not measure pod or image-build usage. The heavyweight image build has its own substantial explicit budget and does not rely on a small generic fallback.

PaC embeds the remote `git-clone` Task in both pipelines. Netcup-apps PR #153 adds a named `step-clone` resource default for this resolved task and injected helper defaults, separately from cluster Task budgets. These settings become effective after merge to `codex/tekton-bootstrap` and Operator reconciliation; they do not replace inline workload budgets.

Betula #5, netcup-apps #152 and cap #151 are released; Betula's Repository limit is 1 and cancellation is false. This workload-specific budget change precedes shared defaults #153, which remains pending Operator reconciliation. Supersession cancellation is unsupported together with positive caps in PaC v0.51.0.

## Checks completed for review

- Ruby YAML parsing passed for both PipelineRuns.
- `kubectl create --dry-run=client --validate=false` rendered both PipelineRuns. This was client-side only; it did not call the Kubernetes API or check server schema/admission.
- `git diff --check` passed.
- `go vet ./...` passed.
- `go test ./...` passed for `cmd/server` and `lib`.
- The image build was unavailable and was not run.

The cluster API has recovered and Betula's Repository cap is effective. Automatic publication remains deferred to avoid a build burst; a skipped check is not a passing test. After shared-task/helper reconciliation, exercise representative pull-request and master builds and retain run URLs, peak memory, CPU throttling, storage use, OOM and eviction evidence. Betula-specific runtime acceptance remains pending. Trigger and security behavior is preserved.
