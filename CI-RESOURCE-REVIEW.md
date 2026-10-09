# CI resource continuation

2026-10-09: keep existing lower Go budgets and change the provisional 8Gi
Buildah memory limit to 4Gi. General build nodes have approximately 7.75Gi
allocatable RAM; an 8Gi step limit cannot contain one build before node
exhaustion. CPU, memory requests and ephemeral-storage budgets stay unchanged.

These are interim bounds. Successful representative Amauta Go/frontend and
Buildah runs establish class-level evidence, not Betula-specific sizing.
Keep per-application peak/storage and startup measurements pending. The
skip marker defers an automatic publication burst; it is not a passed test.

2026-10-09 full v1 PipelineRun API check: preserve the configured CI service account under spec.taskRunTemplate.serviceAccountName. The legacy spec.serviceAccountName field is not part of the installed v1 schema. Complete PipelineRun specs now validate against the installed API, in addition to the strict Step resource check. [skip tkn]

Release update (2026-10-09): PaC v0.51.0 and all fourteen repository caps are live. Complete PipelineRun specs and inline Step budgets passed installed Tekton v1 schemas; resource keys are computeResources and service accounts use taskRunTemplate. These are interim containment budgets. Representative Go, frontend and Buildah classes passed in Amauta; this does not establish sizing or successful application tests for other repositories. CPU throttling was observed and ephemeral-storage peaks remain unavailable. Per-path runtime sizing stays pending after rollout.

Automatic publication uses [skip tkn] to avoid a build burst. A skipped status is not a passing test. Application-specific budgets precede shared remote-clone/helper defaults in netcup-apps #153. Pipeline triggers and security gates remain enabled.
