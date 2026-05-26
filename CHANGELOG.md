# Changelog

## [0.4.0](https://github.com/juancavallotti/copilot-k8s-adk-go/compare/recipes-v0.3.0...recipes-v0.4.0) (2026-05-26)


### Features

* **backend:** added semantic search to backend ([50e7a00](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/50e7a00cfc126557dbf8997e80ff54b18504414b))
* **cli:** added trace indexing ([28ff7df](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/28ff7df2f25be8d57ea25258fffd783282e4691c))
* **cli:** cli now index recipes for semantic search asynchronously ([8dae7db](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/8dae7db5fb01861cd35096b3332de9a2f97f650b))
* **embeddings:** Added embedding capabilities to the backend both for openai and gemini ([a0b7b7d](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/a0b7b7d8d5c6e824d3b12652f4325924f8705a2b))
* **helm:** added ssl support to helm chart ([4582793](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/4582793ab2c89049d98e71931cc54e65f566e80a))
* Semantic search ([65bcd3f](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/65bcd3f711adedf0446abe918e5050d26f47bec2))
* **traces:** Added user prompt to trace events so they can be located more easily ([2922bc5](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/2922bc56dcd548ab6f03c8c63ce4b1e73c0fb7cb))
* **web:** added recipe search to layout ([4c0f54a](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/4c0f54a20ed5f6f9f2ed1b8ad548101d77146ffa))


### Bug Fixes

* **backend:** added shudown hooks to the backend api ([cf8edfd](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/cf8edfdce031671145fdb449013376ea22816822))
* **backend:** added slim semantic search on cli so searches don't return full recipes ([8110c17](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/8110c17c72ae099012412c3ac84e17cc9de84b44))
* **backend:** correctly implemented gemini embedding api call ([a019598](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/a0195981ce43aa3b7b4cb757738a00035f450776))
* **cli:** added wait group to cli so it doesn't exi t before goroutines finish ([ba15e0a](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/ba15e0a3c49fd345688bbf9b21b9d3b7d2526eb6))
* **dev:** fixed devspace to inject secrets into the right namespace ([88f7c76](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/88f7c76ee57490c5dbdbcfa7e2d6844f312969c0))


### Chores

