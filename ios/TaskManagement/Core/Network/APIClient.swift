import Foundation

public enum APIError: LocalizedError {
    case invalidURL
    case unauthorized
    case serverError(String)
    case decodingError(Error)
    case networkError(Error)

    public var errorDescription: String? {
        switch self {
        case .invalidURL: return "Invalid server endpoint URL."
        case .unauthorized: return "Session expired. Please sign in again."
        case .serverError(let msg): return msg
        case .decodingError(let err): return "Failed to parse data: \(err.localizedDescription)"
        case .networkError(let err): return err.localizedDescription
        }
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

        if let token = keychain.get(key: "access_token") {
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

        // Auto Refresh Token on 401
        if httpResponse.statusCode == 401 {
            if retryOn401 {
                AppLogger.warn("401 Unauthorized encountered. Attempting token refresh...", category: .auth)
                let refreshed = await refreshToken()
                if refreshed {
                    AppLogger.success("Token refreshed successfully. Retrying request...", category: .auth)
                    return try await request(endpoint: endpoint, method: method, body: body, retryOn401: false)
                }
            }

            await handleSessionExpired()
            throw APIError.unauthorized
        }

        guard (200...299).contains(httpResponse.statusCode) else {
            if let baseRes = try? decoder.decode(BaseResponse<String>.self, from: data) {
                AppLogger.error("HTTP error response: \(baseRes.message)", category: .network, metadata: ["statusCode": httpResponse.statusCode])
                throw APIError.serverError(baseRes.message)
            }
            let msg = "Server returned code \(httpResponse.statusCode)"
            AppLogger.error("HTTP error response: \(msg)", category: .network, metadata: ["statusCode": httpResponse.statusCode])
            throw APIError.serverError(msg)
        }

        do {
            let envelope = try decoder.decode(BaseResponse<T>.self, from: data)
            if let payload = envelope.data {
                return payload
            }
            if envelope.success, let empty = EmptyResponse() as? T {
                return empty
            }
            throw APIError.serverError(envelope.message)
        } catch {
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
        AppLogger.warn("Token refresh failed. Clearing credentials and broadcasting session expiration", category: .auth)
        keychain.clearAll()
        await MainActor.run {
            NotificationCenter.default.post(name: .sessionExpired, object: nil)
        }
    }

    private func refreshToken() async -> Bool {
        if let currentTask = refreshTask {
            return await currentTask.value
        }

        let task = Task<Bool, Never> { [keychain, baseURL, session, encoder, decoder] in
            guard let refreshToken = keychain.get(key: "refresh_token"),
                  let url = URL(string: "\(baseURL)/auth/refresh") else {
                return false
            }

            var req = URLRequest(url: url)
            req.httpMethod = "POST"
            req.setValue("application/json", forHTTPHeaderField: "Content-Type")
            req.httpBody = try? encoder.encode(RefreshTokenRequest(refreshToken: refreshToken))

            NetworkLogger.logRequest(req)
            let startTime = CFAbsoluteTimeGetCurrent()

            guard let (data, response) = try? await session.data(for: req),
                  let httpRes = response as? HTTPURLResponse else {
                NetworkLogger.logError(url: url.absoluteString, method: "POST", error: APIError.serverError("Token refresh network failed"))
                return false
            }

            let duration = CFAbsoluteTimeGetCurrent() - startTime
            NetworkLogger.logResponse(url: url.absoluteString, method: "POST", statusCode: httpRes.statusCode, data: data, duration: duration)

            guard httpRes.statusCode == 200,
                  let envelope = try? decoder.decode(BaseResponse<TokenPair>.self, from: data),
                  let tokens = envelope.data else {
                return false
            }

            keychain.save(key: "access_token", value: tokens.accessToken)
            keychain.save(key: "refresh_token", value: tokens.refreshToken)
            return true
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
        }
        keychain.save(key: "user_name", value: res.user.name)
        keychain.save(key: "user_email", value: res.user.email)
        return res
    }

    public func getProfile() async throws -> User {
        return try await request(endpoint: "auth/profile")
    }

    public func updateProfile(req: UpdateProfileRequest) async throws -> User {
        let body = try encoder.encode(req)
        let user: User = try await request(endpoint: "auth/profile", method: "PUT", body: body)
        keychain.save(key: "user_name", value: user.name)
        keychain.save(key: "user_email", value: user.email)
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
}
