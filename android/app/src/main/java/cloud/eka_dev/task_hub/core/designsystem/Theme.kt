package cloud.eka_dev.task_hub.core.designsystem

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color

// Frontend Web Brand Colors (Indigo to Violet)
val BrandIndigo = Color(0xFF4F46E5)
val BrandPrimary = Color(0xFF6366F1)
val BrandViolet = Color(0xFF7C3AED)
val BrandSky = Color(0xFF38BDF8)

// Dark Surfaces (matches frontend #090D16 and #0F172A)
val DarkSurfaceBackground = Color(0xFF090D16)
val DarkCardSurface = Color(0xFF0F172A)
val DarkCardBorder = Color(0xFF1E293B)
val LightSurfaceBackground = Color(0xFFF8FAFC)
val LightCardSurface = Color(0xFFFFFFFF)

// Status Colors (matches frontend STATUS_CLASSES)
val StatusBacklog = Color(0xFF64748B)
val StatusTodo = Color(0xFF3B82F6)
val StatusInProgress = Color(0xFFD97706)
val StatusReview = Color(0xFF9333EA)
val StatusDone = Color(0xFF10B981)

// Priority Colors (matches frontend PRIORITY_META)
val PriorityUrgent = Color(0xFFF43F5E)
val PriorityHigh = Color(0xFFF59E0B)
val PriorityMedium = Color(0xFF3B82F6)
val PriorityLow = Color(0xFF64748B)

val BrandGradientBrush = Brush.linearGradient(
    colors = listOf(BrandIndigo, BrandPrimary, BrandViolet)
)

private val LightColorScheme = lightColorScheme(
    primary = BrandPrimary,
    secondary = BrandViolet,
    background = LightSurfaceBackground,
    surface = LightCardSurface,
    onPrimary = Color.White,
    onBackground = Color(0xFF0F172A),
    onSurface = Color(0xFF0F172A),
    surfaceVariant = Color(0xFFF1F5F9),
    onSurfaceVariant = Color(0xFF64748B)
)

private val DarkColorScheme = darkColorScheme(
    primary = BrandPrimary,
    secondary = BrandViolet,
    background = DarkSurfaceBackground,
    surface = DarkCardSurface,
    onPrimary = Color.White,
    onBackground = Color(0xFFF8FAFC),
    onSurface = Color(0xFFF8FAFC),
    surfaceVariant = Color(0xFF1E293B),
    onSurfaceVariant = Color(0xFF94A3B8)
)

@Composable
fun TaskManagementTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit
) {
    val colorScheme = if (darkTheme) DarkColorScheme else LightColorScheme

    MaterialTheme(
        colorScheme = colorScheme,
        typography = Typography(),
        content = content
    )
}
