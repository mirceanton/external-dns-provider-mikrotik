# Changelog

## [1.4.5](https://github.com/mirceanton/external-dns-provider-mikrotik/compare/v1.4.4...v1.4.5) (2025-04-14)


### 🐛 Fixes

* **go:** update dependency go ( 1.24.1 → 1.24.2 ) ([#220](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/220)) ([d193563](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/d1935630214a39d79128dcd6cb01f8e388363df8))
* **go:** update module github.com/prometheus/client_golang ( v1.21.1 → v1.22.0 ) ([#221](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/221)) ([a3822a9](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/a3822a9e8786a87788066c8a1d361a57cac4ef6b))
* **go:** update module golang.org/x/net ( v0.37.0 → v0.39.0 ) ([#219](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/219)) ([531eec3](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/531eec3d3149697900cd37f5ed9ef84ff0a88d4e))
* **go:** update module sigs.k8s.io/external-dns ( v0.15.1 → v0.16.1 ) ([#213](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/213)) ([ede9ee8](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/ede9ee8450e98288a941db1183e40230fa852456))


### 🧰 Maintenance

* cleanup mise config + implement tasks ([#207](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/207)) ([0eb2fea](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/0eb2fea1c95f8c93a6747a4043fdbdaec6fdcca0))
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


### ♻️ Continuous Integraion

* **codel:** add code scanning workflow ([#227](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/227)) ([1426f60](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/1426f603fe797c5da242941c153a24d70d4b93a2))
* deprecate reusable actions and move to in-repo configs ([#231](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/231)) ([f76fe7f](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/f76fe7f855eaeeeb89b3c9da90dc22041c0f01a0))
* **deps:** update dependency go ( 1.24.0 → 1.24.1 ) ([#210](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/210)) ([61d61b1](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/61d61b16ae97f7b6904423b91a029e29653d31b2))
* **deps:** update dependency golangci/golangci-lint ( v1.64.3 → v1.64.6 ) ([#209](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/209)) ([15ba6e5](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/15ba6e5fd595996040fbe9cc3181d406291c3446))
* fix goreleaser workflow ([#206](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/206)) ([e7a270f](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/e7a270ff57a113d8bb00efc6691ab43c3c21b6ba))
* **github-actions:** Update actions/create-github-app-token action ( v1.12.0 → v2.0.2 ) ([#238](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/238)) ([c23b9ef](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/c23b9efaa88357422d77008e3177e651a29c3a2e))
* **github-actions:** update github/codeql-action action ( v2 → v3 ) ([#239](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/239)) ([d30ace5](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/d30ace53d60c90dd634141e7e09325293f2f55df))
* **github-actions:** update renovatebot/github-action action ( v41.0.18 → v41.0.19 ) ([#237](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/237)) ([ddbc32a](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/ddbc32a16a5d2f89b28eb1570bc2a59308768848))
* **github-actions:** update renovatebot/github-action action ( v41.0.19 → v41.0.20 ) ([#241](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/241)) ([a5fc7ad](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/a5fc7ad85170e5ace42b721a1b75a136c5bbe9ae))
* **release-please:** add custom config ([e80e1e1](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/e80e1e14b2ebd2d5607f3c4f50743f52d1495030))
* **renovate:** fix config ([#234](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/234)) ([815de8d](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/815de8d06e75d8bb9f989a8464b128ef47283b40))
* **semantic-release:** run job on a schedule instead of every commit ([#222](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/222)) ([6805515](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/6805515c406be080cf4bc4b61150f6157611a7c1))
* **trivy:** add docker image scanner workflow ([#233](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/233)) ([6dba887](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/6dba8876ca4f5cde3350486eec129efe886d6ba3))
* **trivy:** add workflow dispatch trigger ([#235](https://github.com/mirceanton/external-dns-provider-mikrotik/issues/235)) ([65d6438](https://github.com/mirceanton/external-dns-provider-mikrotik/commit/65d64385d954532c29d40cf52e878a230b283f17))
