# Changelog

All notable changes to this project will be documented in this file.

## Unreleased (2026-07-22)

### Documentation

- remove DCO sign-off instructions ([b1f2d1d](https://github.com/somaz94/kube-diff/commit/b1f2d1d2e3be4bad1b8788fa95971dbaeacf33e3))

### Continuous Integration

- remove DCO workflow ([f824129](https://github.com/somaz94/kube-diff/commit/f824129b4a5bf0f7a7ec9d36077407727a557864))
- pin kustomize install to fix flaky rate-limited download ([de5396d](https://github.com/somaz94/kube-diff/commit/de5396dba336bd20a0f86e9d54351e058824118b))

### Chores

- **deps:** bump actions/setup-go from 6 to 7 ([9a9296a](https://github.com/somaz94/kube-diff/commit/9a9296a2775a853be28c2897ddd4f95ff38590ef))

### Contributors

- somaz

<br/>

## [v0.5.0](https://github.com/somaz94/kube-diff/compare/v0.4.2...v0.5.0) (2026-07-08)

### Features

- add BytesSource for in-memory YAML manifests ([18914d7](https://github.com/somaz94/kube-diff/commit/18914d7d6ed81a30282f94e62ddee7ab1de2fc1b))
- extract diff engine into importable pkg/ with engine orchestrator ([60627c4](https://github.com/somaz94/kube-diff/commit/60627c47bd0e5abebe8d7f48c84f5362922767ae))

### Documentation

- document DCO sign-off requirement in CONTRIBUTING ([227466d](https://github.com/somaz94/kube-diff/commit/227466d764e8b8919ec08faf68825f63c0d40905))

### Tests

- sanitize kubeconfig fixture token to example value ([19eaee0](https://github.com/somaz94/kube-diff/commit/19eaee0e7c6bb359c18583c5b0dfc76d3383207a))

### Continuous Integration

- adopt semantic-pr, labels, lock-threads, PR size, and auto-assign reusables ([8fd933e](https://github.com/somaz94/kube-diff/commit/8fd933e4ec2743655c50be923b3364c16abdf18d))
- use reusable stale-issues workflow ([54e9e28](https://github.com/somaz94/kube-diff/commit/54e9e282389d52acf035bd757ddc93e827543c0f))
- use reusable issue-greeting workflow ([802570a](https://github.com/somaz94/kube-diff/commit/802570af74dec980c6c4d53a23066386de4d3aeb))
- use reusable dependabot-auto-merge workflow ([d5e3318](https://github.com/somaz94/kube-diff/commit/d5e3318888fb592562d8f545bda6fd2348b724b1))
- use reusable contributors workflow ([25211c4](https://github.com/somaz94/kube-diff/commit/25211c4300368a8214bc4f7b78362158d8f5a3a9))
- gate e2e on fork PRs via ok-to-test label ([3e35d0d](https://github.com/somaz94/kube-diff/commit/3e35d0d6c5d0e7d3ca55c6d6d7ef22ef2efac07a))
- add ok-to-test workflow stub ([9c88879](https://github.com/somaz94/kube-diff/commit/9c88879ad031f004570dbe971b66fb98fc9827c6))
- add PR welcome workflow stub ([46d6348](https://github.com/somaz94/kube-diff/commit/46d6348b60b1efb2399b91189e81feb2a3fd0b21))
- pin Helm version and authenticate setup-helm to reduce CI flakes ([59516c6](https://github.com/somaz94/kube-diff/commit/59516c698da4dc5921fd560ef664d85349e04628))
- add DCO check via shared reusable workflow ([41dbd3d](https://github.com/somaz94/kube-diff/commit/41dbd3dd92138ae8dec2dd657aa3feb11dbdc0e4))
- add concurrency guards to recurring workflows ([cd2bed9](https://github.com/somaz94/kube-diff/commit/cd2bed95520ce12248d8615e0ae936d347ef461d))

### Chores

- **deps:** bump actions/checkout from 6 to 7 ([6ffa209](https://github.com/somaz94/kube-diff/commit/6ffa20962b1c88318937f5eb641bd5f9389fde93))
- **deps:** bump the go-minor group with 2 updates (#10) ([#10](https://github.com/somaz94/kube-diff/pull/10)) ([c4f01bf](https://github.com/somaz94/kube-diff/commit/c4f01bf73798e9ed3c39465d6335e49598ea7ffb))
- **deps:** bump the go-minor group with 2 updates (#9) ([#9](https://github.com/somaz94/kube-diff/pull/9)) ([c4f99ef](https://github.com/somaz94/kube-diff/commit/c4f99efa9332b6557398355c4704d0363760475d))
- **deps:** bump github.com/fsnotify/fsnotify in the go-minor group (#8) ([#8](https://github.com/somaz94/kube-diff/pull/8)) ([30da3ea](https://github.com/somaz94/kube-diff/commit/30da3eaf2913b979878f98add33ea43fa8bae479))
- **deps:** bump the go-minor group with 2 updates (#7) ([#7](https://github.com/somaz94/kube-diff/pull/7)) ([e618c4f](https://github.com/somaz94/kube-diff/commit/e618c4fe3071636457230798671c995220a9cb3d))
- **deps:** bump the go-minor group with 2 updates (#6) ([#6](https://github.com/somaz94/kube-diff/pull/6)) ([e89b0c2](https://github.com/somaz94/kube-diff/commit/e89b0c2545dfdd00c20a5d5aa3f37d5f7a2b6c27))
- **deps:** bump dependabot/fetch-metadata from 2 to 3 ([3d8aefa](https://github.com/somaz94/kube-diff/commit/3d8aefa503f7b183c13744a7f4ec8b398cf5cc7e))
- **deps:** bump actions/github-script from 8 to 9 ([8f70ad6](https://github.com/somaz94/kube-diff/commit/8f70ad6187d8e51b3cf2277a3d673e8ff3496d04))

### Contributors

- somaz

<br/>

## [v0.4.2](https://github.com/somaz94/kube-diff/compare/v0.4.1...v0.4.2) (2026-04-03)

### Features

- add branch and pr workflow targets to Makefile ([49d3363](https://github.com/somaz94/kube-diff/commit/49d3363f956c94ceb0ac4a97e9b498383b2c1308))
- add Scoop bucket support for Windows distribution ([b5a1ccf](https://github.com/somaz94/kube-diff/commit/b5a1ccf2791a07869068b2f580ec26e5d793858f))

### Bug Fixes

- add missing errors import, extract ErrChangesDetected from os.Exit ([53e65c6](https://github.com/somaz94/kube-diff/commit/53e65c6be1185348b3bdbd063a1da91b68aa61b8))

### Documentation

- remove duplicate rules covered by global CLAUDE.md ([8b52108](https://github.com/somaz94/kube-diff/commit/8b52108228e80bbf57613c6e357032d399261619))
- add specific version install instructions to README ([8d7cbf9](https://github.com/somaz94/kube-diff/commit/8d7cbf973131f664949c27c27a9a78407361293b))
- add Uninstall section to README ([89b6ec0](https://github.com/somaz94/kube-diff/commit/89b6ec00a48085dc20ee48f99a56923822c379cd))

### Continuous Integration

- add changelog category groups in goreleaser config ([705700c](https://github.com/somaz94/kube-diff/commit/705700cb76553f17e0eee818885c431d023a0334))
- add auto-generated PR body script and limit push trigger to main ([69dd46e](https://github.com/somaz94/kube-diff/commit/69dd46e27fc9fc1c6b47cb3aff6c4927d8f80bd9))

### Chores

- remove duplicate rules from CLAUDE.md (moved to global) ([cd442f5](https://github.com/somaz94/kube-diff/commit/cd442f5f60c9e613fe4a004584adc15b47224f5a))
- add git config protection to CLAUDE.md ([bdf5b97](https://github.com/somaz94/kube-diff/commit/bdf5b9769b2e30cc503d7948f76a3ee853cff500))
- **deps:** bump azure/setup-helm from 4 to 5 ([8cdb976](https://github.com/somaz94/kube-diff/commit/8cdb976f1c4ae957d7f7eee6981e4bc167c5ab4d))
- **deps:** bump the go-minor group with 2 updates (#2) ([#2](https://github.com/somaz94/kube-diff/pull/2)) ([4f703bc](https://github.com/somaz94/kube-diff/commit/4f703bca8058f5dd585c2fb00056a83140d889db))

### Contributors

- somaz

<br/>

## [v0.4.1](https://github.com/somaz94/kube-diff/compare/v0.4.0...v0.4.1) (2026-03-19)

### Features

- add brew install caveats message ([2805957](https://github.com/somaz94/kube-diff/commit/28059579846bf9ab468e3f1326a61db2a026fa4f))

### Contributors

- somaz

<br/>

## [v0.4.0](https://github.com/somaz94/kube-diff/compare/v0.3.1...v0.4.0) (2026-03-19)

### Bug Fixes

- use GITHUB_TOKEN for dependabot auto merge ([2567c80](https://github.com/somaz94/kube-diff/commit/2567c806d95e4e0af148e6cedfed9f58dbb77c1b))

### Code Refactoring

- extract shared test helper, toStringSet, executeDiff, io.Writer for printReport ([37affd1](https://github.com/somaz94/kube-diff/commit/37affd1bc86ef285632d68adaa955c771f779c8a))
- extract filter helpers, split runDiff, deduplicate summary, remove unused restMapper ([3aa14d1](https://github.com/somaz94/kube-diff/commit/3aa14d19d5b74b6cc8292c86a7fe4f48c967c5fc))

### Documentation

- README.md ([dc4225c](https://github.com/somaz94/kube-diff/commit/dc4225cfe94ef35a3e244824aa7c6672b47d5959))
- add no-push rule to CLAUDE.md ([d85fb5c](https://github.com/somaz94/kube-diff/commit/d85fb5cdc039368f3e8a4626674b084e4f6c0084))
- add missing --name/-N flag to all documentation ([a7090c7](https://github.com/somaz94/kube-diff/commit/a7090c7e621cc6212fb4c75fabb27bcbb2f6171f))
- CLUADE.md ([ddc7396](https://github.com/somaz94/kube-diff/commit/ddc73965a7b24a4eb0df11fd444da6e535bc76be))
- README.md ([e35ab97](https://github.com/somaz94/kube-diff/commit/e35ab97c0ddaa6503f9e69d33b9dcc3247269234))

### Tests

- improve coverage for normalize, executeDiff, watch helpers ([903b01b](https://github.com/somaz94/kube-diff/commit/903b01b571f97ef42503fd7e788ad50f8ae2ce9d))

### Continuous Integration

- remove lint workflow ([2f39e07](https://github.com/somaz94/kube-diff/commit/2f39e07545b7cccb3849c0f6fbb1a68f8a579a69))
- upgrade golangci-lint to v2.11.3 for Go 1.26 compatibility ([fbd1bdd](https://github.com/somaz94/kube-diff/commit/fbd1bdd85c09d8d6d8d3a9df18fb059b2b0c9add))
- enable lint workflow on push and pull_request triggers ([3015a0d](https://github.com/somaz94/kube-diff/commit/3015a0df837a961603ac1384320ae9a6fd0338ee))

### Styles

- apply go fmt formatting fixes ([75d09e2](https://github.com/somaz94/kube-diff/commit/75d09e23307679ab30805744530a0c583f80bcc6))

### Contributors

- somaz

<br/>

## [v0.3.1](https://github.com/somaz94/kube-diff/compare/v0.3.0...v0.3.1) (2026-03-18)

### Features

- add --name (-N) flag to filter resources by name ([67a23b7](https://github.com/somaz94/kube-diff/commit/67a23b78e79497b316a7b6223a2e4b9e8c867950))

### Documentation

- add brew upgrade instructions to README ([15e0db5](https://github.com/somaz94/kube-diff/commit/15e0db5c89261587be976015131af3d5a0a9b12a))

### Contributors

- somaz

<br/>

## [v0.3.0](https://github.com/somaz94/kube-diff/compare/v0.2.1...v0.3.0) (2026-03-18)

### Features

- add --diff-strategy flag and watch command ([365c683](https://github.com/somaz94/kube-diff/commit/365c68320e4f67a122ad0047aad8ef75d13ac283))
- add --ignore-field, --context-lines, --exit-code flags and table output ([ee1cc7c](https://github.com/somaz94/kube-diff/commit/ee1cc7c9b49bd1a5ba3e48c90c0cddb06e2fa8e6))

### Bug Fixes

- JSON output exit code bug and add Job/DaemonSet normalize tests ([6854cf7](https://github.com/somaz94/kube-diff/commit/6854cf738d1b8c29b2bb90a3c272e00cf0266bf1))
- resolve lint warnings and summary-only flag bug ([b9a531d](https://github.com/somaz94/kube-diff/commit/b9a531d0d81372ecb655ec1d605ae8e026eb97fd))

### Code Refactoring

- extract ResourceFetcher interface and compareResources function ([54a6283](https://github.com/somaz94/kube-diff/commit/54a6283213502792b47bb9b412c09e5197fb9a3d))

### Documentation

- update documentation for new features ([8ea967f](https://github.com/somaz94/kube-diff/commit/8ea967f1e4cc621c1f62a8979d2f6e02a6e24e7e))

### Continuous Integration

- add table output and advanced features to demo script ([d1b8d14](https://github.com/somaz94/kube-diff/commit/d1b8d1407bb28a725c5396feb7bccd56844867c5))

### Contributors

- somaz

<br/>

## [v0.2.1](https://github.com/somaz94/kube-diff/compare/v0.2.0...v0.2.1) (2026-03-18)

### Bug Fixes

- CHANGELOG.md ([fa137c8](https://github.com/somaz94/kube-diff/commit/fa137c8a69445977574fcd85cae96a911318ac5a))

### Documentation

- add use-cases guide and fix template creationTimestamp normalization ([f6aca0f](https://github.com/somaz94/kube-diff/commit/f6aca0fee397a0ae973c1c1533cedebfa7bfd42a))

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

