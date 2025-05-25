class ChatApp {
  constructor() {
    this.token = localStorage.getItem("token");
    this.currentUser = null;
    this.selectedUser = null;
    this.ws = null;
    this.users = [];
    this.onlineUsers = [];
    this.selectedBroadcastUsers = [];
    this.broadcastMessages = [];

    this.initializeElements();
    this.attachEventListeners();
    this.checkAuthStatus();
  }

  initializeElements() {
    // Auth elements
    this.authSection = document.getElementById("auth-section");
    this.chatSection = document.getElementById("chat-section");
    this.loginTab = document.getElementById("login-tab");
    this.registerTab = document.getElementById("register-tab");
    this.loginForm = document.getElementById("login-form");
    this.registerForm = document.getElementById("register-form");
    this.authError = document.getElementById("auth-error");

    // Chat elements
    this.currentUserSpan = document.getElementById("current-user");
    this.logoutBtn = document.getElementById("logout-btn");
    this.usersList = document.getElementById("users-list");
    this.onlineUsersList = document.getElementById("online-users");
    this.chatWith = document.getElementById("chat-with");
    this.messagesContainer = document.getElementById("messages");
    this.messageInput = document.getElementById("message-input");
    this.sendBtn = document.getElementById("send-btn");
    this.fileInput = document.getElementById("file-input");
    this.fileBtn = document.getElementById("file-btn");
    this.broadcastMode = document.getElementById("broadcast-mode");
    this.broadcastUsers = document.getElementById("broadcast-users");
    this.broadcastUserList = document.getElementById("broadcast-user-list");
    this.broadcastMessagesList = document.getElementById("broadcast-messages");
  }

  attachEventListeners() {
    // Auth tab switching
    this.loginTab.addEventListener("click", () => this.switchTab("login"));
    this.registerTab.addEventListener("click", () =>
      this.switchTab("register")
    );

    // Form submissions
    this.loginForm.addEventListener("submit", (e) => this.handleLogin(e));
    this.registerForm.addEventListener("submit", (e) => this.handleRegister(e));

    // Chat functionality
    this.logoutBtn.addEventListener("click", () => this.handleLogout());
    this.sendBtn.addEventListener("click", () => this.sendMessage());
    this.messageInput.addEventListener("keypress", (e) => {
      if (e.key === "Enter") this.sendMessage();
    });
    this.fileBtn.addEventListener("click", () => this.fileInput.click());
    this.fileInput.addEventListener("change", (e) => this.handleFileUpload(e));
    this.broadcastMode.addEventListener("change", () =>
      this.toggleBroadcastMode()
    );
  }

  switchTab(tab) {
    if (tab === "login") {
      this.loginTab.classList.add("active");
      this.registerTab.classList.remove("active");
      this.loginForm.classList.remove("hidden");
      this.registerForm.classList.add("hidden");
    } else {
      this.registerTab.classList.add("active");
      this.loginTab.classList.remove("active");
      this.registerForm.classList.remove("hidden");
      this.loginForm.classList.add("hidden");
    }
    this.clearAuthError();
  }

  async handleLogin(e) {
    e.preventDefault();
    const email = document.getElementById("login-email").value;
    const password = document.getElementById("login-password").value;

    try {
      const response = await this.apiCall("/api/auth/login", "POST", {
        email,
        password,
      });

      if (response.success) {
        this.token = response.data.token;
        this.currentUser = response.data.user;
        localStorage.setItem("token", this.token);
        this.showChatSection();
      } else {
        this.showAuthError(response.error.message);
      }
    } catch (error) {
      this.showAuthError("Login failed. Please try again.");
    }
  }

  async handleRegister(e) {
    e.preventDefault();
    const username = document.getElementById("register-username").value;
    const email = document.getElementById("register-email").value;
    const password = document.getElementById("register-password").value;

    try {
      const response = await this.apiCall("/api/auth/register", "POST", {
        username,
        email,
        password,
      });

      if (response.success) {
        this.token = response.data.token;
        this.currentUser = response.data.user;
        localStorage.setItem("token", this.token);
        this.showChatSection();
      } else {
        this.showAuthError(response.error.message);
      }
    } catch (error) {
      this.showAuthError("Registration failed. Please try again.");
    }
  }

  async handleLogout() {
    try {
      await this.apiCall("/api/auth/logout", "POST");
    } catch (error) {
      console.error("Logout error:", error);
    }

    this.token = null;
    this.currentUser = null;
    localStorage.removeItem("token");
    this.closeWebSocket();
    this.showAuthSection();
  }

  async checkAuthStatus() {
    if (!this.token) {
      this.showAuthSection();
      return;
    }

    try {
      const response = await this.apiCall("/api/users/me", "GET");
      if (response.success) {
        this.currentUser = response.data;
        this.showChatSection();
      } else {
        this.showAuthSection();
      }
    } catch (error) {
      this.showAuthSection();
    }
  }

  showAuthSection() {
    this.authSection.classList.remove("hidden");
    this.chatSection.classList.add("hidden");
    this.clearAuthError();
  }

  showChatSection() {
    this.authSection.classList.add("hidden");
    this.chatSection.classList.remove("hidden");
    this.currentUserSpan.textContent = this.currentUser.username;
    this.initializeChat();
  }

  async initializeChat() {
    await this.loadUsers();
    await this.loadOnlineUsers();
    await this.loadBroadcastMessages();
    this.connectWebSocket();
    this.enableMessageInput();
  }

  async loadUsers() {
    try {
      const response = await this.apiCall("/api/users", "GET");
      if (response.success) {
        this.users = response.data.filter(
          (user) => user.id !== this.currentUser.id
        );
        this.renderUsers();
      }
    } catch (error) {
      console.error("Failed to load users:", error);
    }
  }

  async loadOnlineUsers() {
    try {
      console.log("Loading online users...");
      const response = await this.apiCall("/api/users/online", "GET");
      if (response.success) {
        console.log("Online users response:", response.data);
        this.onlineUsers = response.data.filter(
          (userStatus) => userStatus.user_id !== this.currentUser.id
        );
        console.log("Filtered online users:", this.onlineUsers);
        this.renderOnlineUsers();
      }
    } catch (error) {
      console.error("Failed to load online users:", error);
    }
  }

  async loadBroadcastMessages() {
    try {
      console.log("Loading broadcast messages...");
      const response = await this.apiCall(
        "/api/messages?page=1&limit=20",
        "GET"
      );
      if (response.success) {
        // Filter only broadcast messages
        this.broadcastMessages = response.data.messages.filter(
          (message) => message.is_broadcast
        );
        console.log("Loaded broadcast messages:", this.broadcastMessages);
        this.renderBroadcastMessages();
      }
    } catch (error) {
      console.error("Failed to load broadcast messages:", error);
    }
  }

  renderUsers() {
    this.usersList.innerHTML = "";
    this.users.forEach((user) => {
      const userElement = document.createElement("div");
      userElement.className = "user-item";
      userElement.innerHTML = `
                <div class="user-status ${
                  user.is_online ? "" : "offline"
                }"></div>
                <span>${user.username}</span>
            `;
      userElement.addEventListener("click", () => this.selectUser(user));
      this.usersList.appendChild(userElement);
    });
  }

  renderOnlineUsers() {
    this.onlineUsersList.innerHTML = "";
    this.onlineUsers.forEach((userStatus) => {
      const user = this.users.find((u) => u.id === userStatus.user_id);
      if (user) {
        const userElement = document.createElement("div");
        userElement.className = "user-item";
        userElement.innerHTML = `
                    <div class="user-status"></div>
                    <span>${user.username}</span>
                `;
        this.onlineUsersList.appendChild(userElement);
      }
    });
  }

  renderBroadcastMessages() {
    this.broadcastMessagesList.innerHTML = "";
    this.broadcastMessages.slice(0, 10).forEach((message) => {
      const messageElement = document.createElement("div");
      messageElement.className = "broadcast-message-item";
      messageElement.innerHTML = `
        <div class="sender">${message.sender_username}</div>
        <div class="content">${this.escapeHtml(
          message.content.substring(0, 50)
        )}${message.content.length > 50 ? "..." : ""}</div>
        <div class="time">${this.formatTime(message.created_at)}</div>
      `;

      messageElement.addEventListener("click", () => {
        this.showBroadcastMessage(message);
      });

      this.broadcastMessagesList.appendChild(messageElement);
    });
  }

  showBroadcastMessage(message) {
    // Clear current chat and show the broadcast message
    this.selectedUser = null;
    this.chatWith.textContent = `Broadcast from ${message.sender_username}`;
    this.messagesContainer.innerHTML = "";
    this.addMessageToUI(message);
    this.scrollToBottom();

    // Update UI
    document.querySelectorAll(".user-item").forEach((item) => {
      item.classList.remove("active");
    });
  }

  selectUser(user) {
    this.selectedUser = user;
    this.chatWith.textContent = `Chat with ${user.username}`;

    // Update UI
    document.querySelectorAll(".user-item").forEach((item) => {
      item.classList.remove("active");
    });
    event.currentTarget.classList.add("active");

    // Load chat history
    this.loadChatHistory();
    this.enableMessageInput();
  }

  async loadChatHistory() {
    if (!this.selectedUser) return;

    try {
      const response = await this.apiCall(
        `/api/messages/history?user_id=${this.selectedUser.id}&page=1&limit=50`,
        "GET"
      );
      if (response.success) {
        this.renderMessages(response.data.messages);
      }
    } catch (error) {
      console.error("Failed to load chat history:", error);
    }
  }

  renderMessages(messages) {
    this.messagesContainer.innerHTML = "";
    messages.reverse().forEach((message) => {
      this.addMessageToUI(message);
    });
    this.scrollToBottom();
  }

  addMessageToUI(message) {
    // Prevent duplicate messages
    if (document.querySelector(`[data-message-id="${message.id}"]`)) {
      console.log("Message already exists, skipping:", message.id);
      return;
    }

    const messageElement = document.createElement("div");
    messageElement.setAttribute("data-message-id", message.id);
    const isOwnMessage = message.sender_id === this.currentUser.id;
    const messageClass = message.is_broadcast
      ? "broadcast"
      : isOwnMessage
      ? "sent"
      : "received";

    messageElement.className = `message ${messageClass}`;

    let content = `<div class="message-content">${this.escapeHtml(
      message.content
    )}</div>`;

    if (message.media_url) {
      content += this.renderMediaContent(message);
    }

    content += `<div class="message-time">${this.formatTime(
      message.created_at
    )}</div>`;

    if (message.is_broadcast) {
      content =
        `<div class="message-header">Broadcast from ${message.sender_username}</div>` +
        content;
    } else if (!isOwnMessage) {
      content =
        `<div class="message-header">${message.sender_username}</div>` +
        content;
    }

    messageElement.innerHTML = content;
    this.messagesContainer.appendChild(messageElement);
  }

  renderMediaContent(message) {
    if (!message.media_url) return "";

    const fileExtension = message.media_filename
      ? message.media_filename.split(".").pop().toLowerCase()
      : "";

    if (["jpg", "jpeg", "png", "gif"].includes(fileExtension)) {
      return `<div class="message-media"><img src="${message.media_url}" alt="Image" onclick="window.open('${message.media_url}', '_blank')"></div>`;
    } else if (["mp4", "avi", "mov"].includes(fileExtension)) {
      return `<div class="message-media"><video controls><source src="${message.media_url}" type="video/${fileExtension}"></video></div>`;
    } else {
      return `<div class="message-file">📎 <a href="${
        message.media_url
      }" target="_blank">${message.media_filename || "File"}</a></div>`;
    }
  }

  async sendMessage() {
    const content = this.messageInput.value.trim();
    if (!content && !this.pendingFile) return;

    let mediaUrl = null;
    if (this.pendingFile) {
      mediaUrl = await this.uploadFile(this.pendingFile);
      if (!mediaUrl) return;
    }

    const messageData = {
      content: content || "File attachment",
      message_type: mediaUrl ? this.getMessageType(this.pendingFile) : "text",
      media_url: mediaUrl,
    };

    try {
      let response;
      if (this.broadcastMode.checked) {
        // Broadcast to selected users
        if (this.selectedBroadcastUsers.length === 0) {
          alert("Please select at least one user for broadcast message.");
          return;
        }
        messageData.recipient_ids = this.selectedBroadcastUsers;
        response = await this.apiCall(
          "/api/messages/broadcast",
          "POST",
          messageData
        );
      } else {
        // Direct message
        if (!this.selectedUser) {
          alert("Please select a user to send a message to.");
          return;
        }
        messageData.recipient_id = this.selectedUser.id;
        response = await this.apiCall(
          "/api/messages/send",
          "POST",
          messageData
        );
      }

      if (response.success) {
        this.messageInput.value = "";
        this.pendingFile = null;
        this.clearFilePreview();

        // Add message to UI immediately for better UX
        this.addMessageToUI(response.data);
        this.scrollToBottom();
      }
    } catch (error) {
      console.error("Failed to send message:", error);
      alert("Failed to send message. Please try again.");
    }
  }

  async uploadFile(file) {
    const formData = new FormData();
    formData.append("file", file);

    try {
      const response = await fetch("/api/media/upload", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
        },
        body: formData,
      });

      const result = await response.json();
      if (result.success) {
        return result.data.url;
      } else {
        alert("File upload failed: " + result.error.message);
        return null;
      }
    } catch (error) {
      console.error("File upload error:", error);
      alert("File upload failed. Please try again.");
      return null;
    }
  }

  handleFileUpload(e) {
    const file = e.target.files[0];
    if (file) {
      this.pendingFile = file;
      this.showFilePreview(file);
    }
  }

  showFilePreview(file) {
    // You could add a file preview UI here
    console.log("File selected:", file.name);
  }

  clearFilePreview() {
    this.fileInput.value = "";
    // Clear any file preview UI
  }

  getMessageType(file) {
    const extension = file.name.split(".").pop().toLowerCase();
    if (["jpg", "jpeg", "png", "gif"].includes(extension)) {
      return "image";
    } else if (["mp4", "avi", "mov"].includes(extension)) {
      return "video";
    } else {
      return "file";
    }
  }

  toggleBroadcastMode() {
    if (this.broadcastMode.checked) {
      this.chatWith.textContent = "Broadcast message";
      this.messageInput.placeholder = "Type a broadcast message...";
      this.broadcastUsers.classList.remove("hidden");
      this.renderBroadcastUsers();
    } else {
      this.broadcastUsers.classList.add("hidden");
      this.selectedBroadcastUsers = [];
      if (this.selectedUser) {
        this.chatWith.textContent = `Chat with ${this.selectedUser.username}`;
        this.messageInput.placeholder = "Type a message...";
      } else {
        this.chatWith.textContent = "Select a user to start chatting";
        this.messageInput.placeholder = "Type a message...";
      }
    }
  }

  renderBroadcastUsers() {
    this.broadcastUserList.innerHTML = "";
    this.users.forEach((user) => {
      const userItem = document.createElement("div");
      userItem.className = "broadcast-user-item";
      userItem.innerHTML = `
        <input type="checkbox" id="broadcast-${user.id}" ${
        this.selectedBroadcastUsers.includes(user.id) ? "checked" : ""
      }>
        <label for="broadcast-${user.id}">${user.username}</label>
      `;

      const checkbox = userItem.querySelector('input[type="checkbox"]');
      checkbox.addEventListener("change", () => {
        this.toggleBroadcastUser(user.id);
      });

      userItem.addEventListener("click", (e) => {
        if (e.target.type !== "checkbox") {
          checkbox.checked = !checkbox.checked;
          this.toggleBroadcastUser(user.id);
        }
      });

      this.broadcastUserList.appendChild(userItem);
    });
  }

  toggleBroadcastUser(userId) {
    const index = this.selectedBroadcastUsers.indexOf(userId);
    if (index > -1) {
      this.selectedBroadcastUsers.splice(index, 1);
    } else {
      this.selectedBroadcastUsers.push(userId);
    }

    // Update UI
    const userItem = document
      .querySelector(`#broadcast-${userId}`)
      .closest(".broadcast-user-item");
    if (this.selectedBroadcastUsers.includes(userId)) {
      userItem.classList.add("selected");
    } else {
      userItem.classList.remove("selected");
    }

    // Update chat header
    if (this.selectedBroadcastUsers.length > 0) {
      this.chatWith.textContent = `Broadcast to ${this.selectedBroadcastUsers.length} user(s)`;
    } else {
      this.chatWith.textContent = "Broadcast message - select users";
    }
  }

  enableMessageInput() {
    this.messageInput.disabled = false;
    this.sendBtn.disabled = false;
  }

  connectWebSocket() {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/ws?token=${this.token}`;

    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log("WebSocket connected");
      // Reload online users when WebSocket connects
      this.loadOnlineUsers();
    };

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.handleWebSocketMessage(data);
    };

    this.ws.onclose = () => {
      console.log("WebSocket disconnected");
      // Attempt to reconnect after 3 seconds
      setTimeout(() => {
        if (this.token) {
          this.connectWebSocket();
        }
      }, 3000);
    };

    this.ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };
  }

  handleWebSocketMessage(data) {
    switch (data.type) {
      case "new_message":
        this.handleNewMessage(data.data);
        break;
      case "user_status":
        this.handleUserStatus(data.data);
        break;
      case "delivery_update":
        this.handleDeliveryUpdate(data.data);
        break;
      default:
        console.log("Unknown WebSocket message type:", data.type);
    }
  }

  handleNewMessage(message) {
    console.log("Received new message:", message);

    // Always show broadcast messages regardless of selected user
    if (message.is_broadcast) {
      // Add to broadcast messages list
      this.broadcastMessages.unshift(message);
      this.renderBroadcastMessages();

      // Show in main chat area
      this.addMessageToUI(message);
      this.scrollToBottom();
      return;
    }

    // For direct messages, check if we're in the relevant chat
    if (this.selectedUser) {
      const isRelevantToCurrentChat =
        (message.sender_id === this.selectedUser.id &&
          message.recipient_id === this.currentUser.id) ||
        (message.sender_id === this.currentUser.id &&
          message.recipient_id === this.selectedUser.id);

      if (isRelevantToCurrentChat) {
        this.addMessageToUI(message);
        this.scrollToBottom();
      }
    }
  }

  handleUserStatus(status) {
    console.log("Received user status update:", status);

    // Update online users list
    if (status.is_online) {
      if (!this.onlineUsers.find((u) => u.user_id === status.user_id)) {
        this.onlineUsers.push(status);
        console.log("Added user to online list:", status.user_id);
      }
    } else {
      this.onlineUsers = this.onlineUsers.filter(
        (u) => u.user_id !== status.user_id
      );
      console.log("Removed user from online list:", status.user_id);
    }

    // Update users list
    const user = this.users.find((u) => u.id === status.user_id);
    if (user) {
      user.is_online = status.is_online;
      this.renderUsers();
      console.log(
        "Updated user online status:",
        user.username,
        status.is_online
      );
    }

    this.renderOnlineUsers();
    console.log("Current online users:", this.onlineUsers.length);
  }

  handleDeliveryUpdate(update) {
    // You could update message delivery status in the UI here
    console.log("Delivery update:", update);
  }

  closeWebSocket() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  async apiCall(endpoint, method, data = null) {
    const options = {
      method,
      headers: {
        "Content-Type": "application/json",
      },
    };

    if (this.token) {
      options.headers["Authorization"] = `Bearer ${this.token}`;
    }

    if (data) {
      options.body = JSON.stringify(data);
    }

    const response = await fetch(endpoint, options);
    return await response.json();
  }

  showAuthError(message) {
    this.authError.textContent = message;
    this.authError.style.display = "block";
  }

  clearAuthError() {
    this.authError.textContent = "";
    this.authError.style.display = "none";
  }

  scrollToBottom() {
    this.messagesContainer.scrollTop = this.messagesContainer.scrollHeight;
  }

  formatTime(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  }

  escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
  }
}

// Initialize the app when the DOM is loaded
document.addEventListener("DOMContentLoaded", () => {
  new ChatApp();
});
