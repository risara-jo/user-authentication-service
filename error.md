
---

## **1️⃣ Typo in Dockerfile**

```dockerfile
COPY go.su[m] ./
```

* **Error:** `go.su[m]` is not a valid file. It should be `go.sum`.
* **Impact:** Dependencies may not download properly → build may fail.

---

## **2️⃣ Air not detecting code changes (live reload issue)**

**Causes:**

1. `.air.toml` may not be copied inside container or not in `/app`.
2. `root = "."` must point to the container working directory `/app`.
3. `poll = false` (default) prevents Air from detecting file changes inside Docker volumes on some systems.

**Impact:** Adding new routes or changing code doesn’t reload until you restart the container.

**Fix:**

```toml
poll = true
poll_interval = 500
```

and ensure `.air.toml` is in `/app`.

---

## **3️⃣ `depends_on` with `condition: service_healthy` in Compose**

```yaml
depends_on:
  postgres:
    condition: service_healthy
```

* **Error:** `condition: service_healthy` is **not supported in v3+ docker-compose**.
* **Impact:** Go container may start before Postgres is ready → connection errors.

**Fix:** Use a **wait-for-it script** in Go container before starting the app.

---

## **4️⃣ Air exclude/include misconfiguration**

* `.air.toml` has many include/exclude directives:

```toml
exclude_dir = ["assets", "tmp", "vendor", "testdata"]
include_ext = ["go", "tpl", "tmpl", "html"]
```

* **Potential problem:**

  * New `.go` files in directories not watched (or wrongly excluded) may not trigger reload.
* **Fix:** Simplify to:

```toml
exclude_dir = ["tmp", "vendor"]
include_ext = ["go"]
```

---

## **5️⃣ Docker volume mount may not propagate changes**

```yaml
volumes:
  - .:/app
```

* On **Linux/macOS**, inotify may not detect changes → Air won’t reload.
* **Fix:** Use `poll = true` in `.air.toml`.

---

## **6️⃣ `.air.toml` complexity**

* Current config is **very verbose**, many unused settings (`exclude_regex`, `rerun_delay`, `kill_delay`, etc.).
* **Impact:** Harder to debug. Some options like `exclude_unchanged = false` may prevent expected rebuild behavior.
* **Fix:** Minimal config works better in Docker.

---

## **7️⃣ Dockerfile.dev CMD**

```dockerfile
CMD ["air", "-c", ".air.toml"]
```

* **Potential issue:**

  * If `.air.toml` is missing in container or not at working directory, Air falls back to default config → may not watch correct folders.

---

## **8️⃣ Postgres migrations only run once**

```yaml
volumes:
  - ./migrations:/docker-entrypoint-initdb.d
```

* **Impact:** Only executes on first container start. Later changes to migration files are **not applied automatically**.

---

## **9️⃣ Docker Compose network naming**

```yaml
networks:
  app-network:
    driver: bridge
```

* Not an error, but **custom networks require correct hostnames**. You already use `DB_HOST=postgres` → correct.

---

## **10️⃣ Container rebuild required for changes in Dockerfile or Air**

* If you change `.air.toml` or Dockerfile, you must:

```bash
docker-compose down --volumes --remove-orphans
docker-compose up --build
```

* **Impact:** Otherwise, Air may still run old config.

---

### **Summary of Key Errors**

| #  | Issue                               | Impact                               |
| -- | ----------------------------------- | ------------------------------------ |
| 1  | Typo `go.su[m]`                     | Dependencies may fail                |
| 2  | Air not detecting code              | Hot reload doesn’t work              |
| 3  | `depends_on` with `service_healthy` | Go may start before DB               |
| 4  | Air exclude/include misconfig       | Some files not watched               |
| 5  | Docker volume mount inotify         | Changes not detected on some OS      |
| 6  | `.air.toml` verbose                 | Hard to debug, may override defaults |
| 7  | CMD points to missing `.air.toml`   | Air fallback config used             |
| 8  | Postgres migrations only run once   | Later migrations ignored             |
| 9  | Network config minor                | Only matters if wrong hostnames      |
| 10 | Docker rebuild required             | Changes to Dockerfile or Air ignored |

---
