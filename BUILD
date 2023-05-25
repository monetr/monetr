load("@bazel_gazelle//:def.bzl", "gazelle")
load("@io_bazel_rules_go//go:def.bzl", "go_binary")
load("@npm//:defs.bzl", "npm_link_all_packages")
load("@npm//:@rspack/cli/package_json.bzl", rspack_cli = "bin")
load("@aspect_rules_jest//jest:defs.bzl", "jest_test")
# load("@coinbase_rules_ruby//ruby:defs.bzl", "rb_binary")

# gazelle:proto disable_global
# gazelle:prefix github.com/monetr/monetr
gazelle(name = "gazelle")

gazelle(
    name = "gazelle-update-repos",
    args = [
        "-from_file=go.mod",
        "-to_macro=deps.bzl%go_dependencies",
        "-prune",
        "-build_file_proto_mode=disable_global",
    ],
    command = "update-repos",
)

go_binary(
    name = "monetr",
    embed = ["//pkg/cmd:cmd_lib"],
    visibility = ["//visibility:public"],
    gotags = ["icons", "simple_icons"],
    gc_linkopts = [
      "-s",
      "-w",
    ],
    x_defs = {
      "github.com/monetr/monetr/pkg/cmd.buildHost": "{BUILD_HOST}",
      "github.com/monetr/monetr/pkg/cmd.buildRevision": "{STABLE_GIT_REVISION}",
      "github.com/monetr/monetr/pkg/cmd.buildTime": "{BUILD_TIME}",
      "github.com/monetr/monetr/pkg/cmd.buildType": "bazel",
      "github.com/monetr/monetr/pkg/cmd.release": "{STABLE_GIT_RELEASE}",
    },
)

# rb_binary(
#     name = "licensed",
#     srcs = ["go.mod", "go.sum"],
#     args = ["cache"],
#     main = "@gems//:bin/licensed",
#     deps = ["@gems//:licensed"],
# )

# genrule(
#   name = "licensed",
#   srcs = [
#     ".licensed.yaml",
#     "go.mod",
#     "go.sum",
#     "package.json",
#     "pnpm-lock.yaml",
#     "//:node_modules",
#     "@gems//:bin/licensed"
#   ],
#   outs = [".licenses"],
#   cmd = """$(location @gems//:bin/licensed) cache --force""",
# )

## UI STUFF ##

filegroup(
  name = "ui_src",
  srcs = glob([
    'ui/**/*',
  ])
)

filegroup(
    name = "ui_reqs",
    srcs = glob([
      'public/*',
      '.swcrc',
      'package.json',
      'postcss.config.cjs',
      'rspack.config.js',
      'tailwind.config.cjs',
      'tsconfig.json',
    ])
)

npm_link_all_packages()

rspack_cli.rspack(
  name = "ui",
  srcs = [
    "//:ui_src",
    "//:ui_reqs",
    "//:node_modules",
  ],
  args = [
    "build"
  ],
  env = {
    "BUILD_PATH": 'static',
  },
  out_dirs = ["static"],
  visibility = ["//visibility:public"],
  patch_node_fs = True,
)

# Jest wont auto import all of node modules by default. But if you make it a filegroup it will.
# I don't know enough about how bazel works yet. This is probably terrible.
filegroup(
  name = "dev_dependencies",
  srcs = ["//:node_modules"],
)

jest_test(
    # A unique name for this target.
    name = "jest",
    # Label pointing to the linked node_modules target where jest is linked, e.g
    node_modules = "//:node_modules",
    config = "jest.config.js",
    data = [
      "//:ui_src",
      "//:dev_dependencies",
    ]
)

