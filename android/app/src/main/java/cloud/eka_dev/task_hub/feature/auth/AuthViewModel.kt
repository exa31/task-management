package cloud.eka_dev.task_hub.feature.auth

import android.content.Context
import androidx.credentials.CredentialManager
import androidx.credentials.GetCredentialRequest
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.google.android.libraries.identity.googleid.GetGoogleIdOption
import com.google.android.libraries.identity.googleid.GoogleIdTokenCredential
import cloud.eka_dev.task_hub.core.datastore.TokenManager
import cloud.eka_dev.task_hub.core.model.OAuthLoginRequest
import cloud.eka_dev.task_hub.core.network.ApiClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

sealed class AuthState {
    object Idle : AuthState()
    object Loading : AuthState()
    data class Authenticated(val userName: String) : AuthState()
    data class Error(val message: String) : AuthState()
}

class AuthViewModel(
    private val apiClient: ApiClient,
    private val tokenManager: TokenManager
) : ViewModel() {

    private val _state = MutableStateFlow<AuthState>(AuthState.Idle)
    val state: StateFlow<AuthState> = _state.asStateFlow()

    init {
        checkExistingSession()
    }

    private fun checkExistingSession() {
        viewModelScope.launch {
            val token = tokenManager.getAccessToken()
            if (!token.isNullOrEmpty()) {
                _state.value = AuthState.Authenticated("User")
            }
        }
    }

    fun signInWithGoogle(context: Context, serverClientId: String) {
        viewModelScope.launch {
            _state.value = AuthState.Loading
            try {
                val credentialManager = CredentialManager.create(context)
                val googleIdOption = GetGoogleIdOption.Builder()
                    .setFilterByAuthorizedAccounts(false)
                    .setServerClientId(serverClientId)
                    .setAutoSelectEnabled(true)
                    .build()

                val request = GetCredentialRequest.Builder()
                    .addCredentialOption(googleIdOption)
                    .build()

                val result = credentialManager.getCredential(context, request)
                val credential = result.credential

                if (credential.type == GoogleIdTokenCredential.TYPE_GOOGLE_ID_TOKEN_CREDENTIAL) {
                    val googleIdTokenCredential = GoogleIdTokenCredential.createFrom(credential.data)
                    val idToken = googleIdTokenCredential.idToken

                    val response = apiClient.apiService.oauthLogin(
                        OAuthLoginRequest(
                            provider = "google",
                            idToken = idToken,
                            name = googleIdTokenCredential.displayName,
                            email = googleIdTokenCredential.id
                        )
                    )

                    if (response.success && response.data?.tokens != null) {
                        val tokens = response.data.tokens
                        val user = response.data.user
                        tokenManager.saveTokens(tokens.accessToken, tokens.refreshToken)
                        tokenManager.saveUserInfo(user.id, user.name, user.email)
                        _state.value = AuthState.Authenticated(user.name)
                    } else {
                        _state.value = AuthState.Error(response.message.ifEmpty { "Login failed" })
                    }
                }
            } catch (e: Exception) {
                _state.value = AuthState.Error(e.localizedMessage ?: "Google sign in error")
            }
        }
    }

    fun signInWithAppleSimulated(idToken: String, email: String, name: String) {
        viewModelScope.launch {
            _state.value = AuthState.Loading
            try {
                val response = apiClient.apiService.oauthLogin(
                    OAuthLoginRequest(
                        provider = "apple",
                        idToken = idToken,
                        email = email,
                        name = name
                    )
                )
                if (response.success && response.data?.tokens != null) {
                    val tokens = response.data.tokens
                    val user = response.data.user
                    tokenManager.saveTokens(tokens.accessToken, tokens.refreshToken)
                    tokenManager.saveUserInfo(user.id, user.name, user.email)
                    _state.value = AuthState.Authenticated(user.name)
                } else {
                    _state.value = AuthState.Error(response.message.ifEmpty { "Apple login failed" })
                }
            } catch (e: Exception) {
                _state.value = AuthState.Error(e.localizedMessage ?: "Apple login error")
            }
        }
    }

    fun signOut() {
        viewModelScope.launch {
            tokenManager.clearSession()
            _state.value = AuthState.Idle
        }
    }
}
