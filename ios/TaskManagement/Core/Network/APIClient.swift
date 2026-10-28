import Foundation

public enum APIError: LocalizedError, Sendable {
    case invalidURL
    case unauthorized(code: String? = nil, message: String? = nil)
    case serverError(String, code: String? = nil, statusCode: Int? = nil)
    case decodingError(Error)
    case networkError(Error)
    case cancelled

    public var isCancelled: Bool {
        switch self {
        case .cancelled:
            return true
        case .networkError(let err):
            if err is CancellationError { return true }
            let nsErr = err as NSError
            return nsErr.domain == NSURLErrorDomain && nsErr.code == NSURLErrorCancelled
        default:
            return false
        }
    }

    public var code: String {
        switch self {
        case .invalidURL:
            return "BAD_REQUEST"
        case .unauthorized(let code, _):
            return code ?? "UNAUTHORIZED"
        case .serverError(_, let code, let statusCode):
            if let code = code, !code.isEmpty { return code }
            if let statusCode = statusCode { return APIError.fallbackCode(for: statusCode) }
            return "UNKNOWN_ERROR"
        case .decodingError:
            return "INTERNAL_SERVER_ERROR"
        case .networkError(let err):
            let nsErr = err as NSError
            if nsErr.domain == NSURLErrorDomain && (nsErr.code == NSURLErrorNotConnectedToInternet || nsErr.code == NSURLErrorNetworkConnectionLost) {
                return "NETWORK_ERROR"
            }
            return "NETWORK_ERROR"
        case .cancelled:
            return "CANCELLED"
        }
    }

    public var rawMessage: String {
        switch self {
        case .invalidURL: return "Invalid server endpoint URL."
        case .unauthorized(_, let msg): return msg ?? "Session expired. Please sign in again."
        case .serverError(let msg, _, _): return msg
        case .decodingError(let err): return "Failed to parse data: \(err.localizedDescription)"
        case .networkError(let err): return err.localizedDescription
        case .cancelled: return ""
        }
    }

    public static func fallbackCode(for statusCode: Int) -> String {
        switch statusCode {
        case 400: return "BAD_REQUEST"
        case 401: return "UNAUTHORIZED"
        case 403: return "FORBIDDEN"
        case 404: return "NOT_FOUND"
        case 409: return "CONFLICT"
        case 422: return "VALIDATION_ERROR"
        case 429: return "RATE_LIMIT_EXCEEDED"
        case 500: return "INTERNAL_SERVER_ERROR"
        case 503: return "SERVICE_ERROR"
        default: return "UNKNOWN_ERROR"
        }
    }

    public var errorDescription: String? {
        if isCancelled { return "" }
        return L10n.Errors.message(for: self)
    }
}

public extension Error {
    var isTaskOrURLCancelled: Bool {
        if self is CancellationError { return true }
        if let apiError = self as? APIError, apiError.isCancelled { return true }
        let nsErr = self as NSError
        return nsErr.domain == NSURLErrorDomain && nsErr.code == NSURLErrorCancelled
    }
}

extension Notification.Name {
    public static let sessionExpired = Notification.Name("TaskManagement.sessionExpired")
}

