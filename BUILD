load("@io_bazel_rules_go//go:def.bzl", "go_prefix")

go_prefix("k8s.io/test-infra")

filegroup(
    name = "package-srcs",
    srcs = glob(
        ["**"],
        exclude = [
            "bazel-*/**",
            ".git/**",
            "*.db",
            "*.gz",
        ],
    ),
    visibility = ["//visibility:private"],
)

filegroup(
    name = "buckets",
    srcs = ["buckets.yaml"],
    visibility = ["//:__subpackages__"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//boskos:all-srcs",
        "//experiment:all-srcs",
        "//gcsweb/cmd/gcsweb:all-srcs",
        "//gcsweb/pkg/version:all-srcs",
        "//images/pull_kubernetes_bazel:all-srcs",
        "//jenkins:all-srcs",
        "//jobs:all-srcs",
        "//kettle:all-srcs",
        "//kubetest:all-srcs",
        "//logexporter/cmd:all-srcs",
        "//maintenance/githubutil:all-srcs",
        "//maintenance/migratestatus:all-srcs",
        "//metrics:all-srcs",
        "//mungegithub:all-srcs",
        "//prow:all-srcs",
        "//scenarios:all-srcs",
        "//testgrid/config:all-srcs",
        "//testgrid/jenkins_verify:all-srcs",
        "//triage:all-srcs",
        "//velodrome:all-srcs",
        "//vendor:all-srcs",
        "//verify:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)

action_listener(
    name = "gofmt-all",
    mnemonics = ["GoCompile"],
    extra_actions = [":gofmt_test"],
    visibility = ["//visibility:public"],
)

extra_action(
    name = "gofmt_test",
    cmd = "$(location @protobuf//:protoc) --decode_raw < $(EXTRA_ACTION_FILE)",
    tools = [
        "@protobuf//:protoc",
        "@io_bazel_rules_go_toolchain//:toolchain",
    ],
)