* added k8s topology ([7dcece9](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/7dcece9b2da2f1da3935c5fb10835f591bf088af))
* added monorepo structure on the readme file ([64ac118](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/64ac1183b1de88dcd40791ace5aec5b744eb0f8d))
* aded task to render values.yaml ([7f917d3](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/7f917d38ae5a40a25c36375ebf864fa40403f84b))
* **backend:** cleared interface pollution at repo level ([f8c20a0](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/f8c20a09e8054247f0e756d9b5f7f4d289fa799c))
* **build:** added stop task and made 'recipes' the default namespace ([52f994b](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/52f994b0ca2d054c502bc662e02b0718cdead0d7))
* **build:** added timestamp to image tags ([a195943](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/a19594308f584adba50d0ceba85a7dd98b03a2d6))
* **cli:** removed interface pollution by clearing up Runner composite interface ([5855b72](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/5855b72cc9bb15735faff92bd9afccee41621ec9))
* converted unused init function into sync once ([053c113](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/053c11373bf8fd0f5efe6dc745471a837e14c201))
* Docs ([03c0680](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/03c06803b51946ed9297fdd6d40164e8b1f339b7))
* Document Copilot Agent architecture in README ([a475521](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/a475521415064255e763d7e78781321ff0209730))
* Go enhancements ([b486a7c](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/b486a7c981768509323f12e4a722c90fd0deb38c))
* **helm:** added configuration on helm chart to support the new semantic search ([f85bde1](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/f85bde11baebfffbf00d47bb8fd3062f0ef5839e))
* implmemented ingress values rendering ([79a70bb](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/79a70bb58887bf62eb544f6c4edce8515415518b))
* increased ingress size ([3477b1f](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/3477b1f6ddc3c3d774d927bead4deb8f9f726546))
* updated architecture doc ([39a7fb4](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/39a7fb4d3cca7b698af00bc7215f76cd7efd8c6b))
* updated readme ([5810c3b](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/5810c3b97d5c92b0b11432013285cd55a52fc29e))
* **web:** enhanced display on recipe list ([20948bf](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/20948bf20b95b0052b7032e54615d7c9c003117e))
* **web:** made recipe lis reusable on the ui ([ccb562b](https://github.com/juancavallotti/copilot-k8s-adk-go/commit/ccb562b5bc8afc3a1c818e332c2b58840abd4440))

## [0.3.0](https://github.com/juancavallotti/simple-task-monorepo/compare/recipes-v0.2.0...recipes-v0.3.0) (2026-05-23)


### Features

* added agent trace viewer ([600ca84](https://github.com/juancavallotti/simple-task-monorepo/commit/600ca844d3488ca260e6d9e7f61906da6e055682))
* Added backup / restore and ability to delete traces. ([7531bc5](https://github.com/juancavallotti/simple-task-monorepo/commit/7531bc557c1b879ffe6bb7950cecaead6ab0459a))
* added backup / restore functionality and ability to delete traces ([99fc443](https://github.com/juancavallotti/simple-task-monorepo/commit/99fc443736c7fcf1ddd1392bcc7c447903948626))
* added providers api and tests ([52922f2](https://github.com/juancavallotti/simple-task-monorepo/commit/52922f2a883f3c08944d9e51775e038f09119d87))
* added recipe filtering and ordering ([958dbd0](https://github.com/juancavallotti/simple-task-monorepo/commit/958dbd0224fa4cd82ecd95ccc0f2d43ab7838b7b))
* added skills to agent ([6a732b8](https://github.com/juancavallotti/simple-task-monorepo/commit/6a732b87b59904fcec150a242e8bd3b1537ebe22))
* added skills to backend ([7737eda](https://github.com/juancavallotti/simple-task-monorepo/commit/7737edad575b5692cc952c9c983c1be16234238f))
* added skills wiring into web app ([d2ae3b2](https://github.com/juancavallotti/simple-task-monorepo/commit/d2ae3b22faa8743b5703a13eba87d43e32404403))
* added support for openai and anthropic models ([2fae220](https://github.com/juancavallotti/simple-task-monorepo/commit/2fae22001d410ccf16db4d6d49e2032a2e9b8a2c))
* added support to switch models on the ui ([78fde95](https://github.com/juancavallotti/simple-task-monorepo/commit/78fde951dec443b2dc6a0fcbd3b20faea1b2dcb5))
* added tables to hold agent traces on db ([434a8e8](https://github.com/juancavallotti/simple-task-monorepo/commit/434a8e89e6da56449c573810f877e96a032752aa))
* added traces display to the ui ([b206d19](https://github.com/juancavallotti/simple-task-monorepo/commit/b206d1963e2221ab8d94ac26208d7742611b1c17))
* added ui selector for various models ([a10e60d](https://github.com/juancavallotti/simple-task-monorepo/commit/a10e60d588f7d45a3bf312915765d038c4aa873b))
* Added visibility into tool calls directly ([b734af5](https://github.com/juancavallotti/simple-task-monorepo/commit/b734af56b07aa4ea2e1f06abf0ed7e1521e64d44))
* agent bubbles now show raw event data ([bdeb1a5](https://github.com/juancavallotti/simple-task-monorepo/commit/bdeb1a5cefe2934324f2cb64ba08169753d88048))
* Agent improvements ([265bbd4](https://github.com/juancavallotti/simple-task-monorepo/commit/265bbd4df858b3a557ab196651a8817664a8fc61))
* backend now can store traces into the db, added apis to query traces ([a33eb51](https://github.com/juancavallotti/simple-task-monorepo/commit/a33eb514a4779a40c39672776abf1ea15611856f))
* Better agent traces ([4986e80](https://github.com/juancavallotti/simple-task-monorepo/commit/4986e802e0ec901a323e2500dfa4bd8ad0e13ab6))
* chat now renders streamed messages correctly ([7eb6eca](https://github.com/juancavallotti/simple-task-monorepo/commit/7eb6eca86c7ce02730fab570195c8fd474d7893d))
* chat now renders streamed messages correctly ([e626db7](https://github.com/juancavallotti/simple-task-monorepo/commit/e626db77a8d3bb7da1c03953e790c82517c4c32e))
* plug slog with cli so traces go to the db ([ab88c54](https://github.com/juancavallotti/simple-task-monorepo/commit/ab88c543bd745a991e6a706bb48602be072e853b))
* reorganized the agent with its system prompt and skills ([5e7fc6d](https://github.com/juancavallotti/simple-task-monorepo/commit/5e7fc6ddc2dd5c63481eaec16396de92ec409255))
* selectable models ([42603c1](https://github.com/juancavallotti/simple-task-monorepo/commit/42603c1efc28b73970c4893f6e33cfaaf8de199d))
* Tracing and Monitoring ([7aea173](https://github.com/juancavallotti/simple-task-monorepo/commit/7aea1738be2aec49675fb11fbbaa64a49673bdb8))
* UI enhancements ([0b5e9db](https://github.com/juancavallotti/simple-task-monorepo/commit/0b5e9db0cddc66cf28ac7944b27b1fb2f0ae6adb))
* wired skills route ([fef4be1](https://github.com/juancavallotti/simple-task-monorepo/commit/fef4be187f762930a08ea4cfbe015e3bbe65f8a9))


### Bug Fixes

* added ability to display the user prompt on the traces screen ([c23315d](https://github.com/juancavallotti/simple-task-monorepo/commit/c23315da2196620b67db646e6429ecf740da8eac))
* added context window management to agent ([4ab78ec](https://github.com/juancavallotti/simple-task-monorepo/commit/4ab78ecaff563db62d46c982851a8415ce1ea64e))
* added llm api keys to sample .env and fixed ttl.sh script ([f9a9397](https://github.com/juancavallotti/simple-task-monorepo/commit/f9a93973d8591ab06753bd0df88e6142c0df13e6))
* added llm api keys to sample .env and fixed ttl.sh script ([cc4110d](https://github.com/juancavallotti/simple-task-monorepo/commit/cc4110ddd5c36f95c1a874dfc284fa3099a10596))
* added tool to ensure agent spits ui actions correctly ([2840a22](https://github.com/juancavallotti/simple-task-monorepo/commit/2840a226bdddc75296f2b62c4eaabb4e213d62da))
* added ui actions so the agent can understand and navigate the recipe context ([f0fed37](https://github.com/juancavallotti/simple-task-monorepo/commit/f0fed37297f72f4bd6997b85aedcaae6dee3fd0e))
* chat now groups messages correctly and shows tool executions while they're happening ([a32b677](https://github.com/juancavallotti/simple-task-monorepo/commit/a32b6778f50d07dbf955a326818b93e6c9ef3033))
* cli does not return image data when querying for recipes anymore unless requested ([eda22c5](https://github.com/juancavallotti/simple-task-monorepo/commit/eda22c5c6d551850e2a5d361c76977718a3ad10a))
* edit recipe now patches instead of submitting full json with images included ([4e4fa3f](https://github.com/juancavallotti/simple-task-monorepo/commit/4e4fa3fba1ce70f534d09af82ef822598ea6a124))
* enhanced cli so help doesn't show as a failed message ([2113d6f](https://github.com/juancavallotti/simple-task-monorepo/commit/2113d6f34548a98a54618ff8bfdddcc038214595))
* enhanced system prompt to optimize recipe results ([0993549](https://github.com/juancavallotti/simple-task-monorepo/commit/0993549fee4dce1670e1bb67ee164d7fd91c942b))
* enhanced tool grouping on the traces UI ([0af6782](https://github.com/juancavallotti/simple-task-monorepo/commit/0af67827f54e98d186c82a53b7aac567602eb6ba))
* enhanced tracing with more metadata on agent ([0f18ef9](https://github.com/juancavallotti/simple-task-monorepo/commit/0f18ef9d5067261e68dca9867ffade4003b658b7))
* fixed cli consistency ([70b477c](https://github.com/juancavallotti/simple-task-monorepo/commit/70b477cbe13731b746d48db33c51676deb54b977))
* fixed cli consistency ([287d7bb](https://github.com/juancavallotti/simple-task-monorepo/commit/287d7bb2dc6090f4ad1b625c6bdff4eefcd3a1bf))
* fixed recipe steps to show markdown ([23e6bad](https://github.com/juancavallotti/simple-task-monorepo/commit/23e6badf71d9c53140bdf90dc34501358fe895fd))
* openai image now receives the correct parameters ([a89eb28](https://github.com/juancavallotti/simple-task-monorepo/commit/a89eb28675fec654a5c063043c1425697cb8b5f7))
* skills markdown now renders headers ([5aac945](https://github.com/juancavallotti/simple-task-monorepo/commit/5aac9451f9700db1808ff2639792df5ca5f7f1a5))
* tool calls are now grouped by function id ([643100b](https://github.com/juancavallotti/simple-task-monorepo/commit/643100b163dbe2ae218111aeb7d9951a500bc342))
* updated helm chart ([c5ab2eb](https://github.com/juancavallotti/simple-task-monorepo/commit/c5ab2ebc91903fcc684a4dc3a9b6641d79477c23))


### Chores

* add release please run automation when merging into main ([8b81ced](https://github.com/juancavallotti/simple-task-monorepo/commit/8b81cedd2ce9c07fbf091ccb8195bdf2404f3029))
* added icons by the model selectors ([76041b6](https://github.com/juancavallotti/simple-task-monorepo/commit/76041b630260b3f4370b8c585b5129907a7dcd18))
* added logger cli to helm ([278dd2d](https://github.com/juancavallotti/simple-task-monorepo/commit/278dd2d85f127d8144f58b506fb41881f2aac552))
* added skills endpoint on API ([d89062c](https://github.com/juancavallotti/simple-task-monorepo/commit/d89062c33fbcbe84d9fd593aed0aa254b4c1ee57))
* added tests to the repo module ([a26e722](https://github.com/juancavallotti/simple-task-monorepo/commit/a26e7224205a0fd37521527f3440f951c6ada1f8))
* added ttl.sh publish script that renders helm values.yaml ([f3faeb5](https://github.com/juancavallotti/simple-task-monorepo/commit/f3faeb5fbdd8dee5741a57457ba3154161a6df20))
* changed agent logging to slog and logged agent events ([ad21728](https://github.com/juancavallotti/simple-task-monorepo/commit/ad21728afe04ab6a7867a9935615b58aaa563722))
* enabled slog on the backend ([e55029d](https://github.com/juancavallotti/simple-task-monorepo/commit/e55029d799df8bcdc622545904dff33721cfbbf4))
* enhanced prompt so it can discover better the CLI functions ([93b22f7](https://github.com/juancavallotti/simple-task-monorepo/commit/93b22f79b24cf5f9528d6abf9eab30059aea0ca9))
* enhanced release please config ([7a66636](https://github.com/juancavallotti/simple-task-monorepo/commit/7a66636a870b69a9dbca7875136bab6c4f4e373e))
* fixed release please config ([6dfc3b4](https://github.com/juancavallotti/simple-task-monorepo/commit/6dfc3b49c9e127b7abccea0177d48af762f0bad2))
* Merge pull request [#4](https://github.com/juancavallotti/simple-task-monorepo/issues/4) from juancavallotti/build-pipeline ([90604c6](https://github.com/juancavallotti/simple-task-monorepo/commit/90604c676e5a4fe88dd1dc2a56cf735c6ca17a2c))
* refactored recipe list to use a reducer ([bc52b47](https://github.com/juancavallotti/simple-task-monorepo/commit/bc52b47080fb6685f3b9eb356bcc46404f29449b))
* reorganized api hanlders for cohesion ([2b0b5fc](https://github.com/juancavallotti/simple-task-monorepo/commit/2b0b5fcb2e4c5b02abc281f16930f13ecece1957))
* reorganized repo module to be more idiomatic ([4b0807d](https://github.com/juancavallotti/simple-task-monorepo/commit/4b0807dd9b585e47062ded51c8d8fd081963c749))
* tidy up react components and introduce reusable elements to remove duplication ([3abfb51](https://github.com/juancavallotti/simple-task-monorepo/commit/3abfb510f1cf2228ecb94aed0b83815304ce8b40))
* upgraded render script to use local dev env values ([150e55d](https://github.com/juancavallotti/simple-task-monorepo/commit/150e55d7a74cd292939b78b7035ebe6d48f59b1f))
* wired skill detail page ([f2687cb](https://github.com/juancavallotti/simple-task-monorepo/commit/f2687cbc75dae50954e5414fb6a770e0e870f4d7))
* wired skills into cli ([24b7564](https://github.com/juancavallotti/simple-task-monorepo/commit/24b7564acc4e1b634f05ae7dae971d2d8c3e5522))

## [0.2.0](https://github.com/juancavallotti/simple-task-monorepo/compare/recipes-v0.1.5...recipes-v0.2.0) (2026-05-17)


### Features

* recipe descriptions now renders markdown ([f1c8e4d](https://github.com/juancavallotti/simple-task-monorepo/commit/f1c8e4d1b389f22b4a9c2306c4ddf425e02dc605))


### Bug Fixes

* reinforced instructions so agent generates the refresh UI instruction when changing a recipe ([1ca9ea7](https://github.com/juancavallotti/simple-task-monorepo/commit/1ca9ea74e0d07a0f147934697f0658126931e2d4))

## [0.1.5](https://github.com/juancavallotti/simple-task-monorepo/compare/recipes-v0.1.4...recipes-v0.1.5) (2026-05-17)


### Bug Fixes

* changed release please to trigger manually ([4a8f530](https://github.com/juancavallotti/simple-task-monorepo/commit/4a8f53017aa8445ab9e1c210c16171aa3a9ddb62))
