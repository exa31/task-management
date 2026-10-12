// swift-tools-version: 5.10
import PackageDescription

let package = Package(
    name: "TaskManagement",
    defaultLocalization: "en",
    platforms: [
        .iOS(.v17),
        .macOS(.v14)
    ],
    products: [
        .library(
            name: "TaskManagement",
            targets: ["TaskManagement"]
        )
    ],
    dependencies: [],
    targets: [
        .target(
            name: "TaskManagement",
            path: "TaskManagement",
            exclude: [
                "TaskManagement.entitlements",
                "Info.plist",
                "GoogleService-Info.plist",
                "Resources/en.lproj",
                "Resources/id.lproj"
            ]
        )
    ]
)
