import { useEffect, useState } from "react";
import type { FormEvent } from "react";

type Library = {
  id: number;
  name: string;
  folder_path: string;
};

type MediaItem = {
  id: number;
  library_id: number;
  title: string;
  file_path: string;
  container: string;
  duration_seconds: number;
};

type MessageResponse = {
  message: string;
};

type AuthResponse = {
  token: string;
  expires_at: number;
  token_type: string;
  username: string;
  is_admin: boolean;
};

type UserSummary = {
  id: number;
  username: string;
  is_admin: boolean;
  created_at: string;
};

const apiBase = "http://localhost:8080";
const authTokenStorageKey = "paykodec_auth_token";
const authUserStorageKey = "paykodec_auth_user";
const authAdminStorageKey = "paykodec_auth_is_admin";

async function parseError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as MessageResponse;
    return body.message || `request failed (${res.status})`;
  } catch {
    return `request failed (${res.status})`;
  }
}

function parseRoute(pathname: string): { page: "home" | "settings" | "dashboard" | "library"; libraryID?: number } {
  if (pathname === "/settings") {
    return { page: "settings" };
  }
  if (pathname === "/dashboard") {
    return { page: "dashboard" };
  }
  const libMatch = pathname.match(/^\/libraries\/(\d+)$/);
  if (libMatch) {
    return { page: "library", libraryID: Number(libMatch[1]) };
  }
  return { page: "home" };
}

function navigate(path: string) {
  if (window.location.pathname !== path) {
    window.history.pushState({}, "", path);
    window.dispatchEvent(new PopStateEvent("popstate"));
  }
}

