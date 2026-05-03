import { useEffect, useState } from "react";
import type { FormEvent } from "react";

type Library = {
  id: number;
  name: string;
  folder_path: string;
};

type MessageResponse = {
  message: string;
};

const apiBase = "http://localhost:8080";

const spotlight = {
  title: "Paykodec",
  subtitle: "Self-hosted media hub",
  description: "Organize and stream own media collection.",
};

const staticRows = [
  { id: 1, title: "Trending", tags: ["Sci-Fi", "Action", "Neo Noir"] },
  { id: 2, title: "Recommended For You", tags: ["Drama", "Mystery", "Indie"] },
  { id: 3, title: "Keep Watching", tags: ["Comedy", "Anime", "Documentary"] },
];

async function parseError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as MessageResponse;
    return body.message || `request failed (${res.status})`;
  } catch {
    return `request failed (${res.status})`;
  }
}

export default function App() {
  const [libraries, setLibraries] = useState<Library[]>([]);
  const [selectedLibrary, setSelectedLibrary] = useState<Library | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detailError, setDetailError] = useState("");
  const [name, setName] = useState("");
  const [folderPath, setFolderPath] = useState("");
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

  async function loadLibraries() {
    setLoading(true);
    setError("");
    try {
      const res = await fetch(`${apiBase}/libraries`);
      if (!res.ok) {
        setError(await parseError(res));
        return;
      }

      const data = (await res.json()) as Library[];
      setLibraries(data);
    } catch {
      setError("Cannot connect to backend. Start Go server on :8080.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadLibraries();
  }, []);

  async function onCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSubmitting(true);
    setNotice("");
    setError("");

    try {
      const res = await fetch(`${apiBase}/libraries`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name,
          folder_path: folderPath,
        }),
      });

      if (!res.ok) {
        setError(await parseError(res));
        return;
      }

      setName("");
      setFolderPath("");
      setNotice("Library created.");
      await loadLibraries();
    } catch {
      setError("Cannot connect to backend. Start Go server on :8080.");
    } finally {
      setSubmitting(false);
    }
  }

  async function onDelete(id: number) {
    setNotice("");
    setError("");
    try {
      const res = await fetch(`${apiBase}/libraries/${id}`, {
        method: "DELETE",
      });

      if (!res.ok) {
        setError(await parseError(res));
        return;
      }

      setNotice("Library deleted.");
      await loadLibraries();
    } catch {
      setError("Cannot connect to backend. Start Go server on :8080.");
    }
  }

  async function openLibraryDetail(id: number) {
    setDetailLoading(true);
    setDetailError("");
    setSelectedLibrary(null);

    try {
      const res = await fetch(`${apiBase}/libraries/${id}`);
      if (!res.ok) {
        setDetailError(await parseError(res));
        return;
      }

      const data = (await res.json()) as Library;
      setSelectedLibrary(data);
    } catch {
      setDetailError("Cannot load library details right now.");
    } finally {
      setDetailLoading(false);
    }
  }

  function closeLibraryDetail() {
    setSelectedLibrary(null);
    setDetailError("");
    setDetailLoading(false);
  }

  return (
    <div className="page">
      <header className="topbar">
        <div className="brand">
          <span className="brand-dot" />
          <span className="brand-text">PAYKODEC</span>
        </div>
        <div className="status">
          API: {loading ? "Loading..." : "Connected"}
        </div>
      </header>

      <section className="hero">
        <div className="hero-overlay" />
        <div className="hero-content">
          <p className="eyebrow">{spotlight.subtitle}</p>
          <h1>{spotlight.title}</h1>
          <p className="hero-description">{spotlight.description}</p>
        </div>
      </section>

      <main className="content">
        <section className="panel">
          <h2>Library Manager</h2>
          <form className="library-form" onSubmit={onCreate}>
            <label>
              Name
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Movies"
                required
              />
            </label>
            <label>
              Folder Path
              <input
                value={folderPath}
                onChange={(e) => setFolderPath(e.target.value)}
                placeholder="D:/media/movies"
                required
              />
            </label>
            <button type="submit" disabled={submitting}>
              {submitting ? "Creating..." : "Add Library"}
            </button>
          </form>

          {notice && <p className="notice">{notice}</p>}
          {error && <p className="error">{error}</p>}
        </section>

        <section className="rows">
          {staticRows.map((row) => (
            <article key={row.id} className="row-block">
              <h3>{row.title}</h3>
              <div className="cards">
                {libraries.length === 0 && !loading ? (
                  <div className="empty-card">No libraries yet.</div>
                ) : (
                  libraries.map((lib) => (
                    <div className="media-card" key={`${row.id}-${lib.id}`}>
                      <button
                        className="media-poster media-open-btn"
                        type="button"
                        onClick={() => openLibraryDetail(lib.id)}
                        aria-label={`Open ${lib.name} details`}
                      />
                      <div className="media-meta">
                        <strong>{lib.name}</strong>
                        <span>{lib.folder_path}</span>
                        <small>{row.tags[lib.id % row.tags.length]}</small>
                        <button
                          className="delete-btn"
                          type="button"
                          onClick={() => onDelete(lib.id)}
                        >
                          Delete
                        </button>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </article>
          ))}
        </section>
      </main>

      {(detailLoading || detailError || selectedLibrary) && (
        <div className="modal-backdrop" role="presentation" onClick={closeLibraryDetail}>
          <section
            className="modal"
            role="dialog"
            aria-modal="true"
            aria-label="Library detail"
            onClick={(e) => e.stopPropagation()}
          >
            <button type="button" className="modal-close" onClick={closeLibraryDetail}>
              x
            </button>
            {detailLoading && <p>Loading details...</p>}
            {!detailLoading && detailError && <p className="error">{detailError}</p>}
            {!detailLoading && selectedLibrary && (
              <>
                <h3>{selectedLibrary.name}</h3>
                <p>
                  <strong>ID:</strong> {selectedLibrary.id}
                </p>
                <p>
                  <strong>Folder:</strong> {selectedLibrary.folder_path}
                </p>
              </>
            )}
          </section>
        </div>
      )}
    </div>
  );
}
