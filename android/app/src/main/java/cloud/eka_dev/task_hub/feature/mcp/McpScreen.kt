package cloud.eka_dev.task_hub.feature.mcp

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import cloud.eka_dev.task_hub.R
import cloud.eka_dev.task_hub.core.designsystem.AppButton
import cloud.eka_dev.task_hub.core.designsystem.BrandPrimary
import cloud.eka_dev.task_hub.core.designsystem.TaskListSkeleton
import cloud.eka_dev.task_hub.core.model.GenerateMcpTokenRequest
import cloud.eka_dev.task_hub.core.model.McpToken
import cloud.eka_dev.task_hub.core.network.ApiClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class McpViewModel(private val apiClient: ApiClient) : ViewModel() {
    private val _tokens = MutableStateFlow<List<McpToken>>(emptyList())
    val tokens: StateFlow<List<McpToken>> = _tokens.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    private val _createdToken = MutableStateFlow<String?>(null)
    val createdToken: StateFlow<String?> = _createdToken.asStateFlow()

    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error.asStateFlow()

    init {
        loadTokens()
    }

    fun loadTokens() {
        viewModelScope.launch {
            _isLoading.value = true
            try {
                val res = apiClient.apiService.getMcpTokens()
                if (res.success && res.data != null) {
                    _tokens.value = res.data
                }
            } catch (e: Exception) {
                _error.value = e.localizedMessage
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun generateToken(name: String) {
        viewModelScope.launch {
            try {
                val res = apiClient.apiService.generateMcpToken(GenerateMcpTokenRequest(name = name))
                if (res.success && res.data != null) {
                    _createdToken.value = res.data.token
                    loadTokens()
                }
            } catch (e: Exception) {
                _error.value = e.localizedMessage
            }
        }
    }

    fun deleteToken(id: Long) {
        viewModelScope.launch {
            try {
                apiClient.apiService.deleteMcpToken(id)
                loadTokens()
            } catch (e: Exception) {
                _error.value = e.localizedMessage
            }
        }
    }

    fun dismissCreatedToken() {
        _createdToken.value = null
    }

    fun clearError() {
        _error.value = null
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun McpScreen(
    viewModel: McpViewModel,
    onBack: () -> Unit
) {
    val tokens by viewModel.tokens.collectAsState()
    val isLoading by viewModel.isLoading.collectAsState()
    val createdToken by viewModel.createdToken.collectAsState()
    val error by viewModel.error.collectAsState()
    val context = LocalContext.current

    var showCreateDialog by remember { mutableStateOf(false) }
    var deletingToken by remember { mutableStateOf<McpToken?>(null) }
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(error) {
        error?.let {
            snackbarHostState.showSnackbar(it)
            viewModel.clearError()
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.mcp_screen_title), fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = stringResource(R.string.back))
                    }
                }
            )
        },
        floatingActionButton = {
            FloatingActionButton(
                onClick = { showCreateDialog = true },
                containerColor = BrandPrimary,
                contentColor = Color.White,
                shape = CircleShape
            ) {
                Icon(Icons.Default.Add, contentDescription = stringResource(R.string.mcp_generate_btn))
            }
        },
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(14.dp)
        ) {
            // Info Banner
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f))
            ) {
                Column(modifier = Modifier.padding(14.dp)) {
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        Icon(Icons.Default.IntegrationInstructions, contentDescription = null, tint = BrandPrimary, modifier = Modifier.size(20.dp))
                        Text(stringResource(R.string.mcp_info_title), fontWeight = FontWeight.Bold, fontSize = 14.sp)
                    }
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        stringResource(R.string.mcp_info_desc),
                        fontSize = 12.sp,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }

            if (isLoading && tokens.isEmpty()) {
                TaskListSkeleton(count = 3)
            } else if (tokens.isEmpty()) {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Icon(Icons.Default.KeyOff, contentDescription = null, tint = MaterialTheme.colorScheme.onSurfaceVariant, modifier = Modifier.size(48.dp))
                        Text(stringResource(R.string.mcp_empty_title), fontWeight = FontWeight.SemiBold)
                        Text(stringResource(R.string.mcp_empty_desc), fontSize = 12.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            } else {
                LazyColumn(
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                    modifier = Modifier.fillMaxSize()
                ) {
                    items(tokens, key = { it.id }) { token ->
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(14.dp),
                            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                            elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
                        ) {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(14.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.SpaceBetween
                            ) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(10.dp),
                                    modifier = Modifier.weight(1f)
                                ) {
                                    Box(
                                        modifier = Modifier
                                            .size(36.dp)
                                            .clip(CircleShape)
                                            .background(BrandPrimary.copy(alpha = 0.12f)),
                                        contentAlignment = Alignment.Center
                                    ) {
                                        Icon(Icons.Default.Key, contentDescription = null, tint = BrandPrimary, modifier = Modifier.size(18.dp))
                                    }
                                    Column {
                                        Text(token.name, fontWeight = FontWeight.Bold, fontSize = 14.sp)
                                        Text(
                                            "Dibuat: ${token.createdAt.take(10)}",
                                            fontSize = 11.sp,
                                            color = MaterialTheme.colorScheme.onSurfaceVariant
                                        )
                                        if (!token.lastUsedAt.isNullOrEmpty()) {
                                            Text(
                                                "Terakhir aktif: ${token.lastUsedAt.take(10)}",
                                                fontSize = 10.sp,
                                                color = Color(0xFF10B981)
                                            )
                                        }
                                    }
                                }

                                IconButton(onClick = { deletingToken = token }) {
                                    Icon(Icons.Default.DeleteOutline, contentDescription = stringResource(R.string.delete), tint = MaterialTheme.colorScheme.error.copy(alpha = 0.7f))
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    if (showCreateDialog) {
        var tokenName by remember { mutableStateOf("") }
        AlertDialog(
            onDismissRequest = { showCreateDialog = false },
            title = { Text(stringResource(R.string.mcp_create_dialog_title), fontWeight = FontWeight.Bold) },
            text = {
                OutlinedTextField(
                    value = tokenName,
                    onValueChange = { tokenName = it },
                    label = { Text(stringResource(R.string.mcp_token_name_label)) },
                    shape = RoundedCornerShape(12.dp),
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true
                )
            },
            confirmButton = {
                AppButton(
                    text = stringResource(R.string.mcp_generate_btn),
                    enabled = tokenName.isNotBlank(),
                    onClick = {
                        viewModel.generateToken(tokenName.trim())
                        showCreateDialog = false
                    }
                )
            },
            dismissButton = {
                TextButton(onClick = { showCreateDialog = false }) { Text(stringResource(R.string.cancel)) }
            }
        )
    }

    createdToken?.let { tokenString ->
        AlertDialog(
            onDismissRequest = { viewModel.dismissCreatedToken() },
            title = { Text(stringResource(R.string.mcp_created_title), fontWeight = FontWeight.Bold) },
            text = {
                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    Text(
                        stringResource(R.string.mcp_created_desc),
                        fontSize = 12.sp,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Surface(
                        shape = RoundedCornerShape(8.dp),
                        color = MaterialTheme.colorScheme.surfaceVariant,
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Text(
                            text = tokenString,
                            fontSize = 12.sp,
                            fontFamily = FontFamily.Monospace,
                            fontWeight = FontWeight.Medium,
                            modifier = Modifier.padding(10.dp)
                        )
                    }
                }
            },
            confirmButton = {
                AppButton(
                    text = stringResource(R.string.mcp_copy_token),
                    onClick = {
                        val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                        val clip = ClipData.newPlainText("MCP Token", tokenString)
                        clipboard.setPrimaryClip(clip)
                        viewModel.dismissCreatedToken()
                    }
                )
            },
            dismissButton = {
                TextButton(onClick = { viewModel.dismissCreatedToken() }) { Text(stringResource(R.string.close)) }
            }
        )
    }

    deletingToken?.let { token ->
        AlertDialog(
            onDismissRequest = { deletingToken = null },
            title = { Text(stringResource(R.string.mcp_delete_title), fontWeight = FontWeight.Bold) },
            text = { Text(stringResource(R.string.mcp_delete_desc, token.name)) },
            confirmButton = {
                Button(
                    onClick = {
                        viewModel.deleteToken(token.id)
                        deletingToken = null
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.error)
                ) {
                    Text(stringResource(R.string.mcp_revoke_btn), color = Color.White)
                }
            },
            dismissButton = {
                TextButton(onClick = { deletingToken = null }) { Text(stringResource(R.string.cancel)) }
            }
        )
    }
}
