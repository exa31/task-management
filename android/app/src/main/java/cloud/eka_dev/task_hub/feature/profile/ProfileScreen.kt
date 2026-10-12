package cloud.eka_dev.task_hub.feature.profile

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.task_hub.core.datastore.TokenManager
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProfileScreen(
    tokenManager: TokenManager,
    onSignOut: () -> Unit
) {
    val coroutineScope = rememberCoroutineScope()
    var userEmail by remember { mutableStateOf("user@example.com") }

    LaunchedEffect(Unit) {
        // Load stored user info if available
    }

    Scaffold(
        topBar = { TopAppBar(title = { Text("Pengaturan & Profil") }) }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp)
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Akun Pengguna", fontWeight = FontWeight.Bold, fontSize = 16.sp)
                    Spacer(modifier = Modifier.height(8.dp))
                    Text("Status: Terhubung", fontSize = 14.sp)
                }
            }

            Button(
                onClick = {
                    coroutineScope.launch {
                        tokenManager.clearSession()
                        onSignOut()
                    }
                },
                colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.error),
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp)
            ) {
                Text("Keluar (Sign Out)", color = Color.White)
            }
        }
    }
}
