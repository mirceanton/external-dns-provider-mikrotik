# Changelog

## [1.4.5](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.4.4...v1.4.5) (2025-04-14)

### 🐛 Fixes

* **go:** update dependency go ( 1.24.1 → 1.24.2 ) ([#220](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/220)) ([d193563](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/d1935630214a39d79128dcd6cb01f8e388363df8))
* **go:** update module github.com/prometheus/client_golang ( v1.21.1 → v1.22.0 ) ([#221](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/221)) ([a3822a9](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/a3822a9e8786a87788066c8a1d361a57cac4ef6b))
* **go:** update module golang.org/x/net ( v0.37.0 → v0.39.0 ) ([#219](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/219)) ([531eec3](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/531eec3d3149697900cd37f5ed9ef84ff0a88d4e))
* **go:** update module sigs.k8s.io/external-dns ( v0.15.1 → v0.16.1 ) ([#213](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/213)) ([ede9ee8](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/ede9ee8450e98288a941db1183e40230fa852456))

### 🧰 Maintenance

* **container:** update ghcr.io/mirceanton/external-dns-provider-mikrotik docker tag ( v1.4.3 → v1.4.4 ) ([#208](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/208)) ([19f50f0](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/19f50f07061781d9560377944ea700669980900c))
* **deps:** update dependency golangci-lint ( 2.0.2 → 2.1.1 ) ([#246](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/246)) ([7cc6064](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/7cc60644a73c73ad364b2abe76c9bd83a6d2ea83))
* **deps:** update dependency golangci/golangci-lint ( v1.64.6 → v1.64.7 ) ([#212](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/212)) ([d60f7e1](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/d60f7e1c8d13270936a6df7def7ed4bde406e64c))
* **deps:** update dependency golangci/golangci-lint ( v1.64.7 → v2.0.2 ) ([#218](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/218)) ([4fa8a0d](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/4fa8a0d9de6e7e0f0ee500e8a4befed26fb49376))
* **deps:** update dependency goreleaser/goreleaser ( v2.7.0 → v2.8.0 ) ([#214](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/214)) ([5eed827](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/5eed8273943d064a2748b36852779c2ccb4dbd74))
* **deps:** update dependency goreleaser/goreleaser ( v2.8.0 → v2.8.2 ) ([#215](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/215)) ([ded038b](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/ded038bf72918154c55769b737b2b3799af9ac91))
* **deps:** update dependency goreleaser/goreleaser ( v2.8.0 → v2.8.2 ) ([#223](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/223)) ([65c9fa4](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/65c9fa4fe98ffdb45a7cbe10fc7a7e6d6ff98950))
* **goreleaser:** update deprecated fields in config ([#226](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/226)) ([b7ba0cc](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/b7ba0ccb892d95c1d897f1c3cf9cbd7bf85539a4))
* **mise:** include all misc tools ([#243](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/243)) ([5e00b84](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/5e00b842af1d2333b20e3a0e40f88171e2beb9d6))
* **mise:** simplify tools config ([#224](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/224)) ([865c8c8](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/865c8c82cd5cde79b948baf698ca9ec5deddffdc))
* **mise**: cleanup config + implement tasks ([#207](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/207)) ([0eb2fea](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/0eb2fea1c95f8c93a6747a4043fdbdaec6fdcca0))

### ♻️ Continuous Integraion

* **codeql:** add code scanning workflow ([#227](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/227)) ([1426f60](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/1426f603fe797c5da242941c153a24d70d4b93a2))
* deprecate reusable actions and move to in-repo configs ([#231](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/231)) ([f76fe7f](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/f76fe7f855eaeeeb89b3c9da90dc22041c0f01a0))
* **deps:** update dependency go ( 1.24.0 → 1.24.1 ) ([#210](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/210)) ([61d61b1](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/61d61b16ae97f7b6904423b91a029e29653d31b2))
* **deps:** update dependency golangci/golangci-lint ( v1.64.3 → v1.64.6 ) ([#209](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/209)) ([15ba6e5](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/15ba6e5fd595996040fbe9cc3181d406291c3446))
* **goreleaser**: fix workflow ([#206](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/206)) ([e7a270f](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/e7a270ff57a113d8bb00efc6691ab43c3c21b6ba))
* **github-actions:** Update actions/create-github-app-token action ( v1.12.0 → v2.0.2 ) ([#238](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/238)) ([c23b9ef](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/c23b9efaa88357422d77008e3177e651a29c3a2e))
* **github-actions:** update github/codeql-action action ( v2 → v3 ) ([#239](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/239)) ([d30ace5](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/d30ace53d60c90dd634141e7e09325293f2f55df))
* **github-actions:** update renovatebot/github-action action ( v41.0.18 → v41.0.19 ) ([#237](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/237)) ([ddbc32a](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/ddbc32a16a5d2f89b28eb1570bc2a59308768848))
* **github-actions:** update renovatebot/github-action action ( v41.0.19 → v41.0.20 ) ([#241](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/241)) ([a5fc7ad](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/a5fc7ad85170e5ace42b721a1b75a136c5bbe9ae))
* **release-please:** add custom config ([e80e1e1](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/e80e1e14b2ebd2d5607f3c4f50743f52d1495030))
* **renovate:** fix config ([#234](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/234)) ([815de8d](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/815de8d06e75d8bb9f989a8464b128ef47283b40))
* **semantic-release:** run job on a schedule instead of every commit ([#222](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/222)) ([6805515](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/6805515c406be080cf4bc4b61150f6157611a7c1))
* **trivy:** add docker image scanner workflow ([#233](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/233)) ([6dba887](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/6dba8876ca4f5cde3350486eec129efe886d6ba3))
* **trivy:** add workflow dispatch trigger ([#235](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/235)) ([65d6438](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/65d64385d954532c29d40cf52e878a230b283f17))

## [1.4.4](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.4.3...v1.4.4) (2025-03-07)

### Bug Fixes

* **go:** update dependency go ( 1.23.6 → 1.24.0 ) ([#195](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/195)) ([3ea4ac4](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/3ea4ac4760e110ea131ad003eafd5a6682c224f0))
* **go:** update dependency go ( 1.24.0 → 1.24.1 ) ([#203](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/203)) ([a28c8ee](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/a28c8ee579373337938912817bff970b40733a73))
* **go:** update module github.com/prometheus/client_golang ( v1.20.5 → v1.21.1 ) ([#200](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/200)) ([d95da04](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/d95da04cc19f3027f4f158f362385e11176b3f7d))
* **go:** update module golang.org/x/net ( v0.35.0 → v0.37.0 ) ([#204](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/204)) ([6d6dafa](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/6d6dafaa0cfc49ef782d49d9ee0bdd6f53d3a19e))

## [1.4.3](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.4.2...v1.4.3) (2025-02-10)

### Bug Fixes

* **go:** update module golang.org/x/net ( v0.34.0 → v0.35.0 ) ([#192](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/192)) ([9b3f96c](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/9b3f96c62742af2cfd789b0978d247c10e0ea8fc))

## [1.4.2](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.4.1...v1.4.2) (2025-02-04)

### Bug Fixes

* **go:** update dependency go ( 1.23.5 → 1.23.6 ) ([#190](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/190)) ([48d2c1b](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/48d2c1b15b8653976e25cdc0f4c78c5a2ebbbcca))
* **go:** update module github.com/go-chi/chi/v5 ( v5.2.0 → v5.2.1 ) ([#189](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/189)) ([7d2e359](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/7d2e35904c0935f715e2c6c574aad0d402c839de))

## [1.4.1](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.4.0...v1.4.1) (2025-01-21)

### Bug Fixes

* Provider-specific configuration from other eDNS instances are causing errors ([#181](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/181)) ([3bd5182](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/3bd5182c29d6e8a0c2009a5ce291dc9f63c38356)) by @mircea-pavel-anton

## [1.4.0](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.3.0...v1.4.0) (2025-01-21)

### Features

* Add support for setting a default comment to all records that do not have one set ([#180](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/180)) ([55a7ce0](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/55a7ce058c256b0cd62d6c6b8528d34a3964caf3)) by @mircea-pavel-anton

**Full Changelog**: <https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.3.0...v1.4.0>

## [1.3.0](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.2.7...v1.3.0) (2025-01-21)

### Features

* Add default TTL for Records ([#173](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/173)) ([213a25d](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/213a25dfb8b874eb05aa03bc2334273105e2f8d7)) by @nobleess

### Chores

* ci(github-actions): update mirceanton/reusable-workflows action ( v3.4.46 → v3.4.47 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/177>
* chore(container): update ghcr.io/mirceanton/external-dns-provider-mikrotik docker tag ( v1.2.6 → v1.2.7 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/176>

## [1.2.7](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.2.6...v1.2.7) (2025-01-19)

### Bug Fixes

* don't run release workflows on forks ([8fad0bd](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/8fad0bda6af7de3a9dd2df9175288f42c877506a))
* don't run renovate on PRs (it still uses config from main. it makes no sense) ([5923ebb](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/5923ebb35db88a4e5c27e9ee801c0c002b335d2a))
* **go:** update dependency go ( 1.23.4 → 1.23.5 ) ([#174](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/174)) ([6cde5c9](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/6cde5c94dd724df93811bca3731f72a8a7d4e7ac))
* **go:** update module golang.org/x/net ( v0.33.0 → v0.34.0 )([#169](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/169)) ([c865cb0](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/c865cb02c10dbc017a4c6a0e53a9f0476f81819e))

## [1.2.6](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.2.5...v1.2.6) (2025-01-05)

### Bug Fixes

* Webhook cannot import records with regexp ([#164](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/164)) ([2d287bc](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/2d287bc476f4575b287ce0bee5e5c94ad8f461a7))

### Misc

* Update golangci/golangci-lint docker tag ( v1.62.2 → v1.63.1 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/158>
* Update golangci/golangci-lint docker tag ( v1.63.1 → v1.63.2 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/159>
* Update goreleaser/goreleaser docker tag ( v2.5.0 → v2.5.1 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/160>
* Update golangci/golangci-lint docker tag ( v1.63.3 → v1.63.4 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/163>
* Update mirceanton/reusable-workflows action ( v3.4.37 → v3.4.38 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/155>
* Update mirceanton/reusable-workflows action ( v3.4.38 → v3.4.39 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/156>
* Update mirceanton/reusable-workflows action ( v3.4.39 → v3.4.40 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/157>
* Update mirceanton/reusable-workflows action ( v3.4.40 → v3.4.41 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/165>

## [1.2.5](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.2.4...v1.2.5) (2024-12-21)

### Dependency Updates

* **go:** update module github.com/caarlos0/env/v11 ( v11.3.0 → v11.3.1 ) ([#152](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/152)) ([c2ebefc](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/c2ebefc9346117ae3d7ae40f6e8b60fb3b82277d))

## [1.2.4](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.2.3...v1.2.4) (2024-12-19)

### Bug Fixes

* **goreleaser:** Version and Git SHA (still) not properly passed in 😅 ([fcad658](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/fcad6583ccb8331b94cbfb8555ea7d542f5cbc80))

## [1.2.3](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.2.2...v1.2.3) (2024-12-19)

### Bug Fixes

* **goreleaser:** Version and Git SHA not properly passed in ([#150](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/150)) ([c115a83](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/c115a83028a1def347e2722887bf69c87317670a))

### Dependency Updates

* **golang:** update module github.com/go-chi/chi/v5 ( v5.1.0 → v5.2.0 ) (#149)
* **golang:** update module sigs.k8s.io/external-dns ( v0.15.0 → v0.15.1 ) (#147)
* **golang:** update module github.com/caarlos0/env/v11 ( v11.2.2 → v11.3.0 ) (#148)
* **devcontainer:** update mcr.microsoft.com/devcontainers/go:1.23-bookworm docker digest ( 2e00578 → a417a34 ) (#138)
* **devcontainer:** update goreleaser/goreleaser docker tag ( v2.4.8 → v2.5.0 ) (#143)
* **github-actions:** update mirceanton/reusable-workflows action ( v3.4.35 → v3.4.36 ) (#144)

## [1.2.2](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.2.1...v1.2.2) (2024-12-11)

Previously the webhook was sending an API request to the Mikrotik RouterOS API to get a list of all of the configured DNS records. This would cause an issue since Mikrotik does support some record types that external DNS does NOT support.

As of this release, the webhook now adds some filtering to only get a list of supported DNS record types back from the RouterOS API.

### Bug Fixes

* only fetch supported record types from the Mikrotik API ([#135](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/135)) ([58df033](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/58df033e87eb873a8095f52a0bdd58d92f2be035))

## [1.2.1](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.2.0...v1.2.1) (2024-12-04)

### Dependency Updates

* update dependency go ( 1.23.2 → 1.23.3 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/108>
* update module golang.org/x/net ( v0.30.0 → v0.31.0 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/110>
* update module github.com/stretchr/testify ( v1.9.0 → v1.10.0 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/122>
* update dependency go ( 1.23.3 → 1.23.4 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/127>
* update module golang.org/x/net ( v0.31.0 → v0.32.0 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/128>

## [1.2.0](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.1.0...v1.2.0) (2024-10-29)

### Features

* Add support for setting `disabled` field in Mikrotik DNS Records via Provider Specific ([#101](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/101)) ([e3d6a13](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/e3d6a13cfc6662136bb14800e81328ecb27db248))

## [1.1.0](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.0.0...v1.1.0) (2024-10-29)

### Features

* Add support for ingress and service annotations ([#100](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/100)) ([95f5b64](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/95f5b64cbd7af73b86c9170498dba675121e2c56))

## [1.0.0](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v0.2.5...v1.0.0) (2024-10-27)

### 💣 BREAKING CHANGES 💣

* The `MIKROTIK_HOST` and `MIKROTIK_PORT` environment variables have been replaced with a single one called `MIKROTIK_BASEURL`

### Features

* Add support for `NS` records ([#98](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/98)) ([2e5776c](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/2e5776cd203fd5dfbee32c2a15a7f94b595e99a3))
* Add support for `SRV` records ([#97](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/97)) ([c195471](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/c195471a67be8f89316587323d5a9546179177c2))
* Add support for `MX` records ([#95](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/95)) ([2a9ff3c](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/2a9ff3cc200fea0a507c74ace9c0cdcde31df328))

### CI

* deps(github-actions): update mirceanton/reusable-workflows action ( v3.4.18 → v3.4.21 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/89>
* deps(github-actions): update mirceanton/reusable-workflows action ( v3.4.21 → v3.4.24 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/93>

### Misc

* chore(container): update mcr.microsoft.com/devcontainers/go:1.23-bookworm docker digest ( 16c623a → 2e00578 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/92>
* chore: Release Cleanup by @mircea-pavel-anton in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/99>

## [0.2.5](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v0.2.4...v0.2.5) (2024-10-15)

### Dependency Updates

* deps(golang): update module github.com/prometheus/client_golang ( v1.20.4 → v1.20.5 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/86>

## [0.2.4](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v0.2.3...v0.2.4) (2024-10-15)

### Bug Fixes

* TTL Conversion ([#88](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/88)) ([4847b35](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/4847b3540ac4fe21db896db49a4f1002602e991d))

## [0.2.3](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v0.2.2...v0.2.3) (2024-10-13)

### Fixes

* patch(deps): update module sigs.k8s.io/external-dns ( v0.14.2 → v0.15.0 ) by @mr-borboto in #82

## [0.2.2](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v0.2.1...v0.2.2) (2024-10-13)

### Fixes

* patch(deps): update module github.com/prometheus/client_golang ( v1.19.1 → v1.20.4 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/77>

## [0.2.1](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v0.2.0...v0.2.1) (2024-10-13)

### Fixes

* patch(deps): update module golang.org/x/net ( v0.28.0 → v0.30.0 ) by @mr-borboto in <https://github.com/mirceanton/external-dns-provider-mikrotik/pull/83>

## [0.2.0](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v0.1.4...v0.2.0) (2024-10-13)

### Features

* Add support for A, AAAA, CNAME and TXT Record Types ([#14](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/14)) ([7a68266](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/7a68266cf2bc1da90200d247fbee59fb0348820d)), closes [#61](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/61)

## [0.1.4](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v0.1.3...v0.1.4) (2024-08-07)

### Bug Fixes

* **deps:** update module github.com/caarlos0/env/v11 to v11.2.2 ([#68](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/68)) ([7f5e0bc](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/7f5e0bc5bd417e28cf65af289d22b061db57c838))

## [0.1.3](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v0.1.2...v0.1.3) (2024-08-06)

### Bug Fixes

* **deps:** update module github.com/caarlos0/env/v11 to v11.2.0 ([#63](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/63)) ([7b2a367](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/7b2a3670ad029f7f9d31a17bca6c748afcb4f7dc))

## [0.1.2](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v0.1.1...v0.1.2) (2024-08-06)

### Bug Fixes

* **deps:** update module golang.org/x/net to v0.28.0 ([#65](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/65)) ([548c022](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/548c022cb971b9b0f79d08a4b4755b7043a6ce32))

## [0.1.1](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v0.1.0...v0.1.1) (2024-08-06)

### Bug Fixes

* **deps:** update module golang.org/x/net to v0.27.0 ([#37](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/37)) ([f66ada8](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/f66ada891f9d90e6d2b249325698eff8404a13c2))
* **docker:** run as non-root ([489a95a](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/489a95adeed6082c6e2d77c717c2c6215202fcf8))

### Misc

* remove hash pins from devcontainer config ([5de8592](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/5de8592abe634c2fa3760eaa09bc4db64e442009))

## [0.1.0](https://github.com/mirceanton/external-dns-provider-mikrotik/releases/tag/v0.1.0) (2024-06-05)

Initial "working?" version
