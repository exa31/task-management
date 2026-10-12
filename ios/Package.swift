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
    dependencies: [
        .package(url: "https://github.com/google/GoogleSignIn-iOS.git", from: "7.1.0")
    ],
    targets: [
        .target(
            name: "TaskManagement",
            dependencies: [
                .product(name: "GoogleSignIn", package: "GoogleSignIn-iOS"),
                .product(name: "GoogleSignInSwift", package: "GoogleSignIn-iOS")
            ],
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
