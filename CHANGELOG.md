# Changelog

All notable changes to this project will be documented in this file.

## [v0.2.1](https://github.com/somaz94/kube-diff/compare/v0.2.0...v0.2.1) (2026-03-18)

### Bug Fixes

- CHANGELOG.md ([fa137c8](https://github.com/somaz94/kube-diff/commit/fa137c8a69445977574fcd85cae96a911318ac5a))

### Documentation

- update changelog ([1408ae2](https://github.com/somaz94/kube-diff/commit/1408ae2b40945beefb8a9f6c249b2bfb6a76ca5c))
- update changelog ([e5430a8](https://github.com/somaz94/kube-diff/commit/e5430a85b619e59cd2777290fac39725ad21f695))
- update changelog ([d5d84cf](https://github.com/somaz94/kube-diff/commit/d5d84cf4b2d2596d2f72ae364fd05e1daa5e906e))
- add use-cases guide and fix template creationTimestamp normalization ([f6aca0f](https://github.com/somaz94/kube-diff/commit/f6aca0fee397a0ae973c1c1533cedebfa7bfd42a))
- update changelog ([5142eab](https://github.com/somaz94/kube-diff/commit/5142eabe1a338f143d240fe879932a9e8e45c79f))

### Contributors

- somaz

<br/>

## [v0.2.0](https://github.com/somaz94/kube-diff/compare/v0.1.0...v0.2.0) (2026-03-18)

### Features

- normalize Kubernetes default fields for cleaner diffs ([4cd7114](https://github.com/somaz94/kube-diff/commit/4cd7114d0374860cd3425a0d74cb4b9a629dda4c))
- add label selector filtering (-l/--selector flag) ([32ca436](https://github.com/somaz94/kube-diff/commit/32ca4361c8ad5314bddbd9468685b13354a03939))
- wire up CLI commands to source/cluster/diff/report pipeline ([b4dd313](https://github.com/somaz94/kube-diff/commit/b4dd31355dc26ed0a22f2869a3e2b2b521013d6e))
- add demo examples, scripts, and examples documentation ([fb18f64](https://github.com/somaz94/kube-diff/commit/fb18f647e8bd3f4fa30020c5e128bf7be99ac844))

### Bug Fixes

- correct fetcher.Get parameter order and improve test coverage ([62e2a0a](https://github.com/somaz94/kube-diff/commit/62e2a0af7d3d2ce4551cb792b0675b141f2b4d53))

### Documentation

- remove roadmap section and update filtering description ([9b271c3](https://github.com/somaz94/kube-diff/commit/9b271c30bc1057f81a8fac3a40c7622a50dfeb8a))
- update changelog ([1bef627](https://github.com/somaz94/kube-diff/commit/1bef6276793d89f79bfe081f671190e419527abd))

### Continuous Integration

- add e2e workflow with kind cluster and demo-all Makefile target ([7653f96](https://github.com/somaz94/kube-diff/commit/7653f966400434b3d668e112c098be52989a373d))

### Contributors

- somaz

<br/>

## [v0.1.0](https://github.com/somaz94/kube-diff/releases/tag/v0.1.0) (2026-03-18)

### Features

- add core source code with CLI, source loaders, cluster fetcher, diff engine, and report output ([e9a96d6](https://github.com/somaz94/kube-diff/commit/e9a96d67867e84622eeea94622baf2355b0d295e))

### Bug Fixes

- use PAT_TOKEN for GoReleaser cross-repo push and fix deprecated format keys ([60adacd](https://github.com/somaz94/kube-diff/commit/60adacde757c2e48681e486c5b1845833c290a1c))

### Documentation

- update CONTRIBUTORS.md ([7c9b7cd](https://github.com/somaz94/kube-diff/commit/7c9b7cd629f2ca772f8ae7fcf88c8da9cb295600))
- update changelog ([1134969](https://github.com/somaz94/kube-diff/commit/1134969bcfb84f2217e5f80c3cc61e9242d65a3a))
- add README, CONTRIBUTING, CODEOWNERS, and docs/ documentation ([21e089a](https://github.com/somaz94/kube-diff/commit/21e089a1d22d1c2f63b3608227aaab2542b1232b))

### Tests

- add edge case tests and update roadmap checkboxes ([bf8522d](https://github.com/somaz94/kube-diff/commit/bf8522d8c4fab543eeedcec37d630adc69c38558))
- add unit tests for all packages (92.6% coverage) ([31f855c](https://github.com/somaz94/kube-diff/commit/31f855c7d3a582bd5962640a3ef4acb46a6a6084))

### Continuous Integration

- add integration tests with helm template and kustomize build ([21ee50e](https://github.com/somaz94/kube-diff/commit/21ee50e8682d39c94c6d4c665b2782065665ffb9))
- enhance CI with coverage threshold, race detection, and go mod tidy check ([eff13ed](https://github.com/somaz94/kube-diff/commit/eff13ed0eb86b53cb3d85015c30f29429dfdabd4))
- add GitHub workflows, dependabot, and release config ([c37607d](https://github.com/somaz94/kube-diff/commit/c37607dfb9bcec83d3d93c8683f340bda16a0706))

### Chores

- **deps:** bump goreleaser/goreleaser-action from 6 to 7 ([1debc92](https://github.com/somaz94/kube-diff/commit/1debc92c456202bd2ba236af656e02d971af3ea9))
- add Makefile, GoReleaser config, .gitignore, .dockerignore ([879b864](https://github.com/somaz94/kube-diff/commit/879b864aefc6ec27e68178c1a6a3c95043e90eb8))

### Contributors

- somaz

<br/>

