// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "trustready-agent-url-handler",
    platforms: [
        .macOS(.v11)
    ],
    products: [
        .library(name: "TrustReadyAgentShared", targets: ["TrustReadyAgentShared"]),
        .library(name: "HelperClient", targets: ["HelperClient"]),
        .executable(name: "com.trustready.agent.helper", targets: ["com.trustready.agent.helper"]),
        .executable(name: "trustready-agent-url-handler", targets: ["trustready-agent-url-handler"]),
    ],
    targets: [
        .target(
            name: "TrustReadyAgentShared",
            path: "Shared",
            exclude: [
                "HelperVersion.generated.swift.tmpl",
                "SigningConstants.generated.swift.tmpl",
            ],
            linkerSettings: [
                .linkedFramework("Security")
            ]
        ),
        .target(
            name: "HelperClient",
            dependencies: ["TrustReadyAgentShared"],
            path: "HelperClient"
        ),
        .executableTarget(
            name: "com.trustready.agent.helper",
            dependencies: ["TrustReadyAgentShared"],
            path: "HelperTool",
            exclude: ["Info.plist.tmpl", "Launchd.plist.tmpl"],
            linkerSettings: [
                .linkedFramework("Security")
            ]
        ),
        .executableTarget(
            name: "trustready-agent-url-handler",
            dependencies: ["HelperClient", "TrustReadyAgentShared"],
            path: "URLHandlerSources"
        ),
    ]
)
