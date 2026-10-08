# Betula PipelineRun resource budgets

Both Betula PipelineRuns now set explicit CPU, memory, and ephemeral-storage requests and limits on the Go test and privileged Buildah steps. These are provisional review budgets: this repository has no retained representative CI measurements, and a local Go test is not a measurement of the Tekton pod or the image build. The Buildah allocation is deliberately sized for the heavyweight image build rather than relying on the small generic Tekton fallback.

The pipelines still use the shared `git-clone` Task. Its resource bounds and Tekton helper-container defaults are supplied by netcup-apps PR #153. Keep #5's cancellation annotation (`cancel-in-progress: "false"`) first, refresh this source branch with that annotation change, and only then apply the positive repository concurrency cap from core netcup-apps PR #151. The annotation change is already carried here as a separate cherry-pick commit `2681b5586d8c6b3f8bcebd031486bea1180fa474`, from PR #5's source commit `d281be8596e3d94e78dd454115089f04f9366f7d`. Merge the app-specific resource PRs before #153 so its generic defaults cannot become the only resource bounds for a heavyweight app step. The shared PR supplies explicit clone/helper bounds and defaults for remaining helper containers.

## Checks completed for review

- Ruby YAML parsing passed for both PipelineRuns.
- `kubectl create --dry-run=client --validate=false` rendered both PipelineRuns. This was client-side only; it did not call the Kubernetes API or check server schema/admission.
- `git diff --check` passed.
- `go vet ./...` passed.
- `go test ./...` passed for `cmd/server` and `lib`.
- The image build was unavailable and was not run.

Cluster CI is intentionally deferred for review because Kubernetes API reads are failing. A deferred or skipped CI check is not a passing test. Before merge or rollout acceptance, after API recovery, exercise each pipeline independently with representative pull-request and master builds; retain run URLs and pod measurements for CPU throttling, peak memory, ephemeral-storage use, OOMs, and evictions. Revisit these provisional budgets against those observations. No trigger, security, or deployment behavior is changed here.
