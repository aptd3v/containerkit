# 🚀 WordPress Multi-Service Scale Example

This example shows how to use **go-contain** to programmatically build, deploy, and manage a scaled WordPress environment — all with native Go code.

---

## 🛠️ What It Does

* 📦 Creates a Docker Compose project with:

  * 1 MySQL database container
  * Multiple individual WordPress containers (scaled separately by `NumWordPress`)
  * An HAProxy container for load balancing
  * Optional Portainer container (skipped on Windows)
* ⚙️ Dynamically generates HAProxy config for round-robin load balancing
* 📝 Exports a `docker-compose.yaml` compatible with Docker Compose CLI
* ▶️ Runs `docker compose up` in detached mode with:

  * Orphan container cleanup
  * Custom colored log output
* 📡 Streams container logs and supports graceful shutdown on Ctrl+C
* 🧹 Runs `docker compose down` on exit, cleaning containers, images, volumes, and orphans

---

## 💡 Key Highlights

* 📌 **Declarative & reusable** container/service definitions in Go functions
* 🔄 **Dynamic scaling** with distinct WordPress services chained by dependencies
* 🩺 Built-in **health checks** for critical services like MySQL and WordPress
* 🧩 **Conditional logic** for cross-platform compatibility
* ⚡ **Graceful lifecycle management** with Go contexts and OS signal handling
* 🎨 **Custom logger** adds colored action prefixes (`[up]`, `[logs]`, `[down]`) for clarity

---

## ▶️ How to Run

*In project root directory.*
```bash
go run ./examples/wordpress/main.go
```

* Exports `docker-compose.yaml` to `./examples/wordpress/`
* Starts all containers and streams logs live (with custom logger)
* Press **Ctrl+C** to stop and clean up all resources automatically

## 📝 Note
The `wordpress.Up` function does not require exporting the Compose project to a YAML file beforehand. You can run and manage your multi-container application entirely in-memory and programmatically, giving you full dynamic control without ever writing YAML to disk.

Exporting YAML is optional and primarily for compatibility, sharing, or debugging purposes.

---

This example proves that **go-contain** isn’t just for simple apps — it can handle complex, scaled multi-service setups with full programmatic control, dynamic config generation, and smooth lifecycle orchestration — all in idiomatic Go! 🎉

---
