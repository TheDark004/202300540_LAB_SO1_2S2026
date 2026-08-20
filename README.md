# 202300540_LAB_SO1_2S2026

Proyecto 1 – Sistemas Operativos 1 (USAC)
**Desarrollo, Conexión y Gestión de Contenedores en Entornos Virtualizados**

Entorno virtualizado con 3 VMs (KVM), cada una con un runtime de contenedores distinto (Containerd, Podman, Docker), 3 APIs REST en Go comunicándose entre sí, y un registro privado de imágenes (Zot).

**Estudiante:** René Sebastian Gutiérrez Contreras
**Carnet:** 202300540

---

## Estructura del repositorio

```
.
├── api1/           # API1 — Go + Fiber, corre en VM1 (Containerd)
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── api2/           # API2 — Go + Fiber, corre en VM1 (Containerd)
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── api3/           # API3 — Go + Fiber, corre en VM2 (Podman)
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
└── docs/           # Documentación (no versionada en git, ver .gitignore)
    ├── MANUAL_TECNICO.md
    └── capturas/
```

## Arquitectura

| VM  | Runtime    | Contenedor(es)                 | IP                |
| --- | ---------- | ------------------------------ | ----------------- |
| VM1 | Containerd | API1 (`:8081`), API2 (`:8082`) | `192.168.122.227` |
| VM2 | Podman     | API3 (`:8083`)                 | `192.168.122.245` |
| VM3 | Docker     | Zot – registro (`:5000`)       | `192.168.122.201` |

## Guía rápida de instalación

1. **Clonar el repo** dentro de cada VM:

   ```bash
   git clone https://github.com/TheDark004/202300540_LAB_SO1_2S2026.git
   ```

2. **Construir y levantar cada API** en su VM correspondiente:

   **VM1 (Containerd/nerdctl):**

   ```bash
   cd api1 && sudo nerdctl build -t api1-202300540 . && sudo nerdctl run -d --name api1 -p 8081:8081 api1-202300540
   cd ../api2 && sudo nerdctl build -t api2-202300540 . && sudo nerdctl run -d --name api2 -p 8082:8082 api2-202300540
   ```

   **VM2 (Podman):**

   ```bash
   cd api3 && podman build -t api3-202300540 . && podman run -d --name api3 -p 8083:8083 api3-202300540
   ```

   **VM3 (Docker + Zot):**

   ```bash
   docker run -d --name zot -p 5000:5000 -v ~/zot/data:/var/lib/registry ghcr.io/project-zot/zot-linux-amd64:latest
   ```

3. **Probar salud de cada API:**

   ```bash
   curl http://<IP_VM>:<PUERTO>/health
   ```

4. **Probar comunicación cruzada** (ejemplo desde VM1):
   ```bash
   curl http://localhost:8081/api1/202300540/call-api2
   curl http://localhost:8081/api1/202300540/call-api3
   ```

Para la guía detallada paso a paso (incluyendo instalación de cada VM desde cero, configuración de dependencias como buildkit/CNI/plugins, y el proceso de push/pull al registro Zot), ver **[docs/MANUAL_TECNICO.md](docs/MANUAL_TECNICO.md)**.

## Endpoints por API

Cada API expone:

- `GET /health` — estado propio.
- `GET /api#/{carnet}/call-api#` — verifica el estado de otra API mediante una llamada HTTP interna a su `/health`.

Ver el detalle completo del formato JSON de cada respuesta en el manual técnico.

## Colaboradores

- @JoseLorenzana272
- @KINGR0X
