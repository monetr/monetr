# Changelog

## [0.11.21](https://github.com/monetr/monetr/compare/v0.11.20...v0.11.21) (2022-10-15)


### Features

* **funding:** Exposing Exclude Weekends for funding schedules. ([2b120c7](https://github.com/monetr/monetr/commit/2b120c7cb0d076dc342001ae43e657e6b56c9719))
* **ui:** Add help link to the sidebar. ([25acb22](https://github.com/monetr/monetr/commit/25acb224c274f7ff8c5d474e7a5e2db8eab06516))


### Bug Fixes

* **ui:** Fixed sizing on the updated expenses view. ([c7ccaec](https://github.com/monetr/monetr/commit/c7ccaec109a0678e075d7d62c1ad6e4cb0856265))
* **ui:** Fixed transfer dialog crash. ([#1134](https://github.com/monetr/monetr/issues/1134)) ([480af70](https://github.com/monetr/monetr/commit/480af700cc5a32b5503790396738b7e0dc09db7e)), closes [#1133](https://github.com/monetr/monetr/issues/1133)


### Dependencies

* **api:** update module github.com/getsentry/sentry-go to v0.14.0 ([#1128](https://github.com/monetr/monetr/issues/1128)) ([14b5ff9](https://github.com/monetr/monetr/commit/14b5ff9c64e5aa9d3ea04d8a1ad34d7b2280b712))
* **api:** update module github.com/spf13/cobra to v1.6.0 ([#1129](https://github.com/monetr/monetr/issues/1129)) ([9d4e54a](https://github.com/monetr/monetr/commit/9d4e54a19774cbe461476ebfadc77982e5536456))
* **api:** update module github.com/spf13/viper to v1.13.0 ([#1130](https://github.com/monetr/monetr/issues/1130)) ([48a3738](https://github.com/monetr/monetr/commit/48a37384fb5587dda2dd5e489c78225e6928bdbf))


### Miscellaneous

* **docs:** Added beta note :tada: ([052504b](https://github.com/monetr/monetr/commit/052504b9130f29aa4779f6da63aca389868c4e72))
* **docs:** Adding discord link to doc site. ([6562756](https://github.com/monetr/monetr/commit/65627560171af89296b8b3fca34911365b1233bd))
* **docs:** Adding documentation for adding funding schedules. ([e60306e](https://github.com/monetr/monetr/commit/e60306e828ce6c15dbf456982939ac89d3325511))
* More sentry improvements to tracing. ([dac0b87](https://github.com/monetr/monetr/commit/dac0b874e4bb9292b49fd2a47a1b9cc418ba6792))
* Update third party NOTICE ([96d344e](https://github.com/monetr/monetr/commit/96d344e6482fc65a5c5e18a2de2e8222672edd93))
* Updated third party NOTICE ([40bb2bc](https://github.com/monetr/monetr/commit/40bb2bc3bbda4dd38cb6252da5369f537d3d6c67))

## [0.11.20](https://github.com/monetr/monetr/compare/v0.11.19...v0.11.20) (2022-10-14)


### Features

* **ui:** New funding schedule UI. ([#1122](https://github.com/monetr/monetr/issues/1122)) ([4aea72f](https://github.com/monetr/monetr/commit/4aea72f0d356aa453191a8841e5aedc410bcedca))


### Bug Fixes

* **ui:** Fixed Sentry crash dialog during registration. ([#1125](https://github.com/monetr/monetr/issues/1125)) ([58bb97b](https://github.com/monetr/monetr/commit/58bb97b516647c3edf5bc02a9ff92201b4686ceb))


### Miscellaneous

* **sentry:** Improvements to sentry tracing. ([57a11e3](https://github.com/monetr/monetr/commit/57a11e3e4e373feef0acff0efaca1216ba86ad72))
* Updating license notice. ([c98f046](https://github.com/monetr/monetr/commit/c98f046d2cd82ec4d4afc204dce155f30489ca14))

## [0.11.19](https://github.com/monetr/monetr/compare/v0.11.18...v0.11.19) (2022-10-08)


### Features

* **sentry:** Significantly improved sentry trace spans. ([b1d6a2f](https://github.com/monetr/monetr/commit/b1d6a2fb8da52ea353bcfba05de04a9711d06291))


### Bug Fixes

* **sentry:** Improved how errors from API requests are reported. ([75c1a2c](https://github.com/monetr/monetr/commit/75c1a2c088bca5b0013ad2aef65dff74b7e4e0dc))


### Dependencies

* **api:** Bumped go-pg to `v10.10.7` ([6cee643](https://github.com/monetr/monetr/commit/6cee64304224451b02507609ba8d45f5a7900211))


### Miscellaneous

* Add discord to README. ([8e04d3a](https://github.com/monetr/monetr/commit/8e04d3a3d0a2821f054aa6357421f99eade8eb98))

## [0.11.18](https://github.com/monetr/monetr/compare/v0.11.17...v0.11.18) (2022-10-02)


### Features

* **notices:** Include third party notices in application. ([#1118](https://github.com/monetr/monetr/issues/1118)) ([c282028](https://github.com/monetr/monetr/commit/c282028480880ec47a07735dd6ca72fd2345caeb))


### Bug Fixes

* **build:** Fixed license task ([2078a24](https://github.com/monetr/monetr/commit/2078a24d77439f1cc77568f8c1dc117cb9bb1a72))


### Miscellaneous

* Makefile improvements. ([63e9470](https://github.com/monetr/monetr/commit/63e9470a1e4f92cc86777a4307cf5e355263e743))
* **ui:** Minor improvements around TOTP MFA code. ([e386c57](https://github.com/monetr/monetr/commit/e386c57da24ed0e1ae9daf0f4838769926067467))

## [0.11.17](https://github.com/monetr/monetr/compare/v0.11.16...v0.11.17) (2022-10-02)


### Features

* Added taks to validate licenses. ([#1096](https://github.com/monetr/monetr/issues/1096)) ([9391004](https://github.com/monetr/monetr/commit/93910045e76b23233651da9748adf32884f0fef1))
* **ui:** Create Dialogs rewrite. ([#1080](https://github.com/monetr/monetr/issues/1080)) ([978178d](https://github.com/monetr/monetr/commit/978178dafe5fa6acf143e0e13103c8b439da1d17))


### Bug Fixes

* **deps:** Fixed RRule import. ([c73a288](https://github.com/monetr/monetr/commit/c73a28880510a0ae9fae455a0be5bb5eba3f9042))


### Documentation

* **funding:** Minor documentation improvements. ([1c1c920](https://github.com/monetr/monetr/commit/1c1c920f63b3a3eeaf70a29b31e665044a3f239e))


### Dependencies

* **api:** update golang.org/x/crypto digest to eccd636 ([#1085](https://github.com/monetr/monetr/issues/1085)) ([d78f833](https://github.com/monetr/monetr/commit/d78f8332e163669dce1cd099c0db1bf10052cc21))
* **api:** update google.golang.org/genproto digest to c98284e ([#1094](https://github.com/monetr/monetr/issues/1094)) ([65eddb5](https://github.com/monetr/monetr/commit/65eddb5c47d35d4395ea34a90fda68e46aa33c3b))
* **ci:** Upgrading + using monetr's own build containers. ([d33d24e](https://github.com/monetr/monetr/commit/d33d24e5533d9de7e0d2550054776799bfa1058c))
* **containers:** update squidfunk/mkdocs-material docker tag to v8.5.5 ([#1097](https://github.com/monetr/monetr/issues/1097)) ([53b91e8](https://github.com/monetr/monetr/commit/53b91e84db72f2bc050bc9f7594d0d636c2e46aa))
* **ui:** update dependency @swc/core to v1.3.4 ([#1105](https://github.com/monetr/monetr/issues/1105)) ([ee0afb1](https://github.com/monetr/monetr/commit/ee0afb1bcef26f2121087588e06a17528ff04b5b))
* **ui:** update dependency @testing-library/jest-dom to v5.16.5 ([#1091](https://github.com/monetr/monetr/issues/1091)) ([52dd8b3](https://github.com/monetr/monetr/commit/52dd8b3b9a2b152e9c1df9e98b413fda5f09c4e8))
* **ui:** update dependency eslint to v8.24.0 ([#1107](https://github.com/monetr/monetr/issues/1107)) ([9016877](https://github.com/monetr/monetr/commit/9016877a9a9b51886a9d94242a1b9da88268b636))
* **ui:** update dependency react-router-dom to v6.4.1 ([#1110](https://github.com/monetr/monetr/issues/1110)) ([3c285f5](https://github.com/monetr/monetr/commit/3c285f5b86798d1c74f957d8b1393e1833d974d8))
* **ui:** update dependency rrule to v2.7.1 ([#1004](https://github.com/monetr/monetr/issues/1004)) ([1edddb3](https://github.com/monetr/monetr/commit/1edddb307854ddc61eac7fa114cae009dc10eeef))
* **ui:** update dependency sass to v1.55.0 ([#1111](https://github.com/monetr/monetr/issues/1111)) ([73d1686](https://github.com/monetr/monetr/commit/73d1686cb1c70be8c68521466f86cb0dc527960c))
* **ui:** update dependency ts-loader to v9.4.1 ([#1112](https://github.com/monetr/monetr/issues/1112)) ([a5b5e35](https://github.com/monetr/monetr/commit/a5b5e35dfed0b24c4183f6034dfbe87dad744a16))
* **ui:** update dependency typescript to v4.8.4 ([#1100](https://github.com/monetr/monetr/issues/1100)) ([0ade516](https://github.com/monetr/monetr/commit/0ade5161ed3e92f7e48094a89ad4818c706bd2bf))
* **ui:** update dependency webpack-dev-server to v4.11.1 ([#1113](https://github.com/monetr/monetr/issues/1113)) ([1a19d30](https://github.com/monetr/monetr/commit/1a19d30e8e4a3937f4ed70418d43110356bdf813))
* **ui:** update jest monorepo to v29 ([#1116](https://github.com/monetr/monetr/issues/1116)) ([269e165](https://github.com/monetr/monetr/commit/269e16537f786abc5fbb244a997401e18b7c6ca8))


### Miscellaneous

* Adding NOTICE ([f55a226](https://github.com/monetr/monetr/commit/f55a226d805781d6b731b72a187dd71784d12bf6))
* **deps:** Fixing yarn.lock ([665f5b4](https://github.com/monetr/monetr/commit/665f5b47d0fa6b36df60489ed7afc2524b99fc17))
* Fixing gofakeit race condition. ([83d86c5](https://github.com/monetr/monetr/commit/83d86c591acf2b00018ed713983f8fb636a44768))
* **pubsub:** Tweaking test to be more correct. ([cee8384](https://github.com/monetr/monetr/commit/cee83844f20ab3850ebc091f70dc375772839df5))
* Updated NOTICE ([398bec3](https://github.com/monetr/monetr/commit/398bec3aed76fc3e8c0de362e1234ad33da5160a))
* Updating NOTICE ([7401ed3](https://github.com/monetr/monetr/commit/7401ed3a4bbaa36e54b2eac635c84012d7c97348))
* Upgraded local development node to the latest version. ([6c323a5](https://github.com/monetr/monetr/commit/6c323a53e230c03e49c863c2a140bf1bf56d12d0))

## [0.11.16](https://github.com/monetr/monetr/compare/v0.11.15...v0.11.16) (2022-10-01)


### Bug Fixes

* **spending:** Fixed actual spending calculation issue. ([599abc1](https://github.com/monetr/monetr/commit/599abc1bb4b3f87895f4465c277be004ca8f58c2))


### Dependencies

* **api:** update module github.com/kataras/iris/v12 to v12.2.0-beta5 ([#1090](https://github.com/monetr/monetr/issues/1090)) ([5f64bf8](https://github.com/monetr/monetr/commit/5f64bf8fc8f73d7c6b8107c52c55e3118b186d56))

## [0.11.15](https://github.com/monetr/monetr/compare/v0.11.14...v0.11.15) (2022-09-30)


### Miscellaneous

* **api:** Moving request logger into API scope. ([e0d33ae](https://github.com/monetr/monetr/commit/e0d33aeec43575b14a09a1826bfb37f9d6ed5e0b))

## [0.11.14](https://github.com/monetr/monetr/compare/v0.11.13...v0.11.14) (2022-09-30)


### Bug Fixes

* **test:** Fixed failing test build due to bad embed. ([3513134](https://github.com/monetr/monetr/commit/3513134fcd985b1a252825d9b43ecd6aa5dd4191))
* **ui:** Fixed warning about `ReactDOM.render`. ([ef21ddc](https://github.com/monetr/monetr/commit/ef21ddc7e903b909ed815778eb9dea03f3ee2b9c))


### Dependencies

* **ui:** update babel monorepo to v7.19.3 ([#1086](https://github.com/monetr/monetr/issues/1086)) ([d24396c](https://github.com/monetr/monetr/commit/d24396c3d51892eb0f1bf6838112cd2c62d693e6))

## [0.11.13](https://github.com/monetr/monetr/compare/v0.11.12...v0.11.13) (2022-09-29)


### Features

* **ui:** Implemented swc-loader to improve build times. ([#1087](https://github.com/monetr/monetr/issues/1087)) ([c40b6ee](https://github.com/monetr/monetr/commit/c40b6ee507bf1c3b605d35138e1e549b233d3287))


### Bug Fixes

* **api:** Improve some logging around availability. ([ef4907a](https://github.com/monetr/monetr/commit/ef4907a247ffc8f3a9d9dd944df7c03eeb017d88))

## [0.11.12](https://github.com/monetr/monetr/compare/v0.11.11...v0.11.12) (2022-09-29)


### Bug Fixes

* **plaid:** Fixing syncing after closing bank account. ([#1083](https://github.com/monetr/monetr/issues/1083)) ([7019c41](https://github.com/monetr/monetr/commit/7019c413150806d772d4beca6f1f083537d0359b))


### Dependencies

* **api:** update github.com/iris-contrib/middleware/cors digest to d3e7dfb ([#1057](https://github.com/monetr/monetr/issues/1057)) ([aa34927](https://github.com/monetr/monetr/commit/aa349275020d2240622592e028f683dfffe75dd3))
* **api:** update module github.com/spf13/cobra to v1.5.0 ([#991](https://github.com/monetr/monetr/issues/991)) ([31bf7d9](https://github.com/monetr/monetr/commit/31bf7d9fddcf9c082ca95d16ecf66640ea05ceff))
* **containers:** update squidfunk/mkdocs-material docker tag to v8.5.3 ([#1051](https://github.com/monetr/monetr/issues/1051)) ([d9512fe](https://github.com/monetr/monetr/commit/d9512fe8038bea2361b12b8eaa4e7026cc38a663))

## [0.11.11](https://github.com/monetr/monetr/compare/v0.11.10...v0.11.11) (2022-09-27)


### Bug Fixes

* **dev:** Fixed local dev build failing. ([#1079](https://github.com/monetr/monetr/issues/1079)) ([e44c630](https://github.com/monetr/monetr/commit/e44c630d8d82382a571819ce5e20e6919d9b24b1)), closes [#1047](https://github.com/monetr/monetr/issues/1047)
* **spending:** Fix spending contribution calculation. ([#1078](https://github.com/monetr/monetr/issues/1078)) ([fe53024](https://github.com/monetr/monetr/commit/fe53024966b05d7b20ff21c41e75592a569db64a)), closes [#1077](https://github.com/monetr/monetr/issues/1077)


### Miscellaneous

* Fix Sentry version ([370dd6c](https://github.com/monetr/monetr/commit/370dd6c91eac6c58474cbe6c2d805b87fa41d1ca))
* General code cleanup for UI backend. ([f0706a6](https://github.com/monetr/monetr/commit/f0706a6d9359928b82cee6b4958a9df8b6831f1c))
* **ui:** Increased refetch time for react-query. ([e020c57](https://github.com/monetr/monetr/commit/e020c57678ef7cd3de03510988eea9d9e29f819e))

## [0.11.10](https://github.com/monetr/monetr/compare/v0.11.9...v0.11.10) (2022-09-25)


### Bug Fixes

* **csp:** Fixing how CSP headers are set for UI serving. ([812139b](https://github.com/monetr/monetr/commit/812139b7c793aca8d01b4ad552e8f297d0d87fc1)), closes [#1028](https://github.com/monetr/monetr/issues/1028)
* **sentry:** Allow Sentry feedback dialog in CSP. ([4d64571](https://github.com/monetr/monetr/commit/4d64571a26e6c3f97e2279294ca6bf4e39b1526b))
* **sentry:** Allow Sentry URLs in the CSP. ([7b5cc1d](https://github.com/monetr/monetr/commit/7b5cc1d5332b585edafcce6c3ae2a4a964ee6315))
* **ui:** Make sure sentry.io is on the CSP list. ([dd9875a](https://github.com/monetr/monetr/commit/dd9875a726097ce2c1378c142e7acd2175dcb5b3))


### Miscellaneous

* Updated README. ([8128db8](https://github.com/monetr/monetr/commit/8128db89cea51ec6dfdc706366f2f805aafc3cd7))

## [0.11.9](https://github.com/monetr/monetr/compare/v0.11.8...v0.11.9) (2022-09-24)


### Bug Fixes

* **cloud:** Fixed cloud based request IDs. ([#1068](https://github.com/monetr/monetr/issues/1068)) ([c68ee1a](https://github.com/monetr/monetr/commit/c68ee1a9ee438ef2e5dbfff401204e74ee361b75)), closes [#1067](https://github.com/monetr/monetr/issues/1067)
* **docs:** Fixed sign in URL on the documentation site. ([#1071](https://github.com/monetr/monetr/issues/1071)) ([7d5ab20](https://github.com/monetr/monetr/commit/7d5ab20a21bf4b787f6dc4a72cc1cfa0349e9e47)), closes [#1070](https://github.com/monetr/monetr/issues/1070)

## [0.11.8](https://github.com/monetr/monetr/compare/v0.11.7...v0.11.8) (2022-09-24)


### Features

* **cli:** Adding `test-kms` command. ([ca8ab76](https://github.com/monetr/monetr/commit/ca8ab769dbc41b0c2d2fc66b0ca424a1a811d629))
* **monetr:** General production preperation. ([#1054](https://github.com/monetr/monetr/issues/1054)) ([572ac15](https://github.com/monetr/monetr/commit/572ac15af3e72d322e116f7a1716e1cb7baef029))
* **ui:** Added edit expense funding schedule dialog. ([#1063](https://github.com/monetr/monetr/issues/1063)) ([2d3de55](https://github.com/monetr/monetr/commit/2d3de55f37f9ab976a35064fd46e58a50d71e2a8)), closes [#276](https://github.com/monetr/monetr/issues/276)


### Bug Fixes

* **api:** Don't return internal errors to the client. ([#1062](https://github.com/monetr/monetr/issues/1062)) ([35098a2](https://github.com/monetr/monetr/commit/35098a2b75657d40a790be99d70009dd1ddd8bfa)), closes [#95](https://github.com/monetr/monetr/issues/95)
* **api:** Fixed parsing of `X-Forwarded-For` ([#1060](https://github.com/monetr/monetr/issues/1060)) ([732ee19](https://github.com/monetr/monetr/commit/732ee19fbd0f0414dc6e9bcabccb2ddee05350ac)), closes [#1052](https://github.com/monetr/monetr/issues/1052)
* **ci:** Fixed uploading sourcemaps to sentry. ([58b7f69](https://github.com/monetr/monetr/commit/58b7f694cca758c28c10d6682aed2190cb0c02ab))
* **ci:** Fixing image tag for Google Cloud Registry. ([8c9faec](https://github.com/monetr/monetr/commit/8c9faecfbdfafcfe5e846fddd6265476acff4419))
* **funding:** Fixed funding schedule calculation. ([#1065](https://github.com/monetr/monetr/issues/1065)) ([ffe0065](https://github.com/monetr/monetr/commit/ffe006506b69df2e3157c5dec1cf006444b6a6e0)), closes [#1064](https://github.com/monetr/monetr/issues/1064)
* **plaid:** Fixed bug with link token caching. ([#1061](https://github.com/monetr/monetr/issues/1061)) ([62bcfff](https://github.com/monetr/monetr/commit/62bcfff2b7972f4cba794f794a67814fe7b538a7)), closes [#1025](https://github.com/monetr/monetr/issues/1025)
* **plaid:** Fixed failure trying to reauthenticate account. ([#1059](https://github.com/monetr/monetr/issues/1059)) ([23b8d86](https://github.com/monetr/monetr/commit/23b8d86ed6ef57ed5e1f767ee259b783b2a5209a)), closes [#1058](https://github.com/monetr/monetr/issues/1058)
* **ui:** Fixed bug in bank account sink. ([#1066](https://github.com/monetr/monetr/issues/1066)) ([ec308e8](https://github.com/monetr/monetr/commit/ec308e8ea0e2cb1dc9415422b6707b881a6c5cd5)), closes [#1053](https://github.com/monetr/monetr/issues/1053)


### Miscellaneous

* **ci:** Deploy my.monetr.dog on main. ([5e011d6](https://github.com/monetr/monetr/commit/5e011d636e08f3b44d652b7656b9a379405844cb))
* **ci:** Making tweaks to CI to build image for each push. ([51cfa28](https://github.com/monetr/monetr/commit/51cfa283d1b4232b06191e042c90834c77423de9))
* **ci:** Push unstable images to Google Container Registry. ([5aad674](https://github.com/monetr/monetr/commit/5aad67437b85f4a25c276669787d8f5ab1ea2a3c))
* **ci:** Switch release CI over to deploy to production. ([fc9aaa5](https://github.com/monetr/monetr/commit/fc9aaa58f75ee5a54c59384ceeeebcea722da172))
* **deps:** Bump mkdocs-material (insiders) to the latest version. ([ff937a2](https://github.com/monetr/monetr/commit/ff937a2a2e31848acdab78be83b9233a1836a397))
* **docs:** Fix indention on authentication docs. ([966dc39](https://github.com/monetr/monetr/commit/966dc39a33dea56f0e6e774cfc8d2449954227af))
* **dog:** Changed repository base for containers in my.monetr.dog ([25ad20a](https://github.com/monetr/monetr/commit/25ad20a023d35de4c5a8f420967591c62540439d))

## [0.11.7](https://github.com/monetr/monetr/compare/v0.11.6...v0.11.7) (2022-09-16)


### Bug Fixes

* **plaid:** Updated account selections are now reflected. ([#1049](https://github.com/monetr/monetr/issues/1049)) ([aeada29](https://github.com/monetr/monetr/commit/aeada29c899b0ca700ca4177da4305a3dff214ad))


### Dependencies

* **containers:** update golang docker tag ([#1040](https://github.com/monetr/monetr/issues/1040)) ([3bf97e1](https://github.com/monetr/monetr/commit/3bf97e1f6ab996bab5946a9e083b36344ec5e6c5))

## [0.11.6](https://github.com/monetr/monetr/compare/v0.11.5...v0.11.6) (2022-09-14)


### Features

* **plaid:** Allow account selection updates at any time. ([23f070e](https://github.com/monetr/monetr/commit/23f070e67dcdf2597e477ef6bd71ec1ca44a667f))


### Miscellaneous

* **ci:** Add timeout to local development build task. ([a3282f5](https://github.com/monetr/monetr/commit/a3282f547975d131d829cdb8e7368a9431735e8a))


### Dependencies

* **api:** update github.com/iris-contrib/middleware/cors digest to 754509e ([#967](https://github.com/monetr/monetr/issues/967)) ([9fb90be](https://github.com/monetr/monetr/commit/9fb90befa1398c0cff52589ff9cb21193d81896e))
* **api:** update golang.org/x/crypto digest to c86fa9a ([#1041](https://github.com/monetr/monetr/issues/1041)) ([d18454f](https://github.com/monetr/monetr/commit/d18454f872fa6aa8ddac93c4ec94b92450f66a5b))
* **api:** update module github.com/brianvoe/gofakeit/v6 to v6.18.0 ([#989](https://github.com/monetr/monetr/issues/989)) ([190a4e6](https://github.com/monetr/monetr/commit/190a4e654d326c384838aa7a1fae094275617945))
* **api:** update module github.com/golang-jwt/jwt/v4 to v4.4.2 ([#694](https://github.com/monetr/monetr/issues/694)) ([0e7c856](https://github.com/monetr/monetr/commit/0e7c856f947b82e1eeb7900f0bf87d0adc0d7bde))
* **api:** update module github.com/sirupsen/logrus to v1.9.0 ([#990](https://github.com/monetr/monetr/issues/990)) ([8c5efac](https://github.com/monetr/monetr/commit/8c5efac35faa4bcb36876cf88f2c61f69d4a91e9))
* **api:** update module github.com/stretchr/testify to v1.8.0 ([#992](https://github.com/monetr/monetr/issues/992)) ([7b86b2e](https://github.com/monetr/monetr/commit/7b86b2ea6b3573ed4fb7279b68cfdd4508a18e01))
* **api:** update module github.com/stripe/stripe-go/v72 to v72.122.0 ([#993](https://github.com/monetr/monetr/issues/993)) ([b9e2669](https://github.com/monetr/monetr/commit/b9e266991e6be594b53b3473c902262576a33e41))
* **containers:** update squidfunk/mkdocs-material docker tag to v8.5.0 ([#1045](https://github.com/monetr/monetr/issues/1045)) ([5071e5a](https://github.com/monetr/monetr/commit/5071e5a100325284f695a4b7d808ac12aa2cd66e))
* **go:** Bumped bluemonday to v1.0.20 ([fdf967c](https://github.com/monetr/monetr/commit/fdf967c2fa68e6704375977addf5ce0748c08454))
* **go:** Upgrade to Golang 1.19.1 ([#1039](https://github.com/monetr/monetr/issues/1039)) ([d6c8545](https://github.com/monetr/monetr/commit/d6c8545854bd12cab12bfed6242a534a1193666b))
* **ui:** update babel monorepo to v7.19.0 ([#964](https://github.com/monetr/monetr/issues/964)) ([d324332](https://github.com/monetr/monetr/commit/d3243328efe6daa0b7cc8fc4d999c4459620e425))
* **ui:** update dependency @testing-library/react to v13 ([#836](https://github.com/monetr/monetr/issues/836)) ([07d5f43](https://github.com/monetr/monetr/commit/07d5f4387d47f382f84a87b93f8e8198ae054dd3))
* **ui:** update dependency @testing-library/user-event to v14 ([#837](https://github.com/monetr/monetr/issues/837)) ([2d73431](https://github.com/monetr/monetr/commit/2d73431481c085976e701f5cc6c77744cef11dc3))
* **ui:** update dependency @types/react to v18 ([#838](https://github.com/monetr/monetr/issues/838)) ([412c5cd](https://github.com/monetr/monetr/commit/412c5cd53292eaacc90847e7f5c2e058f56e7564))
* **ui:** update dependency axios to v0.27.2 ([3c9928c](https://github.com/monetr/monetr/commit/3c9928c4892f996eaa33ff2fdf7cd17a2f9f693f))
* **ui:** update dependency eslint to v8.23.1 ([#1042](https://github.com/monetr/monetr/issues/1042)) ([959de11](https://github.com/monetr/monetr/commit/959de119956caf412e556085474ba35e00912e8b))
* **ui:** update dependency eslint-plugin-jest to v26.9.0 ([#997](https://github.com/monetr/monetr/issues/997)) ([3172e66](https://github.com/monetr/monetr/commit/3172e66961dc73a64f8b7464fe1b4de60b0237d9))
* **ui:** update dependency eslint-plugin-jsx-a11y to v6.6.1 ([#998](https://github.com/monetr/monetr/issues/998)) ([3c36dd5](https://github.com/monetr/monetr/commit/3c36dd5777d98188570ab1d35ecdd250f79db666))
* **ui:** update dependency eslint-webpack-plugin to v3.2.0 ([#1001](https://github.com/monetr/monetr/issues/1001)) ([f1b2a32](https://github.com/monetr/monetr/commit/f1b2a32e121bc3e31f6952882bf89a1c24c9a6c2))
* **ui:** update dependency postcss to v8.4.16 ([#986](https://github.com/monetr/monetr/issues/986)) ([b413631](https://github.com/monetr/monetr/commit/b413631342a0fe411a2a2cdd2838cf785e24f53f))
* **ui:** update dependency sass to v1.54.9 ([#1043](https://github.com/monetr/monetr/issues/1043)) ([160bc56](https://github.com/monetr/monetr/commit/160bc5687759c979fd9a840672305a4e7e1b73e9))
* **ui:** update dependency typescript to v4.8.3 ([#1044](https://github.com/monetr/monetr/issues/1044)) ([94f6fd1](https://github.com/monetr/monetr/commit/94f6fd1f049a445b6d512124ab8d126bf82f02c1))
* **ui:** update dependency webpack-bundle-analyzer to v4.6.1 ([#1008](https://github.com/monetr/monetr/issues/1008)) ([4cb77ec](https://github.com/monetr/monetr/commit/4cb77ec5290f85c16c8e372d70bf56998bcdad21))
* **ui:** update dependency webpack-manifest-plugin to v5 ([#842](https://github.com/monetr/monetr/issues/842)) ([c553c5a](https://github.com/monetr/monetr/commit/c553c5a63a2cad0d746c09d4b63e526d0e3c46fc))
* **ui:** update react monorepo to v18 ([#843](https://github.com/monetr/monetr/issues/843)) ([7704445](https://github.com/monetr/monetr/commit/7704445eac61ceae63ec96e0044e98579c3dafcb))

## [0.11.5](https://github.com/monetr/monetr/compare/v0.11.4...v0.11.5) (2022-09-13)


### Features

* **spending:** Significantly improve spending calculations. ([#1031](https://github.com/monetr/monetr/issues/1031)) ([0b8b597](https://github.com/monetr/monetr/commit/0b8b5972d2b78a8e0e4a77a65ebd440b2895f80e))


### Bug Fixes

* **ui:** Fix transactions page crash. ([#1033](https://github.com/monetr/monetr/issues/1033)) ([7942082](https://github.com/monetr/monetr/commit/7942082cf361931cac5cde6aa99c75b9479517c0)), closes [#1032](https://github.com/monetr/monetr/issues/1032)

## [0.11.4](https://github.com/monetr/monetr/compare/v0.11.3...v0.11.4) (2022-09-12)


### Features

* **dev:** Allow Google Cloud KMS for local development. ([#1023](https://github.com/monetr/monetr/issues/1023)) ([636e71b](https://github.com/monetr/monetr/commit/636e71b724b3267e7b1cd15d977f933b405ba9ae))
* **legal:** Adding references to Terms of Use and Privacy Policy. ([#1029](https://github.com/monetr/monetr/issues/1029)) ([124279e](https://github.com/monetr/monetr/commit/124279e9b990e2a025c230b1a22e0055f9da484c))
* **plaid:** Implemented Plaid Account Select V2. ([#1022](https://github.com/monetr/monetr/issues/1022)) ([b9ab47c](https://github.com/monetr/monetr/commit/b9ab47cc199e690db15d4042966fe4126140e51e)), closes [#780](https://github.com/monetr/monetr/issues/780)


### Dependencies

* **ui:** update dependency eslint to v8.23.0 ([#996](https://github.com/monetr/monetr/issues/996)) ([d9c2a46](https://github.com/monetr/monetr/commit/d9c2a463ad8fb50d96b98dff2b439cde7b6cbf40))
* **ui:** update dependency react-select to v5.4.0 ([#1003](https://github.com/monetr/monetr/issues/1003)) ([a6d879a](https://github.com/monetr/monetr/commit/a6d879a5424bc2d329c730bdf736b69066a53712))


### Miscellaneous

* **api:** Deprecate old login function. ([#1021](https://github.com/monetr/monetr/issues/1021)) ([88111bd](https://github.com/monetr/monetr/commit/88111bd4109aa2986c3361992e4cbf56d80f3867))
* **icons:** Include simple icons `LICENSE.md` file in embed. ([a0f045f](https://github.com/monetr/monetr/commit/a0f045fe4756dcbc6d69b70e53dd434fca5752de))
* **kms:** Add Google KMS to admin commands. ([e879584](https://github.com/monetr/monetr/commit/e879584933b1a1d7ef20ea242d3d72665f8f5268))
* **ui:** General code cleanup. ([0ad45de](https://github.com/monetr/monetr/commit/0ad45de96020ae4bddce05a699de739a251cdf76))

## [0.11.3](https://github.com/monetr/monetr/compare/v0.11.2...v0.11.3) (2022-09-09)


### Features

* **bcrypt:** Use bcrypt for passwords going forward. ([#1017](https://github.com/monetr/monetr/issues/1017)) ([3db5e40](https://github.com/monetr/monetr/commit/3db5e4015d1368d8cfc56132d9091e0771b84e5c))


### Bug Fixes

* **docs:** Fixed table rendering in documentation. ([#1018](https://github.com/monetr/monetr/issues/1018)) ([cac54e5](https://github.com/monetr/monetr/commit/cac54e50b215bf02a4f4683795731097bc20a36b)), closes [#974](https://github.com/monetr/monetr/issues/974)
* **ui:** Improved `Cache-Control` of UI content files. ([81e86e4](https://github.com/monetr/monetr/commit/81e86e47ad6f85db3ebaf3e2b77cd36fa303d546))


### Dependencies

* **ui:** update dependency react-refresh to v0.14.0 ([#1002](https://github.com/monetr/monetr/issues/1002)) ([a11f8cf](https://github.com/monetr/monetr/commit/a11f8cf10dcb224d9ec8f7468fc6038e64fb88b9))
* **ui:** update dependency ts-loader to v9.3.1 ([#1006](https://github.com/monetr/monetr/issues/1006)) ([43f938d](https://github.com/monetr/monetr/commit/43f938d098c623e166b19ff94359bb7eda34f1ab))

## [0.11.2](https://github.com/monetr/monetr/compare/v0.11.1...v0.11.2) (2022-08-28)


### Bug Fixes

* **icons:** Further improve icon searching. ([e8d2fe0](https://github.com/monetr/monetr/commit/e8d2fe0e71597816b770a72787aa2bc72809b22e))
* **icons:** Slightly improve icon search. ([4c35527](https://github.com/monetr/monetr/commit/4c35527586a427a34a4daf277b61dbd587efd2cc))
* **ui:** Fixed goals and expenses not being sorted alphabetically. ([2f8b2e2](https://github.com/monetr/monetr/commit/2f8b2e24ba782983032137f659715eeab21bf856)), closes [#978](https://github.com/monetr/monetr/issues/978)


### Miscellaneous

* Updated bluemonday because of retraction. ([9d66f35](https://github.com/monetr/monetr/commit/9d66f3570b4216b1c32e308a4f3a9d400f81f413))


### Dependencies

* **api:** update google.golang.org/genproto digest to 9e6da59 ([#980](https://github.com/monetr/monetr/issues/980)) ([65ce471](https://github.com/monetr/monetr/commit/65ce4718736922b411770d530c63b85c8d1cc54f))
* **api:** update module github.com/aws/aws-sdk-go to v1.44.86 ([#981](https://github.com/monetr/monetr/issues/981)) ([9f980fc](https://github.com/monetr/monetr/commit/9f980fc8fd569350b5899680141f0213758c8a1b))
* **renovate:** update jamesives/github-pages-deploy-action action to v4.4.0 ([#994](https://github.com/monetr/monetr/issues/994)) ([eb8b1b7](https://github.com/monetr/monetr/commit/eb8b1b73fcfb2619bab7046fddaf506dbedbf741))
* **ui:** update dependency @fontsource/roboto to v4.5.8 ([#984](https://github.com/monetr/monetr/issues/984)) ([43cd670](https://github.com/monetr/monetr/commit/43cd670305f8dc208b8df7dfe9d998b91f2abd25))
* **ui:** update dependency @types/react to v17.0.48 ([#985](https://github.com/monetr/monetr/issues/985)) ([49f19f0](https://github.com/monetr/monetr/commit/49f19f095d4f156ffda8b45566e528906c84a7ed))
* **ui:** update dependency postcss-loader to v7 ([#1014](https://github.com/monetr/monetr/issues/1014)) ([9144fd5](https://github.com/monetr/monetr/commit/9144fd5d8d314e085fef15a6814070952b9eae16))
* **ui:** update dependency sass to v1.54.5 ([#1005](https://github.com/monetr/monetr/issues/1005)) ([9f880d6](https://github.com/monetr/monetr/commit/9f880d6861f299ba541d07f77bb9bb2942775f47))
* **ui:** update dependency sass-loader to v13 ([#1015](https://github.com/monetr/monetr/issues/1015)) ([7d79b7c](https://github.com/monetr/monetr/commit/7d79b7ca62f33d2243721fb233d5ce8b40b2b831))
* **ui:** update dependency typescript to v4.8.2 ([#1007](https://github.com/monetr/monetr/issues/1007)) ([60c81c8](https://github.com/monetr/monetr/commit/60c81c87a3ff54f4389b0e6bd17af7cb9a824319))
* **ui:** update dependency webpack-dev-server to v4.10.0 ([#1009](https://github.com/monetr/monetr/issues/1009)) ([136d731](https://github.com/monetr/monetr/commit/136d731ea3fbe7b9fb7486d755e6a4f6723f7d53))
* **ui:** update dependency zustand to v4.1.1 ([#1010](https://github.com/monetr/monetr/issues/1010)) ([f390a19](https://github.com/monetr/monetr/commit/f390a1902a4cd9078246c0a5165ec57605f30137))
* **ui:** update emotion monorepo to v11.10.0 ([#1011](https://github.com/monetr/monetr/issues/1011)) ([6650b50](https://github.com/monetr/monetr/commit/6650b50f764b421f88cebb274d65d70fe7cda096))

## [0.11.1](https://github.com/monetr/monetr/compare/v0.11.0...v0.11.1) (2022-08-28)


### Bug Fixes

* **icons:** Fixed page crash for icon without color. ([f7f3ab7](https://github.com/monetr/monetr/commit/f7f3ab7d4c5bbea5270cf2bf823645c80938c11d)), closes [#976](https://github.com/monetr/monetr/issues/976)

## [0.11.0](https://github.com/monetr/monetr/compare/v0.10.21...v0.11.0) (2022-08-28)


### ⚠ BREAKING CHANGES

* **ui:** Migration to react-query from redux.

### Features

* **dev:** Adding some stuff for gitpod development. ([8fbe3c6](https://github.com/monetr/monetr/commit/8fbe3c687b10c3f97cf2c3dc9df3f2b0404b8788))
* **github:** Adding support for GitHub Codespaces ([35266f3](https://github.com/monetr/monetr/commit/35266f30ae45ed95c3e605bf67a64cda2695eed9))
* **gitpod:** Adding named ports for gitpod. ([f157554](https://github.com/monetr/monetr/commit/f157554d87d5f07d4193b69cb08613e7f796b123))
* **gitpod:** Further improving support for GitPod. ([a1249d8](https://github.com/monetr/monetr/commit/a1249d8c191e0b48399e3fd33fe5b3cc25937d09))
* **icons:** Improve icon matching, case insensitivity + whitespace. ([eb51ce8](https://github.com/monetr/monetr/commit/eb51ce803f5265e104369a68f4d82ea83cd95136))
* **ui:** Migration to react-query from redux. ([9552ff5](https://github.com/monetr/monetr/commit/9552ff5ec1792486f5907e93767f087d0bff71f8))
* **ui:** Moving more things to react-query, general cleanup. ([1c72463](https://github.com/monetr/monetr/commit/1c72463a5085d9d1720045f1cc6230920410784d))


### Bug Fixes

* **billing:** Resolved Stripe subscription delay issue. ([796d13c](https://github.com/monetr/monetr/commit/796d13ce382d601ca416d805c3aa6579bed80d0d)), closes [#966](https://github.com/monetr/monetr/issues/966)
* **dev:** Fixed GitPod/CodeSpaces development URLs ([ebc6b59](https://github.com/monetr/monetr/commit/ebc6b59851e357addf00378e5c87634c12e79b67))
* **icons:** Fixing icon search query. ([ea49963](https://github.com/monetr/monetr/commit/ea499635f9b529f7f88f8076cbe2d3ac61f76a73)), closes [#961](https://github.com/monetr/monetr/issues/961)
* **job:** Improving debugging around removing transactions. ([31efb5c](https://github.com/monetr/monetr/commit/31efb5c6b8f3ffbcb7a0384720ab4007e82b2f69))
* **login:** Fixing cookie not being cleared after server error. ([d4455c0](https://github.com/monetr/monetr/commit/d4455c0b2d9ff7f4c1c2fad5179ce10ae40c91de))
* **logs:** Fixed stackdriver logs not being formatted correctly. ([9f31299](https://github.com/monetr/monetr/commit/9f3129958e49e2ef94497e1daa7a3595b6d3f89d))


### Dependencies

* **icons:** Bump simple-icons to 7.7.0 ([b9c0304](https://github.com/monetr/monetr/commit/b9c030497374816674c094da389641f5fe3e4d0d))
* **ui:** update dependency tailwindcss to v3.1.8 ([a155cf1](https://github.com/monetr/monetr/commit/a155cf17224dd05355606ca167a839eae6523fb7))
* **ui:** update dependency terser-webpack-plugin to v5.3.3 ([a07505e](https://github.com/monetr/monetr/commit/a07505e8e4c61cc0cd2a08f745e4b22de14304f3))
* **ui:** update dependency webpack to v5.74.0 ([14cae1a](https://github.com/monetr/monetr/commit/14cae1a7080f035eb17510dba1eb9f21e966def3))
* **ui:** update dependency webpack-cli to v4.10.0 ([8210afa](https://github.com/monetr/monetr/commit/8210afaf95daf01fa2c30698c4c0dd60f879ec6d))
* **ui:** update dependency webpack-dev-server to v4.9.3 ([efb5254](https://github.com/monetr/monetr/commit/efb5254681070219d0b971d3670c99670cd093b9))
* **ui:** update emotion monorepo to v11.9.3 ([d060463](https://github.com/monetr/monetr/commit/d060463b16f1aa2f14c8977e725f1ad6ac2392b5))


### Miscellaneous

* Add Open in GitPod button. ([b0733cd](https://github.com/monetr/monetr/commit/b0733cded49d4d283eb36c3a577b6c4bd6058225))
* **deps:** Bump terser from 5.9.0 to 5.14.2 ([776891c](https://github.com/monetr/monetr/commit/776891c4c4410781640057c09e0b49f79b9d06b1))
* **deps:** Removed unused dependencies. ([7110064](https://github.com/monetr/monetr/commit/711006421feb78fc7d3bad91e8d451ab0ecd8be9))
* **deps:** Removing `reselect` dependency. ([566b484](https://github.com/monetr/monetr/commit/566b48466489774862ec5770d51196ce8f73ea81))
* **docs:** Updated screenshot in README ([b8bb345](https://github.com/monetr/monetr/commit/b8bb3456c8f19a74ac6d877113ed7d03a3f5d35f))
* Include version information for icons, hook improvement. ([b8698d6](https://github.com/monetr/monetr/commit/b8698d6276f69edf32eddbfca3794e33cbd99214))
* **lint:** Linting fixes + react-query buy-in ([bd389bd](https://github.com/monetr/monetr/commit/bd389bddf28e12b1d820693aff9f4acdd91934c4))
* release 0.11.0 ([8948829](https://github.com/monetr/monetr/commit/8948829b302cd79b7391ec08e1a73f0b9123f610))
* Start in GitPod immediately ([9a18cc1](https://github.com/monetr/monetr/commit/9a18cc1b894c903a9bad984a96a6989b36d3553f))
* Update readme with new dev instructions ([d164a30](https://github.com/monetr/monetr/commit/d164a301bda17af494ac5e6bb86879d9e8bb0c39))

## [0.10.21](https://github.com/monetr/monetr/compare/v0.10.20...v0.10.21) (2022-07-16)


### Bug Fixes

* **container:** Make sure icons make their way into the container. ([7ae7166](https://github.com/monetr/monetr/commit/7ae7166bdd2174e86f11dbb98090451cf369a7b7))
* **icons:** Don't request icons if they are not embedded. ([16cd6bd](https://github.com/monetr/monetr/commit/16cd6bd94e3d674275ad15f0b46a2ae24625922e))

## [0.10.20](https://github.com/monetr/monetr/compare/v0.10.19...v0.10.20) (2022-07-16)


### Features

* **helm:** Getting ready to rollout KMS. ([6070c48](https://github.com/monetr/monetr/commit/6070c4844e18c4984037eddc78bd9c20afeda074))
* **icons:** Adding basic icon support. ([bb4b105](https://github.com/monetr/monetr/commit/bb4b105c6de258f8e1f355a117eb700b0556a10c))
* **icons:** Show icons on transaction rows. ([3b4d961](https://github.com/monetr/monetr/commit/3b4d9611b533af5f93e8ab4f821f43a8a2786f9d))
* **settings:** Added settings read endpoint, tests, docs. ([6ca91c1](https://github.com/monetr/monetr/commit/6ca91c1a3b5294b94d2b2df30441c39a1db2ba03))


### Bug Fixes

* **test:** Fixed flaky funding schedule test. ([9d004fa](https://github.com/monetr/monetr/commit/9d004fa3ea8a5084e8944819fc906c5a9967f1d0))
* **test:** Fixed tests around funding schedules. ([65bdf76](https://github.com/monetr/monetr/commit/65bdf761182c256c211f3ab1754b231daf7954f2))
* **ui:** Fixed types for retrieving icons. ([093e3c4](https://github.com/monetr/monetr/commit/093e3c45ad6067e2c92756cd0c3127398f41123e))


### Miscellaneous

* Fix failing test, updating container image to latest debian. ([d06d0dc](https://github.com/monetr/monetr/commit/d06d0dcf84fc647cf2601fbe34ce12073f6bee42))
* Fixing renovate ([31ce92c](https://github.com/monetr/monetr/commit/31ce92cb0a6fc1cb368b0ffd48a6c1cdd9b4eb11))
* **icons:** Pin simple icons version. ([7d8b6c8](https://github.com/monetr/monetr/commit/7d8b6c813ea66fb06ae423672e51414060239d36)), closes [#947](https://github.com/monetr/monetr/issues/947)
* **lint:** Fixing the ESLint config and fixing many files. ([78f0d1a](https://github.com/monetr/monetr/commit/78f0d1a5ea30919c3fc3f7b985cadf3104c87dbd))


### Dependencies

* **api:** update module github.com/aws/aws-sdk-go to v1.44.56 ([940308f](https://github.com/monetr/monetr/commit/940308f4249c8eafb8e04228e63a5a51c5265bd9))
* **api:** update module github.com/gomodule/redigo to v1.8.9 ([aa527e0](https://github.com/monetr/monetr/commit/aa527e035579b628a565b8451066082991fa8772))
* **renovate:** update jamesives/github-pages-deploy-action action to v4.3.4 ([64c5fdc](https://github.com/monetr/monetr/commit/64c5fdc2b056e101d9b7c980a9f8f6984dfb3501))
* **ui:** pin dependency react-query to 3.39.2 ([1020d03](https://github.com/monetr/monetr/commit/1020d030423d282e4f7525054e4d77b97e5c1c33))
* **ui:** update babel monorepo to v7.18.6 ([6946af0](https://github.com/monetr/monetr/commit/6946af0a75d2310a466e76ccf81efb4ec407b13b))
* **ui:** update dependency @types/react to v17.0.47 ([42d3f07](https://github.com/monetr/monetr/commit/42d3f0758c6063df1900e634d192438be4521368))
* **ui:** update dependency jest-mock-axios to v4.6.1 ([e7c927a](https://github.com/monetr/monetr/commit/e7c927a841f2a9229e0945bd6db12816639cace0))
* **ui:** update dependency moment to v2.29.4 [security] ([383b61c](https://github.com/monetr/monetr/commit/383b61c1e4f85ce839c8d360e0fff2a1c64ec7e7))
* **ui:** update dependency react-refresh-typescript to v2.0.7 ([662b3d5](https://github.com/monetr/monetr/commit/662b3d55842bf2ffa1e24183d742f2ac5cf98fb6))

## [0.10.19](https://github.com/monetr/monetr/compare/v0.10.18...v0.10.19) (2022-06-15)


### Bug Fixes

* **spending:** Fixed divide by zero bug in calculation. ([c39fe7a](https://github.com/monetr/monetr/commit/c39fe7a10710e0dbb2377ec7c525be10e43e15d2)), closes [#940](https://github.com/monetr/monetr/issues/940)

## [0.10.18](https://github.com/monetr/monetr/compare/v0.10.17...v0.10.18) (2022-06-15)


### Bug Fixes

* **spending:** Fixed contribution count being off. ([082f29b](https://github.com/monetr/monetr/commit/082f29ba93a47d753ff9a54befd2a3ee4dbada89)), closes [#937](https://github.com/monetr/monetr/issues/937)

## [0.10.17](https://github.com/monetr/monetr/compare/v0.10.16...v0.10.17) (2022-06-15)


### Features

* Added global `--config` flag. ([d5aa091](https://github.com/monetr/monetr/commit/d5aa091829b19b0d2a3190a9a38dc1cf2f35be4f)), closes [#933](https://github.com/monetr/monetr/issues/933)
* **cli:** Adding command to help debug secrets. ([cd2e584](https://github.com/monetr/monetr/commit/cd2e58418ed34db794d8febb14c208270d556626))
* **kms:** Adding command to migrate secrets from vault. ([156bce2](https://github.com/monetr/monetr/commit/156bce2a17530144a4b3122e4e321a2101d13b72))
* **kms:** Adding support for decrypting using Google KMS. ([b2dc807](https://github.com/monetr/monetr/commit/b2dc8078a2fd42bf0ec27af55a2cd3e0fc0c5afd))
* **kms:** Check permissions when initializing Google KMS. ([b1133e6](https://github.com/monetr/monetr/commit/b1133e60f6cc7c0b6fcdf8749ff73894862555b4))
* **kms:** Laying groundwork for key-management-system. ([b94d5c9](https://github.com/monetr/monetr/commit/b94d5c9ca4e731411bc4bbdada9a1dbc71e2cdcb))
* **kms:** Laying groundwork for managing secrets via CLI. ([70ab0f3](https://github.com/monetr/monetr/commit/70ab0f316a95aa21fef080bf35ae73dbfd05bbda))
* **kms:** Use KMS provider from configuration. ([e88d2c6](https://github.com/monetr/monetr/commit/e88d2c64f075f60d0ff32c29094a916487103ab2))
* **local:** Allow Golang hot reload to be disabled. ([567fceb](https://github.com/monetr/monetr/commit/567fcebe231649bfc77ebc55938fdf7d9ddde23c))


### Bug Fixes

* **bug:** Spending objects will no longer have negative contributions. ([4d487ad](https://github.com/monetr/monetr/commit/4d487ad7fd80ebd0e6b40a988221828884e5e7e8)), closes [#930](https://github.com/monetr/monetr/issues/930)
* **docs:** Fixed build failing because of missing plugin. ([6aae759](https://github.com/monetr/monetr/commit/6aae759d907c5071a72a015b022830ecfe00b73a))
* **spending:** Fixed spending contribution miscalculation. ([4433ac3](https://github.com/monetr/monetr/commit/4433ac392eaf00a0b90f4b9504f6da45936d56b8)), closes [#937](https://github.com/monetr/monetr/issues/937)


### Dependencies

* **api:** update github.com/iris-contrib/middleware/cors digest to e50b808 ([10e3c5c](https://github.com/monetr/monetr/commit/10e3c5cfaeb6f595c27955366534c33261c60ec6))
* **docs:** Bumped mkdocs material to its latest version. ([b7d6bbb](https://github.com/monetr/monetr/commit/b7d6bbb13c8ee08f09472135d4ea6fc4083f7db7))


### Miscellaneous

* **docs:** Laying groundwork for versioning. ([5d4679f](https://github.com/monetr/monetr/commit/5d4679f3e185120fd8509718dc3a4a1eb1d5bd86))
* General cleanup of unused tools. ([eb2fb5d](https://github.com/monetr/monetr/commit/eb2fb5de764be18b0d0004010dd1ecda80bd3c32))
* **kms:** Progress on AWS and Google KMS implementations. ([ccb0418](https://github.com/monetr/monetr/commit/ccb0418bce10543ffce006b0bb45f63d0845c990))
* Starting to integrate the KMS into the PG secrets provider. ([f58f044](https://github.com/monetr/monetr/commit/f58f044e39da674881ad492aafb9203b78605716))
* Tweaking local dev + kms init. ([a4e80c3](https://github.com/monetr/monetr/commit/a4e80c3c994b610f2aeccd5df9d0aeb2b02749eb))

## [0.10.16](https://github.com/monetr/monetr/compare/v0.10.15...v0.10.16) (2022-06-04)


### Features

* Adding new settings table, adding metrics to improve safe to spend ([c89be4d](https://github.com/monetr/monetr/commit/c89be4d318b5abe94cbf246c1f4b6a26d24dbc75))
* **api:** Added support for updating funding schedules ([63ce246](https://github.com/monetr/monetr/commit/63ce246b6ee2718c5e429346314581bce8ac2686))
* **funding:** Adding basic support for waiting for deposits. ([0ea5d1a](https://github.com/monetr/monetr/commit/0ea5d1a0517263aca60528cd10c9036c68f6ae81))


### Dependencies

* **api:** update module github.com/alicebob/miniredis/v2 to v2.21.0 ([5ea2e5d](https://github.com/monetr/monetr/commit/5ea2e5dd9c4802f8704e3e9415e8a0ea5d58916b))
* **api:** update module github.com/jarcoal/httpmock to v1.2.0 ([bb12fbf](https://github.com/monetr/monetr/commit/bb12fbf8eb7e3c03f7dfcc59ed6e72a2db414556))
* **api:** update module github.com/spf13/viper to v1.12.0 ([d064ea0](https://github.com/monetr/monetr/commit/d064ea0bdec7b8f1edd331550f47805670303888))
* **api:** update module github.com/stripe/stripe-go/v72 to v72.111.0 ([f6ec576](https://github.com/monetr/monetr/commit/f6ec57659fe4f9496c70db78fd4fc6dc1c2ab569))
* **containers:** update dependency squidfunk/mkdocs-material to v8.2.16 ([4df6fb3](https://github.com/monetr/monetr/commit/4df6fb34716da08e3e3ef2aa4769295bf59904ec))
* **ui:** update dependency @date-io/moment to v2.14.0 ([ab7c6c4](https://github.com/monetr/monetr/commit/ab7c6c4f136abaec0a4411756ad8b70d0b3ce443))
* **ui:** update dependency webpack to v5.72.1 ([dfe5e92](https://github.com/monetr/monetr/commit/dfe5e9234eb6cfc3ba660eb698f9acf76f90727c))
* **ui:** update sentry-javascript monorepo to v6.19.7 ([466df41](https://github.com/monetr/monetr/commit/466df41aed50f37b75aa0df9b9a3d83ec2f6415e))


### Miscellaneous

* **api:** Improving Links API + docs. ([ae972ba](https://github.com/monetr/monetr/commit/ae972ba0ce8bc5320c9c51eeba0e5504ae382329))
* **dev:** Removing minikube local development. ([2771138](https://github.com/monetr/monetr/commit/2771138d85c7713dfdba95bd7d2005ed62001c92))
* **docs:** Updated local development hostname. ([e1a8724](https://github.com/monetr/monetr/commit/e1a87243fd05574b276c7b26482b519190243be5))

### [0.10.15](https://github.com/monetr/monetr/compare/v0.10.14...v0.10.15) (2022-05-24)


### Features

* **api:** Adding support for removing funding schedules. ([8f194f4](https://github.com/monetr/monetr/commit/8f194f431611c05642c654781f8bdef67f1e4e9a))
* **ci:** Test the local development environment inside CI. ([697ac46](https://github.com/monetr/monetr/commit/697ac4622e368d6b2c9e5f465c070e7837525d66))
* **cli:** Adding tool to export your data to JSON. ([aac02d6](https://github.com/monetr/monetr/commit/aac02d658e46c9cd776f1dc7f4f9ff13b66e42c1))


### Bug Fixes

* **docs:** Fixed "Getting Started" link for development docs. ([7a3347a](https://github.com/monetr/monetr/commit/7a3347aa69b9fe87ab51a6910050718af3994821))
* **test:** Fixing failing test from auth error changes. ([2f24ace](https://github.com/monetr/monetr/commit/2f24ace78835979fb70801ee5b2c234611013ed6))
* **ui:** Fixed crash when trying to edit a goal. ([a067dc3](https://github.com/monetr/monetr/commit/a067dc33ffa404632dc7a8cac57926c9407ece5d))


### Dependencies

* **containers:** update dependency redis to v6.2.7 ([151f0d0](https://github.com/monetr/monetr/commit/151f0d07f49b6b74b9b85cb7bf0831612a55f2ac))
* **containers:** update dependency squidfunk/mkdocs-material to v8.2.15 ([8153709](https://github.com/monetr/monetr/commit/815370944cca5ed4df5886f4e4af63202fd3c8da))
* **renovate:** update jamesives/github-pages-deploy-action action to v4.3.3 ([7afdc4d](https://github.com/monetr/monetr/commit/7afdc4d3c993a68b67c198d5c921fe49a11722dd))
* **ui:** update babel monorepo ([b3b9e07](https://github.com/monetr/monetr/commit/b3b9e07d4949debc14a4d917c6ec36bc6a0f279b))
* **ui:** update babel monorepo to v7.17.12 ([8dee34a](https://github.com/monetr/monetr/commit/8dee34a9174a25194be30d518880cdedb4b5a6fa))
* **ui:** update dependency @fontsource/roboto to v4.5.7 ([b33ec2e](https://github.com/monetr/monetr/commit/b33ec2e3907810b7129e861be84d9fd42a2c2e0e))
* **ui:** update dependency @pmmmwh/react-refresh-webpack-plugin to v0.5.6 ([cd05ad3](https://github.com/monetr/monetr/commit/cd05ad3a45e5b34fa448375b7a90566a96ad01f8))
* **ui:** update dependency @types/react to v17.0.45 ([2fab942](https://github.com/monetr/monetr/commit/2fab94227e6dd9d60c2984bae5588466329df723))
* **ui:** update dependency @types/react-dom to v17.0.17 ([639fbd8](https://github.com/monetr/monetr/commit/639fbd81e5d70fb88fea1fbb63db397f36a2b006))
* **ui:** update dependency autoprefixer to v10.4.7 ([2e4d27e](https://github.com/monetr/monetr/commit/2e4d27e37aadee5e82d8f12965f3ecbc2431661e))
* **ui:** update dependency dotenv to v16.0.1 ([3ef493a](https://github.com/monetr/monetr/commit/3ef493a227faa09dadc7e1cfe3733f783013d8e6))
* **ui:** update dependency immer to v9.0.14 ([1c6310f](https://github.com/monetr/monetr/commit/1c6310f4e1c22043c2942f202e85a9df18e91221))
* **ui:** update dependency notistack to v2.0.5 ([e64bbfd](https://github.com/monetr/monetr/commit/e64bbfd7568ad689d566b727a5a1217aaabb1d6c))
* **ui:** update dependency path-to-regexp to v6.2.1 ([616d2d4](https://github.com/monetr/monetr/commit/616d2d4d964a3ba20a82141f339b97df686d6cd4))
* **ui:** update dependency postcss to v8.4.14 ([ded5383](https://github.com/monetr/monetr/commit/ded5383303624e4f650792c1e60360f9c4a6169f))
* **ui:** update dependency react-plaid-link to v3.3.2 ([543513b](https://github.com/monetr/monetr/commit/543513b2a9e3427a0b15b31ed9cefceb930854cc))
* **ui:** update dependency react-select to v5.3.2 ([6279feb](https://github.com/monetr/monetr/commit/6279febc232143b3d8217cbae78c9716065a3882))


### Miscellaneous

* Adding tests for edit goals view. ([63f37d5](https://github.com/monetr/monetr/commit/63f37d5ce676c200d276a3746b78692809386e05))
* **deps:** Bump async from 2.6.3 to 2.6.4 ([093cde9](https://github.com/monetr/monetr/commit/093cde99c0d02f5c094dc3120779484eade7fb75))
* **deps:** Bump ejs from 3.1.6 to 3.1.8 ([b6ee2f0](https://github.com/monetr/monetr/commit/b6ee2f01c37752d0dede0f247e6f623a9bfa1f51))
* **docs:** Adding a not found error to docs for funding schedules. ([7b6e454](https://github.com/monetr/monetr/commit/7b6e454b635b38de7ad28b232e4301fdb04a744b))
* **docs:** Adding documentation on reading funding schedules. ([83c3ff7](https://github.com/monetr/monetr/commit/83c3ff7dc0a8e71283a78a03e1ddf90f8bfc9c35))
* **docs:** Adding github actions docs page. ([96e0983](https://github.com/monetr/monetr/commit/96e0983fdabcb725a6bb170d0ff0973a2beb1d4c))
* **tests:** Adding tests for deleting funding schedules. ([98a9818](https://github.com/monetr/monetr/commit/98a98183089e55052652073eb443deaa47017337))

### [0.10.14](https://github.com/monetr/monetr/compare/v0.10.13...v0.10.14) (2022-05-16)


### Features

* **funding:** Allow funding schedules to exclude weekends. ([ede6ecb](https://github.com/monetr/monetr/commit/ede6ecb8dd8ce710cbd1f04adf8f5ee88d3dade6))


### Bug Fixes

* **ui/transactions:** Fixed padding issue for pending transactions. ([a033ea9](https://github.com/monetr/monetr/commit/a033ea960d34535e0b60ed3d541c28c8ec199251))
* **ui:** Fixed loading state on initial Plaid setup. ([ad74acf](https://github.com/monetr/monetr/commit/ad74acfdadea5e55ca55328effca9249b9a8b981))
* **ui:** Removed some buggy ui code for subscriptions. ([0637009](https://github.com/monetr/monetr/commit/06370098d3b50de45dda42b7e70277457ddc9168))


### Dependencies

* **api:** update module github.com/fsnotify/fsnotify to v1.5.4 ([5a2228c](https://github.com/monetr/monetr/commit/5a2228ce59c9f34feb51f577b40c1685bbf46453))
* **api:** update module github.com/kataras/iris/v12 to v12.2.0-beta2 ([1b12002](https://github.com/monetr/monetr/commit/1b12002510808ec4aa0db44b24fd89706fdfb8c2))
* **api:** update module github.com/nyaruka/phonenumbers to v1.0.75 ([612db24](https://github.com/monetr/monetr/commit/612db24cdb3a667d335dac99056cf68dc2934c32))


### Miscellaneous

* **docs:** Adding basic docs for funding schedules. ([0d5a549](https://github.com/monetr/monetr/commit/0d5a549f2b1fbe719f2680513b357dc93d08059e))
* Updated screenshot. ([01cedb6](https://github.com/monetr/monetr/commit/01cedb6e73cd202760aaa93017d365e2249a7fcf))

### [0.10.13](https://github.com/monetr/monetr/compare/v0.10.12...v0.10.13) (2022-04-18)


### Bug Fixes

* **rrule:** Fixed failing tests due to DTSTART. ([033550a](https://github.com/monetr/monetr/commit/033550afc874103faa9b2a6225c224b98cc27b9a))


### Dependencies

* **api:** update module github.com/brianvoe/gofakeit/v6 to v6.16.0 ([00ed758](https://github.com/monetr/monetr/commit/00ed75846bc9d42d878bd3b7074885e6620f6ee5))
* **api:** update module github.com/kataras/iris/v12 to v12.2.0-beta1 ([3625d66](https://github.com/monetr/monetr/commit/3625d6667aa8650ab177b3b0afcb3749c37a7120))
* **api:** update module github.com/spf13/viper to v1.11.0 ([1bf6059](https://github.com/monetr/monetr/commit/1bf60599b7236b0029c9c342b00534475bc4689d))
* **renovate:** update guyarb/golang-test-annotations action to v0.5.1 ([42527d5](https://github.com/monetr/monetr/commit/42527d56d3425855a943d187255f80ffe2a67161))
* **ui:** update dependency @testing-library/react to v12.1.5 ([86273c2](https://github.com/monetr/monetr/commit/86273c2395664683415c9f5a4c94a598194a59a3))
* **ui:** update dependency @types/react-redux to v7.1.24 ([d83b7c5](https://github.com/monetr/monetr/commit/d83b7c5dfc7046c021b2646ec703a5408a546111))
* **ui:** update dependency eslint-config-react-app to v7.0.1 ([7502cbf](https://github.com/monetr/monetr/commit/7502cbf93413951d8f3fa87addc796ccf6faf7b7))
* **ui:** update dependency eslint-plugin-testing-library to v5.3.1 ([ef6f73b](https://github.com/monetr/monetr/commit/ef6f73b60c5d98ff732076e1f2594c77adf8f656))
* **ui:** update dependency fs-extra to v10.1.0 ([71c7f5b](https://github.com/monetr/monetr/commit/71c7f5b500a1e43d969fe91f73466d0085a43d43))
* **ui:** update dependency moment to v2.29.3 ([5ba62d3](https://github.com/monetr/monetr/commit/5ba62d3fa369c19982995f6ddce9a3b37f329cd0))
* **ui:** update dependency notistack to v2.0.4 ([afb1561](https://github.com/monetr/monetr/commit/afb15610d6e15f7a51b3ddc03d45c4c1603a1fc8))
* **ui:** update dependency react-select to v5.3.0 ([28cf3a1](https://github.com/monetr/monetr/commit/28cf3a177879350df0722b619a105bf3ac4f08f9))
* **ui:** update dependency semver to v7.3.7 ([481b9bd](https://github.com/monetr/monetr/commit/481b9bd39ffc6ee76eff88d85b2f5b75c852bbc2))
* **ui:** update dependency tailwindcss to v3.0.24 ([ee5f3ca](https://github.com/monetr/monetr/commit/ee5f3ca98d3091d282cb0eb4ff9396973e33861e))
* **ui:** update dependency workbox-webpack-plugin to v6.5.3 ([5ae784e](https://github.com/monetr/monetr/commit/5ae784e2894a4a92e7e72a8238bbb586111c94ef))
* **ui:** update material-ui monorepo ([123c042](https://github.com/monetr/monetr/commit/123c042b517b5d3c4f84106a7f2ef3147149f0df))
* **ui:** update typescript-eslint monorepo to v5.19.0 ([10f794b](https://github.com/monetr/monetr/commit/10f794b37279aae808818c3901746f3e4481ced8))

### [0.10.12](https://github.com/monetr/monetr/compare/v0.10.11...v0.10.12) (2022-04-10)


### Features

* **spending:** Completely rewrote spending calculation code. ([50af9fa](https://github.com/monetr/monetr/commit/50af9fa1ba6db290055f9e4a3b591133a8976cbe))
* **spending:** Support more frequent spending than funding. ([f230702](https://github.com/monetr/monetr/commit/f2307027353ab7d7d6ee3a6cd547ae95d465da59)), closes [#141](https://github.com/monetr/monetr/issues/141)


### Bug Fixes

* **docs:** Removed references to the swagger file from documentation. ([9cb2e4b](https://github.com/monetr/monetr/commit/9cb2e4beaa52105b1ab577603dc66986701c2140))
* **sentry:** Fixed sentry transport for new version of SDK. ([6caa36d](https://github.com/monetr/monetr/commit/6caa36dcd6a86d289751dab9cc0e72f70f7efe4a))


### Dependencies

* **ui:** update sentry-javascript monorepo to v6.19.6 ([cb24ace](https://github.com/monetr/monetr/commit/cb24acefcd2cc146416a9444be41ccb68b0583f7))


### Miscellaneous

* **ci:** Remove swagger docs generation in pipeline. ([71697e8](https://github.com/monetr/monetr/commit/71697e8ab5b59e8e65d19c89b574b134531f5190))
* **development:** Adding support for ReCAPTCHA for local development. ([89f6831](https://github.com/monetr/monetr/commit/89f683112565ce6f5bd085ffc08d5eaf10114a48))
* **docs:** Added plaid links to resource index. ([2bf31d3](https://github.com/monetr/monetr/commit/2bf31d337f91372b4f890ffb3207068ea44e574a))
* **docs:** Adding documentation around Plaid Links. ([e80295c](https://github.com/monetr/monetr/commit/e80295c957ec18edba8acb2c7b925bf8b8cd54f3))
* Updated go.mod ([66c52b7](https://github.com/monetr/monetr/commit/66c52b7d1b0b39429f2b1329931a5cbe9b228161))

### [0.10.11](https://github.com/monetr/monetr/compare/v0.10.10...v0.10.11) (2022-04-09)


### Features

* **ui:** Added bank logos to all accounts view. ([47c298b](https://github.com/monetr/monetr/commit/47c298b9f056bffcbfcc163219e3d53b3441bdc0))


### Bug Fixes

* **institutions:** Fixed issue introduced by institution change. ([f90872f](https://github.com/monetr/monetr/commit/f90872fc45782bf50f1f49cf9599fdf1aa45da70))
* **security:** Trying to get CSP headers working properly. ([af2f5bc](https://github.com/monetr/monetr/commit/af2f5bc772fa0ac09e94bda27cce0e8f8c659f75))

### [0.10.10](https://github.com/monetr/monetr/compare/v0.10.9...v0.10.10) (2022-04-08)


### Bug Fixes

* **security:** Fixed CSP middleware. ([f60d163](https://github.com/monetr/monetr/commit/f60d16346550c5ebc304576f2ff147d303e00ed2))
* **ui/transactions:** Fixed request spam on transactions page. ([20fc4f1](https://github.com/monetr/monetr/commit/20fc4f1fe5086c6ec44bf78733f2a92636e54755)), closes [#852](https://github.com/monetr/monetr/issues/852)

### [0.10.9](https://github.com/monetr/monetr/compare/v0.10.8...v0.10.9) (2022-04-08)


### Features

* **security:** Improving how CSP works with deployment. ([6403fdc](https://github.com/monetr/monetr/commit/6403fdc52d569253beefc00ec59f7e8eac27a0a5))


### Dependencies

* **ui:** pin dependency @fontsource/roboto to 4.5.5 ([6f3a215](https://github.com/monetr/monetr/commit/6f3a215d1a67ed42001bf1c0125452c4ce5c0564))

### [0.10.8](https://github.com/monetr/monetr/compare/v0.10.7...v0.10.8) (2022-04-08)


### Features

* **ui:** Significant improvements for mobile. ([684ec36](https://github.com/monetr/monetr/commit/684ec369da3b00a14b60917c41411c791176fece))


### Dependencies

* **api:** update module github.com/stripe/stripe-go/v72 to v72.101.0 ([0643a48](https://github.com/monetr/monetr/commit/0643a484969956e97ce51b5b934ae2652e6a511d))

### [0.10.7](https://github.com/monetr/monetr/compare/v0.10.6...v0.10.7) (2022-04-08)


### Features

* **jobs:** Added job to deactivate plaid links for expired accounts. ([baba144](https://github.com/monetr/monetr/commit/baba144c268d20840ade1a21281d32438cfc144c))
* **ui:** Added infinite scrolling to transactions. ([cf1e72f](https://github.com/monetr/monetr/commit/cf1e72f116411a6cbd561c57dcfa0efd9d7c93ad)), closes [#89](https://github.com/monetr/monetr/issues/89)


### Bug Fixes

* **jobs:** Fixed cron job schedule for link deactivation. ([85d1cc0](https://github.com/monetr/monetr/commit/85d1cc0475f4fe2c32d4821e82aa6c981f7be237))
* **ui:** Fixed Sentry crash report showing when not logged in. ([3e2fcd8](https://github.com/monetr/monetr/commit/3e2fcd8fc5bd55849d857b9d4c18907d1d64778b))


### Miscellaneous

* Tweaking makefile and renovate. ([3c34d2b](https://github.com/monetr/monetr/commit/3c34d2bf512ead9073362eb368e80d71130f0947))


### Dependencies

* **api:** update module github.com/stripe/stripe-go/v72 to v72.100.0 ([bd70c37](https://github.com/monetr/monetr/commit/bd70c3737710c3b99d4c49cdb015fc1b6f22573e))
* **renovate:** update actions/cache action to v3 ([fe890dc](https://github.com/monetr/monetr/commit/fe890dcbf94ebb8e724a21927e80b18faa234c3d))
* **renovate:** update actions/checkout action to v3 ([741ec72](https://github.com/monetr/monetr/commit/741ec72e3a18827c89ec40c8e1231269b093e05a))
* **renovate:** update actions/download-artifact action to v3 ([5070dfa](https://github.com/monetr/monetr/commit/5070dfa412cf4191cb30798f136eef3733863e81))
* **renovate:** update actions/upload-artifact action to v3 ([ab0ab63](https://github.com/monetr/monetr/commit/ab0ab631807b7692edd71aa82682c64bb76df3cd))
* **renovate:** update codecov/codecov-action action to v3 ([ca38bcc](https://github.com/monetr/monetr/commit/ca38bccf4f2439f7e060273cbc5121dc7f716195))
* **renovate:** update jamesives/github-pages-deploy-action action to v4.3.0 ([2d28738](https://github.com/monetr/monetr/commit/2d28738231ca44061f01463b87b3f46b33586426))
* **ui:** update dependency @emotion/react to v11.9.0 ([64c3aed](https://github.com/monetr/monetr/commit/64c3aed7f21943441956f39ecfa6e3b36d703f4c))
* **ui:** update dependency @types/react to v17.0.44 ([8c0f983](https://github.com/monetr/monetr/commit/8c0f983e387313b4d4a5abb499e471c61c3e6067))
* **ui:** update dependency @types/react-dom to v17.0.15 ([05257a6](https://github.com/monetr/monetr/commit/05257a6c8523068b810707fa989ed0129189e4bc))
* **ui:** update dependency css-what to v6.1.0 ([73c2733](https://github.com/monetr/monetr/commit/73c273319231a85d5e7302f17c7cb668923f8552))
* **ui:** update dependency eslint-plugin-import to v2.26.0 ([85835b0](https://github.com/monetr/monetr/commit/85835b0657e7077741505de98d2998d11cbbef3e))
* **ui:** update dependency eslint-plugin-testing-library to v5.2.1 ([e3aeeaf](https://github.com/monetr/monetr/commit/e3aeeaf8144ebead3558bc5fd56b6fbfb62271be))
* **ui:** update dependency react-router-dom to v6.3.0 ([7656aa9](https://github.com/monetr/monetr/commit/7656aa98977ea13958f76094fef10beda4017ab7))
* **ui:** update dependency sass to v1.49.11 ([da412d5](https://github.com/monetr/monetr/commit/da412d5f32b8845ad10c36844857aaeb9deb1de7))
* **ui:** update dependency sass to v1.50.0 ([6783e12](https://github.com/monetr/monetr/commit/6783e12dc465afbae2cf494ad50ec6c8a8d34dec))
* **ui:** update dependency semver to v7.3.6 ([69973c4](https://github.com/monetr/monetr/commit/69973c4c402705a27039ea52f3acb472db5ca84a))
* **ui:** update dependency webpack to v5.71.0 ([80c8522](https://github.com/monetr/monetr/commit/80c8522efc884cdb4742fe9c382fd26eb8d911cc))
* **ui:** update dependency webpack to v5.72.0 ([f7e3067](https://github.com/monetr/monetr/commit/f7e3067bd1db870a62067a7ddb53d8f95126e74e))
* **ui:** update dependency webpack-dev-server to v4.8.1 ([b43c6c9](https://github.com/monetr/monetr/commit/b43c6c93840b70b0b2c6f5ad8f37b3a67474a432))
* **ui:** update material-ui monorepo ([84b193a](https://github.com/monetr/monetr/commit/84b193acffccbce7ba7b80c10cd548ef5858137e))
* **ui:** update react monorepo ([2c84936](https://github.com/monetr/monetr/commit/2c84936d91ea8526da30bb1875edbadd45996409))
* **ui:** update typescript-eslint monorepo to v5.18.0 ([03d9b26](https://github.com/monetr/monetr/commit/03d9b2625fb8e11065666648d446a7321a0a335c))

### [0.10.6](https://github.com/monetr/monetr/compare/v0.10.5...v0.10.6) (2022-04-06)


### Features

* **docs:** Added backdrop to documentation site. ([a80a1b5](https://github.com/monetr/monetr/commit/a80a1b5ea8c97c1226c4b0ba7ae7ddcb1d7f097d))
* **docs:** Adding basic onboarding documentation. ([416fc9d](https://github.com/monetr/monetr/commit/416fc9d674d6834a67e97f8a1d5967ce5d37773a))
* **docs:** Adding documentation on development credentials. ([5d69a58](https://github.com/monetr/monetr/commit/5d69a581a79bb97e9a394f2d2b81ecab3e4cf516))
* **docs:** Adding information on removing a bank account. ([31468e1](https://github.com/monetr/monetr/commit/31468e1bb62688affdd88e4c05b91827fb566c3f))
* **docs:** Adding mkdocs-material-insider. ([0654329](https://github.com/monetr/monetr/commit/065432904b9bb1e967e94ef38269c2e3cf688dff))
* **docs:** Building out more documentation structure. ([1be8c5f](https://github.com/monetr/monetr/commit/1be8c5f13c1e404a1c923ba8818edc8b52f656e0))


### Bug Fixes

* **api:** Fixing status codes for authorization. ([62dcf78](https://github.com/monetr/monetr/commit/62dcf78cfccb5284cc22edf1f72f0e99dee90f64))
* **docs:** Fixed `Developer > Local` link, reordered sidebar. ([a21e2ce](https://github.com/monetr/monetr/commit/a21e2ce8bbbd5108655bce085d2cec3e30fccb1b))
* **docs:** Fixed debugger screenshot for local dev. ([ae7e97e](https://github.com/monetr/monetr/commit/ae7e97e41615ce6470b07fa9424a372581018342))
* **test:** Fixed flaky JWT test. ([fe40835](https://github.com/monetr/monetr/commit/fe408354c311a120bc7c4a9b4b4927cfc61a2e54))


### Miscellaneous

* **build:** Increase PR limit for renovate. ([70f3ce6](https://github.com/monetr/monetr/commit/70f3ce6d4c1e5ba09abc7ac3445e02b992fa3116))
* **ci:** Adding stuff for new static site. ([8aedbea](https://github.com/monetr/monetr/commit/8aedbeaaafa115e5f01c9aaf9fab5e17bc52aa7b))
* **docs:** Add confirm screenshot to remove account doc. ([ed31510](https://github.com/monetr/monetr/commit/ed31510e7408d968594eafd5c2af8e2ec2e8247c))
* **docs:** Adding endpoint to list in index. ([ecbfd4d](https://github.com/monetr/monetr/commit/ecbfd4df8e905eb4d4c95e42e5f09817534c280d))
* **docs:** Adding more documentation around authentication. ([326fe9d](https://github.com/monetr/monetr/commit/326fe9d9d925dcfca4ceec43aaa879f2ec3d7a4a))
* **docs:** Fix sign up link. ([bc072e5](https://github.com/monetr/monetr/commit/bc072e5b589a68635aaf595f38a506b1f30ff3fb))
* **docs:** Reference github issues for missing documentation. ([2bc7753](https://github.com/monetr/monetr/commit/2bc7753a2facf22b1ffd74239bc78ba609e7d73d))
* Experimenting with background. ([a212ac6](https://github.com/monetr/monetr/commit/a212ac655373e938cb955b24f32836f41f1c89b4))
* Fixed readme referencing outdated url. ([a878d39](https://github.com/monetr/monetr/commit/a878d39cb0746bed2f755c5e7faabc2581bae5cd))
* Reduce background intensity. ([7c16dfd](https://github.com/monetr/monetr/commit/7c16dfdbe69c875cd51528f0907bb5c62dbfcb59))
* **tests:** Adding test for retrieving transactions. ([2178392](https://github.com/monetr/monetr/commit/217839223bd3a8fed5ea4503efefb468317ece1e))


### Dependencies

* **ui:** update dependency @babel/core to v7.17.9 ([60bbe49](https://github.com/monetr/monetr/commit/60bbe49772c571501fec20eda91de5c655929eab))
* **ui:** update dependency @testing-library/jest-dom to v5.16.4 ([39ae2b4](https://github.com/monetr/monetr/commit/39ae2b407e35b529cab183e7b74e083f116fdd90))
* **ui:** update dependency react-refresh-typescript to v2.0.4 ([160693f](https://github.com/monetr/monetr/commit/160693f30d450701cf5a9b8127ef58c96c9fdd8f))
* **ui:** update dependency redoc-cli to v0.13.10 ([411fcc7](https://github.com/monetr/monetr/commit/411fcc74f64646a47d7b09cf8778a220d0e7e892))

### [0.10.5](https://github.com/monetr/monetr/compare/v0.10.4...v0.10.5) (2022-04-05)


### Features

* Adding new documentation site groundwork. ([5b33df8](https://github.com/monetr/monetr/commit/5b33df864d1726884a458f790e87f7f6240e5033))
* **docs:** Adding more documentation for API. ([02fb643](https://github.com/monetr/monetr/commit/02fb643b8280cc03c1c430e433f17acda415ec92))
* **docs:** Building out completely new docs site. ([862fd30](https://github.com/monetr/monetr/commit/862fd30e763c301d962a997114303de1007164c8))


### Bug Fixes

* Fixing axios in tests, I hate you jest. ([fcd8087](https://github.com/monetr/monetr/commit/fcd808771bbe8d87492f16d9a0f1a8beb3172d7f))
* **ui:** Fixed background colors on other routes. ([61330f3](https://github.com/monetr/monetr/commit/61330f30193efe00f05fe4c7f1d528a7d6f9ab55))
* **ui:** Fixed not using the global axios instance. ([916e059](https://github.com/monetr/monetr/commit/916e059e04c55c671e9c82eda1cc537421e6d9ff))


### Dependencies

* **api:** update module github.com/teambition/rrule-go to v1.8.0 ([486a47c](https://github.com/monetr/monetr/commit/486a47c17c0b6c28e14924a90aa5ef3d4e587d7f))
* **ui:** update dependency prettier to v2.6.2 ([755e88c](https://github.com/monetr/monetr/commit/755e88cd54231b8f55c6e74d9ff2e2be51846be6))
* **ui:** update dependency rrule to v2.6.9 ([80217c2](https://github.com/monetr/monetr/commit/80217c28a6e1262db81c3ac9d9c468ef09d16600))


### Miscellaneous

* Adding more tests because I can. ([63d5941](https://github.com/monetr/monetr/commit/63d5941d5e4f29ba4c518370ed5dccdcea3a1bbb))
* **development:** Improved Go hot-reload watchlist. ([ff553fb](https://github.com/monetr/monetr/commit/ff553fbe1f1427aac66bf9d341b3a636806a8dcd))
* Improving tests. ([af74440](https://github.com/monetr/monetr/commit/af74440da4681007368efa1db72563f03cceb0c6))
* **ui:** Move transactions view to component tree. ([57b1fa7](https://github.com/monetr/monetr/commit/57b1fa72113540b4d4ac1da1ac38c53eb0993a3e))

### [0.10.4](https://github.com/monetr/monetr/compare/v0.10.3...v0.10.4) (2022-04-04)


### Features

* **ui/layout:** Show which area is currently active in sidebar. ([de8ae13](https://github.com/monetr/monetr/commit/de8ae137c3532115724b50c85fbb24d2ccc7df06))
* **ui:** Greatly improving UI structure in code. ([d99ed48](https://github.com/monetr/monetr/commit/d99ed48173efffad2501a79ccf5bd2a897497a17))


### Bug Fixes

* **ui:** Added logo back to subscribe page. ([4f0dca8](https://github.com/monetr/monetr/commit/4f0dca837245e052bed026f6ba78e691a6b706ee))


### Dependencies

* **ui:** Removing unused dependencies. ([bd8c993](https://github.com/monetr/monetr/commit/bd8c9934bdd63187029ecf9482d39edc819287a5))
* **ui:** update dependency moment to v2.29.2 ([f44681f](https://github.com/monetr/monetr/commit/f44681f3d577004a8733a80ca9ca1089834618c4))


### Miscellaneous

* Adding screenshots to docs folder. ([261e555](https://github.com/monetr/monetr/commit/261e555368ea812ca5a22eac96fe24d3e63f9599))
* Don't remove docs dir on clean. ([13c48ec](https://github.com/monetr/monetr/commit/13c48ec8e1707e7c6b21851dfa28b5ccc6464c60))
* Improved local dev, improved documentation folder structure. ([0714b38](https://github.com/monetr/monetr/commit/0714b38ef059e6e8368d8e6757760e86c3b3525f))
* Minor improvements. ([17ced87](https://github.com/monetr/monetr/commit/17ced87490b946e6b1460d4b8bdfb4f6909a01af))
* Remove docs folder from git ignore. ([6f9278f](https://github.com/monetr/monetr/commit/6f9278f1067c854cfc07be7e76ba51dd907d68a1))
* Remove unused import in LinkedAccountItem. ([c5fe7d8](https://github.com/monetr/monetr/commit/c5fe7d8cf9cee8d5a933ddbb8338fbc6fb73d3b3))
* **ui:** Minor sidebar improvement. ([1878415](https://github.com/monetr/monetr/commit/1878415bbf668f465011089ab386181ffd86cb1e))

### [0.10.3](https://github.com/monetr/monetr/compare/v0.10.2...v0.10.3) (2022-03-31)


### Features

* **development:** Custom MailHog container to support arm64. ([47e9143](https://github.com/monetr/monetr/commit/47e914347129342237b49c91de58136d17b1545f))
* **development:** Significantly improved local development. ([7033ecf](https://github.com/monetr/monetr/commit/7033ecf20bb5723961b62633a5c134db85970402))
* **ui/about:** Include Node version in about screen. ([36cdb99](https://github.com/monetr/monetr/commit/36cdb99d9dfb2ba5823a3bc114e5a540f7c513db))
* **ui:** All accounts view improvements. ([708000b](https://github.com/monetr/monetr/commit/708000bab6966d12ea9fcadd76ec902145f1a724))
* **ui:** Building out basic About screen. ([aa3b94d](https://github.com/monetr/monetr/commit/aa3b94d5d53dad9d2dbed655af92b4164a45b3aa))


### Bug Fixes

* **development:** Allow for development on arm64. ([85b1a58](https://github.com/monetr/monetr/commit/85b1a5812d8f625e40d1afb68fc5d71b5b4997ff))
* **ui:** Fixed bad practive with UI settings component. ([8713793](https://github.com/monetr/monetr/commit/87137934eb0b34b787ef71729392581f98eb89f6))
* **ui:** Resolve Chrome warning about password forms. ([98335ff](https://github.com/monetr/monetr/commit/98335ff85d2d2bef6f262ad345ae2000a170b62b))


### Dependencies

* **api:** update module github.com/stripe/stripe-go/v72 to v72.96.0 ([303ccf9](https://github.com/monetr/monetr/commit/303ccf9e73a62122d66152df66d0730203eb43c2))
* **containers:** update dependency redis to v6.2.6 ([c151895](https://github.com/monetr/monetr/commit/c151895aa633f467d299a8ad9f832f10b137486c))
* **containers:** update node.js to v17.8.0 ([ebade0e](https://github.com/monetr/monetr/commit/ebade0e13813642a2e2c38b706438e765e059425))
* **ui:** update dependency @testing-library/jest-dom to v5.16.3 ([ff876d6](https://github.com/monetr/monetr/commit/ff876d631296cbd39297bd4cf6266abb5788aee0))
* **ui:** update dependency @types/react to v17.0.43 ([3bf8c86](https://github.com/monetr/monetr/commit/3bf8c863cdbd33a64c521e54c5297f70f084740e))
* **ui:** update dependency eslint-plugin-jest to v26.1.3 ([69a6c57](https://github.com/monetr/monetr/commit/69a6c57f23a15a8821e87650e5e1e5aca3104be3))
* **ui:** update dependency prettier to v2.6.1 ([a7d42d9](https://github.com/monetr/monetr/commit/a7d42d9447c12aeaa17881f80918c61f015e6c2d))
* **ui:** update dependency typescript to v4.6.3 ([37fc321](https://github.com/monetr/monetr/commit/37fc3218f0174f4cb4fe98f499fb6169cd252c0e))
* **ui:** update dependency webpack to v5.70.0 ([f91ea5c](https://github.com/monetr/monetr/commit/f91ea5c0d7c551d8f9486b1ca32349d0b49fe1a0))
* **ui:** update dependency workbox-webpack-plugin to v6.5.2 ([aad7968](https://github.com/monetr/monetr/commit/aad7968a2a1387b30a6eb9a7b92e2aaf506db202))
* **ui:** update material-ui monorepo ([07de381](https://github.com/monetr/monetr/commit/07de3818d45d8d3b8551fe43251a6438f7804b34))
* **ui:** update material-ui monorepo ([409fed5](https://github.com/monetr/monetr/commit/409fed56427afd1d67619b256d54055d6c70826f))


### Miscellaneous

* Added monetr screenshot. ([5f99330](https://github.com/monetr/monetr/commit/5f99330d8fc6d6e931b65513a0373b31cbe03f5d))
* **development:** Shutdown compose if its running on clean. ([8fa5f11](https://github.com/monetr/monetr/commit/8fa5f11cb720bc3ef07adedf2e4b6d5620cbc616))
* **ui:** General codebase improvements. ([3426e89](https://github.com/monetr/monetr/commit/3426e89e076023e47ff738030a1656dd076b9626))

### [0.10.2](https://github.com/monetr/monetr/compare/v0.10.1...v0.10.2) (2022-03-27)


### Bug Fixes

* **billing:** Fixed return URL for billing portal. ([dde7232](https://github.com/monetr/monetr/commit/dde72324b2dbc80d7a1901cb6d7826fe093be1db))

### [0.10.1](https://github.com/monetr/monetr/compare/v0.10.0...v0.10.1) (2022-03-27)


### Features

* **billing:** Added new billing navigation. ([39c18dd](https://github.com/monetr/monetr/commit/39c18dddc7702cdea8a12ca3deffa8a3b8566eef))


### Bug Fixes

* **ui:** Fixed change password fields type. ([9963344](https://github.com/monetr/monetr/commit/996334441e22c1e709ebbbd2013cfc780c9a560f))

## [0.10.0](https://github.com/monetr/monetr/compare/v0.9.10...v0.10.0) (2022-03-27)


### Features

* Adding docker compose for local development. ([b20620b](https://github.com/monetr/monetr/commit/b20620b2e0218c2b66311f43d9ca6fa17432c8e3))
* **authentication:** Adding support for changing passwords. ([83417c3](https://github.com/monetr/monetr/commit/83417c35d6545ffe70426d5502fc8f5dbcff6403)), closes [#565](https://github.com/monetr/monetr/issues/565)
* **go:** Upgrading to Go 1.18. ([6648000](https://github.com/monetr/monetr/commit/664800017dfe409442e399c01392265c676db7c9))


### Bug Fixes

* Fixed failing test, imroved dev documentation. ([05f8c9a](https://github.com/monetr/monetr/commit/05f8c9af9acf8580b8577dcdf8d2c217876761f0))


### Dependencies

* **api:** update module github.com/alicebob/miniredis/v2 to v2.19.0 ([c6c8455](https://github.com/monetr/monetr/commit/c6c8455d47954525553fe28b3a6721cb443c8df5))
* **api:** update module github.com/alicebob/miniredis/v2 to v2.20.0 ([d4e505e](https://github.com/monetr/monetr/commit/d4e505e25f694e07a92972f9e9b151f5ffa07b46))
* **api:** update module github.com/brianvoe/gofakeit/v6 to v6.15.0 ([ac0bf07](https://github.com/monetr/monetr/commit/ac0bf07115281f80571824fb3025037109876005))
* **api:** update module github.com/getsentry/sentry-go to v0.13.0 ([4c8716c](https://github.com/monetr/monetr/commit/4c8716ce37a327f6c67a3ca9e14b5960f3781760))
* **api:** update module github.com/spf13/cobra to v1.4.0 ([53b9c80](https://github.com/monetr/monetr/commit/53b9c80a05104add4febf10591628784998aa859))
* **containers:** update dependency golang to v1.18 ([792fbb1](https://github.com/monetr/monetr/commit/792fbb177604c87a2c7ef68b7a528c988fb0de85))
* **ui:** update dependency @types/react to v17.0.42 ([b183059](https://github.com/monetr/monetr/commit/b18305971281ed56a24f52f8388c6260f4f7b55e))
* **ui:** update dependency autoprefixer to v10.4.4 ([c2041b5](https://github.com/monetr/monetr/commit/c2041b583d1a65cfeb5ec9d43a863b7045432e8d))
* **ui:** update dependency axios to v0.26.1 ([58a7d76](https://github.com/monetr/monetr/commit/58a7d764ef4db2abc14e91c1d77834f1c5aa816f))
* **ui:** update dependency babel-loader to v8.2.4 ([fe6a319](https://github.com/monetr/monetr/commit/fe6a319bf58f94e46f0651f7d1a0baf7de07481f))
* **ui:** update dependency css-loader to v6.7.1 ([0462c09](https://github.com/monetr/monetr/commit/0462c09d52fc8432e1e4e66185cd0d101ab70b90))
* **ui:** update dependency dotenv-expand to v8.0.3 ([1f80e23](https://github.com/monetr/monetr/commit/1f80e23024b4c854a9c75c77277d08dc47577909))
* **ui:** update dependency eslint to v8.11.0 ([a2a1101](https://github.com/monetr/monetr/commit/a2a1101842d121c4546b8efe2c6ab0db4d6b6447))
* **ui:** update dependency eslint-plugin-jest to v26.1.2 ([b78bf8d](https://github.com/monetr/monetr/commit/b78bf8dd5d75011ba5bc805d7fe573d37cf037bf))
* **ui:** update dependency eslint-plugin-testing-library to v5.1.0 ([52793c7](https://github.com/monetr/monetr/commit/52793c76aa40e23abad637c91c2c229f24e10cdb))
* **ui:** update dependency mini-css-extract-plugin to v2.6.0 ([6fbf18b](https://github.com/monetr/monetr/commit/6fbf18b942bd4042276d58c516f5794ec87d9827))
* **ui:** update dependency postcss to v8.4.12 ([97b747f](https://github.com/monetr/monetr/commit/97b747f7052ddbc1406b925e87842d3267189840))
* **ui:** update dependency postcss-preset-env to v7.4.3 ([b96a8d9](https://github.com/monetr/monetr/commit/b96a8d904c35511214a6c052bf0bcbeb7b48b04a))
* **ui:** update dependency ts-loader to v9.2.8 ([a22128a](https://github.com/monetr/monetr/commit/a22128a58b1682709d0a218e9228e6920f03959a))
* **ui:** update dependency workbox-webpack-plugin to v6.5.1 ([b2e81bf](https://github.com/monetr/monetr/commit/b2e81bf97c328df4c0bb892576b335390c38002a))


### Miscellaneous

* **ci:** Upgrading CI pipelines to Go 1.18 ([284c887](https://github.com/monetr/monetr/commit/284c887ad00a7121ffbf02d17b27104aeea64df8))
* **development:** Adding comments to the docker compose file. ([fafb808](https://github.com/monetr/monetr/commit/fafb80884d77f495f9ac107a053e3fbc53bd61f1))
* **development:** Huge improvements to local development. ([211c412](https://github.com/monetr/monetr/commit/211c41201965bddf739b15041901cbe819e642ed))
* **docs:** Add information on how to develop. ([8827855](https://github.com/monetr/monetr/commit/8827855eab09cd59a610c729ff12195bb8f6f66a))
* **make:** Fix golang version in dependencies scripts. ([9a46a4c](https://github.com/monetr/monetr/commit/9a46a4c0f73713f0ecada43ca98b656942b28ccd))
* Minor cleanup and logging fix. ([68d2785](https://github.com/monetr/monetr/commit/68d2785d300fb63bcb4736de2750cb46f9335371))
* release 0.10.0 ([1fabf31](https://github.com/monetr/monetr/commit/1fabf312117bf5641ba2352abe226d1fceb73c36))

### [0.9.10](https://github.com/monetr/monetr/compare/v0.9.9...v0.9.10) (2022-03-21)


### Features

* Laying the groundwork for TOTP. ([46f712e](https://github.com/monetr/monetr/commit/46f712e287e7a6a0f151f28c33a4a85ff4a87ed1))


### Bug Fixes

* Adding theme color to manifest.json. ([e30de2d](https://github.com/monetr/monetr/commit/e30de2d81c885d305bbdf772276632ba651b825c))
* Close background job processor on shutdown. ([89e1084](https://github.com/monetr/monetr/commit/89e10845a2120ea3914c4b87275aa11457405e4e)), closes [#744](https://github.com/monetr/monetr/issues/744)
* Fixing timezone for "last sync" date. ([a1682da](https://github.com/monetr/monetr/commit/a1682dabc1cf6059d433e44555facecf6b3ce720))


### Miscellaneous

* Add ISC license to allow list. ([13dc6f1](https://github.com/monetr/monetr/commit/13dc6f15fa26683d84b0104c74f135eaadaf02cc))
* Add vscode to git ignore. ([a0dca61](https://github.com/monetr/monetr/commit/a0dca612cf7fc6e6f628e954720eac41fc040658))
* Minor code cleanup ([38ae436](https://github.com/monetr/monetr/commit/38ae436b2d32d667a425b69a784c8b4f5c03733e))
* Minor makefile improvements. ([8bdb59d](https://github.com/monetr/monetr/commit/8bdb59da0ee6101336c29b4b2f34db2ede64aab1))
* Remove unused unix socket code. ([5ff0304](https://github.com/monetr/monetr/commit/5ff030466bdb84ac8aa8072edb0524ee73de8823))
* Start using the proper Captcha interface. ([2143afc](https://github.com/monetr/monetr/commit/2143afc0135af4e5d65c4efcfe29f495467462a3))


### Dependencies

* **api:** update module github.com/stretchr/testify to v1.7.1 ([6e53b6f](https://github.com/monetr/monetr/commit/6e53b6f90f76e02219bed1ece335e12e16234169))
* **api:** update module github.com/stripe/stripe-go/v72 to v72.94.0 ([f48da3d](https://github.com/monetr/monetr/commit/f48da3d0f47742e81038b73212044c53d895af9f))
* **ui:** update dependency @babel/core to v7.17.8 ([3b5571c](https://github.com/monetr/monetr/commit/3b5571c9947719a458e91667160dbad8eb3e26bf))
* **ui:** update dependency @emotion/react to v11.8.2 ([f7dd0eb](https://github.com/monetr/monetr/commit/f7dd0ebe904d8940ba293afd2d01cd1cf44f4774))
* **ui:** update dependency @testing-library/react to v12.1.4 ([f0ee3e6](https://github.com/monetr/monetr/commit/f0ee3e66350a964a0f607679d1cb5f6837173972))
* **ui:** update dependency @types/react to v17.0.41 ([40d514a](https://github.com/monetr/monetr/commit/40d514a9ad8448b857f368dd06252e89ddace033))
* **ui:** update dependency @types/react-dom to v17.0.14 ([66fccc3](https://github.com/monetr/monetr/commit/66fccc351ef6e962d9262f4251fa09dde52dcd89))
* **ui:** update dependency @types/react-google-recaptcha to v2.1.5 ([00366ea](https://github.com/monetr/monetr/commit/00366ea92c3627ee7148c3c906b22fb8185d98cd))
* **ui:** update dependency @types/react-redux to v7.1.23 ([494195d](https://github.com/monetr/monetr/commit/494195d1d0950a63dc4b8e683ae032d6abc16f5b))
* **ui:** update dependency eslint-plugin-react to v7.29.4 ([208410a](https://github.com/monetr/monetr/commit/208410aeea79d48e3010a501010ac0f12bdfd973))
* **ui:** update dependency redoc-cli to v0.13.9 ([43d369e](https://github.com/monetr/monetr/commit/43d369ea286346de2fe048b2a1f6e09ecf159421))

### [0.9.9](https://github.com/monetr/monetr/compare/v0.9.8...v0.9.9) (2022-03-09)


### Bug Fixes

* Fixed N/A string for empty expenses/goals. ([f5a4a3b](https://github.com/monetr/monetr/commit/f5a4a3bdf6ba5325f95c4c1446ffc336b347fc92))


### Dependencies

* **api:** update github.com/iris-contrib/middleware/cors commit hash to 27fa0f6 ([bf06340](https://github.com/monetr/monetr/commit/bf063407f453c6f4fe9ee2a9210fe8b6f26dda83))
* **ui:** update dependency @types/react-dom to v17.0.13 ([d4cd55e](https://github.com/monetr/monetr/commit/d4cd55ee7bbf11cddcf3effd0a48a7f346775350))
* **ui:** update dependency eslint to v8.10.0 ([4a890c9](https://github.com/monetr/monetr/commit/4a890c9f296d372b6eb1a39f12fb467bc1667257))
* **ui:** update dependency react-router-dom to v6.2.2 ([325b5f3](https://github.com/monetr/monetr/commit/325b5f37e637bc3a1ce33e2177ef4e03dbfe6ff1))
* **ui:** update dependency sass to v1.49.9 ([fb6c872](https://github.com/monetr/monetr/commit/fb6c8721d51ce0e27826024b34f60320efd5f817))
* **ui:** update dependency typescript to v4.6.2 ([589c67a](https://github.com/monetr/monetr/commit/589c67a6162fe63c7deee2a7ad64137b1b79ab4f))
* **ui:** update material-ui monorepo ([927c5b9](https://github.com/monetr/monetr/commit/927c5b9f4c3fb1e5b06ebc8cc38f10a5c4f4e476))

### [0.9.8](https://github.com/monetr/monetr/compare/v0.9.7...v0.9.8) (2022-02-26)


### Bug Fixes

* Prevent multiple subscriptions from being made. ([802e888](https://github.com/monetr/monetr/commit/802e88848d674ebbd96b45060b4d38bf372479aa)), closes [#717](https://github.com/monetr/monetr/issues/717)


### Miscellaneous

* Improving documentation and link errors. ([c7a28e5](https://github.com/monetr/monetr/commit/c7a28e560dcbdb18d23b0810790611ccb9847737))


### Dependencies

* **api:** update github.com/iris-contrib/middleware/cors commit hash to 8e282f2 ([80185c0](https://github.com/monetr/monetr/commit/80185c031a30eb3ba51c11f9a7fec7f9b332d8dd))
* **api:** update module github.com/kataras/iris/v12 to v12.2.0-alpha6 ([b2c6733](https://github.com/monetr/monetr/commit/b2c673316985871b75b43bc27764722a3011c5c5))
* **api:** update module github.com/stripe/stripe-go/v72 to v72.88.0 ([3a7174f](https://github.com/monetr/monetr/commit/3a7174ff6ceab1512bb37ea89e1eb8a42eacd658))
* **ui:** update material-ui monorepo ([0eb073b](https://github.com/monetr/monetr/commit/0eb073b090a227041d425b2b9af26f22b9b0445e))

### [0.9.7](https://github.com/monetr/monetr/compare/v0.9.6...v0.9.7) (2022-02-18)


### Bug Fixes

* Fix page crash on changing bank account. ([b349db2](https://github.com/monetr/monetr/commit/b349db252c68ed9271ea5fbaab8fd1db22eff71e)), closes [#700](https://github.com/monetr/monetr/issues/700)


### Miscellaneous

* Improve testing of captcha interface. ([12d1143](https://github.com/monetr/monetr/commit/12d11439c3aca2bd6b204aeb3338d5c008028be3))


### Dependencies

* **api:** update module github.com/stripe/stripe-go/v72 to v72.87.0 ([7522d2f](https://github.com/monetr/monetr/commit/7522d2f80931d137879d9c33c56a9c61a464729f))
* **ui:** update dependency @babel/core to v7.17.5 ([3c444cd](https://github.com/monetr/monetr/commit/3c444cd8df6473868921c671ea9933a39c776692))
* **ui:** update dependency @testing-library/react to v12.1.3 ([5ed74fc](https://github.com/monetr/monetr/commit/5ed74fc57fb386e44b83a509ad651853918ac895))
* **ui:** update dependency axios to v0.26.0 ([380a8b4](https://github.com/monetr/monetr/commit/380a8b4d43272b419827180acaa4abb74d642ec5))
* **ui:** update dependency eslint-plugin-jest to v26 ([2d2e36a](https://github.com/monetr/monetr/commit/2d2e36a35cd896149c4f84ea5b3dc754d71661c2))
* **ui:** update dependency postcss-preset-env to v7.4.1 ([e36e3d7](https://github.com/monetr/monetr/commit/e36e3d7cd0683425ea8ca1ffb8d591e5d6d79a59))
* **ui:** update dependency sass to v1.49.8 ([653546b](https://github.com/monetr/monetr/commit/653546b223310e004f825f153082c481a5302c9b))
* **ui:** update dependency sass-loader to v12.6.0 ([e4458a0](https://github.com/monetr/monetr/commit/e4458a0993c180385960222b24940bef876092f8))
* **ui:** update dependency tailwindcss to v3.0.23 ([77c5c3e](https://github.com/monetr/monetr/commit/77c5c3ea4027d5abffa98d53c1bb3876d33e6fcb))
* **ui:** update dependency webpack to v5.69.1 ([81a5a8d](https://github.com/monetr/monetr/commit/81a5a8da012e2a3ae260be51bda27008ec24d00a))

### [0.9.6](https://github.com/monetr/monetr/compare/v0.9.5...v0.9.6) (2022-02-16)


### Features

* Improving UI appearance. ([fae66b9](https://github.com/monetr/monetr/commit/fae66b94f5a2216cabbd399640e080232332e09a))
* Laying ground-work for manual syncing. ([90a2f3e](https://github.com/monetr/monetr/commit/90a2f3e6bf85b90a4d1372373e87eb02e9f54c34))


### Bug Fixes

* Exclude paused spending from contribution totals. ([315d9c8](https://github.com/monetr/monetr/commit/315d9c865c9a01f62a9fb93eb97470fcaba2b97f))


### Miscellaneous

* **containers:** Bumping golang to 1.17.7 and node to 17.5.0. ([027722b](https://github.com/monetr/monetr/commit/027722baf4bcdba722dabf1488795061325ae7ed))
* Improved local development for UI. ([5dde8c0](https://github.com/monetr/monetr/commit/5dde8c01cb2d2c021f2e9f290f336afa4b2d38bf))


### Dependencies

* **renovate:** update jamesives/github-pages-deploy-action action to v4.2.5 ([c3e187b](https://github.com/monetr/monetr/commit/c3e187bb2f306d520a86bd29b81e8308bdaa3d32))
* **ui:** update dependency @babel/core to v7.17.4 ([5c7b94a](https://github.com/monetr/monetr/commit/5c7b94a50d10753f4deb10b35845e4f3c4079232))
* **ui:** update dependency @mui/icons-material to v5.4.2 ([8410a7b](https://github.com/monetr/monetr/commit/8410a7b557673072abbf63bf2120748445582b88))
* **ui:** update dependency @mui/lab to v5.0.0-alpha.69 ([4f3bd30](https://github.com/monetr/monetr/commit/4f3bd308f9ee613dac862216a23f57b62387148d))
* **ui:** update dependency @mui/material to v5.4.2 ([b519ac8](https://github.com/monetr/monetr/commit/b519ac8ff88be815e274c48564523cf7b3f62ebb))
* **ui:** update dependency @mui/styles to v5.4.2 ([f8c8e46](https://github.com/monetr/monetr/commit/f8c8e467d7692a8323eeea0d27479aad121ae627))
* **ui:** update dependency @types/react-google-recaptcha to v2.1.4 ([39bba1e](https://github.com/monetr/monetr/commit/39bba1ea8a7cd2ab4301cb8ff4c4cd2fd6ee401d))
* **ui:** update dependency dotenv to v16 ([a493072](https://github.com/monetr/monetr/commit/a49307263d6ad57955dd404c48ac78d57a738d58))
* **ui:** update dependency dotenv-expand to v8 ([c69941f](https://github.com/monetr/monetr/commit/c69941fbed4ec5c99a27b5192715982fd6bed0e5))
* **ui:** update dependency postcss-preset-env to v7.4.0 ([fdd7547](https://github.com/monetr/monetr/commit/fdd7547e3585720e874db29183cb642da718c70a))
* **ui:** update jest monorepo to v27.5.1 ([29a6645](https://github.com/monetr/monetr/commit/29a66451e3bedc2414f8b46e7f347d7cf0c97e15))
* **ui:** update typescript-eslint monorepo to v5.12.0 ([c5ce714](https://github.com/monetr/monetr/commit/c5ce714164be595888bc3ee3398526f5cfcb4e08))

### [0.9.5](https://github.com/monetr/monetr/compare/v0.9.4...v0.9.5) (2022-02-16)


### Bug Fixes

* Fixed build revision and version not being embedded in container. ([88620a0](https://github.com/monetr/monetr/commit/88620a0023a69866fb9b622ad3f8479de5548688)), closes [#683](https://github.com/monetr/monetr/issues/683)

### [0.9.4](https://github.com/monetr/monetr/compare/v0.9.3...v0.9.4) (2022-02-16)


### Bug Fixes

* Fixed deploy dependencies. ([75e845b](https://github.com/monetr/monetr/commit/75e845be950e453423c84db7294bddf3e9c6888c))
* Fixed release pipeline. ([f787c6c](https://github.com/monetr/monetr/commit/f787c6c9f1daaf588c46cb9fd7b121e8659bfe4a))
* Move back to docker for container builds. ([29be5fd](https://github.com/monetr/monetr/commit/29be5fd683b5e46aeb76649b35cf02e8247de5e3))
* Refresh balances when transactions change. ([c9cf04a](https://github.com/monetr/monetr/commit/c9cf04a5050263852057cfaffaafc84ee374eff3)), closes [#680](https://github.com/monetr/monetr/issues/680)

### [0.9.3](https://github.com/monetr/monetr/compare/v0.9.2...v0.9.3) (2022-02-14)


### Bug Fixes

* Fixed container-push make task ([6d65b44](https://github.com/monetr/monetr/commit/6d65b4433de1fbafb056fec73a042b46798a6b08))

### [0.9.2](https://github.com/monetr/monetr/compare/v0.9.1...v0.9.2) (2022-02-14)


### Miscellaneous

* Add ability to build container using docker or podman. ([3924a66](https://github.com/monetr/monetr/commit/3924a669fd8b14fea8fbec6c1dc3ba39039550d8))

### [0.9.1](https://github.com/monetr/monetr/compare/v0.9.0...v0.9.1) (2022-02-14)


### Dependencies

* **ui:** update dependency eslint to v8.9.0 ([00d3ea9](https://github.com/monetr/monetr/commit/00d3ea97662032810655927ac89cd837be85feec))

## [0.9.0](https://github.com/monetr/monetr/compare/v0.8.11...v0.9.0) (2022-02-14)


### Features

* **jobs:** Rewriting background job implementation. ([961ab0f](https://github.com/monetr/monetr/commit/961ab0fa8b105cd50ac3816615245ccc05ff5ac7))


### Bug Fixes

* **container:** Fixed ca-certificates version, upgraded bookworm. ([ebba23c](https://github.com/monetr/monetr/commit/ebba23ce1a8bae6cd50cb157b6246c57ca201c39))


### Miscellaneous

* release 0.9.0 ([0ea9ca8](https://github.com/monetr/monetr/commit/0ea9ca84f2184576d3b9aed37a21b799e4083365))

### [0.8.11](https://github.com/monetr/monetr/compare/v0.8.10...v0.8.11) (2022-02-09)


### Features

* Improve spent from dropdown, converted to react-select. ([391b6f4](https://github.com/monetr/monetr/commit/391b6f4dc4d9ec61ddecf33c0fd04c5209218918))


### Bug Fixes

* Added icon for dark-mode menu. ([a058b14](https://github.com/monetr/monetr/commit/a058b14f5fdcebe14f71005c7cb81d87b83e41a1))
* Close bank account menu after selecting "View All Accounts". ([5634930](https://github.com/monetr/monetr/commit/5634930a2916fa7e6393942ee3dc0b7c3a00e98f))
* Convert `updateTransaction` to a react hook. ([780541b](https://github.com/monetr/monetr/commit/780541b48de74fd55e7dfdefae84dd78d5f681ff))
* Fixed funding schedule arrow button appearance. ([a48e5a9](https://github.com/monetr/monetr/commit/a48e5a9456687de562140d83ba06bc3043d98937))
* Fixed missing key on funding schedules list. ([3da4e9e](https://github.com/monetr/monetr/commit/3da4e9e1ee4e6089ad2a8ddc8794c15f73c59a52))
* Improved spacing of transaction row columns. ([88371e2](https://github.com/monetr/monetr/commit/88371e2c84ec9023da50790d1f676f5db032c85d))
* Prevent spending from being specified for deposits. ([79ad36a](https://github.com/monetr/monetr/commit/79ad36ab4c19529a0c8f3a4740e5355f852bd202))


### Miscellaneous

* Added test for `formatAmount`. ([72bb618](https://github.com/monetr/monetr/commit/72bb618b569c27a329f2c2b457073fb3d0fc8c39))


### Dependencies

* **api:** update module github.com/plaid/plaid-go to v1.10.0 ([8cabfac](https://github.com/monetr/monetr/commit/8cabfac43b44fa3ff1b28b31ea89ad059deb4742))
* **api:** update module github.com/stripe/stripe-go/v72 to v72.86.0 ([56a7347](https://github.com/monetr/monetr/commit/56a734703dea8a522609610242ef40b5a236bbe8))
* **ui:** update dependency @babel/core to v7.17.2 ([6e013df](https://github.com/monetr/monetr/commit/6e013df2aac8df65376f88936307ea7c6e8aa4e3))
* **ui:** update dependency @date-io/moment to v2.13.1 ([e8fc6b9](https://github.com/monetr/monetr/commit/e8fc6b989606fcef2b3ace7db1989181d98047f5))
* **ui:** update dependency @mui/icons-material to v5.4.1 ([791e733](https://github.com/monetr/monetr/commit/791e733f5f2d83d888033740ea0e86bbb84ab1ce))
* **ui:** update dependency @mui/material to v5.4.1 ([701a14e](https://github.com/monetr/monetr/commit/701a14e612c528e7967997edfd60262030b489a4))
* **ui:** update dependency @mui/styles to v5.4.1 ([68d5a15](https://github.com/monetr/monetr/commit/68d5a15ecdefc92bf084c351397315ac01870916))
* **ui:** update dependency css-loader to v6.6.0 ([6721f10](https://github.com/monetr/monetr/commit/6721f1085a16bf34bab31f49240fb51b2ee26737))
* **ui:** update dependency postcss to v8.4.6 ([dccff62](https://github.com/monetr/monetr/commit/dccff62159f493a9a049e3740801575b332377a7))
* **ui:** update dependency react-plaid-link to v3.3.0 ([a82ac16](https://github.com/monetr/monetr/commit/a82ac16393a9a2ed93547e382e6a249dc753917b))
* **ui:** update dependency sass to v1.49.7 ([6af25d8](https://github.com/monetr/monetr/commit/6af25d88b44b63e3216395e5a32bc7241fa95088))
* **ui:** update dependency tailwindcss to v3.0.19 ([266af34](https://github.com/monetr/monetr/commit/266af348a6bd6e48ca55847ffa8a3f932630e4f4))
* **ui:** update dependency terser-webpack-plugin to v5.3.1 ([447258a](https://github.com/monetr/monetr/commit/447258aa568a14224dab75a1aaf13cb870fe4dce))
* **ui:** update dependency webpack-dev-server to v4.7.4 ([c794bf4](https://github.com/monetr/monetr/commit/c794bf479917489ea642abdddae99cf5b6febb80))
* Upgrading to golang 1.17.6 ([b79f06a](https://github.com/monetr/monetr/commit/b79f06a9c284c3255f5b57336833cc44de7286e0))

### [0.8.10](https://github.com/monetr/monetr/compare/v0.8.9...v0.8.10) (2022-02-08)


### Features

* Include contribution amount on the funding schedule item. ([32b6bb2](https://github.com/monetr/monetr/commit/32b6bb239ed058240921e39720a2ad2bc7520aff))


### Dependencies

* **ui:** update dependency @mui/lab to v5.0.0-alpha.68 ([f0620c0](https://github.com/monetr/monetr/commit/f0620c0a412df81bea7a5962988d8dc4cdf74d48))
* **ui:** update dependency @svgr/webpack to v6.2.1 ([2ddd02d](https://github.com/monetr/monetr/commit/2ddd02d1bdbfac1df106c91eae85d1741ec95e88))
* **ui:** update dependency @testing-library/jest-dom to v5.16.2 ([a6af260](https://github.com/monetr/monetr/commit/a6af260d449052ef5c7a0c4613fffebaa6ce28e4))
* **ui:** update dependency @types/react to v17.0.39 ([032bdfa](https://github.com/monetr/monetr/commit/032bdfa61cd7dbdfccf638d5e06177560085872f))
* **ui:** update dependency eslint-plugin-testing-library to v5.0.5 ([b9dce45](https://github.com/monetr/monetr/commit/b9dce452dc92c7b7db93251cc20a3334a19d9cbc))
* **ui:** update dependency node-notifier to v10.0.1 ([c9de21d](https://github.com/monetr/monetr/commit/c9de21d747f399b2a4cfc441e58c4b961b02a888))

### [0.8.9](https://github.com/monetr/monetr/compare/v0.8.8...v0.8.9) (2022-02-08)


### Bug Fixes

* Adding in a basic completed goals view. ([e06af85](https://github.com/monetr/monetr/commit/e06af85fef765da4a6f319c4341dcd070c6d02da))
* Fix additional Goal crash on delete. ([e332833](https://github.com/monetr/monetr/commit/e3328334ec5d2b398c03f624b15c57cfa9af8a73))
* Fixed code coverage being random on mock. ([ad6244a](https://github.com/monetr/monetr/commit/ad6244af2c30ed7cc7dc326634b03e34291e6d33))

### [0.8.8](https://github.com/monetr/monetr/compare/v0.8.7...v0.8.8) (2022-02-08)


### Bug Fixes

* Resolved issue causing page crash when deleting a goal. ([bac640e](https://github.com/monetr/monetr/commit/bac640e2b1564567aaea60b047a68ac8fb7441f9)), closes [#640](https://github.com/monetr/monetr/issues/640)


### Dependencies

* **renovate:** update jamesives/github-pages-deploy-action action to v4.2.3 ([b97e1da](https://github.com/monetr/monetr/commit/b97e1daf900dc643b3d6e2ffdca13cb9edc48320))

### [0.8.7](https://github.com/monetr/monetr/compare/v0.8.6...v0.8.7) (2022-02-03)


### Dependencies

* **api:** update module github.com/nleeper/goment to v1.4.3 ([2e6c892](https://github.com/monetr/monetr/commit/2e6c892590581f3c3bb264c41f64bcd9993890af))
* **api:** update module github.com/nleeper/goment to v1.4.4 ([53c9a5e](https://github.com/monetr/monetr/commit/53c9a5eea735f3f706a5267be089c001095194de))
* **ui:** update dependency @babel/core to v7.16.12 ([3904afb](https://github.com/monetr/monetr/commit/3904afbb7be61627609b0fd3bde1171b5dc839c6))
* **ui:** update dependency tailwindcss to v3.0.18 ([d2d8d0f](https://github.com/monetr/monetr/commit/d2d8d0fd736953c86148836dd2b360a873d26728))
* **ui:** update dependency webpack-cli to v4.9.2 ([ffc1a66](https://github.com/monetr/monetr/commit/ffc1a668a07a3e6ee59e3c4fbf7268cc089b252c))

### [0.8.6](https://github.com/monetr/monetr/compare/v0.8.5...v0.8.6) (2022-01-23)


### Dependencies

* **api:** update module github.com/alicebob/miniredis/v2 to v2.18.0 ([fe6733d](https://github.com/monetr/monetr/commit/fe6733dd5903500509ced1c34c84d2e17f7c5ff7))
* **api:** update module github.com/brianvoe/gofakeit/v6 to v6.13.2 ([2c5e041](https://github.com/monetr/monetr/commit/2c5e041b0388664efcdc5782d6147f2890f0c15d))
* **api:** update module github.com/brianvoe/gofakeit/v6 to v6.14.0 ([8a12e23](https://github.com/monetr/monetr/commit/8a12e23c5dec2cd8e488f3ad0a606e1cb0189ce6))
* **api:** update module github.com/brianvoe/gofakeit/v6 to v6.14.2 ([7d8595b](https://github.com/monetr/monetr/commit/7d8595b6a0ecb43c79dba25f1f8bd9859e02dc09))
* **api:** update module github.com/prometheus/client_golang to v1.12.0 ([5b91a4a](https://github.com/monetr/monetr/commit/5b91a4a8d9725e7728b845ca2accee7e02faff5e))
* **api:** update module github.com/stripe/stripe-go/v72 to v72.85.0 ([bf28cf9](https://github.com/monetr/monetr/commit/bf28cf93ba1c07bc6903048cc4bbaf203b1f1b86))
* **ui:** update babel monorepo ([4f9438a](https://github.com/monetr/monetr/commit/4f9438a102f5af11015f52566d31f40443637490))
* **ui:** update dependency @hot-loader/react-dom to v17.0.2 ([6390144](https://github.com/monetr/monetr/commit/639014478889beac3a64a04150eedef62bd79873))
* **ui:** update dependency axios to v0.25.0 ([0da6d60](https://github.com/monetr/monetr/commit/0da6d60cb4430bd2246684d5ea151dd09bbafdae))
* **ui:** update dependency dotenv to v12 ([7bb255d](https://github.com/monetr/monetr/commit/7bb255de757b29506506b39f4aae2ff600529a8b))
* **ui:** update dependency dotenv to v12.0.4 ([d2e793d](https://github.com/monetr/monetr/commit/d2e793dd7e826d9caa7351b6858e471eac356f49))
* **ui:** update dependency dotenv to v14 ([fa65ea4](https://github.com/monetr/monetr/commit/fa65ea48e1240db26caeef4dff6a34aee620621b))
* **ui:** update dependency eslint to v8.7.0 ([ea408cb](https://github.com/monetr/monetr/commit/ea408cb83f186c42f598bd1451f33b324ab10f75))
* **ui:** update dependency eslint-plugin-jest to v25.7.0 ([c7d8880](https://github.com/monetr/monetr/commit/c7d8880a5102611103ed4b568cb8ce5246426bdd))
* **ui:** update dependency mini-css-extract-plugin to v2.5.0 ([39838d8](https://github.com/monetr/monetr/commit/39838d83326ecfebb61a2e850f4aee37813423c1))
* **ui:** update dependency mini-css-extract-plugin to v2.5.2 ([44cf427](https://github.com/monetr/monetr/commit/44cf427747aecc6ac003ac61774e4f3d7371cf63))
* **ui:** update dependency resolve-url-loader to v5 ([1021cda](https://github.com/monetr/monetr/commit/1021cda06f9f0dfec445b51763e30b2d57e9a18c))
* **ui:** update dependency sass to v1.49.0 ([304dcfd](https://github.com/monetr/monetr/commit/304dcfd3680139d6318134c4eea9f3e3f359a73f))
* **ui:** update dependency tailwindcss to v3.0.14 ([c4f0289](https://github.com/monetr/monetr/commit/c4f028904358b6b9eb64989dba3b3fb8cf99d61e))
* **ui:** update dependency tailwindcss to v3.0.15 ([847e79e](https://github.com/monetr/monetr/commit/847e79e13fab14165e2c85d2fa242fce69dd3365))
* **ui:** update material-ui monorepo ([a6483cc](https://github.com/monetr/monetr/commit/a6483cc95cfb51f73588cd029348df88fb37bfa8))
* **ui:** update typescript-eslint monorepo to v5.10.0 ([203b64e](https://github.com/monetr/monetr/commit/203b64e7ca91e43ac0cc094beacf5c2affc8956c))

### [0.8.5](https://github.com/monetr/monetr/compare/v0.8.4...v0.8.5) (2022-01-14)


### Bug Fixes

* Significantly improve simplicity of versioning. ([f5abdef](https://github.com/monetr/monetr/commit/f5abdef84d53aec927d1de500003f78458bb182d))


### Miscellaneous

* **local:** Bumped local development kube version. ([54c6bba](https://github.com/monetr/monetr/commit/54c6bba544591c09dabe77d697d7cd8067ac908c))


### Dependencies

* **api:** update module github.com/brianvoe/gofakeit/v6 to v6.12.1 ([4b16782](https://github.com/monetr/monetr/commit/4b1678223a071a084f3b6774ba8d7aec400d8665))
* **api:** update module github.com/brianvoe/gofakeit/v6 to v6.12.2 ([282a7b1](https://github.com/monetr/monetr/commit/282a7b11174da338fe919c9d75c61c57e0c62e69))
* **api:** update module github.com/brianvoe/gofakeit/v6 to v6.13.0 ([9d24220](https://github.com/monetr/monetr/commit/9d2422034a474193ff66be7969bed6c0e0f485d8))
* **api:** update module github.com/stripe/stripe-go/v72 to v72.82.0 ([de45c1a](https://github.com/monetr/monetr/commit/de45c1a6e5faed6d056bab59dae8bced4bf49f52))
* **api:** update module github.com/stripe/stripe-go/v72 to v72.83.0 ([c4617d3](https://github.com/monetr/monetr/commit/c4617d3d4aa0d2c37a1add5487a6853ec77f59c5))
* **api:** update module github.com/teambition/rrule-go to v1.7.3 ([be96ec3](https://github.com/monetr/monetr/commit/be96ec380d0bb470cbd89c8f60cf326781575e32))
* **ui:** update dependency @types/react-redux to v7.1.22 ([a4f1945](https://github.com/monetr/monetr/commit/a4f19457ee65a1874c5d635fa917f92fc6160121))
* **ui:** update dependency dotenv to v11 ([99567c6](https://github.com/monetr/monetr/commit/99567c689bee4a9a0bc1d49314085cffb6e70563))
* **ui:** update dependency immer to v9.0.12 ([0975b09](https://github.com/monetr/monetr/commit/0975b091fa7c79800eb9ca700f7a6acc58f95d0d))
* **ui:** update dependency mini-css-extract-plugin to v2.4.7 ([8690f6c](https://github.com/monetr/monetr/commit/8690f6c4246f9b53a824e5af246cf00d1dbc2317))
* **ui:** update dependency postcss-preset-env to v7.2.3 ([5dc27e5](https://github.com/monetr/monetr/commit/5dc27e5c123f981a468a959b9c7d2809d58859cc))
* **ui:** update dependency react-select to v5.2.2 ([a0300b9](https://github.com/monetr/monetr/commit/a0300b9e1cccdc057e5bdc67882bfc5c6d97804a))
* **ui:** update dependency sass to v1.48.0 ([35d4561](https://github.com/monetr/monetr/commit/35d4561e5b138b19538928afd07ca2871188f4a0))
* **ui:** update dependency tailwindcss to v3.0.13 ([2eb7ea4](https://github.com/monetr/monetr/commit/2eb7ea4e4f6ec94b05027cb2b9932d90ceed6e3d))
* **ui:** update dependency webpack to v5.66.0 ([40b2f8b](https://github.com/monetr/monetr/commit/40b2f8b8ecae367d09b93bc4dc721a1a26c8fb5d))
* **ui:** update dependency webpack-dev-server to v4.7.3 ([dceb3ad](https://github.com/monetr/monetr/commit/dceb3ad114de7b91539477f28a99b5e1460fcb87))
* **ui:** update dependency webpack-manifest-plugin to v4.1.1 ([83693d9](https://github.com/monetr/monetr/commit/83693d925bc7d81ca8f758e27c24e16e97f4373d))

### [0.8.4](https://github.com/monetr/monetr/compare/v0.8.3...v0.8.4) (2022-01-10)


### Dependencies

* **api:** update github.com/xlzd/gotp commit hash to 8b477b0 ([6810b12](https://github.com/monetr/monetr/commit/6810b1252a17795808f8a61c1644ebfb3a9ef6fd))
* **api:** update github.com/xlzd/gotp commit hash to fab697c ([0defdaf](https://github.com/monetr/monetr/commit/0defdaf95f93c1db56288143c56ad6bb83bcdd95))
* **api:** update module github.com/brianvoe/gofakeit/v6 to v6.12.0 ([4e39903](https://github.com/monetr/monetr/commit/4e3990304f45f479518f1d87e4d1e0c84a5282dd))
* **renovate:** update jamesives/github-pages-deploy-action action to v4.2.1 ([6b6680b](https://github.com/monetr/monetr/commit/6b6680b92b444eedb361b31f3a269f11264e6a40))
* **renovate:** update jamesives/github-pages-deploy-action action to v4.2.2 ([56f85ed](https://github.com/monetr/monetr/commit/56f85ed4c0dbe339ae2dd8c682a07f2ea7e90c71))
* **ui:** update dependency @babel/preset-env to v7.16.8 ([22773f3](https://github.com/monetr/monetr/commit/22773f330cffd6d7ccc1dd104d9e012baf034a2a))
* **ui:** update dependency @svgr/webpack to v6.2.0 ([fd3d986](https://github.com/monetr/monetr/commit/fd3d9868e4f3d62e552bfb6c28a91ee58a3953cc))
* **ui:** update dependency autoprefixer to v10.4.2 ([dd30b55](https://github.com/monetr/monetr/commit/dd30b558da5e594a2fd647aee9f003735d7c40cf))
* **ui:** update dependency eslint-plugin-testing-library to v5.0.2 ([0a347ab](https://github.com/monetr/monetr/commit/0a347ab89d1d8802373be2f79420d4b82e194b81))
* **ui:** update dependency eslint-plugin-testing-library to v5.0.3 ([e350376](https://github.com/monetr/monetr/commit/e350376ed6d6646459e3a9eb6bf310ba9874c612))
* **ui:** update dependency sass to v1.47.0 ([3f0de34](https://github.com/monetr/monetr/commit/3f0de3492d22833859ad2b123d3def7073385b42))
* **ui:** update dependency tailwindcss to v3.0.12 ([c5988b0](https://github.com/monetr/monetr/commit/c5988b03a9617ff21795f2b0c09bb50fdf575999))
* **ui:** update dependency web-vitals to v2.1.3 ([df1bcb0](https://github.com/monetr/monetr/commit/df1bcb06fdb3dcf054fdba7bf20411a1b0d5bf87))
* **ui:** update material-ui monorepo ([a67913c](https://github.com/monetr/monetr/commit/a67913c5ead9e6c795791d585db9a68c01262388))
* **ui:** update typescript-eslint monorepo to v5.9.1 ([c97bd77](https://github.com/monetr/monetr/commit/c97bd77007722caf1db358fff4d96859dde48930))

### [0.8.3](https://www.github.com/monetr/monetr/compare/v0.8.2...v0.8.3) (2022-01-06)


### Dependencies

* **renovate:** update jamesives/github-pages-deploy-action action to v4.1.9 ([68cd345](https://www.github.com/monetr/monetr/commit/68cd3459fe2a6ec714bbb8c2f0cbc90ab414bbf3))
* **renovate:** update jamesives/github-pages-deploy-action action to v4.2.0 ([269c08c](https://www.github.com/monetr/monetr/commit/269c08c7fbe38fc8b8f5a4a3cf4112a95a5a67a6))
* **ui:** update dependency @types/jest to v27.4.0 ([c425e65](https://www.github.com/monetr/monetr/commit/c425e65b6da3ebb8c2a5a9735b5f3cd0cc79c50e))
* **ui:** update dependency camelcase to v6.3.0 ([2863bde](https://www.github.com/monetr/monetr/commit/2863bdeb8b06ea266ee6e51430bee3e7b22869b6))
* **ui:** update dependency eslint to v8.6.0 ([0b53cc5](https://www.github.com/monetr/monetr/commit/0b53cc5cd153c40990e1a5f08c65384a852ed2c2))
* **ui:** update dependency jest to v27.4.7 ([cee903c](https://www.github.com/monetr/monetr/commit/cee903cbf3e7241f60aed4af6f177a2fdca2b0d8))
* **ui:** update dependency mini-css-extract-plugin to v2.4.6 ([d2a0d12](https://www.github.com/monetr/monetr/commit/d2a0d122bf604123f9a81c1816cbe9fe137a6861))
* **ui:** update dependency postcss-preset-env to v7.2.0 ([dcba2c9](https://www.github.com/monetr/monetr/commit/dcba2c9d60fcc03e5c0f25177c2377616999cd99))
* **ui:** update dependency prop-types to v15.8.1 ([94e5412](https://www.github.com/monetr/monetr/commit/94e5412f8edaaba52db06022fdff48035c9aa6e2))
* **ui:** update dependency resolve to v1.21.0 ([6c00059](https://www.github.com/monetr/monetr/commit/6c00059be54e527602693ac434a9605ed397e79e))
* **ui:** update dependency sass to v1.45.2 ([2d96422](https://www.github.com/monetr/monetr/commit/2d9642254cc26ad411b9fd5c2b2137d326973e3f))
* **ui:** update dependency sass to v1.46.0 ([672f149](https://www.github.com/monetr/monetr/commit/672f1496d11bf6e4be6d1fff6bf26c4171baf2cb))
* **ui:** update dependency tailwindcss to v3.0.10 ([5a47a0d](https://www.github.com/monetr/monetr/commit/5a47a0d6f325e8045b54b792a4c121868ff15447))
* **ui:** update dependency tailwindcss to v3.0.11 ([be11f1a](https://www.github.com/monetr/monetr/commit/be11f1a727bad45b6ee68e46cfdae0881545e493))
* **ui:** update jest monorepo to v27.4.6 ([b2f89a0](https://www.github.com/monetr/monetr/commit/b2f89a087106155f2eead1d80b88f2e713b62ff0))

### [0.8.2](https://www.github.com/monetr/monetr/compare/v0.8.1...v0.8.2) (2022-01-04)


### Bug Fixes

* **temp:** Removed `linux/arm64` containers from builds ([462a53f](https://www.github.com/monetr/monetr/commit/462a53f3619c1167879aa5ecad1ad227bc7a0883))


### Miscellaneous

* Updated License for 2022. ([f8cfc08](https://www.github.com/monetr/monetr/commit/f8cfc08c3c538f1a8bd965f8b13d8db3565c1f56))


### Dependencies

* **api:** update module github.com/brianvoe/gofakeit/v6 to v6.11.1 ([f0947f5](https://www.github.com/monetr/monetr/commit/f0947f59b8118a8d91ed4e6d3cd272da04a8a173))
* **api:** update module github.com/gomodule/redigo to v1.8.8 ([9cc8c69](https://www.github.com/monetr/monetr/commit/9cc8c6935c4389faeb0f6a10949ae7f0ab1bb776))
* **api:** update module github.com/nyaruka/phonenumbers to v1.0.74 ([35b0258](https://www.github.com/monetr/monetr/commit/35b0258c6174c8957b750602e965aa373cd309e5))
* **ui:** update babel monorepo to v7.16.7 ([917f248](https://www.github.com/monetr/monetr/commit/917f248cd6626c6ba5d09300ed500ee48be6193a))
* **ui:** update dependency autoprefixer to v10.4.1 ([4fbece8](https://www.github.com/monetr/monetr/commit/4fbece8c72e5395bf0091ff0671533981b626f22))
* **ui:** update dependency css-what to v6 ([de53a0f](https://www.github.com/monetr/monetr/commit/de53a0f0f30831d83d3e1aacdb78d816b77ba4b8))
* **ui:** update dependency eslint-plugin-import to v2.25.4 ([4e63f9c](https://www.github.com/monetr/monetr/commit/4e63f9c986b032abebc9c512c362a15929011227))
* **ui:** update dependency eslint-plugin-jest to v25.3.4 ([3f9f53d](https://www.github.com/monetr/monetr/commit/3f9f53d6b594d334151734f9afc3d8b31becf2ad))
* **ui:** update dependency http-status-codes to v2.2.0 ([391abee](https://www.github.com/monetr/monetr/commit/391abee272733f7c10c5a0accd49108f90baded5))
* **ui:** update dependency tailwindcss to v3.0.9 ([e0177e0](https://www.github.com/monetr/monetr/commit/e0177e03cf3ff031599e202fbc5b4a35c5adab5c))
* **ui:** update dependency webpack-dev-server to v4.7.2 ([b01c457](https://www.github.com/monetr/monetr/commit/b01c45793ceb60271fda9280277e96d71fe178f2))
* **ui:** update material-ui monorepo ([f55f20f](https://www.github.com/monetr/monetr/commit/f55f20f4d934cda945433119bc96fe991a215667))
* **ui:** update typescript-eslint monorepo to v5.9.0 ([e2fe61d](https://www.github.com/monetr/monetr/commit/e2fe61d2c7573327e1f7e762830b0b567412c939))

### [0.8.1](https://www.github.com/monetr/monetr/compare/v0.8.0...v0.8.1) (2021-12-24)


### Bug Fixes

* **container:** Fixed container not being tagged with latest. ([11aa14e](https://www.github.com/monetr/monetr/commit/11aa14e7d470280a19e58b41177830725f438d5b)), closes [#501](https://www.github.com/monetr/monetr/issues/501)
* **local:** Remove PGAdmin from local development for now. ([469dc8d](https://www.github.com/monetr/monetr/commit/469dc8d831264a66d73554149d62fdbc94c8c3b4))


### Dependencies

* **api:** update module github.com/getsentry/sentry-go to v0.12.0 ([4ae70df](https://www.github.com/monetr/monetr/commit/4ae70dfad0534e1ee4bf0471626ef3819b5fff93))
* **api:** update module github.com/hashicorp/vault/api to v1.3.1 ([e02f22b](https://www.github.com/monetr/monetr/commit/e02f22b53fbbdd6d80bbef4c6000d82145956772))
* **api:** update module github.com/jarcoal/httpmock to v1.1.0 ([6cb70d4](https://www.github.com/monetr/monetr/commit/6cb70d42c7c4285e6e33f51459c5b69d30890ed0))
* **api:** update module github.com/stripe/stripe-go/v72 to v72.81.0 ([0ab0c4d](https://www.github.com/monetr/monetr/commit/0ab0c4d68e1f5ba804312ce42c3513b88c2ecaf4))
* **renovate:** update terraform vault to v3.1.0 ([86ef22e](https://www.github.com/monetr/monetr/commit/86ef22e8cad11bedba93bae580e05e32e0e6462c))
* **renovate:** update terraform vault to v3.1.1 ([a61371e](https://www.github.com/monetr/monetr/commit/a61371efaf8416cc6457097b3972aab8d7da4df4))
* **ui:** update dependency @pmmmwh/react-refresh-webpack-plugin to v0.5.4 ([0108517](https://www.github.com/monetr/monetr/commit/01085172a56be15f17e3a635125f1dde2cdccb5f))
* **ui:** update dependency @types/react to v17.0.38 ([021eb5b](https://www.github.com/monetr/monetr/commit/021eb5bb5399ddff8aef1299cc5e3074d8adbc17))
* **ui:** update dependency @types/react-redux to v7.1.21 ([d350a4c](https://www.github.com/monetr/monetr/commit/d350a4c0828b1f1a419b89412c36059f813a6556))
* **ui:** update dependency eslint-plugin-react to v7.28.0 ([aeebe37](https://www.github.com/monetr/monetr/commit/aeebe3743abaea67de4e4be38289c0377a22c45e))
* **ui:** update dependency postcss-preset-env to v7.1.0 ([6ab599a](https://www.github.com/monetr/monetr/commit/6ab599a930a0504a25137b2971cb8ec716c0b70f))
* **ui:** update dependency prop-types to v15.8.0 ([2036c9b](https://www.github.com/monetr/monetr/commit/2036c9b5ed64c2e674de1c46e573326c9452fad7))
* **ui:** update dependency sass to v1.45.1 ([0c6126a](https://www.github.com/monetr/monetr/commit/0c6126a2b19a7442431810cd30fd2cbb23949cf6))
* **ui:** update dependency webpack-dev-server to v4.7.1 ([e99ff63](https://www.github.com/monetr/monetr/commit/e99ff630366f765c6dc26efb29b10aab554886ac))
* **ui:** update material-ui monorepo ([8616db5](https://www.github.com/monetr/monetr/commit/8616db50a9699b960b9fcf19c7ed34ab86b5c3d6))
* **ui:** update typescript-eslint monorepo to v5.8.0 ([bcd85a2](https://www.github.com/monetr/monetr/commit/bcd85a24dcb6b1c405b081234dd773485be4a8c3))

## [0.8.0](https://www.github.com/monetr/monetr/compare/v0.7.10...v0.8.0) (2021-12-18)


### Features

* **container:** Changed container to slim debian from ubuntu. ([5df2b34](https://www.github.com/monetr/monetr/commit/5df2b3482895fdc7f865fbe15e689318485251d2))


### Miscellaneous

* Add wakatime badge to readme. ([4feef04](https://www.github.com/monetr/monetr/commit/4feef040ad0c3c308a095f776a166ab7fc2aecee))


### Dependencies

* **renovate:** update jamesives/github-pages-deploy-action action to v4.1.8 ([dc4df70](https://www.github.com/monetr/monetr/commit/dc4df70e3f7ea5d5af905198ad59cedc229394b4))
* **ui:** update dependency eslint to v8.5.0 ([efe74f5](https://www.github.com/monetr/monetr/commit/efe74f501f8c7436a3293aafcb9bbd3b24afe177))
* **ui:** update dependency react-router-dom to v6.2.0 ([f2d86ce](https://www.github.com/monetr/monetr/commit/f2d86ce31edcab863971bd3e3d9e05e8f7ffb87c))
* **ui:** update dependency react-router-dom to v6.2.1 ([26c30ea](https://www.github.com/monetr/monetr/commit/26c30ea9601df472188a34063ff1dd968d1f62c7))
* **ui:** update dependency tailwindcss to v3.0.7 ([59d8979](https://www.github.com/monetr/monetr/commit/59d8979398caaebd55a5cee612200956583df3ad))

### [0.7.10](https://www.github.com/monetr/monetr/compare/v0.7.9...v0.7.10) (2021-12-16)


### Bug Fixes

* **ci:** Artifacts being uploaded for binaries for release. ([0a5fd0d](https://www.github.com/monetr/monetr/commit/0a5fd0d2be19110a8bb94013329cc0d1390fbe8b))
* **ci:** Fixed other paths for built binaries. ([9085563](https://www.github.com/monetr/monetr/commit/908556395c8536eee051606df2426cb154f35b92))


### Miscellaneous

* Improving local development, cleanup. ([4147a2b](https://www.github.com/monetr/monetr/commit/4147a2b2b8fe2b06d117d7c36d6873cf53c17c37))


### Dependencies

* **api:** update module github.com/spf13/cobra to v1.3.0 ([8e3a0c9](https://www.github.com/monetr/monetr/commit/8e3a0c9f904e2518f1b888bf966f44c986527408))
* **api:** update module github.com/spf13/viper to v1.10.1 ([37d48d8](https://www.github.com/monetr/monetr/commit/37d48d86208c8664c4382d277e5cff3cf2726c9c))
* **api:** update module github.com/stripe/stripe-go/v72 to v72.80.0 ([fd174b3](https://www.github.com/monetr/monetr/commit/fd174b37840b412ea17bbe8e799b21dcff761b59))
* **ui:** update dependency babel-plugin-named-asset-import to v0.3.8 ([93a94b8](https://www.github.com/monetr/monetr/commit/93a94b8696bb7ef353f3ea8a31fbe9db1294c857))
* **ui:** update dependency babel-preset-react-app to v10.0.1 ([c826436](https://www.github.com/monetr/monetr/commit/c826436b09941cf18846c9e9d5d319156433cd7b))
* **ui:** update dependency eslint-config-react-app to v7 ([4459987](https://www.github.com/monetr/monetr/commit/445998747c43636ccb44d211af7f6d39be3c5547))
* **ui:** update dependency postcss-preset-env to v7.0.2 ([053ddc3](https://www.github.com/monetr/monetr/commit/053ddc39056fc498bbb8d3e60f2bf3213271a6ea))
* **ui:** update dependency react-app-polyfill to v3 ([d74cd05](https://www.github.com/monetr/monetr/commit/d74cd05a5c85088562b41bc98494ebe141fb41d3))
* **ui:** update dependency tailwindcss to v3.0.5 ([f162ad0](https://www.github.com/monetr/monetr/commit/f162ad0c4bc4dd21acbcaa28c3656cc9a9357038))
* **ui:** update dependency tailwindcss to v3.0.6 ([2ff6b99](https://www.github.com/monetr/monetr/commit/2ff6b9952d56c6e19f4f5bc2d572d8d5ccdb0f0f))
* **ui:** update dependency terser-webpack-plugin to v5.3.0 ([75a6157](https://www.github.com/monetr/monetr/commit/75a61573250905aa74364fea193dc7c1c8cf6f04))
* **ui:** update material-ui monorepo ([31c577d](https://www.github.com/monetr/monetr/commit/31c577d6c0d52e49b4db0c80ba9850311bb4739f))

### [0.7.9](https://www.github.com/monetr/monetr/compare/v0.7.8...v0.7.9) (2021-12-14)


### Bug Fixes

* Re-impelemented basic testing for components. ([3504b4a](https://www.github.com/monetr/monetr/commit/3504b4aa9608c8ae1eb50a5eef3649102dee1221))
* **ui:** Fixed failing build due to redux state type issue. ([357622f](https://www.github.com/monetr/monetr/commit/357622f21412faf6d69d26799b56b583ecad2ea7))
* **ui:** Improving transactions, testing and hooks. ([6307ca1](https://www.github.com/monetr/monetr/commit/6307ca1af2293b6547fc4a18edb06cb0ff76c798))


### Dependencies

* **api:** update module github.com/alicebob/miniredis/v2 to v2.17.0 ([104e0a3](https://www.github.com/monetr/monetr/commit/104e0a336bf6a426c2da2c75aeacb062ba8fc1f0))
* **api:** update module github.com/spf13/viper to v1.10.0 ([6019ec5](https://www.github.com/monetr/monetr/commit/6019ec5fd562f0d40ea4273bfc7b76a4a7d77371))
* **ui:** pin dependency @types/jest to 27.0.3 ([3f1634a](https://www.github.com/monetr/monetr/commit/3f1634a12e4fad92a63c8eae41d786eb51ecde1c))
* **ui:** pin dependency jest-environment-jsdom to 27.4.4 ([f3faa76](https://www.github.com/monetr/monetr/commit/f3faa76036503814db351b888ffb556fed43f69e))
* **ui:** update babel monorepo to v7.16.5 ([14f8e53](https://www.github.com/monetr/monetr/commit/14f8e535b91eaaca2916fa114df345432e986943))
* **ui:** update dependency @emotion/react to v11.7.1 ([ee3ac2a](https://www.github.com/monetr/monetr/commit/ee3ac2a3a29f09ead48387414bf933fac1f3c6f6))
* **ui:** update dependency @svgr/webpack to v6.1.2 ([0e1ff7b](https://www.github.com/monetr/monetr/commit/0e1ff7b9081d4d01671fd1cc0f698d6c518947bd))
* **ui:** update dependency postcss to v8.4.5 ([b967173](https://www.github.com/monetr/monetr/commit/b9671732c0f25ff0d47cdae3fff037e1416205e0))
* **ui:** update dependency react-router-dom to v6.1.1 ([0d61e82](https://www.github.com/monetr/monetr/commit/0d61e82491621c64097e3a8c99ff36884a5985ac))
* **ui:** update dependency tailwindcss to v3.0.2 ([5f3e172](https://www.github.com/monetr/monetr/commit/5f3e172dd837e11e782f8d0c176aac980f20b3bc))
* **ui:** update dependency typescript to v4.5.4 ([7232a3c](https://www.github.com/monetr/monetr/commit/7232a3c940e0c713dd3998112d3b052d5a93cc74))
* **ui:** update jest monorepo to v27.4.5 ([3985386](https://www.github.com/monetr/monetr/commit/3985386ada6946d245f6d4b4ca9f312d7d7d89a2))
* **ui:** update typescript-eslint monorepo to v5.7.0 ([f143969](https://www.github.com/monetr/monetr/commit/f1439694487ccc306aa010e456d671577be44f2e))


### Miscellaneous

* Add code of conduct ([ce2e804](https://www.github.com/monetr/monetr/commit/ce2e80464a5c02d826b1ad927f1f5671351b3f15))
* Improvements to building container locally. ([61cb6df](https://www.github.com/monetr/monetr/commit/61cb6dfee25378df1c500a4ae2029eb63ca634a2))
* Improving testing. ([4561d61](https://www.github.com/monetr/monetr/commit/4561d619135158cc319fc621f95c508d0dc8b7e5))
* Local development cleanup ([641b7be](https://www.github.com/monetr/monetr/commit/641b7be9630812e557ea2e95a014c855dd6905b6))

### [0.7.8](https://www.github.com/monetr/monetr/compare/v0.7.7...v0.7.8) (2021-12-11)


### Miscellaneous

* **build:** Improving development builds. ([e091005](https://www.github.com/monetr/monetr/commit/e091005c36d1f8ae7d38120675450813ee69b0c9))


### Dependencies

* **container:** Upgrading to the latest golang and node. ([e9d8718](https://www.github.com/monetr/monetr/commit/e9d87188b20dc3693f24f81d06cb99cdbf4ea3b8))
* **ui:** update dependency tailwindcss to v3.0.1 ([2a3b32c](https://www.github.com/monetr/monetr/commit/2a3b32c471a69ffa4a0b06ad755b9e32da1fcb0f))

### [0.7.7](https://www.github.com/monetr/monetr/compare/v0.7.6...v0.7.7) (2021-12-11)


### Dependencies

* **ui:** update dependency react-router-dom to v6.1.0 ([2e6d116](https://www.github.com/monetr/monetr/commit/2e6d1167d0e5729d69a2ed4ab840a28abfbbc835))
* **ui:** update dependency sass to v1.45.0 ([7397a1a](https://www.github.com/monetr/monetr/commit/7397a1a9ed0b79475604eaceb33529aad7a92fb8))
* **ui:** update sentry-javascript monorepo to v6.16.1 ([57a2b2b](https://www.github.com/monetr/monetr/commit/57a2b2b2122fbcee58d9d0cd22527d04333deb8d))
* **ui:** Upgraded jest to latest version and improved config. ([28094d4](https://www.github.com/monetr/monetr/commit/28094d4cba7baa40f885e9ccb4fbfc1b5f6209cd))

### [0.7.6](https://www.github.com/monetr/monetr/compare/v0.7.5...v0.7.6) (2021-12-10)


### Dependencies

* **api:** update module github.com/plaid/plaid-go to v1.9.0 ([3d448bd](https://www.github.com/monetr/monetr/commit/3d448bdd83f8a415fb6118b03bfa587150fba2f5))
* **api:** update module github.com/stripe/stripe-go/v72 to v72.78.0 ([66c5b35](https://www.github.com/monetr/monetr/commit/66c5b35435755595389732355b7e94e32872a1d4))
* **ui:** update dependency redoc-cli to v0.13.2 ([9306464](https://www.github.com/monetr/monetr/commit/93064641470086da9af992d956dbb4d93fc75390))
* **ui:** update dependency typescript to v4.5.3 ([27ad9f9](https://www.github.com/monetr/monetr/commit/27ad9f9a524f3932f9842b24a42cfdd56c77ecc0))

### [0.7.5](https://www.github.com/monetr/monetr/compare/v0.7.4...v0.7.5) (2021-12-08)


### Bug Fixes

* **plaid:** Make sure webhook unauthorized errors are reported to sentry ([a698307](https://www.github.com/monetr/monetr/commit/a698307ccb3cb26c7ca2b099bde40ec8d34b62af))


### Dependencies

* **api:** update github.com/iris-contrib/middleware/cors commit hash to 081c558 ([3d577bf](https://www.github.com/monetr/monetr/commit/3d577bf254f49ff9528cc9c367b43db893029872))
* **ui:** update dependency eslint to v8.4.1 ([a21ab8c](https://www.github.com/monetr/monetr/commit/a21ab8c185b41304470a1ebb24cfe90800a8d665))
* **ui:** update dependency sass-loader to v12.4.0 ([fdd55ed](https://www.github.com/monetr/monetr/commit/fdd55ede7efdc21b7562bfb057e41a6e12d0a934))
* **ui:** update sentry-javascript monorepo to v6.16.0 ([166a4ac](https://www.github.com/monetr/monetr/commit/166a4ac45f5af3d7ee870c22ec6af09f5df1d4fc))

### [0.7.4](https://www.github.com/monetr/monetr/compare/v0.7.3...v0.7.4) (2021-12-06)


### Bug Fixes

* Fixed references to keyfunc with version update. ([143025e](https://www.github.com/monetr/monetr/commit/143025e506bd14192a795d8065e01a465a5560d5))
* **sentry:** Improved span information for sentry. ([219dcd7](https://www.github.com/monetr/monetr/commit/219dcd76a2225efc1a862c3baa8c1a97b61d3d77))


### Miscellaneous

* **containers:** Improved order of labels in Dockerfile. ([3a3a18d](https://www.github.com/monetr/monetr/commit/3a3a18dd501be0164ab1756f0739d98200773d58))


### Dependencies

* **api:** update github.com/iris-contrib/middleware/cors commit hash to a287965 ([1b65b08](https://www.github.com/monetr/monetr/commit/1b65b082eecd6afb25ab2f15d39abb19e51d5bd9))
* **api:** update github.com/iris-contrib/middleware/cors commit hash to cd41492 ([8fd5a74](https://www.github.com/monetr/monetr/commit/8fd5a744308d2b92ec6d21a6a23446e15167d280))
* **api:** update module github.com/micahparks/keyfunc to v1 ([4a5ec0e](https://www.github.com/monetr/monetr/commit/4a5ec0e6f41400e662d310165123ecc9ccec3eda))
* **api:** update module github.com/micahparks/keyfunc to v1.0.1 ([8103f44](https://www.github.com/monetr/monetr/commit/8103f44a05c3de981a23f9a41bdd47b590f93412))
* **ui:** update dependency @testing-library/jest-dom to v5.16.1 ([9be130f](https://www.github.com/monetr/monetr/commit/9be130f9b077427fb16664ebedd50a4ed15fa791))
* **ui:** update dependency webpack to v5.65.0 ([1761d22](https://www.github.com/monetr/monetr/commit/1761d22400f3c83c10ed395830473db788815c6c))
* **ui:** update material-ui monorepo ([8c7c92d](https://www.github.com/monetr/monetr/commit/8c7c92d8a7b2bfaf943e3c10ae3604344cd1485e))
* **ui:** update typescript-eslint monorepo to v5.6.0 ([1ad1584](https://www.github.com/monetr/monetr/commit/1ad1584adeb70a6102bab7148bcf5a04a77d593a))

### [0.7.3](https://www.github.com/monetr/monetr/compare/v0.7.2...v0.7.3) (2021-12-05)


### Bug Fixes

* **deps:** update dependency @svgr/webpack to v6.1.1 ([d7115cf](https://www.github.com/monetr/monetr/commit/d7115cf3d32e9669c096f20338d6dbb43b0e1c4a))
* **deps:** update dependency eslint to v8.4.0 ([ece4fa2](https://www.github.com/monetr/monetr/commit/ece4fa26f8a241751f289d354806177677292566))
* Fixed forgot password endpoint failing with ReCAPTCHA. ([a6293d1](https://www.github.com/monetr/monetr/commit/a6293d129b1f25f460cd50df90fbbbb6494158f5))


### Documentation

* Adding some docs for forgot password. ([6fd272e](https://www.github.com/monetr/monetr/commit/6fd272e9b61aeba60b0f236e24e88fede7600e19))


### Build Automation

* Add commit types to release please action. ([5671cf6](https://www.github.com/monetr/monetr/commit/5671cf629c380f1941f6d290cf61189f227f9f0f))
* Added multiple commit types for `feature`/`docs` ([4e4fe87](https://www.github.com/monetr/monetr/commit/4e4fe87377ef9accc1aa79785654f8d7db32dd59))
* Try to group dependencies together in release notes. ([907fbf8](https://www.github.com/monetr/monetr/commit/907fbf8957f8f1cdfaa48adf906f7fdda42bcce3))


### Dependencies

* **ci:** Upgrade GitHub actions to golang:1.17.4 ([a6fd620](https://www.github.com/monetr/monetr/commit/a6fd620451b9f5d8c5b5f6ee583a63152f9c7cb7))
* **containers:** update golang docker tag to v1.17.4 ([3fdb0e2](https://www.github.com/monetr/monetr/commit/3fdb0e226baefaa10f1f32d687016fae12572192))
* **ui:** update dependency @date-io/moment to v2 ([4ef8c1d](https://www.github.com/monetr/monetr/commit/4ef8c1dbf3cf03a4551adaacb5e331fc6e76e173))
* **ui:** update dependency jest-mock-axios to v4.5.0 ([70c9f93](https://www.github.com/monetr/monetr/commit/70c9f93e3ce352dbb5419f8aab92f25a0c1321f5))


### Miscellaneous

* Change release-please type to `helm` ([a1229d7](https://www.github.com/monetr/monetr/commit/a1229d7d1920db26fb384fa7c0d686d2cc1580ba))
* **ci:** Cleaned up (removed) JUnit job steps. ([acd7160](https://www.github.com/monetr/monetr/commit/acd7160c967ef4e5182149c4b4bc4f31cb756b78))
* **deps:** update dependency prettier to v2.5.1 ([b1a13ae](https://www.github.com/monetr/monetr/commit/b1a13ae0c158dbb2df33a2c768ed55c568d8abdd))
* Finalize semantic config for renovate ([af71df6](https://www.github.com/monetr/monetr/commit/af71df6c9f8c3a857ccc96b736d7800ac967bdf8))
* Fixing semantic commit types ([5bc057b](https://www.github.com/monetr/monetr/commit/5bc057b1fb1e12cd04141ee9fd9edc407c948d6c))
* Trying to get semantic commits working. ([4b614d8](https://www.github.com/monetr/monetr/commit/4b614d873f98e0aa698f8278ccf4abdd7e0e014b))
* Trying to improve renovate config. ([662b1e1](https://www.github.com/monetr/monetr/commit/662b1e1024d550e32577afcd40fc988f5ef95268))

### [0.7.2](https://www.github.com/monetr/monetr/compare/v0.7.1...v0.7.2) (2021-12-04)


### Bug Fixes

* Fixed verify forgot password in helm chart. ([5cc6ac2](https://www.github.com/monetr/monetr/commit/5cc6ac2883197f08644f8533b848ffdd9ce7de57))

### [0.7.1](https://www.github.com/monetr/monetr/compare/v0.7.0...v0.7.1) (2021-12-04)


### Bug Fixes

* Enabled forgot password in testing environment ([7bcdfdc](https://www.github.com/monetr/monetr/commit/7bcdfdc204799bdbb06faa43abbabdf127091ede))

## [0.7.0](https://www.github.com/monetr/monetr/compare/v0.6.16...v0.7.0) (2021-12-03)


### Features

* Adding `Forgot Password` to login page. ([48b0757](https://www.github.com/monetr/monetr/commit/48b075728fad9b877f3f7f0831ffef48653988d2)), closes [#92](https://www.github.com/monetr/monetr/issues/92)


### Bug Fixes

* **deps:** update module github.com/golang-jwt/jwt/v4 to v4.2.0 ([30ea8e6](https://www.github.com/monetr/monetr/commit/30ea8e6da49637a7d41ca9aff6518b13fcdc46be))

### [0.6.16](https://www.github.com/monetr/monetr/compare/v0.6.15...v0.6.16) (2021-12-03)


### Bug Fixes

* **deps:** update dependency @svgr/webpack to v6.1.0 ([134c39c](https://www.github.com/monetr/monetr/commit/134c39cade4c2ffe1e02c48546b48a32e9d4ffe2))
* **deps:** update dependency eslint-plugin-testing-library to v5.0.1 ([cbcc1c4](https://www.github.com/monetr/monetr/commit/cbcc1c4f532fbe93dc2d2ba761ab2cd2494327e8))
* **deps:** update dependency workbox-webpack-plugin to v6.4.2 ([8a355ae](https://www.github.com/monetr/monetr/commit/8a355ae13ba7031ce60f36c0d67bd32907507399))
* **deps:** update module github.com/gomodule/redigo to v1.8.6 ([298dc96](https://www.github.com/monetr/monetr/commit/298dc96bdf72270cd16a46c3c9ab2ada45f4b18e))
* **deps:** update module github.com/kataras/iris/v12 to v12.2.0-alpha5 ([0d59b95](https://www.github.com/monetr/monetr/commit/0d59b9524e4eb0d5e258175fca6a830b5bc9cecf))
* **deps:** update module github.com/plaid/plaid-go to v1.8.0 ([91a1119](https://www.github.com/monetr/monetr/commit/91a1119b8061c3199d28e45f8efc898b3c931547))
* Improving documentation for API. ([ebf3f42](https://www.github.com/monetr/monetr/commit/ebf3f42e63b49c13bd5717af5ad8c35cd6c98f22))

### [0.6.15](https://www.github.com/monetr/monetr/compare/v0.6.14...v0.6.15) (2021-11-30)


### Bug Fixes

* **deps:** update dependency sass to v1.44.0 ([c94be80](https://www.github.com/monetr/monetr/commit/c94be8068852b0e157fc88e20e988cdf05b74185))

### [0.6.14](https://www.github.com/monetr/monetr/compare/v0.6.13...v0.6.14) (2021-11-29)


### Bug Fixes

* **deps:** update material-ui monorepo ([751d4f7](https://www.github.com/monetr/monetr/commit/751d4f7cb022b7a8b81db42dd9bed8051086a79f))
* **deps:** update module github.com/brianvoe/gofakeit/v6 to v6.10.0 ([514e1bd](https://www.github.com/monetr/monetr/commit/514e1bdb361183e027f6fd381b2a486763b0816a))
* **deps:** update typescript-eslint monorepo to v5.5.0 ([47f2ef3](https://www.github.com/monetr/monetr/commit/47f2ef38ab7d8e96383385b5c14b0582d36ce5d8))

### [0.6.13](https://www.github.com/monetr/monetr/compare/v0.6.12...v0.6.13) (2021-11-29)


### Bug Fixes

* added basic config generate, local dev config relocated. ([2d0cd8d](https://www.github.com/monetr/monetr/commit/2d0cd8d6b42d7963cb3e0696b7e2e2ea283fffdb))
* Adding way to store secrets outside working dir. ([45b7dde](https://www.github.com/monetr/monetr/commit/45b7dde6d32d1846ba502568070f6dc656d99c74))
* Cleaned up unused makefile stuff. ([aa4f596](https://www.github.com/monetr/monetr/commit/aa4f5967fd39b5e77dcf7f93d722a5977478d890))
* Improve sentry span status reporting. ([3dfcb14](https://www.github.com/monetr/monetr/commit/3dfcb149e9e4e08d826afb42b389069da092fc75))
* Improved development.yaml support. ([66f2a9f](https://www.github.com/monetr/monetr/commit/66f2a9f39e05a4a6f66ae8a4c154e1c560e9a831))

### [0.6.12](https://www.github.com/monetr/monetr/compare/v0.6.11...v0.6.12) (2021-11-28)


### Bug Fixes

* Added some basic Sentry transaction to bootstrap. ([51c7bcf](https://www.github.com/monetr/monetr/commit/51c7bcf41c172e965e935dda231516b296058ff1))
* Don't report to sentry for unauthorization errors. ([df02e88](https://www.github.com/monetr/monetr/commit/df02e88cb299a846b1c443e444ae6a67688c7b51))
* **stripe:** Deprecate Stripe public key from UI. ([7e1b38c](https://www.github.com/monetr/monetr/commit/7e1b38c7862739999400d86ee73aae52f2d79993))

### [0.6.11](https://www.github.com/monetr/monetr/compare/v0.6.10...v0.6.11) (2021-11-28)


### Bug Fixes

* Fixed how bootstrapLogin would throw exceptions. ([e6add7b](https://www.github.com/monetr/monetr/commit/e6add7bbd9c6dd46e9694e00d573d2b142a6ec27))
* Fixed sentry submit feedback, fixed Sentry sourcemaps. ([aaf46a5](https://www.github.com/monetr/monetr/commit/aaf46a5f2bbc3a671fbb0568d279ae80cfa9a344))

### [0.6.10](https://www.github.com/monetr/monetr/compare/v0.6.9...v0.6.10) (2021-11-28)


### Bug Fixes

* Fix (hopefully) rollout bug with v0.6.9 ([9e5e067](https://www.github.com/monetr/monetr/commit/9e5e067dd26bf5895329d0dd797175964332c12b))
* Improving notificcations for other components. ([af482bc](https://www.github.com/monetr/monetr/commit/af482bc81261ba724ab4d58f5e0a2fed4c0ce4ce))
* Improving snackbar notifications. ([51dab3d](https://www.github.com/monetr/monetr/commit/51dab3d9dc36a6e3c5f2eb5aceed509ef6f42744))

### [0.6.9](https://www.github.com/monetr/monetr/compare/v0.6.8...v0.6.9) (2021-11-27)


### Bug Fixes

* Upgrading to react router v6 + more. ([5ab4cbd](https://www.github.com/monetr/monetr/commit/5ab4cbd7e7ee4a443cbf3218c159807d185ce20d))

### [0.6.8](https://www.github.com/monetr/monetr/compare/v0.6.7...v0.6.8) (2021-11-27)


### Bug Fixes

* **deps:** update dependency @emotion/react to v11.7.0 ([2178922](https://www.github.com/monetr/monetr/commit/2178922a025bd43b2dec1269c0f949b0fc4fae82))
* **deps:** update dependency immer to v9.0.7 ([77dff34](https://www.github.com/monetr/monetr/commit/77dff34266766d77f31b59c2ecc8d4a2aadb96d1))
* **deps:** update dependency redux-thunk to v2.4.1 ([0a90e9e](https://www.github.com/monetr/monetr/commit/0a90e9ee6cd905e853e5c53715aaf8525b314615))
* **deps:** update dependency sass to v1.43.5 ([11898a8](https://www.github.com/monetr/monetr/commit/11898a85f592b19ba1660d499e0a6210c63c1dd8))
* **deps:** update material-ui monorepo ([015721f](https://www.github.com/monetr/monetr/commit/015721f37df0fbd43069bab6e544586bb1098a6f))
* Disable sentry locally, it is not useful without credentials. ([2f33469](https://www.github.com/monetr/monetr/commit/2f33469e201c494b577eeef79090643c69a0daab))

### [0.6.7](https://www.github.com/monetr/monetr/compare/v0.6.6...v0.6.7) (2021-11-24)


### Bug Fixes

* Added ability to limit the number of Plaid links. ([8eff39f](https://www.github.com/monetr/monetr/commit/8eff39f48088085c580c9f396d98e3e73e9cbfbd)), closes [#341](https://www.github.com/monetr/monetr/issues/341)
* **deps:** update dependency is-svg to v4.3.2 ([042e585](https://www.github.com/monetr/monetr/commit/042e585f71740d62531af96a4ba5d202490dd347))

### [0.6.6](https://www.github.com/monetr/monetr/compare/v0.6.5...v0.6.6) (2021-11-24)


### Bug Fixes

* **deps:** update dependency eslint-plugin-jest to v25.3.0 ([9195d71](https://www.github.com/monetr/monetr/commit/9195d71cc2ad3b0f9791299532e8839cb3b5f862))
* **deps:** update material-ui monorepo ([54dc2a5](https://www.github.com/monetr/monetr/commit/54dc2a59468ec34617f9e33b49a9bcb2a22bad7d))
* Improve some errors and performance monitoring. ([2e683cb](https://www.github.com/monetr/monetr/commit/2e683cb276bac0a2806033d3f891deddfff34e47))
* Improving embedded ui makefile. ([a559ec7](https://www.github.com/monetr/monetr/commit/a559ec776fba62b954001007f83f2d0e4212a822))

### [0.6.5](https://www.github.com/monetr/monetr/compare/v0.6.4...v0.6.5) (2021-11-22)


### Bug Fixes

* **deps:** update dependency eslint to v8.3.0 ([64641f2](https://www.github.com/monetr/monetr/commit/64641f2944b0e53380fa7ece77116030abec8f25))

### [0.6.4](https://www.github.com/monetr/monetr/compare/v0.6.3...v0.6.4) (2021-11-21)


### Bug Fixes

* Fixed container being tagd with a `v` prefix. ([d3e128e](https://www.github.com/monetr/monetr/commit/d3e128e2f611d45624661f7f8aab323b5d3db0cc))

### [0.6.3](https://www.github.com/monetr/monetr/compare/v0.6.2...v0.6.3) (2021-11-21)


### Bug Fixes

* Purge unused tailwind CSS styles. ([d47f544](https://www.github.com/monetr/monetr/commit/d47f5448fd127df542da4a27e29f88f5704c47b9))

### [0.6.2](https://www.github.com/monetr/monetr/compare/v0.6.1...v0.6.2) (2021-11-20)


### Bug Fixes

* **deps:** update dependency eslint-plugin-react to v7.27.1 ([5ef5619](https://www.github.com/monetr/monetr/commit/5ef56190410043bf61cadb62008ecc68df1c345b))
* Fixed version being passed incorrectly to build. ([6f0b40c](https://www.github.com/monetr/monetr/commit/6f0b40cbf18575187b5b31007987f4484f1427e0))

### [0.6.1](https://www.github.com/monetr/monetr/compare/v0.6.0...v0.6.1) (2021-11-20)


### Bug Fixes

* Fixed buildah container pushing. ([ae9c8fa](https://www.github.com/monetr/monetr/commit/ae9c8fa8515c617228329b914f1431e67737bba4))

## [0.6.0](https://www.github.com/monetr/monetr/compare/v0.5.1...v0.6.0) (2021-11-20)


### Features

* Switching to buildah for container builds. ([dbc370d](https://www.github.com/monetr/monetr/commit/dbc370d97cfacdc0d9419519ee6a6b4f8fb94ba2))


### Bug Fixes

* **deps:** update dependency mini-css-extract-plugin to v2.4.5 ([30521d0](https://www.github.com/monetr/monetr/commit/30521d06206b3ceb40ac1b010e6bca128c7a519d))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.76.0 ([ca68675](https://www.github.com/monetr/monetr/commit/ca686755f2fc5826f577de39d9435b6db1b8c4a1))
* Fixed release workflow mistake. ([e5f2db2](https://www.github.com/monetr/monetr/commit/e5f2db2802d902a15771b061180c7f8114de832f))

### [0.5.1](https://www.github.com/monetr/monetr/compare/v0.5.0...v0.5.1) (2021-11-17)


### Bug Fixes

* **deps:** update dependency postcss-preset-env to v7 ([ec533a8](https://www.github.com/monetr/monetr/commit/ec533a8d03568bf9f09d0208606350b8f25cb21c))
* **deps:** update material-ui monorepo ([434033c](https://www.github.com/monetr/monetr/commit/434033c5aa15448be6ce150e3e8dbe89ba4a4f9f))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.75.0 ([79bcb65](https://www.github.com/monetr/monetr/commit/79bcb6558c1a7b52f71ffa3cbe1af2ed6dfb6ae9))
* **deps:** update sentry-javascript monorepo to v6.15.0 ([043b66a](https://www.github.com/monetr/monetr/commit/043b66aebbee25b8b6ab91c480752186d91ecd00))

## [0.5.0](https://www.github.com/monetr/monetr/compare/v0.4.17...v0.5.0) (2021-11-16)


### ⚠ BREAKING CHANGES

* **sentry:** Created an axios based Sentry transport.

### Bug Fixes

* **deps:** update dependency workbox-webpack-plugin to v6.4.1 ([8a313e4](https://www.github.com/monetr/monetr/commit/8a313e4c1e401df764ee5e5d7cdb8b15f5026d59))
* **sentry:** Created an axios based Sentry transport. ([c4217a4](https://www.github.com/monetr/monetr/commit/c4217a4cdd0625086925ae0a7734633183f3cd88))


### Miscellaneous Chores

* release 0.5.0 ([7fac0d1](https://www.github.com/monetr/monetr/commit/7fac0d1c988491fd7ff1cb2c5994b5ca038bdbc4))

### [0.4.17](https://www.github.com/monetr/monetr/compare/v0.4.16...v0.4.17) (2021-11-15)


### Bug Fixes

* Changed HTTP testing to use httpexpect. ([26a6982](https://www.github.com/monetr/monetr/commit/26a69828ab1c1ae65d90f6078129f3e87a608d48))
* Cleaning up dependencies. ([2d3456f](https://www.github.com/monetr/monetr/commit/2d3456f21d544baf84bcaed38bfe0c240b83a92a))
* **deps:** update dependency camelcase to v6.2.1 ([e5b4f3f](https://www.github.com/monetr/monetr/commit/e5b4f3f974dd89be98126a59811cef1ab0d394df))
* **deps:** update google.golang.org/genproto commit hash to 271947f ([39c84df](https://www.github.com/monetr/monetr/commit/39c84dfe3bf1ea57a3724e3ae46ce7558029e311))
* **deps:** update typescript-eslint monorepo to v5.4.0 ([f15c12c](https://www.github.com/monetr/monetr/commit/f15c12ca1a8b47f191b482241c58f08acae78753))
* Updated renovate config. ([882a853](https://www.github.com/monetr/monetr/commit/882a853060b77629233d025d345814dbfca3f215))

### [0.4.16](https://www.github.com/monetr/monetr/compare/v0.4.15...v0.4.16) (2021-11-15)


### Bug Fixes

* Allow separate Sentry DSN for UI. ([c01dac1](https://www.github.com/monetr/monetr/commit/c01dac163e8d68adc19f0f58aa051de9206558a9))
* Sentry relay testing locally. Updated content policy. ([7efa38c](https://www.github.com/monetr/monetr/commit/7efa38cafc88b2cbf1ce46022666fe00680ddb0b))

### [0.4.15](https://www.github.com/monetr/monetr/compare/v0.4.14...v0.4.15) (2021-11-14)


### Bug Fixes

* Adding security headers and removing stripe-js. ([fc0fa6a](https://www.github.com/monetr/monetr/commit/fc0fa6a89177e59273d53fddc3024411d5a5e7e4)), closes [#304](https://www.github.com/monetr/monetr/issues/304)

### [0.4.14](https://www.github.com/monetr/monetr/compare/v0.4.13...v0.4.14) (2021-11-14)


### Bug Fixes

* **deps:** update emotion monorepo to v11.6.0 ([79d202d](https://www.github.com/monetr/monetr/commit/79d202dc4de150d3554a5cfdf70d73e4b883f32d))

### [0.4.13](https://www.github.com/monetr/monetr/compare/v0.4.12...v0.4.13) (2021-11-13)


### Bug Fixes

* All Stackdriver label values must be strings. ([29fa381](https://www.github.com/monetr/monetr/commit/29fa381dad577b037e4f66a0c833ae4d9f267230))
* Fixed test for stack driver formatter wrapper. ([bffb038](https://www.github.com/monetr/monetr/commit/bffb0387ece4e89a6d5d88060b0faa8b98ca8dea))
* Fixing labels in Stackdriver. ([2b48517](https://www.github.com/monetr/monetr/commit/2b48517bdd87e5f5711ce97f1abac1ad69796929))

### [0.4.12](https://www.github.com/monetr/monetr/compare/v0.4.11...v0.4.12) (2021-11-13)


### Bug Fixes

* **deps:** update module github.com/micahparks/keyfunc to v0.10.0 ([6b2ea09](https://www.github.com/monetr/monetr/commit/6b2ea0940f3fc6ff442f2695d0438bcff120e640))

### [0.4.11](https://www.github.com/monetr/monetr/compare/v0.4.10...v0.4.11) (2021-11-13)


### Bug Fixes

* Fixed failing test for stackdriver logging. ([b25d851](https://www.github.com/monetr/monetr/commit/b25d851190c5c81f29b2fc779f5a91728ebb922e))

### [0.4.10](https://www.github.com/monetr/monetr/compare/v0.4.9...v0.4.10) (2021-11-12)


### Bug Fixes

* Really really really really fixed logging this time. ([6e5667e](https://www.github.com/monetr/monetr/commit/6e5667e8ec31e1e69cdeef197c17624d5b816bcf))

### [0.4.9](https://www.github.com/monetr/monetr/compare/v0.4.9...v0.4.9) (2021-11-12)


### Miscellaneous Chores

* release 0.4.9 ([ba57c98](https://www.github.com/monetr/monetr/commit/ba57c9860ab6d7d5b200a2c2246fbae8433a31dd))

### [0.4.8](https://www.github.com/monetr/monetr/compare/v0.4.7...v0.4.8) (2021-11-12)


### Bug Fixes

* Added better Sentry error reporting. ([aa7d312](https://www.github.com/monetr/monetr/commit/aa7d312b50f14c4ec529bc32301c686654b730b9))
* Removed `docker` from Brewfile. ([b392127](https://www.github.com/monetr/monetr/commit/b39212720d3703f0e9c647c8a0cce0c5107f395f))

### [0.4.7](https://www.github.com/monetr/monetr/compare/v0.4.6...v0.4.7) (2021-11-12)


### Bug Fixes

* Added windows/arm64 to regular binary build. ([ade0365](https://www.github.com/monetr/monetr/commit/ade036579e005039cfab22d6f0af25bcfc5cf7d5))
* Fixed logging issue due to bad duplication. ([484caa4](https://www.github.com/monetr/monetr/commit/484caa4dd54239813c6b047f02344dcc3cb85bde))

### [0.4.6](https://www.github.com/monetr/monetr/compare/v0.4.5...v0.4.6) (2021-11-11)


### Bug Fixes

* Added source maps and JobID logging. ([a96b900](https://www.github.com/monetr/monetr/commit/a96b900d306429599a2d69913a5c209896f91a17))
* Adding better support for Stackdriver logging. ([694c86d](https://www.github.com/monetr/monetr/commit/694c86d96aeae47795008459bfbb0e9f03f03c79))
* **deps:** update dependency eslint-webpack-plugin to v3.1.1 ([3289e65](https://www.github.com/monetr/monetr/commit/3289e65103fafa72d17b3301e14959d92f63a9cf))
* **deps:** update sentry-javascript monorepo to v6.14.2 ([f52b5fc](https://www.github.com/monetr/monetr/commit/f52b5fc2dad1bf3b23ea3b0b55507d8ba1d3bbce))
* **deps:** update sentry-javascript monorepo to v6.14.3 ([1c73ae8](https://www.github.com/monetr/monetr/commit/1c73ae865e6be520c0a8ed151971980c18eec2e6))

### [0.4.5](https://www.github.com/monetr/monetr/compare/v0.4.4...v0.4.5) (2021-11-11)


### Bug Fixes

* Bit of cleanup and memo of nav bar. ([2119c4e](https://www.github.com/monetr/monetr/commit/2119c4e8af0d722428e89e83b2ec1e3d1c54ed24))
* Converting a lot more components to hooks/functional. ([3b7aeea](https://www.github.com/monetr/monetr/commit/3b7aeeaca854a8347972c214d6623e9e28d55c6a))
* Declared variables for webpack define. ([d4f21eb](https://www.github.com/monetr/monetr/commit/d4f21ebe5b394315ba2a5c5ef0ca7c4284edcfce))
* **deps:** update dependency eslint-plugin-import to v2.25.3 ([95c6b8f](https://www.github.com/monetr/monetr/commit/95c6b8fbefc159ed664b443e3f92057f22a08090))
* **deps:** update dependency eslint-plugin-jsx-a11y to v6.5.1 ([0f50b05](https://www.github.com/monetr/monetr/commit/0f50b05b083cf4c6903db23c3b3a09190211fc5d))
* **deps:** update dependency eslint-plugin-react to v7.27.0 ([f980114](https://www.github.com/monetr/monetr/commit/f980114c2806f03111d05e4f742c31954383f405))

### [0.4.4](https://www.github.com/monetr/monetr/compare/v0.4.3...v0.4.4) (2021-11-09)


### Bug Fixes

* Added `noreferrer` to github release link. ([c8f46ee](https://www.github.com/monetr/monetr/commit/c8f46ee834a5457061da92093f1443f9f1351add))
* Converting more components to functional components. ([6b7dd54](https://www.github.com/monetr/monetr/commit/6b7dd54e829379c302ed545114a188961098def9))
* **deps:** update dependency eslint-plugin-react-hooks to v4.3.0 ([410d9dc](https://www.github.com/monetr/monetr/commit/410d9dcc6f9e45c3bece2fb3780a9ec7ed2698f7))
* Reduce memory requests in testing environment. ([6323808](https://www.github.com/monetr/monetr/commit/63238080940e439d2d542091cd8181885dd70346))
* Removed old unused setup view. ([9f01501](https://www.github.com/monetr/monetr/commit/9f0150112c88f302c8943339ab0b7906016c01d7))

### [0.4.3](https://www.github.com/monetr/monetr/compare/v0.4.2...v0.4.3) (2021-11-09)


### Bug Fixes

* Added release version to footer. ([58e1921](https://www.github.com/monetr/monetr/commit/58e1921c974d56cddfbdd87b89d37e558ea13625))

### [0.4.2](https://www.github.com/monetr/monetr/compare/v0.4.1...v0.4.2) (2021-11-09)


### Bug Fixes

* **deps:** update dependency eslint-plugin-jest to v25.2.4 ([b5ba346](https://www.github.com/monetr/monetr/commit/b5ba3466f967b67f318559bfd9ce5e291059c0f6))
* **deps:** update dependency terser-webpack-plugin to v5.2.5 ([5772a2a](https://www.github.com/monetr/monetr/commit/5772a2a1ece9b1e85acdb436b46be2301737aa01))
* **deps:** update material-ui monorepo ([0be2aa7](https://www.github.com/monetr/monetr/commit/0be2aa707cbe592aad152688234439c698b91b88))
* **deps:** update typescript-eslint monorepo to v5.3.1 ([cb92603](https://www.github.com/monetr/monetr/commit/cb92603417659c22a3875aab806c7da0dbff942c))
* Improvements for sentry, consistency. ([8576b15](https://www.github.com/monetr/monetr/commit/8576b159bc025784cb2909a7db449bdd9e1a8902))
* Massive Typescript improvements. ([7919941](https://www.github.com/monetr/monetr/commit/7919941d3c441b24fc6268bc8aebe67f5da7091d))

### [0.4.1](https://www.github.com/monetr/monetr/compare/v0.4.0...v0.4.1) (2021-11-08)


### Bug Fixes

* Added DST test to make sure I'm not going insane. ([bb40ab4](https://www.github.com/monetr/monetr/commit/bb40ab49adf74f9978393bd5fa1e4dac03b4f8c4))
* Fixed race condition in pubsub test. ([59a5d78](https://www.github.com/monetr/monetr/commit/59a5d782afe8a37d9c5df79f0565451d93054e50)), closes [#272](https://www.github.com/monetr/monetr/issues/272)
* Improve code coverage for testutils. ([b287949](https://www.github.com/monetr/monetr/commit/b287949b2991cd6f322cdc80a933b3fe671c2d14))
* Increased notification delay, fixed React issue. ([637ca66](https://www.github.com/monetr/monetr/commit/637ca66da5e372187bb0cb3b1d1ee43a78ca6656))
* Laying the groundwork for password resetting. ([06db23a](https://www.github.com/monetr/monetr/commit/06db23a1956f809768e8825454a2576702c77a7a))
* Log `item_id` and `bank_account_id` for Plaid requests. ([d9acf8d](https://www.github.com/monetr/monetr/commit/d9acf8d2b36abe71e48f53f850852605412c7d54)), closes [#269](https://www.github.com/monetr/monetr/issues/269)
* Minor improvements, testing pub sub. ([aa26bf7](https://www.github.com/monetr/monetr/commit/aa26bf75c3d8cc7ef1ae1cd4256ecaceb6d6035c))
* Testing improvements and login documentation. ([a5d64e4](https://www.github.com/monetr/monetr/commit/a5d64e4d3a59df8fb6da0f52db1f43b33d16ace0))

## [0.4.0](https://www.github.com/monetr/monetr/compare/v0.3.9...v0.4.0) (2021-11-07)


### Features

* Adding support for tax collection via Stripe. ([ef385e5](https://www.github.com/monetr/monetr/commit/ef385e5a9ea232a1ea504ea49544da9c58b8b04d)), closes [#261](https://www.github.com/monetr/monetr/issues/261)


### Bug Fixes

* Added failure documentation for `/api/health` ([b7202e4](https://www.github.com/monetr/monetr/commit/b7202e418486d80e5b93f18511d1ce5d1d875098))
* Added transactions index to improve query performance. ([2e0f0b6](https://www.github.com/monetr/monetr/commit/2e0f0b697054efcac5b8bbf11f19aa00f75b509c)), closes [#265](https://www.github.com/monetr/monetr/issues/265)
* Fixed (hopefully) stripe requests showing missing implementation. ([6589ee0](https://www.github.com/monetr/monetr/commit/6589ee0281ab05c32869acaa89733d665fc7299e)), closes [#251](https://www.github.com/monetr/monetr/issues/251)
* Fixed cleanup job name in sentry. ([f67bd17](https://www.github.com/monetr/monetr/commit/f67bd17a0bc4b8362441e42e776002c7ef56d509)), closes [#250](https://www.github.com/monetr/monetr/issues/250)
* Fixed setup notification sending before ready. ([e1e341f](https://www.github.com/monetr/monetr/commit/e1e341f7fcddf0fa83490d6894ddffbb9c22bbf2)), closes [#262](https://www.github.com/monetr/monetr/issues/262)
* Improved health endpoint + logging. ([83c2989](https://www.github.com/monetr/monetr/commit/83c29892eceb295b0039a06c4a0fbc064f6e65f6))
* Tweaked default log level for commands. ([8f4731a](https://www.github.com/monetr/monetr/commit/8f4731a8d699da94360b824bf655afdf7ed51eba))
* Upgraded container images. ([c8df3b5](https://www.github.com/monetr/monetr/commit/c8df3b5918421387b283a1d0b08c7e6b0539fbcd))

### [0.3.9](https://www.github.com/monetr/monetr/compare/v0.3.8...v0.3.9) (2021-11-06)


### Bug Fixes

* Significantly improved logging and metadata on log entries. ([7f3d35e](https://www.github.com/monetr/monetr/commit/7f3d35ebba7f79993bc461fd0c26df3607455477))

### [0.3.8](https://www.github.com/monetr/monetr/compare/v0.3.7...v0.3.8) (2021-11-06)


### Bug Fixes

* Adjusted SMTP port for sendgrid. ([7ec1502](https://www.github.com/monetr/monetr/commit/7ec1502b53ef3b48d2074aa1634e488d5dbaaa1b))

### [0.3.7](https://www.github.com/monetr/monetr/compare/v0.3.6...v0.3.7) (2021-11-06)


### Bug Fixes

* **deps:** update dependency eslint to v8.2.0 ([2efd618](https://www.github.com/monetr/monetr/commit/2efd618515fd56941be2aa3828148e89eeae53f8))
* **deps:** update dependency react-select to v5.2.1 ([fcc63de](https://www.github.com/monetr/monetr/commit/fcc63de8cf80a396a0dd22fe730dc06dead2f211))
* **deps:** update sentry-javascript monorepo to v6.14.1 ([3e6a737](https://www.github.com/monetr/monetr/commit/3e6a7379647be2b1cf6316fe5d82c6abcdd4cd99))

### [0.3.6](https://www.github.com/monetr/monetr/compare/v0.3.5...v0.3.6) (2021-11-05)


### Bug Fixes

* Reverted change to how release is articulated to Sentry. ([6062475](https://www.github.com/monetr/monetr/commit/60624751d3d9f0074f477b45449d26987026272a))

### [0.3.5](https://www.github.com/monetr/monetr/compare/v0.3.4...v0.3.5) (2021-11-05)


### Bug Fixes

* Fixed incorrect build args for container build. ([3080eb4](https://www.github.com/monetr/monetr/commit/3080eb47c1442b02e80c991b3a2f323c40c6588e)), closes [#241](https://www.github.com/monetr/monetr/issues/241)

### [0.3.4](https://www.github.com/monetr/monetr/compare/v0.3.3...v0.3.4) (2021-11-04)


### Bug Fixes

* Allow log format to be configured. ([0aacf0a](https://www.github.com/monetr/monetr/commit/0aacf0a5fa8ab865b78233fb8d0a12b06d73042d)), closes [#237](https://www.github.com/monetr/monetr/issues/237)
* **deps:** update dependency eslint-plugin-jest to v25.2.3 ([90a35f0](https://www.github.com/monetr/monetr/commit/90a35f01e1aca87583f21d2cd38aafc0a18c1a65))

### [0.3.3](https://www.github.com/monetr/monetr/compare/v0.3.2...v0.3.3) (2021-11-04)


### Bug Fixes

* Added Sentry releases to release flow. ([0229715](https://www.github.com/monetr/monetr/commit/0229715fb4a9c139d75979e0e4958295f4550f01)), closes [#232](https://www.github.com/monetr/monetr/issues/232)

### [0.3.2](https://www.github.com/monetr/monetr/compare/v0.3.1...v0.3.2) (2021-11-04)


### Bug Fixes

* **deps:** update dependency mini-css-extract-plugin to v2.4.4 ([45a2aa6](https://www.github.com/monetr/monetr/commit/45a2aa64e4cb32ec960071ed4a030735dcd491ed))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.73.1 ([cd836a2](https://www.github.com/monetr/monetr/commit/cd836a2b1dfdab12d23cd67332173dd13862a136))
* Fixed SMTP port used for SendGrid. ([1f4f8fe](https://www.github.com/monetr/monetr/commit/1f4f8feb1208a7d3b495a0dc663051a26ed3108d)), closes [#231](https://www.github.com/monetr/monetr/issues/231)
* Improve error reporting to sentry for failed requests. ([71b1018](https://www.github.com/monetr/monetr/commit/71b1018b1af9443dc7812160f1d5925b0d0f5f63))

### [0.3.1](https://www.github.com/monetr/monetr/compare/v0.3.0...v0.3.1) (2021-11-04)


### Bug Fixes

* Added `X-Real-Ip` header to derive client IP address. ([894c1a9](https://www.github.com/monetr/monetr/commit/894c1a98a5d48cffe784bfebe4fb2cd53957f96a))

## [0.3.0](https://www.github.com/monetr/monetr/compare/v0.2.1...v0.3.0) (2021-11-04)


### Features

* Added "dogfooding" environment for alpha-testing. ([6f02d2d](https://www.github.com/monetr/monetr/commit/6f02d2d91a0c54fe52a0abfc54c8b351e13c096e))


### Bug Fixes

* **deps:** update dependency @stripe/stripe-js to v1.21.1 ([cb3c740](https://www.github.com/monetr/monetr/commit/cb3c740ca10a784562a7732b7b13c749733f0e03))
* **deps:** update module github.com/plaid/plaid-go to v1.7.0 ([7516299](https://www.github.com/monetr/monetr/commit/75162999fdbb19c0e44be8f43a3dee97826b2469))
* **deps:** update sentry-javascript monorepo to v6.14.0 ([80649ff](https://www.github.com/monetr/monetr/commit/80649fff5bfc32a7284410cbf7223c697e55dd1e))

### [0.2.1](https://www.github.com/monetr/monetr/compare/v0.2.0...v0.2.1) (2021-11-02)


### Bug Fixes

* Added terraform to local dependencies. ([4221f14](https://www.github.com/monetr/monetr/commit/4221f140954bca7709865e9f5df5d48b7d294c20))
* **deps:** update dependency @stripe/stripe-js to v1.21.0 ([18e146f](https://www.github.com/monetr/monetr/commit/18e146fa2938e8a4a2150b2ac974abf84d2ce0cd))
* **deps:** update dependency @testing-library/jest-dom to v5.15.0 ([1506f9f](https://www.github.com/monetr/monetr/commit/1506f9fd5333b5425f46d946701ffccc4afaafad))
* Ensure `client_name` sent to Plaid is correct. ([6af5f72](https://www.github.com/monetr/monetr/commit/6af5f72238abdc3b725c16b7ff24fe36ca507421)), closes [#185](https://www.github.com/monetr/monetr/issues/185)
* Ensure `products` parameter of `/link/token/create` is correct. ([28c8fa1](https://www.github.com/monetr/monetr/commit/28c8fa1a4c30d15c71bf41005aecd2102c65d9a8)), closes [#186](https://www.github.com/monetr/monetr/issues/186)
* Fixed `ARCH` not being set in Makefile. ([7389583](https://www.github.com/monetr/monetr/commit/738958380102816714f76a3c6f3ea2511fcded72))
* Fixed `make init-mini` vault bug. ([6129da6](https://www.github.com/monetr/monetr/commit/6129da6719e475e1d0f09d35989f0a8f8134460a))
* Fixed `RedirectToCheckoutOptions` import change in Stripe upgrade. ([b967871](https://www.github.com/monetr/monetr/commit/b9678714d664b2820dbff32fca2aa0aeb4d685a5))
* Improving Dockerfile build efficiency. ([e9277db](https://www.github.com/monetr/monetr/commit/e9277db2ba2863c17754542668bb2f21070ef7a5))
* Moved institution details endpoint and allow institution ID param. ([c82e8bd](https://www.github.com/monetr/monetr/commit/c82e8bd6deaff6d5154246b40aaf918f1242ac10))
* No longer require `docker` for local development. ([7bd3afb](https://www.github.com/monetr/monetr/commit/7bd3afbf23a74310214ae57ff3c55a064ea2bfa4))
* Removed PostgreSQL tests, makefile and github actions cleanup. ([e0dfab9](https://www.github.com/monetr/monetr/commit/e0dfab907f41d805bb3760c6e4772978b298d4ae))
* Removed unusued go tools. ([5d6c35b](https://www.github.com/monetr/monetr/commit/5d6c35bdc60f106f827c88d05256061990cf9b54))

## [0.2.0](https://www.github.com/monetr/monetr/compare/v0.1.1...v0.2.0) (2021-11-02)


### Features

* Preventing duplicate item adds. ([76b3036](https://www.github.com/monetr/monetr/commit/76b30367232bbc8edeacdf5521f587c7c4506341)), closes [#193](https://www.github.com/monetr/monetr/issues/193)


### Bug Fixes

* Added vault parameters to the helm chart environment variables. ([a14b73d](https://www.github.com/monetr/monetr/commit/a14b73d4aacd8a027413467eb98d6a4e12c2f46b)), closes [#200](https://www.github.com/monetr/monetr/issues/200)
* **deps:** update dependency react-select to v5.2.0 ([167a102](https://www.github.com/monetr/monetr/commit/167a102cda9df5529efd1443f75dbd51513faa91))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.73.0 ([a3aaa68](https://www.github.com/monetr/monetr/commit/a3aaa68cdec986422726216874a535cac03ef6c5))
* **deps:** update typescript-eslint monorepo to v5.3.0 ([f1aa47e](https://www.github.com/monetr/monetr/commit/f1aa47e2891448c2cf9d1919ec6156352f2125f8))
* Fixed failing test due to Link table changes. ([2869daa](https://www.github.com/monetr/monetr/commit/2869daa3acd35589b34e449c9dd58034ce6d2685))
* Fixed vault secret expiring and not being refreshed. ([c10c731](https://www.github.com/monetr/monetr/commit/c10c731c663b6483bac01e61a3adf40256396123))
* **vault:** Fixed vault authentication not refreshing. ([c989267](https://www.github.com/monetr/monetr/commit/c9892678319ebca8328bf99a9057a478e75bc09b))

### [0.1.1](https://www.github.com/monetr/monetr/compare/v0.1.0...v0.1.1) (2021-10-31)


### Bug Fixes

* **container:** Delete the apt-get lists after installing something ([edd4a38](https://www.github.com/monetr/monetr/commit/edd4a38859148e5edc1d3bbb649f4fbbc22727be))
* **container:** Pin apt-get package versions. ([d887f52](https://www.github.com/monetr/monetr/commit/d887f52662857d9c1eb02c6aa37074ab65d053d4))
* **deps:** Pinned react-refresh. ([c282814](https://www.github.com/monetr/monetr/commit/c282814af1d1adddf287d81dfd215ae14bb29477))
* **deps:** Updated kataras/iris to the latest version. ([1cb0fed](https://www.github.com/monetr/monetr/commit/1cb0fed806dc163c595ce3c1a80a301505796dec))
* **deps:** Upgrading to MUI V5. ([adb3294](https://www.github.com/monetr/monetr/commit/adb3294056b6e8def3cb62aca5ba555143f88841))
* Fixed key error on goals view. ([6f90ab1](https://www.github.com/monetr/monetr/commit/6f90ab160838b29fb084d645d1e58e4701e8dfd1))
* Improved CodeCov reporting on the main pipeline. ([2725235](https://www.github.com/monetr/monetr/commit/27252352d74936534998f9148f1ae696d65e75d5))
* Improved CodeCov uploading. ([0368752](https://www.github.com/monetr/monetr/commit/0368752be9fdd7f9ee56f027869c874f5af98bbc))
* Pinned dependencies for material-ui ([c624908](https://www.github.com/monetr/monetr/commit/c6249086dcf7d8ed98b5e17e56185259bb76efca))
* **webpack:** Adding react-refresh, refactor webpack config. ([75cdc09](https://www.github.com/monetr/monetr/commit/75cdc090f9f31e54552819c0e8f4f38ff93bb616))

## [0.1.0](https://www.github.com/monetr/monetr/compare/v0.0.11...v0.1.0) (2021-10-30)


### Features

* Added API endpoint to retrieve institution details via Link ID. ([7f96d6c](https://www.github.com/monetr/monetr/commit/7f96d6c1c9e345061b81e34efca41ecd644bb9c3))
* Added basic toggling of dark mode via local storage. ([c2c4728](https://www.github.com/monetr/monetr/commit/c2c47284a77a60ab64b931246b8195349e5f8518))
* Added CodeCov report to github actions. ([48c4cf3](https://www.github.com/monetr/monetr/commit/48c4cf37d751487524aff4d746b1863419d5adee))
* Adding support for institution statuses. ([73a59e8](https://www.github.com/monetr/monetr/commit/73a59e8b472af881dfef355e83bb1500bff052a0))


### Bug Fixes

* Actually fixed expenses being created on the current day. ([6155d91](https://www.github.com/monetr/monetr/commit/6155d91e2859a71abd20044d6479c433d1cf1ba0))
* Added build version and revision to generic binary builds. ([f015e7c](https://www.github.com/monetr/monetr/commit/f015e7cfddd86f62e13ec6798287d05e7089669e))
* Clean apt cache in container image. ([cda274c](https://www.github.com/monetr/monetr/commit/cda274c6482dcf5c2dbcfbdfff95382bac483482))
* Fixed helm template stats port issue. Improved local dev. ([224326d](https://www.github.com/monetr/monetr/commit/224326d53e7a5f7ddb9844d0e78f2579be8bd150))
* Refactored typescript models, fixed webpack. ([bc538d3](https://www.github.com/monetr/monetr/commit/bc538d39e727e9fa1d5e593ad855293056cf85e1))

### [0.0.11](https://www.github.com/monetr/monetr/compare/v0.0.10...v0.0.11) (2021-10-29)


### Bug Fixes

* Cleaned up unused transaction components. ([9517f7e](https://www.github.com/monetr/monetr/commit/9517f7e9133bb1c2bbfa5cc4bce8c95e8e7e13be))
* **deps:** update dependency eslint-plugin-flowtype to v8.0.3 ([02b3db6](https://www.github.com/monetr/monetr/commit/02b3db647d28a5b443636ff1231e3bf25a86fbd7))
* Fixed failing test from removed component. ([c118105](https://www.github.com/monetr/monetr/commit/c118105ef75a3f10c5031b58f4df2db2214c0ab4))
* Fixing funding stats view again. Should be correct now. ([5983c35](https://www.github.com/monetr/monetr/commit/5983c350459854ac2d60338d2cc7240312283162))

### [0.0.10](https://www.github.com/monetr/monetr/compare/v0.0.9...v0.0.10) (2021-10-28)


### Bug Fixes

* Checkout repository for the `gh` CLI for artifacts. ([b223497](https://www.github.com/monetr/monetr/commit/b2234970e0f4ec55a8a68f07e1fdf16fb4b3dbd6))

### [0.0.9](https://www.github.com/monetr/monetr/compare/v0.0.8...v0.0.9) (2021-10-28)


### Bug Fixes

* Added proper port config to helm chart. ([04c3a55](https://www.github.com/monetr/monetr/commit/04c3a55dd26405119f4eddbc1d4b7d57b08fe7e0))
* **deps:** update dependency redux to v4.1.2 ([9a1f280](https://www.github.com/monetr/monetr/commit/9a1f280846488d86d866866c82da5e88a407c6c9))
* **deps:** update module github.com/hashicorp/vault/api to v1.3.0 ([88e53b4](https://www.github.com/monetr/monetr/commit/88e53b43e903f17840099dfeafd2c28268031f40))
* Fixed build that broke due to refactor. ([ba560c5](https://www.github.com/monetr/monetr/commit/ba560c5ee171f9a73d58f2c9c183bfa20735f0a0))
* Hopefully fixing the release asset pipeline. ([4ce4b17](https://www.github.com/monetr/monetr/commit/4ce4b17d7eb95246557398134b84dc40a374e2f2))
* Improved testing around Plaid create links. ([f80b2a8](https://www.github.com/monetr/monetr/commit/f80b2a898b92b3d7b09e90b1121bd7646e91513e))
* Removed unused RedirectURI param for creating links. ([3a99c64](https://www.github.com/monetr/monetr/commit/3a99c64c01913d18a9dbf20889349e56667a07cd))

### [0.0.8](https://www.github.com/monetr/monetr/compare/v0.0.7...v0.0.8) (2021-10-28)


### Bug Fixes

* Allow configuring port for prometheus metrics. ([279edee](https://www.github.com/monetr/monetr/commit/279edeea06832e11646d7625f3c12c35b079fa4a))
* Cleaned up old files, trying to get actions working still. ([fc938ae](https://www.github.com/monetr/monetr/commit/fc938ae670db304070f0e514da389f67b4ec3fa3))

### [0.0.7](https://www.github.com/monetr/monetr/compare/v0.0.6...v0.0.7) (2021-10-27)


### Bug Fixes

* Hopefully fixed release asset uploading in pipelines. ([b6af463](https://www.github.com/monetr/monetr/commit/b6af4638d1322409fca494fa01675263d9b7f2db))

### [0.0.6](https://www.github.com/monetr/monetr/compare/v0.0.5...v0.0.6) (2021-10-27)


### Bug Fixes

* Fixed release flow for multiple OS's. Assets. ([757030c](https://www.github.com/monetr/monetr/commit/757030ca8ee7a5a0f0291d47314c3621b63ded3e))
* Hopefully fixed the GitHub release assets uploading. ([a3aa79c](https://www.github.com/monetr/monetr/commit/a3aa79c023b4ca042a33b449cddd81c12775aa77))

### [0.0.5](https://www.github.com/monetr/monetr/compare/v0.0.4...v0.0.5) (2021-10-27)


### Features

* Added ability to change listen port via config. ([c9481cd](https://www.github.com/monetr/monetr/commit/c9481cdbf7f4f56f0df4fb32b350997c057e977c))


### Miscellaneous Chores

* release 0.0.5 ([3958271](https://www.github.com/monetr/monetr/commit/395827115e102a5a870657be3f2efa6d0b02ddc1))

### [0.0.4](https://www.github.com/monetr/monetr/compare/v0.0.3...v0.0.4) (2021-10-27)


### Bug Fixes

* Added some super basic logging improvements. ([491c404](https://www.github.com/monetr/monetr/commit/491c404afeb6dc1444f08c96a5d825dea1ee7c6a))
* Added year to spending objects that are due a different year. ([bff0c7b](https://www.github.com/monetr/monetr/commit/bff0c7b6359d06ce1648a1307cbe04d0021ed07e)), closes [#147](https://www.github.com/monetr/monetr/issues/147)
* **deps:** update dependency eslint-webpack-plugin to v3.1.0 ([04902f7](https://www.github.com/monetr/monetr/commit/04902f7c40d750f901a5cca0779b7f391a31dfd8))
* **deps:** update dependency sass to v1.43.4 ([369f1ac](https://www.github.com/monetr/monetr/commit/369f1ac92dcbb2cd5c39d57ab29d4d91ba6cae5a))
* **deps:** update google.golang.org/genproto commit hash to 4688e4c ([b528a1b](https://www.github.com/monetr/monetr/commit/b528a1bdf39b2a4fe258292edfb97f5a2cafcd77))
* **deps:** update module github.com/plaid/plaid-go to v1.6.0 ([81bc574](https://www.github.com/monetr/monetr/commit/81bc574d32664711d138870842ae784398b9b7f9))
* Removed `plaid.ACCOUNTSUBTYPE_HOME` due to plaid upgrade. ([d04efaf](https://www.github.com/monetr/monetr/commit/d04efaf1cf2482f5470d7130d81c898f53923883))

### [0.0.3](https://www.github.com/monetr/monetr/compare/v0.0.2...v0.0.3) (2021-10-26)


### Bug Fixes

* Added Expires header to the static content handler. ([a70aeac](https://www.github.com/monetr/monetr/commit/a70aeacf9358e0c8d55f3320c20edb90b23f422e))
* Added job to cleanup old job records. ([21b89c3](https://www.github.com/monetr/monetr/commit/21b89c388e7a9c0707ec375609c00cdaf26bad67)), closes [#107](https://www.github.com/monetr/monetr/issues/107)
* **deps:** update dependency axios to v0.24.0 ([533ff66](https://www.github.com/monetr/monetr/commit/533ff667ad8b277e4524b46a51b1457e89196be8))
* **deps:** update dependency eslint to v8.1.0 ([936a687](https://www.github.com/monetr/monetr/commit/936a687f8e54a538a7061534e8a17c7eb3a90108))
* **deps:** update dependency eslint-plugin-flowtype to v7 ([31cd3ec](https://www.github.com/monetr/monetr/commit/31cd3ec3f9974629370c2c47af78e707a1dc447b))
* **deps:** update dependency eslint-plugin-flowtype to v8 ([8bd650f](https://www.github.com/monetr/monetr/commit/8bd650f112ce6c9fc5243f2fbb74fcc14769251d))
* **deps:** update dependency eslint-plugin-testing-library to v5 ([eef5f26](https://www.github.com/monetr/monetr/commit/eef5f26687a9a2f5a9e55ac7ac2b595f2d1f8456))
* **deps:** update dependency mini-css-extract-plugin to v2.4.3 ([ab4e977](https://www.github.com/monetr/monetr/commit/ab4e97701b6d8a9f0ac6a63a6aa9234aa52f456e))
* **deps:** update dependency react-redux to v7.2.6 ([81a1e64](https://www.github.com/monetr/monetr/commit/81a1e64a68d0b3ecbe37d729b53c8b5bfb3d3a1f))
* **deps:** update dependency redux-thunk to v2.4.0 ([8b81839](https://www.github.com/monetr/monetr/commit/8b81839f0410a4c78d66a16072b53364b36a25ba))
* **deps:** update dependency sass to v1.43.3 ([409d550](https://www.github.com/monetr/monetr/commit/409d550b6c3fe604082a7b89dba148286eecbf5a))
* **deps:** update google.golang.org/genproto commit hash to 2b14602 ([a48f491](https://www.github.com/monetr/monetr/commit/a48f491dff4df26c023f59cb842741097e86fb5c))
* **deps:** update google.golang.org/genproto commit hash to b7c3a96 ([9f123f4](https://www.github.com/monetr/monetr/commit/9f123f41c0b9d00ba145e4ef731ab1ecafcf38cc))
* **deps:** update module github.com/alicebob/miniredis/v2 to v2.16.0 ([1e38d10](https://www.github.com/monetr/monetr/commit/1e38d104e6fd4858782f85e878f1b4460a96d943))
* **deps:** update module github.com/nyaruka/phonenumbers to v1.0.73 ([5ee8433](https://www.github.com/monetr/monetr/commit/5ee8433b606936c173db6d209221044dd3249467))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.72.0 ([11297ed](https://www.github.com/monetr/monetr/commit/11297edaaf07392ef60b2da32f80bf6f2c42fa48))
* **deps:** update module github.com/vmihailenco/msgpack/v5 to v5.3.5 ([cc5440f](https://www.github.com/monetr/monetr/commit/cc5440f6ec75c38ce67075c71e506262ecf9e514))
* **deps:** update typescript-eslint monorepo to v5.2.0 ([253bde7](https://www.github.com/monetr/monetr/commit/253bde70d2d7a3ea4dff35df0b717652b99c1536))
* Keep cookies longer than the browser being closed. ([130083b](https://www.github.com/monetr/monetr/commit/130083bbc55c3d25e9122bd64f607b158689d9b2))
* You can no longer select the current date when creating an expense or goal. ([c5d0615](https://www.github.com/monetr/monetr/commit/c5d061553a3119605f224e7a26a485d7cf79ec52))

### [0.0.2](https://www.github.com/monetr/monetr/compare/v0.0.1...v0.0.2) (2021-10-20)


### Features

* Rename transactions from the UI. ([5c372c5](https://www.github.com/monetr/monetr/commit/5c372c5be98765438f0b0d31a065362c3bc90b22))
* Upgraded Plaid library for the UI to the latest version. ([f7edcd9](https://www.github.com/monetr/monetr/commit/f7edcd9b0091ef32015491e2106a390e19513e28))


### Bug Fixes

* Added `v` prefix to container RELEASE variable. ([8c4c579](https://www.github.com/monetr/monetr/commit/8c4c57983ed15d1443e5d35576cac5361e316642))
* **deps:** update dependency @stripe/stripe-js to v1.20.3 ([4058542](https://www.github.com/monetr/monetr/commit/4058542129691f411c1fb3cb0186307143c4d1a9))
* **deps:** update dependency @testing-library/user-event to v13.5.0 ([d8fc405](https://www.github.com/monetr/monetr/commit/d8fc4053df43eca921757ae1cb3d6de2afe24d83))
* **deps:** update dependency eslint-plugin-flowtype to v6.1.1 ([3876946](https://www.github.com/monetr/monetr/commit/3876946fc6c208e8b1cbe2d419ff28f5d68df3bf))
* **deps:** update dependency eslint-plugin-jest to v25.2.2 ([2064ab6](https://www.github.com/monetr/monetr/commit/2064ab663825915f9f6d5405b63a48f2eb34cf05))
* **deps:** update dependency react-plaid-link to v3.2.1 ([cb972bf](https://www.github.com/monetr/monetr/commit/cb972bf1511c3eac0e9f88795a3f8c9c2888b798))
* **deps:** update google.golang.org/genproto commit hash to 63b7e35 ([b734d36](https://www.github.com/monetr/monetr/commit/b734d36ad280f1c185001122a2e1b3b85edcc45c))
* **deps:** update google.golang.org/genproto commit hash to cf77aa7 ([70c90f0](https://www.github.com/monetr/monetr/commit/70c90f05e06b46f7741ce45a4e09827dc4fc0d2a))
* **deps:** update module github.com/plaid/plaid-go to v1.5.0 ([c1d97c4](https://www.github.com/monetr/monetr/commit/c1d97c45b0a40a0a5848487644e33039fd6e4cf1))
* **deps:** update typescript-eslint monorepo to v5.1.0 ([f34e68e](https://www.github.com/monetr/monetr/commit/f34e68efa27f8185c893a87697bd9c69e642439a))
* Don't overwrite transaction name on update. ([63ba2c9](https://www.github.com/monetr/monetr/commit/63ba2c9702ec5a7cd1ba78c6e39318a45cf17efd)), closes [#96](https://www.github.com/monetr/monetr/issues/96)
* Fixed issue where the UI file names changed each build. ([bf49a58](https://www.github.com/monetr/monetr/commit/bf49a581888e11ad0ef76a2f524c5791654b5a4a)), closes [#94](https://www.github.com/monetr/monetr/issues/94)
* Include email address verified time when link with Plaid. ([55fb142](https://www.github.com/monetr/monetr/commit/55fb1427513b87dd7e9a2821aaac3bd9b32a3514))
* Transaction name dropdown no longer renders infront of date. ([79c8402](https://www.github.com/monetr/monetr/commit/79c8402ee1bf95e0db19b2af46e8b06de26fbd31)), closes [#91](https://www.github.com/monetr/monetr/issues/91)
* Updating deepsource configuration for new monorepo. ([e67b3ce](https://www.github.com/monetr/monetr/commit/e67b3ce1dfe0fe48ed4b402ce3a2dc3aa4d029b5))


### Miscellaneous Chores

* release 0.0.2 ([0ff4273](https://www.github.com/monetr/monetr/commit/0ff427396047bb8f11bb33131d06c4a3b1949f25))

### 0.0.1 (2021-10-16)


### Features

* Push container image to both Github and DockerHub. ([7dba17f](https://www.github.com/monetr/monetr/commit/7dba17f416a5ea427342bac2d372acb0d219a524))


### Bug Fixes

* Changed architectures for main build. Tweaks. ([9a405fc](https://www.github.com/monetr/monetr/commit/9a405fc09805e57e987d1ee6c94365586cd92799))
* **deps:** update dependency @stripe/react-stripe-js to v1.6.0 ([88b4ee7](https://www.github.com/monetr/monetr/commit/88b4ee73c84b2d86301c51c831611caa99dd3cbe))
* **deps:** update dependency @stripe/stripe-js to v1.20.2 ([6a9b39b](https://www.github.com/monetr/monetr/commit/6a9b39bdd9c985fc3de2ffa66e126f9e7b433889))
* **deps:** update dependency @testing-library/user-event to v13.3.0 ([4f336b2](https://www.github.com/monetr/monetr/commit/4f336b20c125bc14d56d2aeea845f3c7f169a40c))
* **deps:** update dependency @testing-library/user-event to v13.4.1 ([e683a76](https://www.github.com/monetr/monetr/commit/e683a7649c92649ff936785a0ad6cef4fe414603))
* **deps:** update dependency css-what to v5.1.0 ([06ac4aa](https://www.github.com/monetr/monetr/commit/06ac4aa6112941ea4818407fbf2b2b7fdb715003))
* **deps:** update dependency eslint to v8 ([5ac89a0](https://www.github.com/monetr/monetr/commit/5ac89a0b5d08606cb7004e8a97c8cd96c3b91754))
* **deps:** update dependency eslint to v8.0.1 ([5e2d36c](https://www.github.com/monetr/monetr/commit/5e2d36c86e211a56bdff5b4dc50eba30dbc15de2))
* **deps:** update dependency eslint-plugin-import to v2.25.2 ([b5d5d5c](https://www.github.com/monetr/monetr/commit/b5d5d5c787c328b1432b6a82bd88834d3b0eef17))
* **deps:** update dependency eslint-plugin-jest to v24.7.0 ([59611bb](https://www.github.com/monetr/monetr/commit/59611bb8e855727df6de272007e550c267290788))
* **deps:** update dependency eslint-plugin-jest to v25 ([7c9c73f](https://www.github.com/monetr/monetr/commit/7c9c73fe65f45abd4c70a28da27da641f8eeedca))
* **deps:** update dependency eslint-plugin-jest to v25.0.6 ([9fcb8ba](https://www.github.com/monetr/monetr/commit/9fcb8ba7fd8bc5df4029f7d613a3b8b6eeceebad))
* **deps:** update dependency immutable to v4.0.0 ([c5c0d60](https://www.github.com/monetr/monetr/commit/c5c0d602742d12c6aeb3726664776b5f10a7730e))
* **deps:** update dependency jest-watch-typeahead to v1 ([d89054f](https://www.github.com/monetr/monetr/commit/d89054fadc29d5558d3be91ef65e9082e049d0f5))
* **deps:** update dependency mini-css-extract-plugin to v2.4.2 ([906eccb](https://www.github.com/monetr/monetr/commit/906eccb17eece53e421a4752fbd93d97af24054c))
* **deps:** update dependency prompts to v2.4.2 ([f803734](https://www.github.com/monetr/monetr/commit/f8037344d98236e80616878f2390620ef7e36091))
* **deps:** update dependency sass to v1.43.2 ([58456f5](https://www.github.com/monetr/monetr/commit/58456f5dae89ec56ef50a1dbfebc464277d84048))
* **deps:** update dependency typescript to v4.4.3 ([0bbc73e](https://www.github.com/monetr/monetr/commit/0bbc73e2be4e7832806f759c63df56309bb13248))
* **deps:** update dependency web-vitals to v2.1.2 ([0008829](https://www.github.com/monetr/monetr/commit/0008829bba73273798a0e38d884f04acae5c11fc))
* **deps:** update dependency yargs to v17.2.0 ([0e2ab91](https://www.github.com/monetr/monetr/commit/0e2ab91cae55439aa21b697920846e6e730d204f))
* **deps:** update dependency yargs to v17.2.1 ([a321297](https://www.github.com/monetr/monetr/commit/a321297da0b4a0735c31b5f05a73f3f854a1b916))
* **deps:** update google.golang.org/genproto commit hash to 181ce0d ([3b11ec4](https://www.github.com/monetr/monetr/commit/3b11ec46a00f23a2534c97d93875f8962f198d22))
* **deps:** update google.golang.org/genproto commit hash to 2e2e100 ([bdc6172](https://www.github.com/monetr/monetr/commit/bdc6172718b16e1b960049651c4db99bdac5a874))
* **deps:** update google.golang.org/genproto commit hash to 3192f97 ([487e88d](https://www.github.com/monetr/monetr/commit/487e88d257141a6d52c42d8f84831798f3b623a0))
* **deps:** update google.golang.org/genproto commit hash to 3238e09 ([6100df5](https://www.github.com/monetr/monetr/commit/6100df51bbd41821d5adb0641f378b85a2e179bf))
* **deps:** update google.golang.org/genproto commit hash to 37fc393 ([ee58530](https://www.github.com/monetr/monetr/commit/ee58530b2f3c18fdb22c894fb42164d7bceb7823))
* **deps:** update google.golang.org/genproto commit hash to 3dee208 ([ca6014b](https://www.github.com/monetr/monetr/commit/ca6014b6272fd9e4a49918b50bc51cf41b82cb98))
* **deps:** update google.golang.org/genproto commit hash to 433400c ([622401a](https://www.github.com/monetr/monetr/commit/622401a14be7f1966873d26a98a8efb0623ffee3))
* **deps:** update google.golang.org/genproto commit hash to 86cf123 ([fb71b16](https://www.github.com/monetr/monetr/commit/fb71b16a243dc1afef127810252f2d9e59c6624b))
* **deps:** update google.golang.org/genproto commit hash to 896c89f ([54f495a](https://www.github.com/monetr/monetr/commit/54f495adbf9e5c8d9c2acfad5cf511fe709d7091))
* **deps:** update google.golang.org/genproto commit hash to a8c4777 ([71ec38e](https://www.github.com/monetr/monetr/commit/71ec38e6de70d9e3a49b063091a63d1070c89d2a))
* **deps:** update google.golang.org/genproto commit hash to b395a37 ([e4d9548](https://www.github.com/monetr/monetr/commit/e4d9548b8711280ab76cc1f3d5343bee7f1e3cb9))
* **deps:** update google.golang.org/genproto commit hash to bfb93cc ([19cac36](https://www.github.com/monetr/monetr/commit/19cac36380453d6eb03521af21cc8290515a44e1))
* **deps:** update google.golang.org/genproto commit hash to c7af6a1 ([1732f2f](https://www.github.com/monetr/monetr/commit/1732f2f702eba284e412837108528487bb85fafe))
* **deps:** update google.golang.org/genproto commit hash to ce87815 ([7dad67a](https://www.github.com/monetr/monetr/commit/7dad67ae7be321c60370ecf1898febf63b45d71c))
* **deps:** update google.golang.org/genproto commit hash to d08c68a ([71ecf41](https://www.github.com/monetr/monetr/commit/71ecf4144c653a1e69bdc88b21fe08f28de23449))
* **deps:** update google.golang.org/genproto commit hash to fe13028 ([9829362](https://www.github.com/monetr/monetr/commit/9829362f4a7a31de996e1ef705256bd1479bbb43))
* **deps:** update module github.com/brianvoe/gofakeit/v6 to v6.8.0 ([993358c](https://www.github.com/monetr/monetr/commit/993358c3e748d0c5cf21262c1064c40176713238))
* **deps:** update module github.com/brianvoe/gofakeit/v6 to v6.9.0 ([7002d26](https://www.github.com/monetr/monetr/commit/7002d2685d9d19513a14ebcad779ce4ec145aee0))
* **deps:** update module github.com/go-pg/pg/v10 to v10.10.5 ([e4f8962](https://www.github.com/monetr/monetr/commit/e4f89627a8366b81c73d22f7a78b7657781780d7))
* **deps:** update module github.com/go-pg/pg/v10 to v10.10.6 ([eac3876](https://www.github.com/monetr/monetr/commit/eac3876849501342d69ef39f5da905d765b2c1ca))
* **deps:** update module github.com/google/go-github/v38 to v39 ([347b555](https://www.github.com/monetr/monetr/commit/347b5554e61e6c9a636c11cf3ef1124d39addab5))
* **deps:** update module github.com/hashicorp/vault/api to v1.2.0 ([cdf969b](https://www.github.com/monetr/monetr/commit/cdf969b51fafaa3a4e8d2f93bc59992eee6b39bf))
* **deps:** update module github.com/kataras/iris/v12 to v12.2.0-alpha3 ([69a02bc](https://www.github.com/monetr/monetr/commit/69a02bc987138601a4969f0b0199981cb1f606f3))
* **deps:** update module github.com/kataras/iris/v12 to v12.2.0-alpha4 ([9cdbe11](https://www.github.com/monetr/monetr/commit/9cdbe11543ac5db6830a17e3f7f2d5431d8c6415))
* **deps:** update module github.com/micahparks/keyfunc to v0.9.0 ([410bcd8](https://www.github.com/monetr/monetr/commit/410bcd813b7ddb34169d4029df2bf6daab599c7a))
* **deps:** update module github.com/nyaruka/phonenumbers to v1.0.72 ([f509875](https://www.github.com/monetr/monetr/commit/f509875695dca38848e3431e7973862a63b7832c))
* **deps:** update module github.com/oneofone/xxhash to v1.2.8 ([7e3060d](https://www.github.com/monetr/monetr/commit/7e3060d8f8e53151cf78436e5f0a18e34080dbca))
* **deps:** update module github.com/plaid/plaid-go to v1.2.0 ([c7d38e2](https://www.github.com/monetr/monetr/commit/c7d38e289de88d857103d77ae2359f26d19d8880))
* **deps:** update module github.com/plaid/plaid-go to v1.3.0 ([9120829](https://www.github.com/monetr/monetr/commit/9120829819a1e619281ea73a6593ad02ec63687f))
* **deps:** update module github.com/plaid/plaid-go to v1.4.0 ([711db28](https://www.github.com/monetr/monetr/commit/711db28a36472be4bfa6f788df47a0520f96fc6d))
* **deps:** update module github.com/spf13/viper to v1.9.0 ([5aad7c5](https://www.github.com/monetr/monetr/commit/5aad7c52d12b2d35bf13f8e96f2a12f28d2eec75))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.63.0 ([bbb2b08](https://www.github.com/monetr/monetr/commit/bbb2b0884dbe177ad56c1da917ed774309e99552))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.64.0 ([242ac15](https://www.github.com/monetr/monetr/commit/242ac157ad0a89102ff06a0f2a27410e8456f3e0))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.64.1 ([a0fcb4e](https://www.github.com/monetr/monetr/commit/a0fcb4ea599d362115a1c6dd5aeaea0995ad665e))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.65.0 ([791c9de](https://www.github.com/monetr/monetr/commit/791c9de9400f49420c6c864fb46e864054c0bee0))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.67.0 ([aa6618a](https://www.github.com/monetr/monetr/commit/aa6618a28775a8bba208edce2fd1542aa747995c))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.70.0 ([286743a](https://www.github.com/monetr/monetr/commit/286743a1dc981451f67cd19073052b04da5663bb))
* **deps:** update module github.com/stripe/stripe-go/v72 to v72.71.0 ([f3275fe](https://www.github.com/monetr/monetr/commit/f3275fed84b44be3e3ce53e1174f9395bfe9413b))
* **deps:** update sentry-javascript monorepo to v6.13.3 ([80ccab8](https://www.github.com/monetr/monetr/commit/80ccab85d0d844d7d6829b17a526fdedd47aaa85))
* **deps:** update typescript-eslint monorepo to v4.33.0 ([9a048bb](https://www.github.com/monetr/monetr/commit/9a048bbed1adf784804b32c27a00757368155db1))
* **deps:** update typescript-eslint monorepo to v5 ([ce8def1](https://www.github.com/monetr/monetr/commit/ce8def11b55b1e5a9085e7e0d8d54b169d2706b4))
* Fixed bug where funding schedules were not being processed. ([5e0f509](https://www.github.com/monetr/monetr/commit/5e0f5094021d0289421273518d3b24b9cd6d807a))
* Fixed github action context variable for release. ([95971c9](https://www.github.com/monetr/monetr/commit/95971c9feb742600b06f5179b2371cf98d023bf4))
* Removed additional architectures from binary. Adjusting release. ([e01c613](https://www.github.com/monetr/monetr/commit/e01c613827ebb2ae002673e3e5ce9cd5d2e018df))
* Stripe race condition. ([be401a2](https://www.github.com/monetr/monetr/commit/be401a2a617c74e8fcacd8dfc3ab53d4a8c43bf6))


### Miscellaneous Chores

* release 0.0.1 ([f31604c](https://www.github.com/monetr/monetr/commit/f31604cbccfc00b531b19d8a03fac453888b2d4f))
