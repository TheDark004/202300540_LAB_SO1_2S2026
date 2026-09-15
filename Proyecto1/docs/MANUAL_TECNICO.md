# Manual Técnico y Guía de Instalación

**Proyecto 1 – Sistemas Operativos 1**
Desarrollo, Conexión y Gestión de Contenedores en Entornos Virtualizados

**Estudiante:** René Sebastian Gutiérrez Contreras
**Carnet:** 202300540
**Universidad de San Carlos de Guatemala – Facultad de Ingeniería**
**Fecha:** 20/08/2026

---

## Índice

1. [Introducción](#1-introducción)
2. [Arquitectura General](#2-arquitectura-general)
3. [VM1 – Containerd (API1 y API2)](#3-vm1--containerd-api1-y-api2)
4. [VM2 – Podman (API3)](#4-vm2--podman-api3)
5. [VM3 – Docker y Registro Zot](#5-vm3--docker-y-registro-zot)
6. [Comunicación entre APIs](#6-comunicación-entre-apis)
7. [Registro Zot: Push y Pull de Imágenes](#7-registro-zot-push-y-pull-de-imágenes)
8. [Pruebas Funcionales](#8-pruebas-funcionales)
9. [Problemas Encontrados y Soluciones](#9-problemas-encontrados-y-soluciones)
10. [Conclusiones](#10-conclusiones)

---

## 1. Introducción

El presente proyecto tiene como objetivo principal el diseño e implementación de un
entorno virtualizado que integre el uso de máquinas virtuales (VMs) y contenedores,
empleando tecnologías modernas como Docker, Containerd, Podman, Go y Zot. Esta
arquitectura permite simular entornos reales de desarrollo utilizados en la industria,
enfocados en la contenerización, el almacenamiento de imágenes de contenedores,
conexión entre API´s y la gestión eficiente de recursos.

---

## 2. Arquitectura General

![arquitectura](/docs/capturas/arquitectura.png)

**Distribución de la infraestructura:**

| Máquina Virtual | Runtime de Contenedores | Contenedor(es) Ejecutado(s)            | Disco Asignado | IP              |
| --------------- | ----------------------- | -------------------------------------- | -------------- | --------------- |
| VM1             | Containerd              | API1 (puerto 8081), API2 (puerto 8082) | 9 GB           | 192.168.122.227 |
| VM2             | Podman                  | API3 (puerto 8083)                     | 7 GB           | 192.168.122.245 |
| VM3             | Docker                  | Zot (puerto 5000)                      | 8 GB           | 192.168.122.201 |

**Herramientas del host:**

- Hipervisor: KVM/QEMU con virt-manager (libvirt)
- Sistema operativo de las VMs: Ubuntu Server 24.04 LTS
- Lenguaje de las APIs: Go 1.25.1 con framework Fiber v2

---

## 3. VM1 – Containerd (API1 y API2)

### 3.1 Instalación de la Máquina Virtual

Creación de la VM en virt-manager, 9GB de disco, 2GB RAM, 2 vCPU, Ubuntu Server 24.04 LTS.

**Captura: Instalación completa de Ubuntu Server**
![instalacion](/docs/capturas/instalacion-vm1.png)

**Captura: Login exitoso y datos del sistema (hostname, IP)**
![](/docs/capturas/inicio-sesion-vm1.png)

### 3.2 Instalación de Containerd

Comando utilizado:

```bash
sudo apt install -y containerd
```

**Captura: containerd --version y systemctl status containerd**
![](/docs/capturas/containerd-status.png)

### 3.3 Instalación de nerdctl, buildkit y plugins CNI

Containerd es un runtime de bajo nivel y no incluye por sí solo herramientas de build ni de red, a diferencia de Docker y Podman. Fue necesario instalar tres componentes adicionales:

- **nerdctl**: CLI compatible con la sintaxis de Docker que interactúa con containerd.
- **buildkit**: motor de construcción de imágenes requerido por `nerdctl build`.
- **containernetworking-plugins (CNI)**: plugins de red requeridos por `nerdctl run` para crear la red del contenedor y mapear puertos.

```bash
# nerdctl
wget https://github.com/containerd/nerdctl/releases/download/v2.0.5/nerdctl-2.0.5-linux-amd64.tar.gz
sudo tar Cxzvf /usr/local/bin nerdctl-2.0.5-linux-amd64.tar.gz

# buildkit
wget https://github.com/moby/buildkit/releases/download/v0.19.0/buildkit-v0.19.0.linux-amd64.tar.gz
sudo tar Cxzvf /usr/local buildkit-v0.19.0.linux-amd64.tar.gz
sudo buildkitd &

# plugins CNI (necesarios para nerdctl run)
sudo apt install -y containernetworking-plugins
sudo mkdir -p /opt/cni/bin
sudo ln -s /usr/lib/cni/* /opt/cni/bin/
```

**Captura: nerdctl --version**
![](/docs/capturas/nerdctl--version.png)

### 3.4 Clonado del Repositorio

```bash
git clone https://github.com/TheDark004/202300540_LAB_SO1_2S2026.git
```

### 3.5 Construcción de las Imágenes

```bash
cd api1
sudo nerdctl build -t api1-202300540 .

cd ../api2
sudo nerdctl build -t api2-202300540 .
```

**Captura: salida completa del build de API1**
![](/docs/capturas/api1-202300540.png)

**Captura: nerdctl images mostrando api1-202300540 y api2-202300540**
![](/docs/capturas/images-vm1.png)

### 3.6 Ejecución de los Contenedores

```bash
sudo nerdctl run -d --name api1 -p 8081:8081 api1-202300540
sudo nerdctl run -d --name api2 -p 8082:8082 api2-202300540
```

**Captura: nerdctl ps mostrando api1 y api2 corriendo**
![](/docs/capturas/api1-api2-run.png)

---

## 4. VM2 – Podman (API3)

### 4.1 Instalación de la Máquina Virtual

Mismo proceso que VM1, ajustando a 7GB de disco, hostname vm2-podman.

### 4.2 Instalación de Podman

```bash
sudo apt install -y podman
```

**Captura: podman version**

![](/docs/capturas/podman-v.png)

### 4.3 Clonado del Repositorio

```bash
git clone https://github.com/TheDark004/202300540_LAB_SO1_2S2026.git
```

### 4.4 Construcción de la Imagen

```bash
cd api3
podman build -t api3-202300540 .
```

**Captura: salida del build de API3**

![](/docs/capturas/build-api3.png)

**Captura: podman images mostrando api3-202300540**
![](/docs/capturas/podman-img.png)

### 4.5 Ejecución del Contenedor

```bash
podman run -d --name api3 -p 8083:8083 api3-202300540
```

**Captura: api3 corriendo y respondiendo /health**
![](/docs/capturas/api3-run.png)

---

## 5. VM3 – Docker y Registro Zot

### 5.1 Instalación de la Máquina Virtual

Mismo proceso, 8GB de disco, hostname vm3-docker.

### 5.2 Instalación de Docker

Se instaló Docker CE desde el repositorio oficial (no el paquete `docker.io` de los repos de Ubuntu, para obtener la versión más reciente):

```bash
sudo apt install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
```

**Captura: docker --version y docker ps**
`[ESPACIO PARA CAPTURA]`

### 5.3 Instalación y Configuración de Zot

Zot se ejecuta como un contenedor Docker, exponiendo el puerto 5000 (estándar para registros OCI/Docker):

```bash
mkdir -p ~/zot/data
docker run -d --name zot -p 5000:5000 -v ~/zot/data:/var/lib/registry ghcr.io/project-zot/zot-linux-amd64:latest
```

**Captura: docker ps mostrando el contenedor zot corriendo**
![](/docs/capturas/docker-version.png)

**Captura: curl -v http://localhost:5000/v2/ (confirma HTTP 200 OK)**
![](/docs/capturas/url-zot.png)

---

## 6. Comunicación entre APIs

Cada una de las 3 APIs expone tres endpoints:

- `GET /health`: responde con su propio estado (`status`, `message`, `timestamp`, `VM`, `carnet`).
- `GET /api#/{carnet}/call-api#`: realiza una petición HTTP interna al endpoint `/health` de otra API. Si la respuesta llega correctamente y su campo `status` es `"UP"`, se devuelve `connection: true`; si hay un error de conexión o el estado no es válido, se devuelve `connection: false` junto con un mensaje de error.

### 6.1 Código de las APIs

Las tres APIs están escritas en Go utilizando el framework **Fiber v2**, con una estructura idéntica entre sí (solo cambian el puerto propio y las URLs de las APIs a las que llaman). Cada una se empaqueta con un **Dockerfile multi-stage**:

1. Etapa `builder` (`golang:1.25-alpine`): compila el binario estático con `CGO_ENABLED=0`.
2. Etapa final (`alpine:3.22`): copia únicamente el binario compilado, resultando en imágenes de ~20MB.

La comunicación entre APIs que residen en la misma VM (API1 y API2, ambas en VM1) también se realiza mediante la IP real de la VM y no mediante `localhost`, ya que cada contenedor posee su propia red aislada (ver sección 9).

### 6.2 Prueba Individual (API1)

```bash
curl http://localhost:8081/health
```

**Captura: respuesta JSON de /health**
![](/docs/capturas/)

### 6.3 Prueba de Comunicación Cruzada

Con las 3 APIs corriendo simultáneamente (API1 y API2 en VM1, API3 en VM2), se probaron las 6 combinaciones posibles de comunicación cruzada, confirmando `connection: true` en todos los casos:

**Captura: GET /api1/202300540/call-api2 (VM1 → VM1, vía IP real)**
![](/docs/capturas/api1-call-api2.png)

**Captura: GET /api1/202300540/call-api3 (VM1 → VM2)**
![](/docs/capturas/api1-call-api3.png)

**Captura: GET /api2/202300540/call-api1 (VM1 → VM1)**
![](/docs/capturas/api2-call-api1.png)

**Captura: GET /api2/202300540/call-api3 (VM1 → VM2)**
![](/docs/capturas/api2-call-api3.png)

**Captura: GET /api3/202300540/call-api1 (VM2 → VM1)**
![](/docs/capturas/api3-call-api1.png)

**Captura: GET /api3/202300540/call-api2 (VM2 → VM1)**
![](/docs/capturas/api3-call-api2.png)

### 6.4 Manejo de Errores

Se simuló la caída de API3 (VM2) deteniendo su contenedor sin borrarlo (`podman stop api3`), y se confirmó que API1 detecta correctamente la falla y responde con `connection: false` y el mensaje de error especificado en el enunciado. Al reiniciar el contenedor (`podman start api3`), la comunicación se restableció automáticamente sin necesidad de reconstruir la imagen.

**Captura: connection:false cuando API3 está detenida**
![](/docs/capturas/podman-stop.png)

**Captura: connection:true tras reiniciar API3**
![](/docs/capturas/api2-call-api3.png)

---

## 7. Registro Zot: Push y Pull de Imágenes

Las 3 imágenes fueron subidas al registro Zot (VM3, `192.168.122.201:5000`) desde sus respectivas VMs, y se confirmó la descarga desde una VM distinta a la que construyó la imagen.

**Configuración previa (registro sin HTTPS):** dado que Zot se desplegó sin certificados TLS, fue necesario configurar cada runtime para aceptar el registro como "inseguro":

- **Docker (VM3):** `insecure-registries` en `/etc/docker/daemon.json`:
  ```json
  { "insecure-registries": ["192.168.122.201:5000"] }
  ```
- **Containerd (VM1):** archivo `/etc/containerd/certs.d/192.168.122.201:5000/hosts.toml`:

  ```toml
  server = "http://192.168.122.201:5000"

  [host."http://192.168.122.201:5000"]
    capabilities = ["pull", "resolve", "push"]
    skip_verify = true
  ```

- **Podman (VM2):** entrada en `/etc/containers/registries.conf`:
  ```toml
  [[registry]]
  location = "192.168.122.201:5000"
  insecure = true
  ```

**Push desde VM1 (containerd/nerdctl) — API1 y API2:**

```bash
sudo nerdctl tag api1-202300540 192.168.122.201:5000/api1-202300540
sudo nerdctl push 192.168.122.201:5000/api1-202300540

sudo nerdctl tag api2-202300540 192.168.122.201:5000/api2-202300540
sudo nerdctl push 192.168.122.201:5000/api2-202300540
```

**Push desde VM2 (Podman) — API3:**

```bash
podman tag api3-202300540 192.168.122.201:5000/api3-202300540
podman push 192.168.122.201:5000/api3-202300540
```

**Pull desde VM2, confirmando que el registro funciona entre VMs distintas:**

```bash
podman pull 192.168.122.201:5000/api1-202300540
```

**Verificación del catálogo completo del registro:**

```bash
curl http://192.168.122.201:5000/v2/_catalog
```

**Captura: push exitoso de api1-202300540 desde VM1**
![](/docs/capturas/push-vm1.png)

**Captura: pull exitoso de api1-202300540 desde VM2 (podman images mostrando la imagen)**
![](/docs/capturas/podman-pull.png)

**Captura: curl .../v2/\_catalog mostrando las 3 imágenes**
![](/docs/capturas/_catalog.png)

---

## 8. Pruebas Funcionales

El sistema completo fue validado de punta a punta:

- Las 3 VMs (Containerd, Podman, Docker) están operativas y accesibles por SSH.
- Las 3 APIs (Go + Fiber) responden correctamente en su endpoint `/health` con el formato JSON especificado.
- La comunicación cruzada entre las 6 combinaciones de APIs funciona correctamente, tanto entre contenedores de la misma VM como entre VMs distintas.
- El manejo de errores fue verificado deteniendo intencionalmente un contenedor y confirmando la respuesta `connection: false`, así como la recuperación automática al reiniciarlo.
- El registro privado Zot recibió las 3 imágenes (subidas desde sus respectivas VMs) y permitió su descarga desde una VM distinta a la de origen.

---

## 9. Problemas Encontrados y Soluciones

| Problema                                                                                                      | Causa                                                                                                                                    | Solución                                                                                                                                                   |
| ------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| virt-manager mostraba "Not Connected"                                                                         | No se había reiniciado sesión tras agregar el usuario a los grupos libvirt/kvm                                                           | Cerrar sesión y volver a entrar                                                                                                                            |
| `nerdctl build` fallaba con error de buildkit                                                                 | containerd no incluye motor de build propio                                                                                              | Instalar buildkit y ejecutar `buildkitd` como daemon                                                                                                       |
| `podman build` fallaba con "short-name did not resolve to an alias" al construir API3                         | Podman, a diferencia de Docker, no asume Docker Hub por defecto para nombres de imagen sin registro explícito (ej. `golang:1.25-alpine`) | Editar `/etc/containers/registries.conf` y descomentar/agregar `unqualified-search-registries = ["docker.io"]`                                             |
| `nerdctl run` fallaba con "needs CNI plugin bridge to be installed" al intentar levantar API1 como contenedor | containerd no incluye plugins de red (CNI) integrados, a diferencia de Docker/Podman                                                     | Instalar el paquete `containernetworking-plugins` vía apt y enlazar los binarios a `/opt/cni/bin` (la ruta que espera nerdctl)                             |
| API1 no podía contactar a API2 usando `127.0.0.1`, aunque ambas corren en VM1                                 | Cada contenedor tiene su propia red aislada; `127.0.0.1` dentro de un contenedor se refiere a sí mismo, no al host que lo contiene       | Usar la IP real de la VM (`192.168.122.227`) en vez de `localhost`/`127.0.0.1` para toda comunicación entre contenedores, incluso si comparten la misma VM |
| `docker push`/`nerdctl push`/`podman push` fallaban al subir imágenes a Zot                                   | Zot se desplegó sin certificados TLS (HTTP plano), y los runtimes rechazan por defecto registros sin HTTPS                               | Configurar cada runtime para tratar `192.168.122.201:5000` como registro "inseguro" (ver sección 7)                                                        |

---

## 10. Conclusiones

Este proyecto permitió comparar de forma práctica tres runtimes de contenedores con filosofías distintas: **Docker**, orientado a facilidad de uso con todas las herramientas integradas (build, red, registro inseguro) desde un único daemon; **Podman**, con un enfoque rootless y compatible con la sintaxis de Docker, pero con diferencias de comportamiento como la resolución de nombres de imagen y la gestión de contenedores por usuario; y **Containerd**, un runtime de bajo nivel que requiere ensamblar manualmente piezas adicionales (nerdctl, buildkit, plugins CNI) para lograr una experiencia equivalente a los otros dos.

La virtualización con KVM permitió aislar cada runtime en su propio entorno, evitando conflictos entre configuraciones y simulando un despliegue distribuido real. El uso de un registro privado (Zot) evidenció la importancia de la configuración de seguridad en el transporte de imágenes (HTTP vs. HTTPS) y cómo cada herramienta maneja esa confianza de forma distinta.

Finalmente, la comunicación REST entre las tres APIs reforzó un concepto clave de redes en contenedores: cada contenedor posee su propio espacio de red aislado, por lo que `localhost` nunca debe asumirse como una forma válida de comunicación entre contenedores distintos, incluso si conviven en la misma máquina virtual.