public actor APIClient {
    public static let shared = APIClient()

    private let baseURL = AppConfig.baseURL
    private let session: URLSession
    private let keychain = KeychainManager.shared
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder
    private var refreshTask: Task<Bool, Never>?

    private init() {
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        self.session = URLSession(configuration: config)
        self.decoder = JSONDecoder()
        self.encoder = JSONEncoder()
    }

    /// Determines if an endpoint is public (unauthenticated) and should not include a Bearer token or auto-refresh.
    private func isPublicEndpoint(_ endpoint: String) -> Bool {
        let publicPrefixes = [
            "auth/login",
            "auth/register",
            "auth/oauth",
            "auth/refresh",
            "auth/dev-login",
            "auth/apple/webhook"
        ]
        return publicPrefixes.contains { endpoint == $0 || endpoint.hasPrefix($0 + "/") || endpoint.hasPrefix($0 + "?") }
    }

    /// Checks if a JWT token has expired or is nearing expiration (within bufferSeconds).
    private func isTokenExpired(_ token: String, bufferSeconds: TimeInterval = 60) -> Bool {
        let parts = token.components(separatedBy: ".")
        guard parts.count >= 2 else { return true }
        var base64 = parts[1]
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")
        while base64.count % 4 != 0 {
            base64.append("=")
        }
        guard let data = Data(base64Encoded: base64),
              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
              let exp = json["exp"] as? TimeInterval else {
            return false
        }
        return Date().timeIntervalSince1970 >= (exp - bufferSeconds)
    }

    /// Retrieves a valid access token. If missing or expiring soon, performs proactive refresh when refresh token is available.
    private func getValidAccessToken() async -> String? {
        let currentToken = keychain.get(key: "access_token") ?? SharedWidgetDataManager.shared.getAuthToken()
        let hasRefreshToken = (keychain.get(key: "refresh_token") ?? SharedWidgetDataManager.shared.getRefreshToken()) != nil

        if let token = currentToken, !isTokenExpired(token) {
            return token
        }

        // Proactive token refresh if token is expired/expiring and refresh token is present
        if hasRefreshToken {
            AppLogger.info("Access token missing or expired; initiating proactive refresh...", category: .auth)
            let refreshed = await refreshToken()
            if refreshed {
                return keychain.get(key: "access_token") ?? SharedWidgetDataManager.shared.getAuthToken()
            }
        }

        return currentToken
    }

    public func request<T: Codable>(
        endpoint: String,
        method: String = "GET",
        body: Data? = nil,
        retryOn401: Bool = true,
        customHeaders: [String: String]? = nil
    ) async throws -> T {
        guard let url = URL(string: "\(baseURL)/\(endpoint)") else {
            let error = APIError.invalidURL
            NetworkLogger.logError(url: "\(baseURL)/\(endpoint)", method: method, error: error)
            throw error
        }

        var req = URLRequest(url: url)
        req.httpMethod = method
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let customHeaders = customHeaders {
            for (key, val) in customHeaders {
                req.setValue(val, forHTTPHeaderField: key)
            }
        }

        // Request Interceptor: Attach Bearer token for protected endpoints
        let isPublic = isPublicEndpoint(endpoint)
        if !isPublic {
            guard let token = await getValidAccessToken() else {
                AppLogger.warn("Unauthorized: No valid session or refresh token available for '\(endpoint)'", category: .auth)
                throw APIError.unauthorized()
            }
            req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        if let body = body {
            req.httpBody = body
        }

        NetworkLogger.logRequest(req)
        let startTime = CFAbsoluteTimeGetCurrent()

        let (data, response): (Data, URLResponse)
        do {
            (data, response) = try await session.data(for: req)
        } catch {
            if Task.isCancelled || (error as? URLError)?.code == .cancelled || (error as NSError).code == NSURLErrorCancelled {
                AppLogger.debug("Request cancelled for endpoint '\(endpoint)'", category: .network)
                throw APIError.cancelled
            }
            let durationMs = Int64((CFAbsoluteTimeGetCurrent() - startTime) * 1000)
            CrashReporterService.shared.recordHttpError(
                method: method,
                path: endpoint,
                statusCode: 0,
                errorMessage: error.localizedDescription,
                responseBody: nil,
                durationMs: durationMs
            )
            NetworkLogger.logError(url: url.absoluteString, method: method, error: error)
            throw APIError.networkError(error)
        }

        guard let httpResponse = response as? HTTPURLResponse else {
            let err = APIError.serverError("Invalid server response")
            NetworkLogger.logError(url: url.absoluteString, method: method, error: err)
            throw err
        }

        let duration = CFAbsoluteTimeGetCurrent() - startTime
        NetworkLogger.logResponse(
            url: url.absoluteString,
            method: method,
            statusCode: httpResponse.statusCode,
            data: data,
            duration: duration
        )

        // Response Interceptor: Auto Refresh Token on 401 for protected endpoints
        if httpResponse.statusCode == 401 && !isPublic {
            if retryOn401 {
                AppLogger.warn("401 Unauthorized encountered on '\(endpoint)'. Attempting token refresh...", category: .auth)
                let refreshed = await refreshToken()
                if refreshed {
                    AppLogger.success("Token refreshed successfully. Retrying request to '\(endpoint)'...", category: .auth)
                    return try await request(endpoint: endpoint, method: method, body: body, retryOn401: false, customHeaders: customHeaders)
                }
            }

            let errorEnvelope = try? decoder.decode(ApiErrorEnvelope.self, from: data)
            let code = errorEnvelope?.code ?? "UNAUTHORIZED"
            let msg = errorEnvelope?.message ?? "Session expired. Please sign in again."
            let durationMs = Int64(duration * 1000)
            CrashReporterService.shared.recordHttpError(
                method: method,
                path: endpoint,
                statusCode: 401,
                errorMessage: "[\(code)] \(msg)",
                responseBody: String(data: data, encoding: .utf8),
                durationMs: durationMs
            )
            throw APIError.unauthorized(code: code, message: msg)
        }

        guard (200...299).contains(httpResponse.statusCode) else {
            let errorEnvelope = try? decoder.decode(ApiErrorEnvelope.self, from: data)
            let code = errorEnvelope?.code ?? APIError.fallbackCode(for: httpResponse.statusCode)
            let msg = errorEnvelope?.message ?? "Server returned code \(httpResponse.statusCode)"
            let durationMs = Int64(duration * 1000)
            CrashReporterService.shared.recordHttpError(
                method: method,
                path: endpoint,
                statusCode: httpResponse.statusCode,
                errorMessage: "[\(code)] \(msg)",
                responseBody: String(data: data, encoding: .utf8),
                durationMs: durationMs
            )
            AppLogger.error("HTTP error response [\(code)] (status: \(httpResponse.statusCode)): \(msg)", category: .network)
            throw APIError.serverError(msg, code: code, statusCode: httpResponse.statusCode)
        }

        do {
            let envelope = try decoder.decode(BaseResponse<T>.self, from: data)
            if let payload = envelope.data {
                return payload
            }
            if envelope.success, let empty = EmptyResponse() as? T {
                return empty
            }
            let code = envelope.code ?? "OPERATION_FAILED"
            throw APIError.serverError(envelope.message, code: code, statusCode: httpResponse.statusCode)
        } catch {
            if let apiError = error as? APIError {
                throw apiError
            }
            if let emptyRes = try? decoder.decode(BaseResponse<EmptyResponse>.self, from: data),
               emptyRes.success,
               let empty = EmptyResponse() as? T {
                return empty
            }
            AppLogger.error("Failed to decode response payload", error: error, category: .network, metadata: ["type": "\(T.self)"])
            throw APIError.decodingError(error)
        }
    }

    private func handleSessionExpired() async {
        AppLogger.warn("Token permanently rejected. Clearing credentials and broadcasting session expiration", category: .auth)
        keychain.clearAll()
        SharedWidgetDataManager.shared.clearAllTokens()
        CrashReporterService.shared.setUserId(nil)
        await MainActor.run {
            NotificationCenter.default.post(name: .sessionExpired, object: nil)
        }
    }

    public func refreshToken() async -> Bool {
        if let currentTask = refreshTask {
            return await currentTask.value
        }

        let task = Task<Bool, Never> { [keychain, baseURL, session, encoder, decoder] in
            let refreshToken = keychain.get(key: "refresh_token") ?? SharedWidgetDataManager.shared.getRefreshToken()
            guard let refreshToken, let url = URL(string: "\(baseURL)/auth/refresh") else {
                AppLogger.warn("Cannot refresh token: no refresh token available in Keychain or SharedWidgetDataManager", category: .auth)
                return false
            }

            var req = URLRequest(url: url)
            req.httpMethod = "POST"
            req.setValue("application/json", forHTTPHeaderField: "Content-Type")
            req.httpBody = try? encoder.encode(RefreshTokenRequest(refreshToken: refreshToken))

            NetworkLogger.logRequest(req)
            let startTime = CFAbsoluteTimeGetCurrent()

            let data: Data
            let response: URLResponse
            do {
                (data, response) = try await session.data(for: req)
            } catch {
                NetworkLogger.logError(url: url.absoluteString, method: "POST", error: error)
                AppLogger.warn("Token refresh network request failed (transient error): \(error.localizedDescription)", category: .auth)
                return false
            }

            guard let httpRes = response as? HTTPURLResponse else {
                return false
            }

            let duration = CFAbsoluteTimeGetCurrent() - startTime
            NetworkLogger.logResponse(url: url.absoluteString, method: "POST", statusCode: httpRes.statusCode, data: data, duration: duration)

            if httpRes.statusCode == 200,
               let envelope = try? decoder.decode(BaseResponse<TokenPair>.self, from: data),
               let tokens = envelope.data {
                keychain.save(key: "access_token", value: tokens.accessToken)
                keychain.save(key: "refresh_token", value: tokens.refreshToken)
                SharedWidgetDataManager.shared.saveTokens(accessToken: tokens.accessToken, refreshToken: tokens.refreshToken)
                AppLogger.success("Token refreshed successfully; synchronized to Keychain and SharedWidgetDataManager", category: .auth)
                return true
            }

            // Permanent invalidation: server explicitly rejected refresh token
            if [400, 401, 403].contains(httpRes.statusCode) {
                AppLogger.error("Server rejected refresh token with status \(httpRes.statusCode). Session expired.", category: .auth)
                await self.handleSessionExpired()
            }

            return false
        }

        refreshTask = task
        let result = await task.value
        refreshTask = nil
        return result
    }

    // MARK: - Specialized API Calls

    public func oauthLogin(req: OAuthLoginRequest) async throws -> AuthResponse {
        let body = try encoder.encode(req)
        let res: AuthResponse = try await request(endpoint: "auth/oauth", method: "POST", body: body, retryOn401: false)
        if let tokens = res.tokens {
            keychain.save(key: "access_token", value: tokens.accessToken)
            keychain.save(key: "refresh_token", value: tokens.refreshToken)
            SharedWidgetDataManager.shared.saveTokens(accessToken: tokens.accessToken, refreshToken: tokens.refreshToken)
        }
        keychain.save(key: "user_name", value: res.user.name)
        keychain.save(key: "user_email", value: res.user.email)
        CrashReporterService.shared.setUserId(res.user.id)
        DateTimeHelper.setUserTimeZone(res.user.timezone)
        return res
    }

    public func getProfile() async throws -> User {
        let user: User = try await request(endpoint: "auth/profile")
        CrashReporterService.shared.setUserId(user.id)
        DateTimeHelper.setUserTimeZone(user.timezone)
        return user
    }

    public func updateProfile(req: UpdateProfileRequest) async throws -> User {
        let body = try encoder.encode(req)
        let user: User = try await request(endpoint: "auth/profile", method: "PUT", body: body)
        keychain.save(key: "user_name", value: user.name)
        keychain.save(key: "user_email", value: user.email)
        DateTimeHelper.setUserTimeZone(user.timezone)
        return user
    }

    public func uploadAvatar(data: Data, mimeType: String = "image/jpeg") async throws -> User {
        let boundary = "Boundary-\(UUID().uuidString)"
        var body = Data()
        body.append("--\(boundary)\r\n".data(using: .utf8)!)
        body.append("Content-Disposition: form-data; name=\"avatar\"; filename=\"avatar.jpg\"\r\n".data(using: .utf8)!)
        body.append("Content-Type: \(mimeType)\r\n\r\n".data(using: .utf8)!)
        body.append(data)
        body.append("\r\n--\(boundary)--\r\n".data(using: .utf8)!)

        let customHeaders = [
            "Content-Type": "multipart/form-data; boundary=\(boundary)"
        ]
        let user: User = try await request(
            endpoint: "auth/avatar",
            method: "POST",
            body: body,
            customHeaders: customHeaders
        )
        return user
    }

    public func bindOAuth(req: BindOAuthRequest) async throws -> User {
        let body = try encoder.encode(req)
        return try await request(endpoint: "auth/bind-oauth", method: "POST", body: body)
    }

    public func unbindOAuth(provider: String) async throws -> User {
        let body = try encoder.encode(UnbindOAuthRequest(provider: provider))
        return try await request(endpoint: "auth/unbind-oauth", method: "POST", body: body)
    }

    public func deleteAccount(confirmation: String = "DELETE") async throws {
        let body = try encoder.encode(DeleteAccountRequest(confirmation: confirmation))
        let _: EmptyResponse = try await request(endpoint: "auth/account", method: "DELETE", body: body)
        keychain.clearAll()
    }

    public func getTasks(status: String? = nil, search: String? = nil, projectId: Int64? = nil, limit: Int? = nil) async throws -> [TaskItem] {
        var queryItems: [URLQueryItem] = []
        if let status = status, !status.isEmpty {
            queryItems.append(URLQueryItem(name: "status", value: status))
        }
        if let search = search, !search.isEmpty {
            queryItems.append(URLQueryItem(name: "search", value: search))
        }
        if let projectId = projectId {
            queryItems.append(URLQueryItem(name: "project_id", value: String(projectId)))
        }
        if let limit = limit {
            queryItems.append(URLQueryItem(name: "limit", value: String(limit)))
        }
        var endpoint = "tasks"
        if !queryItems.isEmpty {
            var components = URLComponents()
            components.queryItems = queryItems
            if let queryString = components.percentEncodedQuery {
                endpoint += "?\(queryString)"
            }
        }
        return try await request(endpoint: endpoint)
    }

    public func getRecurringTasks() async throws -> [RecurringTaskItem] {
        return try await request(endpoint: "recurring-tasks")
    }

    public func createRecurringTask(req: CreateRecurringTaskRequest) async throws -> RecurringTaskItem {
        let body = try encoder.encode(req)
        return try await request(endpoint: "recurring-tasks", method: "POST", body: body)
    }

    public func updateRecurringTask(id: Int64, req: UpdateRecurringTaskRequest) async throws -> RecurringTaskItem {
        let body = try encoder.encode(req)
        return try await request(endpoint: "recurring-tasks/\(id)", method: "PUT", body: body)
    }

    public func deleteRecurringTask(id: Int64) async throws {
        let _: EmptyResponse = try await request(endpoint: "recurring-tasks/\(id)", method: "DELETE")
    }

    public func toggleActiveRecurringTask(id: Int64) async throws -> RecurringTaskItem {
        return try await request(endpoint: "recurring-tasks/\(id)/toggle-active", method: "POST")
    }

    public func getTasksPaginated(
        status: String? = nil,
        search: String? = nil,
        projectId: Int64? = nil,
        priority: String? = nil,
        limit: Int = 15,
        cursor: Int64? = nil
    ) async throws -> TaskCursorResponse {
        var queryItems: [URLQueryItem] = [
            URLQueryItem(name: "paginate", value: "cursor"),
            URLQueryItem(name: "limit", value: String(limit))
        ]
        if let status = status, !status.isEmpty {
            queryItems.append(URLQueryItem(name: "status", value: status))
        }
        if let search = search, !search.isEmpty {
            queryItems.append(URLQueryItem(name: "search", value: search))
        }
        if let projectId = projectId {
            queryItems.append(URLQueryItem(name: "project_id", value: String(projectId)))
        }
        if let priority = priority, !priority.isEmpty {
            queryItems.append(URLQueryItem(name: "priority", value: priority))
        }
        if let cursor = cursor {
            queryItems.append(URLQueryItem(name: "cursor", value: String(cursor)))
        }
        var endpoint = "tasks"
        var components = URLComponents()
        components.queryItems = queryItems
        if let queryString = components.percentEncodedQuery {
            endpoint += "?\(queryString)"
        }
        return try await request(endpoint: endpoint)
    }

    public func getTask(id: Int64) async throws -> TaskItem {
        return try await request(endpoint: "tasks/\(id)")
    }

    public func createTask(req: CreateTaskRequest) async throws -> TaskItem {
        let body = try encoder.encode(req)
        return try await request(endpoint: "tasks", method: "POST", body: body)
    }

    public func updateTask(id: Int64, req: UpdateTaskRequest) async throws -> TaskItem {
        let body = try encoder.encode(req)
        return try await request(endpoint: "tasks/\(id)", method: "PUT", body: body)
    }

    public func toggleTaskFocus(id: Int64) async throws -> TaskItem {
        return try await request(endpoint: "tasks/\(id)/focus", method: "POST")
    }

    public func deleteTask(id: Int64) async throws {
        let _: EmptyResponse = try await request(endpoint: "tasks/\(id)", method: "DELETE")
    }

    public func updateSubtask(taskId: Int64, subtaskId: Int64, isCompleted: Bool) async throws -> Subtask {
        let body = try encoder.encode(UpdateSubtaskRequest(title: nil, isCompleted: isCompleted))
        return try await request(endpoint: "tasks/\(taskId)/subtasks/\(subtaskId)", method: "PUT", body: body)
    }

    public func updateSubtask(taskId: Int64, subtaskId: Int64, req: UpdateSubtaskRequest) async throws -> Subtask {
        let body = try encoder.encode(req)
        return try await request(endpoint: "tasks/\(taskId)/subtasks/\(subtaskId)", method: "PUT", body: body)
    }

    public func createSubtask(taskId: Int64, req: CreateSubtaskRequest) async throws -> Subtask {
        let body = try encoder.encode(req)
        return try await request(endpoint: "tasks/\(taskId)/subtasks", method: "POST", body: body)
    }

    public func deleteSubtask(taskId: Int64, subtaskId: Int64) async throws {
        let _: EmptyResponse = try await request(endpoint: "tasks/\(taskId)/subtasks/\(subtaskId)", method: "DELETE")
    }

    public func getWidgetBundle() async throws -> WidgetBundleResponse {
        return try await request(endpoint: "tasks/widgets/bundle")
    }

    public func getTaskActivities(taskId: Int64) async throws -> [TaskActivity] {
        return try await request(endpoint: "tasks/\(taskId)/activities")
    }

    public func createTaskActivity(taskId: Int64, req: CreateTaskActivityRequest) async throws -> TaskActivity {
        let body = try encoder.encode(req)
        return try await request(endpoint: "tasks/\(taskId)/activities", method: "POST", body: body)
    }

    public func getProjects() async throws -> [Project] {
        return try await request(endpoint: "projects")
    }

    public func createProject(req: CreateProjectRequest) async throws -> Project {
        let body = try encoder.encode(req)
        return try await request(endpoint: "projects", method: "POST", body: body)
    }

    public func updateProject(id: Int64, req: UpdateProjectRequest) async throws -> Project {
        let body = try encoder.encode(req)
        return try await request(endpoint: "projects/\(id)", method: "PUT", body: body)
    }

    public func deleteProject(id: Int64) async throws {
        let _: EmptyResponse = try await request(endpoint: "projects/\(id)", method: "DELETE")
    }

    public func getNotes() async throws -> [Note] {
        return try await request(endpoint: "notes")
    }

    public func createNote(req: CreateNoteRequest) async throws -> Note {
        let body = try encoder.encode(req)
        return try await request(endpoint: "notes", method: "POST", body: body)
    }

    public func updateNote(id: Int64, req: UpdateNoteRequest) async throws -> Note {
        let body = try encoder.encode(req)
        return try await request(endpoint: "notes/\(id)", method: "PUT", body: body)
    }

    public func deleteNote(id: Int64) async throws {
        let _: EmptyResponse? = try? await request(endpoint: "notes/\(id)", method: "DELETE")
    }

    public func convertNoteToTask(id: Int64, req: ConvertNoteToTaskRequest = ConvertNoteToTaskRequest()) async throws -> ConvertNoteResponse {
        let body = try encoder.encode(req)
        return try await request(endpoint: "notes/\(id)/convert-task", method: "POST", body: body)
    }

    public func getReminders() async throws -> [Reminder] {
        do {
            let res: PaginatedResponse<Reminder> = try await request(endpoint: "reminders")
            return res.items
        } catch {
            if let directList: [Reminder] = try? await request(endpoint: "reminders") {
                return directList
            }
            throw error
        }
    }

    public func createReminder(title: String, description: String? = nil, remindAt: String, taskId: Int64? = nil) async throws -> Reminder {
        let req = CreateReminderRequest(title: title, description: description, remindAt: remindAt, taskId: taskId)
        let body = try encoder.encode(req)
        return try await request(endpoint: "reminders", method: "POST", body: body)
    }

    public func updateReminder(id: Int64, req: UpdateReminderRequest) async throws -> Reminder {
        let body = try encoder.encode(req)
        return try await request(endpoint: "reminders/\(id)", method: "PUT", body: body)
    }

    public func deleteReminder(id: Int64) async throws {
        let _: EmptyResponse? = try? await request(endpoint: "reminders/\(id)", method: "DELETE")
    }

    public func snoozeReminder(id: Int64, minutes: Int = 15) async throws -> Reminder {
        return try await request(endpoint: "reminders/\(id)/snooze?minutes=\(minutes)", method: "POST")
    }

    public func dismissReminder(id: Int64) async throws -> Reminder {
        return try await request(endpoint: "reminders/\(id)/dismiss", method: "POST")
    }

    public func registerDeviceToken(token: String) async throws {
        let req = RegisterDeviceTokenRequest(token: token, platform: "ios")
        let body = try encoder.encode(req)
        let _: String? = try? await request(endpoint: "reminders/device-token", method: "POST", body: body)
    }

    // MARK: - MCP (Model Context Protocol)
    public func getMcpTokens() async throws -> [McpToken] {
        return try await request(endpoint: "tokens")
    }

    public func generateMcpToken(name: String) async throws -> GenerateMcpTokenResponse {
        let req = GenerateMcpTokenRequest(name: name)
        let body = try encoder.encode(req)
        return try await request(endpoint: "tokens", method: "POST", body: body)
    }

    public func deleteMcpToken(id: Int64) async throws {
        let _: EmptyResponse = try await request(endpoint: "tokens/\(id)", method: "DELETE")
    }

    // MARK: - GraphQL Client
    public func graphql<T: Codable>(
        query: String,
        variables: [String: Any]? = nil,
        retryOn401: Bool = true
    ) async throws -> T {
        guard let url = URL(string: "\(baseURL)/graphql") else {
            throw APIError.invalidURL
        }

        var req = URLRequest(url: url)
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")

        guard let token = await getValidAccessToken() else {
            AppLogger.warn("Unauthorized: No valid session or refresh token available for GraphQL", category: .auth)
            throw APIError.unauthorized()
        }
        req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        var payload: [String: Any] = ["query": query]
        if let variables = variables {
            payload["variables"] = variables
        }
        req.httpBody = try JSONSerialization.data(withJSONObject: payload)

        NetworkLogger.logRequest(req)
        let startTime = CFAbsoluteTimeGetCurrent()

        let (data, response): (Data, URLResponse)
        do {
            (data, response) = try await session.data(for: req)
        } catch {
            if Task.isCancelled || (error as? URLError)?.code == .cancelled || (error as NSError).code == NSURLErrorCancelled {
                AppLogger.debug("GraphQL query cancelled", category: .network)
                throw APIError.cancelled
            }
            NetworkLogger.logError(url: url.absoluteString, method: "POST", error: error)
            throw APIError.networkError(error)
        }

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.serverError("Invalid server response")
        }

        let duration = CFAbsoluteTimeGetCurrent() - startTime
        NetworkLogger.logResponse(
            url: url.absoluteString,
            method: "POST",
            statusCode: httpResponse.statusCode,
            data: data,
            duration: duration
        )

        if httpResponse.statusCode == 401 {
            if retryOn401 {
                let refreshed = await refreshToken()
                if refreshed {
                    return try await graphql(query: query, variables: variables, retryOn401: false)
                }
            }
            throw APIError.unauthorized()
        }

        guard (200...299).contains(httpResponse.statusCode) else {
            throw APIError.serverError("Server returned code \(httpResponse.statusCode)")
        }

        let envelope = try decoder.decode(GraphQLEnvelope<T>.self, from: data)
        if let errors = envelope.errors, !errors.isEmpty {
            let combined = errors.map { $0.message }.joined(separator: ", ")
            throw APIError.serverError(combined)
        }

        guard let result = envelope.data else {
            throw APIError.serverError("No data in GraphQL response")
        }

        return result
    }
}

// MARK: - Internal GraphQL Envelope
public struct GraphQLEnvelope<P: Codable>: Codable {
    public let data: P?
    public let errors: [GraphQLErrorItem]?
}

public struct GraphQLErrorItem: Codable {
    public let message: String
}
