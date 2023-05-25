workspace(name = "com_github_monetr_monetr")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "6b65cb7917b4d1709f9410ffe00ecf3e160edf674b78c54a894471320862184f",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.39.0/rules_go-v0.39.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.39.0/rules_go-v0.39.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "727f3e4edd96ea20c29e8c2ca9e8d2af724d8c7778e7923a854b2c80952bc405",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.30.0/bazel-gazelle-v0.30.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.30.0/bazel-gazelle-v0.30.0.tar.gz",
    ],
)

http_archive(
    name = "aspect_bazel_lib",
    sha256 = "e3151d87910f69cf1fc88755392d7c878034a69d6499b287bcfc00b1cf9bb415",
    strip_prefix = "bazel-lib-1.32.1",
    url = "https://github.com/aspect-build/bazel-lib/releases/download/v1.32.1/bazel-lib-v1.32.1.tar.gz",
)

load("@aspect_bazel_lib//lib:repositories.bzl", "aspect_bazel_lib_dependencies")

aspect_bazel_lib_dependencies()

git_repository(
    name = "simple-icons",
    remote = "https://github.com/simple-icons/simple-icons.git",
    # tag = "7.19.0",
    commit = "4dc71b5905b1d3541d688dce6c88bc337613fc07", # 7.19.0
    build_file_content = """
# Used to make the files we need available to the pkg/icons package.
filegroup(
    name = 'icons',
    srcs = glob([
        'LICENSE.md',
        'package.json',
        'slugs.md',
        '_data/simple-icons.json',
        'icons/*.svg',
    ]),
    visibility = ['//visibility:public'],
)
    """,
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

load("//:deps.bzl", "go_dependencies")

# gazelle:repository_macro deps.bzl%go_dependencies
go_dependencies()

go_rules_dependencies()

go_register_toolchains(version = "1.19.4")

gazelle_dependencies()

# Node JS stuff

http_archive(
    name = "aspect_rules_js",
    sha256 = "e3e6c3d42491e2938f4239a3d04259a58adc83e21e352346ad4ef62f87e76125",
    strip_prefix = "rules_js-1.30.0",
    url = "https://github.com/aspect-build/rules_js/releases/download/v1.30.0/rules_js-v1.30.0.tar.gz",
)

load("@aspect_rules_js//js:repositories.bzl", "rules_js_dependencies")

rules_js_dependencies()

load("@rules_nodejs//nodejs:repositories.bzl", "DEFAULT_NODE_VERSION", "nodejs_register_toolchains")

nodejs_register_toolchains(
    name = "nodejs",
    node_version = DEFAULT_NODE_VERSION,
)

load("@aspect_rules_js//npm:repositories.bzl", "npm_translate_lock")

npm_translate_lock(
    name = "npm",
    pnpm_lock = "//:pnpm-lock.yaml",
    npmrc = "@//:.npmrc",
    verify_node_modules_ignored = "//:.bazelignore",
    # prod = True,
)

load("@npm//:repositories.bzl", "npm_repositories")

npm_repositories()


## JS Tests

http_archive(
    name = "aspect_rules_jest",
    sha256 = "098186ffc450f2a604843d8ba14217088a0e259ea6a03294af5360a7f1bcd3e8",
    strip_prefix = "rules_jest-0.19.5",
    url = "https://github.com/aspect-build/rules_jest/releases/download/v0.19.5/rules_jest-v0.19.5.tar.gz",
)

load("@aspect_rules_jest//jest:dependencies.bzl", "rules_jest_dependencies")

rules_jest_dependencies()

## Ruby stuff for licensed.

# git_repository(
#     name = "coinbase_rules_ruby",
#     remote = "https://github.com/coinbase/rules_ruby.git",
#     branch = "master",
# )
#
# load(
#     "@coinbase_rules_ruby//ruby:deps.bzl",
#     "ruby_register_toolchains",
#     "rules_ruby_dependencies",
# )
#
# rules_ruby_dependencies()
#
# ruby_register_toolchains(version = "3.1.2")
#
# load("@coinbase_rules_ruby//ruby:defs.bzl", "rb_bundle")
#
# rb_bundle(
#     name = "gems",
#     bundler_version = '2.3.26',
#     gemfile = "//:scripts/Gemfile",
#     gemfile_lock = "//:scripts/Gemfile.lock",
# )