export default function App() {
  const [token, setToken] = useState(() => localStorage.getItem(authTokenStorageKey) ?? "");
  const [username, setUsername] = useState(() => localStorage.getItem(authUserStorageKey) ?? "");
  const [isAdmin, setIsAdmin] = useState(() => localStorage.getItem(authAdminStorageKey) === "true");
  const [loginID, setLoginID] = useState("admin");
  const [loginPassword, setLoginPassword] = useState("admin");
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginError, setLoginError] = useState("");

  const [route, setRoute] = useState(() => parseRoute(window.location.pathname));
  const [settingsSection, setSettingsSection] = useState<"account" | "library" | "display">("account");
  const [adminSection, setAdminSection] = useState<"users" | "libraries" | "metadata" | "general">("users");

  const [libraries, setLibraries] = useState<Library[]>([]);
  const [loadingLibraries, setLoadingLibraries] = useState(false);
  const [libraryError, setLibraryError] = useState("");
  const [libraryNotice, setLibraryNotice] = useState("");

  const [selectedLibrary, setSelectedLibrary] = useState<Library | null>(null);
  const [mediaItems, setMediaItems] = useState<MediaItem[]>([]);
  const [loadingMediaItems, setLoadingMediaItems] = useState(false);
  const [mediaError, setMediaError] = useState("");
  const [mediaNotice, setMediaNotice] = useState("");

  const [name, setName] = useState("");
  const [folderPath, setFolderPath] = useState("");

  const [newPassword, setNewPassword] = useState("");
  const [currentPassword, setCurrentPassword] = useState("");
  const [accountNotice, setAccountNotice] = useState("");
  const [accountError, setAccountError] = useState("");

  const [adminNotice, setAdminNotice] = useState("");
  const [adminError, setAdminError] = useState("");
  const [users, setUsers] = useState<UserSummary[]>([]);
  const [loadingUsers, setLoadingUsers] = useState(false);
  const [newUserName, setNewUserName] = useState("");
  const [newUserPassword, setNewUserPassword] = useState("");

  function authHeaders(base?: HeadersInit): HeadersInit {
    const headers: Record<string, string> = {
      ...(typeof base === "object" && base ? (base as Record<string, string>) : {}),
    };
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    return headers;
  }

  function clearMessages() {
    setLibraryError("");
    setLibraryNotice("");
    setMediaError("");
    setMediaNotice("");
    setAccountNotice("");
    setAccountError("");
    setAdminNotice("");
    setAdminError("");
  }

  function saveAuth(nextToken: string, nextUsername: string, nextIsAdmin: boolean) {
    setToken(nextToken);
    setUsername(nextUsername);
    setIsAdmin(nextIsAdmin);
    localStorage.setItem(authTokenStorageKey, nextToken);
    localStorage.setItem(authUserStorageKey, nextUsername);
    localStorage.setItem(authAdminStorageKey, String(nextIsAdmin));
  }

  function clearAuth() {
    setToken("");
    setUsername("");
    setIsAdmin(false);
    localStorage.removeItem(authTokenStorageKey);
    localStorage.removeItem(authUserStorageKey);
    localStorage.removeItem(authAdminStorageKey);
    navigate("/");
    clearMessages();
  }

  async function loadLibraries() {
    setLoadingLibraries(true);
    setLibraryError("");
    try {
      const res = await fetch(`${apiBase}/libraries`);
      if (!res.ok) {
        setLibraryError(await parseError(res));
        return;
      }
      const data = (await res.json()) as Library[];
      setLibraries(data);
    } catch {
      setLibraryError("Cannot connect to backend. Start Go server on :8080.");
    } finally {
      setLoadingLibraries(false);
    }
  }

  async function loadMediaItems(libraryID: number) {
    setLoadingMediaItems(true);
    setMediaError("");
    setMediaItems([]);
    try {
      const res = await fetch(`${apiBase}/libraries/${libraryID}/media`);
      if (!res.ok) {
        setMediaError(await parseError(res));
        return;
      }
      const data = (await res.json()) as MediaItem[];
      setMediaItems(data);
    } catch {
      setMediaError("Cannot connect to backend. Start Go server on :8080.");
    } finally {
      setLoadingMediaItems(false);
    }
  }

  async function onLogin(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoginLoading(true);
    setLoginError("");
    try {
      const res = await fetch(`${apiBase}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username: loginID, password: loginPassword }),
      });
      if (!res.ok) {
        setLoginError(await parseError(res));
        return;
      }
      const body = (await res.json()) as AuthResponse;
      saveAuth(body.token, body.username, body.is_admin);
      setLoginPassword("");
      navigate("/");
    } catch {
      setLoginError("Cannot connect to backend. Start Go server on :8080.");
    } finally {
      setLoginLoading(false);
    }
  }

  async function onChangePassword(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setAccountError("");
    setAccountNotice("");
    try {
      const res = await fetch(`${apiBase}/me/password`, {
        method: "PATCH",
        headers: authHeaders({ "Content-Type": "application/json" }),
        body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
      });
      if (!res.ok) {
        setAccountError(await parseError(res));
        return;
      }
      setAccountNotice("Password updated.");
      setCurrentPassword("");
      setNewPassword("");
    } catch {
      setAccountError("Cannot connect to backend. Start Go server on :8080.");
    }
  }

  async function loadUsers() {
    if (!isAdmin) {
      return;
    }
    setLoadingUsers(true);
    setAdminError("");
    try {
      const res = await fetch(`${apiBase}/admin/users`, { headers: authHeaders() });
      if (!res.ok) {
        setAdminError(await parseError(res));
        return;
      }
      const data = (await res.json()) as UserSummary[];
      setUsers(data);
    } catch {
      setAdminError("Cannot connect to backend. Start Go server on :8080.");
    } finally {
      setLoadingUsers(false);
    }
  }

  async function onCreateUserByAdmin(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setAdminError("");
    setAdminNotice("");
    try {
      const res = await fetch(`${apiBase}/admin/users`, {
        method: "POST",
        headers: authHeaders({ "Content-Type": "application/json" }),
        body: JSON.stringify({ username: newUserName, password: newUserPassword }),
      });
      if (!res.ok) {
        setAdminError(await parseError(res));
        return;
      }
      setAdminNotice("General user created.");
      setNewUserName("");
      setNewUserPassword("");
      await loadUsers();
    } catch {
      setAdminError("Cannot connect to backend. Start Go server on :8080.");
    }
  }

  async function onSetAdmin(userID: number, nextIsAdmin: boolean) {
    setAdminError("");
    setAdminNotice("");
    try {
      const res = await fetch(`${apiBase}/admin/users/${userID}/admin`, {
        method: "PATCH",
        headers: authHeaders({ "Content-Type": "application/json" }),
        body: JSON.stringify({ is_admin: nextIsAdmin }),
      });
      if (!res.ok) {
        setAdminError(await parseError(res));
        return;
      }
      setAdminNotice("User role updated.");
      await loadUsers();
    } catch {
      setAdminError("Cannot connect to backend. Start Go server on :8080.");
    }
  }

  async function onDeleteUser(userID: number) {
    if (!window.confirm("정말로 이 계정을 삭제하시겠습니까?")) {
      return;
    }
    setAdminError("");
    setAdminNotice("");
    try {
      const res = await fetch(`${apiBase}/admin/users/${userID}`, {
        method: "DELETE",
        headers: authHeaders(),
      });
      if (!res.ok) {
        setAdminError(await parseError(res));
        return;
      }
      setAdminNotice("User deleted.");
      await loadUsers();
    } catch {
      setAdminError("Cannot connect to backend. Start Go server on :8080.");
    }
  }

  async function onCreateLibraryByAdmin(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLibraryError("");
    setLibraryNotice("");
    try {
      const res = await fetch(`${apiBase}/libraries`, {
        method: "POST",
        headers: authHeaders({ "Content-Type": "application/json" }),
        body: JSON.stringify({ name, folder_path: folderPath }),
      });
      if (!res.ok) {
        setLibraryError(await parseError(res));
        return;
      }
      setLibraryNotice("Library created.");
      setName("");
      setFolderPath("");
      await loadLibraries();
    } catch {
      setLibraryError("Cannot connect to backend. Start Go server on :8080.");
    }
  }

  async function onDeleteLibraryByAdmin(id: number) {
    if (!window.confirm("정말로 이 라이브러리를 삭제하시겠습니까?")) {
      return;
    }
    setLibraryError("");
    setLibraryNotice("");
    try {
      const res = await fetch(`${apiBase}/libraries/${id}`, {
        method: "DELETE",
        headers: authHeaders(),
      });
      if (!res.ok) {
        setLibraryError(await parseError(res));
        return;
      }
      setLibraryNotice("Library deleted.");
      await loadLibraries();
    } catch {
      setLibraryError("Cannot connect to backend. Start Go server on :8080.");
    }
  }

  useEffect(() => {
    void loadLibraries();
  }, []);

  useEffect(() => {
    const onPop = () => setRoute(parseRoute(window.location.pathname));
    window.addEventListener("popstate", onPop);
    return () => window.removeEventListener("popstate", onPop);
  }, []);

  useEffect(() => {
    setRoute(parseRoute(window.location.pathname));
  }, [token]);

  useEffect(() => {
    clearMessages();
  }, [route.page, route.libraryID, settingsSection, adminSection]);

  useEffect(() => {
    if (route.page === "dashboard" && isAdmin) {
      void loadUsers();
    }
  }, [route.page, isAdmin]);

  useEffect(() => {
    if (route.page !== "library") {
      return;
    }
    const targetID = route.libraryID;
    if (!targetID) {
      return;
    }
    const target = libraries.find((lib) => lib.id === targetID);
    if (target) {
      setSelectedLibrary(target);
      void loadMediaItems(targetID);
    }
  }, [route.page, route.libraryID, libraries]);

  if (!token) {
    return (
      <div className="login-page">
        <div className="login-brand">
          <span className="brand-dot" />
          <span className="brand-text">PAYKODEC</span>
        </div>
        <form className="login-card" onSubmit={onLogin}>
          <h1>Login</h1>
          <label>
            ID
            <input value={loginID} onChange={(e) => setLoginID(e.target.value)} required />
          </label>
          <label>
            Password
            <input type="password" value={loginPassword} onChange={(e) => setLoginPassword(e.target.value)} required />
          </label>
          <button type="submit" disabled={loginLoading}>
            {loginLoading ? "Signing in..." : "Sign In"}
          </button>
          {loginError && <p className="error">{loginError}</p>}
        </form>
      </div>
    );
  }

  if (route.page === "dashboard" && !isAdmin) {
    navigate("/");
    return null;
  }

  return (
    <div className="page">
      <header className="topbar">
        <button type="button" className="brand brand-btn" onClick={() => navigate("/")}>
          <span className="brand-dot" />
          <span className="brand-text">PAYKODEC</span>
        </button>
        <div className="toolbar">
          <span className="login-user">{username}</span>
          <button type="button" className="nav-btn" onClick={() => navigate("/settings")}>
            Settings
          </button>
          {isAdmin && (
            <button type="button" className="nav-btn" onClick={() => navigate("/dashboard")}>
              Dashboard
            </button>
          )}
          <button type="button" className="nav-btn" onClick={clearAuth}>
            Logout
          </button>
        </div>
      </header>

      {route.page === "home" && (
        <main className="home-page">
          <section className="home-hero panel">
            <h2>Libraries</h2>
            <p>Your registered media libraries</p>
          </section>
          {loadingLibraries ? (
            <p>Loading...</p>
          ) : libraries.length === 0 ? (
            <section className="panel">
              <p>No libraries yet.</p>
            </section>
          ) : (
            <section className="library-grid">
              {libraries.map((lib) => (
                <button key={lib.id} type="button" className="library-card" onClick={() => navigate(`/libraries/${lib.id}`)}>
                  <span className="card-poster" />
                  <span className="card-title">{lib.name}</span>
                  <span className="card-subtitle">{lib.folder_path}</span>
                </button>
              ))}
            </section>
          )}
          {libraryError && <p className="error">{libraryError}</p>}
        </main>
      )}

      {route.page === "library" && (
        <main className="library-detail-page">
          <section className="panel library-detail-header">
            <h2>{selectedLibrary ? selectedLibrary.name : "Library"}</h2>
            <p>{selectedLibrary ? selectedLibrary.folder_path : ""}</p>
          </section>
          <section className="media-grid">
            {loadingMediaItems ? (
              <p>Loading media files...</p>
            ) : mediaItems.length === 0 ? (
              <p>No media files in this library yet.</p>
            ) : (
              mediaItems.map((item) => (
                <button
                  key={item.id}
                  type="button"
                  className="media-card media-open-btn"
                  onClick={() => setMediaNotice("재생 기능은 다음 단계에서 구현합니다.")}
                >
                  <span className="card-poster" />
                  <strong>{item.title}</strong>
                  <p>{item.file_path}</p>
                </button>
              ))
            )}
          </section>
          {mediaNotice && <p className="notice">{mediaNotice}</p>}
          {mediaError && <p className="error">{mediaError}</p>}
        </main>
      )}

      {route.page === "settings" && (
        <main className="settings-layout">
          <aside className="settings-sidebar panel">
            <h2>Settings</h2>
            <button
              type="button"
              className={settingsSection === "account" ? "nav-btn active" : "nav-btn"}
              onClick={() => setSettingsSection("account")}
            >
              Account
            </button>
            <button
              type="button"
              className={settingsSection === "library" ? "nav-btn active" : "nav-btn"}
              onClick={() => setSettingsSection("library")}
            >
              Library
            </button>
            <button
              type="button"
              className={settingsSection === "display" ? "nav-btn active" : "nav-btn"}
              onClick={() => setSettingsSection("display")}
            >
              Display
            </button>
          </aside>

          <section className="settings-main">
            {settingsSection === "account" && (
              <section className="panel">
                <h2>Change Password</h2>
                <form className="library-form" onSubmit={onChangePassword}>
                  <label>
                    Current Password
                    <input type="password" value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)} required />
                  </label>
                  <label>
                    New Password
                    <input type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} required />
                  </label>
                  <button type="submit">Update Password</button>
                </form>
                {accountNotice && <p className="notice">{accountNotice}</p>}
                {accountError && <p className="error">{accountError}</p>}
              </section>
            )}

            {settingsSection === "library" && (
              <section className="panel">
                <h2>Library Section</h2>
                <p>Library settings UI is reserved for next step.</p>
              </section>
            )}

            {settingsSection === "display" && (
              <section className="panel">
                <h2>Display Section</h2>
                <p>Display settings UI is reserved for next step.</p>
              </section>
            )}
          </section>
        </main>
      )}

      {route.page === "dashboard" && isAdmin && (
        <main className="settings-layout">
          <aside className="settings-sidebar panel">
            <h2>Dashboard</h2>
            <button
              type="button"
              className={adminSection === "users" ? "nav-btn active" : "nav-btn"}
              onClick={() => setAdminSection("users")}
            >
              User Management
            </button>
            <button
              type="button"
              className={adminSection === "libraries" ? "nav-btn active" : "nav-btn"}
              onClick={() => setAdminSection("libraries")}
            >
              Libraries
            </button>
            <button
              type="button"
              className={adminSection === "metadata" ? "nav-btn active" : "nav-btn"}
              onClick={() => setAdminSection("metadata")}
            >
              Metadata
            </button>
            <button
              type="button"
              className={adminSection === "general" ? "nav-btn active" : "nav-btn"}
              onClick={() => setAdminSection("general")}
            >
              General Settings
            </button>
          </aside>

          <section className="settings-main">
            {adminSection === "users" && (
              <section className="panel">
                <h2>User Management</h2>
                {loadingUsers ? (
                  <p>Loading users...</p>
                ) : (
                  <>
                    <form className="library-form" onSubmit={onCreateUserByAdmin}>
                      <label>
                        New User ID
                        <input value={newUserName} onChange={(e) => setNewUserName(e.target.value)} required />
                      </label>
                      <label>
                        New User Password
                        <input type="password" value={newUserPassword} onChange={(e) => setNewUserPassword(e.target.value)} required />
                      </label>
                      <button type="submit">Create General User</button>
                    </form>
                    <ul className="user-list">
                      {users.map((user) => (
                        <li key={user.id}>
                          <div>
                            <strong>
                              {user.username} #{user.id}
                            </strong>
                            <p>{user.is_admin ? "Admin" : "General"}</p>
                          </div>
                          <div className="user-actions">
                            <button type="button" onClick={() => void onSetAdmin(user.id, !user.is_admin)}>
                              {user.is_admin ? "Remove Admin" : "Make Admin"}
                            </button>
                            <button type="button" className="delete-btn" onClick={() => void onDeleteUser(user.id)}>
                              Delete User
                            </button>
                          </div>
                        </li>
                      ))}
                    </ul>
                  </>
                )}
                {adminNotice && <p className="notice">{adminNotice}</p>}
                {adminError && <p className="error">{adminError}</p>}
              </section>
            )}

            {adminSection === "libraries" && (
              <section className="panel">
                <h2>Library Management</h2>
                <form className="library-form" onSubmit={onCreateLibraryByAdmin}>
                  <label>
                    Name
                    <input value={name} onChange={(e) => setName(e.target.value)} required />
                  </label>
                  <label>
                    Folder Path
                    <input value={folderPath} onChange={(e) => setFolderPath(e.target.value)} required />
                  </label>
                  <button type="submit">Add Library</button>
                </form>
                <ul className="library-list">
                  {libraries.map((lib) => (
                    <li key={lib.id}>
                      <div>
                        <strong>{lib.name}</strong>
                        <p>{lib.folder_path}</p>
                      </div>
                      <button type="button" className="delete-btn" onClick={() => void onDeleteLibraryByAdmin(lib.id)}>
                        Delete
                      </button>
                    </li>
                  ))}
                </ul>
                {libraryNotice && <p className="notice">{libraryNotice}</p>}
                {libraryError && <p className="error">{libraryError}</p>}
              </section>
            )}

            {adminSection === "metadata" && (
              <section className="panel">
                <h2>Metadata</h2>
                <p>Metadata settings UI is reserved for next step.</p>
              </section>
            )}

            {adminSection === "general" && (
              <section className="panel">
                <h2>General Settings</h2>
                <p>General settings UI is reserved for next step.</p>
              </section>
            )}
          </section>
        </main>
      )}
    </div>
  );
}
