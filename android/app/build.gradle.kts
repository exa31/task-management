plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.kotlin.serialization)
}

import java.io.FileInputStream
import java.util.Properties

// Read configuration from System Environment, .env (from CI/CD ENV_FILE or root), or local.properties
val envProperties = Properties().apply {
    // 1. Check root .env (from monorepo root or CI/CD ENV_FILE dump)
    val rootEnv = rootProject.file("../.env")
    val localEnv = rootProject.file(".env")
    val targetEnv = when {
        rootEnv.exists() -> rootEnv
        localEnv.exists() -> localEnv
        else -> null
    }
    targetEnv?.forEachLine { line ->
        val trimmed = line.trim()
        if (trimmed.isNotEmpty() && !trimmed.startsWith("#") && trimmed.contains("=")) {
            val parts = trimmed.split("=", limit = 2)
            put(parts[0].trim(), parts[1].trim().trim('"', '\''))
        }
    }

    // 2. Check local.properties (standard Android SDK and local config)
    val localProps = rootProject.file("local.properties")
    if (localProps.exists()) {
        FileInputStream(localProps).use { load(it) }
    }
}

fun getEnv(vararg keys: String, defaultVal: String = ""): String {
    for (key in keys) {
        val sysVal = System.getenv(key)
        if (!sysVal.isNullOrBlank()) return sysVal

        val propVal = envProperties.getProperty(key)
        if (!propVal.isNullOrBlank()) return propVal

        val gradleProp = project.findProperty(key) as? String
        if (!gradleProp.isNullOrBlank()) return gradleProp
    }
    return defaultVal
}

android {
    namespace = "cloud.eka_dev.task_hub"
    compileSdk = 35

    buildFeatures {
        compose = true
        buildConfig = true
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    defaultConfig {
        applicationId = "cloud.eka_dev.task_hub"
        minSdk = 26
        targetSdk = 35
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        // Dynamic Environment Config (matches iOS EnvReader behavior)
        val defaultBaseUrl = getEnv("API_BASE_URL", "BASE_URL", "VITE_API_URL", defaultVal = "https://be-task-hub.eka-dev.cloud/api/v1/")
        val defaultSseUrl = getEnv("SSE_URL", "VITE_WS_URL", defaultVal = "${defaultBaseUrl.trimEnd('/')}/realtime/sse/events")
        val defaultGoogleClientId = getEnv("GOOGLE_CLIENT_ID", "GOOGLE_WEB_CLIENT_ID", "VITE_GOOGLE_CLIENT_ID", defaultVal = "897905079551-86q6g4ov6b966r2t7sbmqsmks5bea682.apps.googleusercontent.com")
        val defaultLokiUrl = getEnv("LOKI_URL", defaultVal = "https://loki.eka-dev.cloud/loki/api/v1/push")

        buildConfigField("String", "BASE_URL", "\"$defaultBaseUrl\"")
        buildConfigField("String", "SSE_URL", "\"$defaultSseUrl\"")
        buildConfigField("String", "GOOGLE_WEB_CLIENT_ID", "\"$defaultGoogleClientId\"")
        buildConfigField("String", "LOKI_URL", "\"$defaultLokiUrl\"")
    }

    signingConfigs {
        create("release") {
            val keystorePath = getEnv(
                "KEYSTORE_FILE",
                "ANDROID_KEYSTORE_FILE",
                "ANDROID_KEYSTORE_PATH",
                defaultVal = rootProject.file("taskhub-release.jks").absolutePath
            )
            val keystoreFile = file(keystorePath)
            if (keystoreFile.exists()) {
                storeFile = keystoreFile
                storePassword = getEnv("KEYSTORE_PASSWORD", "ANDROID_KEYSTORE_PASSWORD", defaultVal = "TaskHubSecureRelease2026!")
                keyAlias = getEnv("KEY_ALIAS", "ANDROID_KEY_ALIAS", defaultVal = "taskhub")
                keyPassword = getEnv("KEY_PASSWORD", "ANDROID_KEY_PASSWORD", defaultVal = "TaskHubSecureRelease2026!")
            }
        }
    }

    buildTypes {
        debug {
            val debugBaseUrl = getEnv("DEV_BASE_URL", "API_BASE_URL", "BASE_URL", defaultVal = "https://be-task-hub.eka-dev.cloud/api/v1/")
            val debugSseUrl = getEnv("DEV_SSE_URL", "SSE_URL", defaultVal = "${debugBaseUrl.trimEnd('/')}/realtime/sse/events")
            buildConfigField("String", "BASE_URL", "\"$debugBaseUrl\"")
            buildConfigField("String", "SSE_URL", "\"$debugSseUrl\"")
        }
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
            val releaseSigning = signingConfigs.getByName("release")
            if (releaseSigning.storeFile?.exists() == true) {
                signingConfig = releaseSigning
            }

            val prodBaseUrl = getEnv("PROD_BASE_URL", "API_BASE_URL", "BASE_URL", defaultVal = "https://be-task-hub.eka-dev.cloud/api/v1/")
            val prodSseUrl = getEnv("PROD_SSE_URL", "SSE_URL", defaultVal = "${prodBaseUrl.trimEnd('/')}/realtime/sse/events")
            val prodGoogleClientId = getEnv("PROD_GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_ID", "GOOGLE_WEB_CLIENT_ID", defaultVal = "897905079551-86q6g4ov6b966r2t7sbmqsmks5bea682.apps.googleusercontent.com")

            buildConfigField("String", "BASE_URL", "\"$prodBaseUrl\"")
            buildConfigField("String", "SSE_URL", "\"$prodSseUrl\"")
            buildConfigField("String", "GOOGLE_WEB_CLIENT_ID", "\"$prodGoogleClientId\"")
        }
    }
}

dependencies {
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.lifecycle.viewmodel.compose)
    implementation(libs.androidx.activity.compose)
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.ui)
    implementation(libs.androidx.ui.graphics)
    implementation(libs.androidx.ui.tooling.preview)
    implementation(libs.androidx.material3)
    implementation(libs.androidx.material.icons)
    implementation(libs.androidx.navigation.compose)

    // DataStore & Auth
    implementation(libs.androidx.datastore.preferences)
    implementation(libs.androidx.credentials)
    implementation(libs.androidx.credentials.play.services)
    implementation(libs.googleid)

    // Network & SSE
    implementation(libs.okhttp)
    implementation(libs.okhttp.logging)
    implementation(libs.okhttp.sse)
    implementation(libs.retrofit)
    implementation(libs.retrofit.kotlinx.serialization)
    implementation(libs.kotlinx.serialization.json)

    debugImplementation(libs.androidx.ui.tooling)
}
