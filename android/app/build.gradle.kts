plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.kotlin.serialization)
}

android {
    namespace = "cloud.eka_dev.task_hub"
    compileSdk = 35

    buildFeatures {
        compose = true
        buildConfig = true
    }

    defaultConfig {
        applicationId = "cloud.eka_dev.task_hub"
        minSdk = 26
        targetSdk = 35
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        // Backend Cloud Environment
        buildConfigField("String", "BASE_URL", "\"https://be-task-hub.eka-dev.cloud/api/v1/\"")
        buildConfigField("String", "SSE_URL", "\"https://be-task-hub.eka-dev.cloud/api/v1/realtime/sse/events\"")
        buildConfigField("String", "GOOGLE_WEB_CLIENT_ID", "\"YOUR_DEV_GOOGLE_CLIENT_ID.apps.googleusercontent.com\"")
    }

    buildTypes {
        debug {
            buildConfigField("String", "BASE_URL", "\"https://be-task-hub.eka-dev.cloud/api/v1/\"")
            buildConfigField("String", "SSE_URL", "\"https://be-task-hub.eka-dev.cloud/api/v1/realtime/sse/events\"")
        }
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
            buildConfigField("String", "BASE_URL", "\"https://be-task-hub.eka-dev.cloud/api/v1/\"")
            buildConfigField("String", "SSE_URL", "\"https://be-task-hub.eka-dev.cloud/api/v1/realtime/sse/events\"")
            buildConfigField("String", "GOOGLE_WEB_CLIENT_ID", "\"YOUR_PROD_GOOGLE_CLIENT_ID.apps.googleusercontent.com\"")
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
    implementation(libs.androidx.credentials.play-services)
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
