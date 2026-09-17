package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	procPath     = "/proc/continfo_pr2_so1_202300540"
	carnet       = "202300540"
	cronSchedule = "*/2 * * * *"
	cronScript   = "/home/thedark004/Escritorio/CURSOS_U/SO1_202300540/202300540_LAB_SO1_2S2026/Proyecto2/scripts/crear_contenedores.sh"
	loopInterval = 30 * time.Second

	minBajoConsumo = 3
	minAltoConsumo = 2

	bpftraceScript = `tracepoint:syscalls:sys_enter_kill /args->sig == 9 || args->sig == 15/ { printf("KILL_EVENT|%d|%d|%d\n", pid, args->pid, args->sig); }
tracepoint:syscalls:sys_enter_tgkill /args->sig == 9 || args->sig == 15/ { printf("KILL_EVENT|%d|%d|%d\n", pid, args->pid, args->sig); }`
)

var ctx = context.Background()

type ebpfEvent struct {
	At       time.Time
	Emitter  int
	Receiver int
	Signal   int
}

var ebpfEvents = make(chan ebpfEvent, 256)

type Proceso struct {
	PID     int
	Nombre  string
	Cmd     string
	VszKB   int
	RssKB   int
	MemPorc int
	CpuPorc int
}

// Grafana

func iniciarGrafana() {
	log.Println("[INIT] Verificando/iniciando contenedor de Grafana via docker compose...")
	cmd := exec.Command("docker", "compose", "-f",
		"/home/thedark004/Escritorio/CURSOS_U/SO1_202300540/202300540_LAB_SO1_2S2026/Proyecto2/grafana-valkey/docker-compose.yml",
		"up", "-d")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[WARN] Error iniciando grafana/valkey: %v - %s", err, string(out))
	} else {
		log.Println("[INIT] Grafana y Valkey activos.")
	}
}

// Cronjob

func instalarCronjob() {
	log.Println("[CRON] Instalando cronjob...")
	linea := fmt.Sprintf("%s bash %s >> /tmp/proyecto2_cron.log 2>&1", cronSchedule, cronScript)

	getCmd := exec.Command("bash", "-c", "crontab -l 2>/dev/null || true")
	actual, _ := getCmd.Output()

	nuevo := string(actual)
	if !strings.Contains(nuevo, cronScript) {
		nuevo += linea + "\n"
	}

	writeCmd := exec.Command("bash", "-c", fmt.Sprintf("echo '%s' | crontab -", strings.ReplaceAll(nuevo, "'", "'\\''")))
	if out, err := writeCmd.CombinedOutput(); err != nil {
		log.Printf("[WARN] Error instalando cronjob: %v - %s", err, string(out))
	} else {
		log.Println("[CRON] Cronjob instalado correctamente.")
	}
}

func eliminarCronjob() {
	log.Println("[CRON] Eliminando cronjob antes de finalizar...")
	getCmd := exec.Command("bash", "-c", "crontab -l 2>/dev/null || true")
	actual, _ := getCmd.Output()

	lineas := strings.Split(string(actual), "\n")
	var nuevasLineas []string
	for _, l := range lineas {
		if !strings.Contains(l, cronScript) && strings.TrimSpace(l) != "" {
			nuevasLineas = append(nuevasLineas, l)
		}
	}
	nuevo := strings.Join(nuevasLineas, "\n")

	writeCmd := exec.Command("bash", "-c", fmt.Sprintf("echo '%s' | crontab -", strings.ReplaceAll(nuevo, "'", "'\\''")))
	writeCmd.Run()
	log.Println("[CRON] Cronjob eliminado.")
}

// Cargar módulo de kernel

func cargarModuloKernel() {
	log.Println("[KERNEL] Cargando módulo de kernel...")
	rutaKo := "/home/thedark004/Escritorio/CURSOS_U/SO1_202300540/202300540_LAB_SO1_2S2026/Proyecto2/kernel-module/continfo.ko"

	// Descargar por si ya estaba cargado de una corrida anterior
	exec.Command("sudo", "rmmod", "continfo").Run()

	out, err := exec.Command("sudo", "insmod", rutaKo).CombinedOutput()
	if err != nil {
		log.Printf("[WARN] Error cargando módulo: %v - %s", err, string(out))
	} else {
		log.Println("[KERNEL] Módulo cargado correctamente.")
	}
}

func descargarModuloKernel() {
	log.Println("[KERNEL] Descargando módulo de kernel...")
	exec.Command("sudo", "rmmod", "continfo").Run()
}

// Sonda eBPF

func iniciarSondaEBPF(rdb *redis.Client) {
	log.Println("[EBPF] Iniciando sonda eBPF (bpftrace) sobre sys_kill...")
	cmd := exec.Command("bpftrace", "-B", "line", "-e", bpftraceScript)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[EBPF][WARN] No se pudo preparar stdout: %v", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("[EBPF][WARN] No se pudo preparar stderr: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[EBPF][WARN] Error iniciando bpftrace: %v", err)
		return
	}

	// Goroutine para ver errores/diagnóstico de bpftrace
	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			log.Printf("[EBPF-STDERR] %s", scannerErr.Text())
		}
		if err := scannerErr.Err(); err != nil {
			log.Printf("[EBPF][WARN] Error leyendo stderr de bpftrace: %v", err)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			linea := scanner.Text()

			if strings.HasPrefix(linea, "KILL_EVENT|") {
				campos := strings.Split(linea, "|")
				if len(campos) != 4 {
					continue
				}
				emisor, _ := strconv.Atoi(campos[1])
				receptor, _ := strconv.Atoi(campos[2])
				senal, _ := strconv.Atoi(campos[3])
				evento := ebpfEvent{At: time.Now(), Emitter: emisor, Receiver: receptor, Signal: senal}
				log.Printf("[EBPF] Señal de terminación observada: %s", linea)
				select {
				case ebpfEvents <- evento:
				default:
					log.Printf("[EBPF][WARN] Cola de eventos llena; se descarta %s", linea)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("[EBPF][WARN] Error leyendo stdout de bpftrace: %v", err)
		}
	}()
	log.Println("[EBPF] Sonda activa, escuchando eventos sys_kill.")
}

// Lectura y parseo de /proc

func leerProc() (memTotal, memFree, memUsed int, procesos []Proceso, err error) {
	f, err := os.Open(procPath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	enProcesos := false
	for scanner.Scan() {
		linea := scanner.Text()

		if strings.HasPrefix(linea, "MEM_TOTAL_KB:") {
			memTotal, _ = strconv.Atoi(strings.TrimPrefix(linea, "MEM_TOTAL_KB:"))
			continue
		}
		if strings.HasPrefix(linea, "MEM_FREE_KB:") {
			memFree, _ = strconv.Atoi(strings.TrimPrefix(linea, "MEM_FREE_KB:"))
			continue
		}
		if strings.HasPrefix(linea, "MEM_USED_KB:") {
			memUsed, _ = strconv.Atoi(strings.TrimPrefix(linea, "MEM_USED_KB:"))
			continue
		}
		if strings.HasPrefix(linea, "---PROCESOS---") {
			enProcesos = true
			continue
		}
		if strings.HasPrefix(linea, "PID|NOMBRE") {
			continue // header
		}
		if enProcesos {
			campos := strings.Split(linea, "|")
			if len(campos) != 7 {
				continue
			}
			pid, _ := strconv.Atoi(campos[0])
			vsz, _ := strconv.Atoi(campos[3])
			rss, _ := strconv.Atoi(campos[4])
			memP, _ := strconv.Atoi(campos[5])
			cpuP, _ := strconv.Atoi(campos[6])
			procesos = append(procesos, Proceso{
				PID: pid, Nombre: campos[1], Cmd: campos[2],
				VszKB: vsz, RssKB: rss, MemPorc: memP, CpuPorc: cpuP,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, 0, nil, fmt.Errorf("leyendo %s: %w", procPath, err)
	}
	return
}

// Gestión de contenedores Docker

type ContenedorDocker struct {
	ID        string
	Nombre    string
	Categoria string
	PID       int
}

func listarContenedoresDocker() []ContenedorDocker {
	out, err := exec.Command("bash", "-c",
		`docker ps --filter "label=proyecto2=carga" --format "{{.ID}}|{{.Names}}|{{.Label \"proyecto2_categoria\"}}"`).Output()
	if err != nil {
		log.Printf("[WARN] Error listando contenedores docker: %v", err)
		return nil
	}

	var contenedores []ContenedorDocker
	for _, linea := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if linea == "" {
			continue
		}
		campos := strings.Split(linea, "|")
		if len(campos) != 3 {
			continue
		}
		pidOut, err := exec.Command("bash", "-c",
			fmt.Sprintf(`docker inspect -f '{{.State.Pid}}' %s`, campos[0])).Output()
		if err != nil {
			continue
		}
		pid, _ := strconv.Atoi(strings.TrimSpace(string(pidOut)))
		contenedores = append(contenedores, ContenedorDocker{
			ID: campos[0], Nombre: campos[1], Categoria: campos[2], PID: pid,
		})
	}
	return contenedores
}

func matarContenedor(id string, pid int, rdb *redis.Client) bool {
	log.Printf("[ACCION] Enviando señal de terminación al contenedor %s (PID host %d)", id, pid)
	inicio := time.Now()

	if out, err := exec.Command("docker", "kill", "--signal", "TERM", id).CombinedOutput(); err != nil {
		log.Printf("[WARN] No se pudo enviar SIGTERM al contenedor %s: %v - %s", id, err, strings.TrimSpace(string(out)))
		return false
	}

	evento, eBPFConfirmado := esperarEventoEBPF(inicio, 2*time.Second)

	for intento := 0; intento < 10; intento++ {
		estado, err := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", id).Output()
		if err != nil || strings.TrimSpace(string(estado)) == "false" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if out, err := exec.Command("docker", "rm", id).CombinedOutput(); err != nil {
		log.Printf("[WARN] TERM no detuvo %s; usando eliminación forzada: %v - %s", id, err, strings.TrimSpace(string(out)))
		if out, err = exec.Command("docker", "rm", "-f", id).CombinedOutput(); err != nil {
			log.Printf("[WARN] No se pudo eliminar forzosamente el contenedor %s: %v - %s", id, err, strings.TrimSpace(string(out)))
			return false
		}
		log.Printf("[EBPF] Eliminación forzada solicitada para %s; se espera confirmación de salida.", id)
	}

	if eBPFConfirmado {
		log.Printf("[EBPF] Confirmada terminación del contenedor %s (PID observado %d, señal %d, emisor %d)", id, evento.Receiver, evento.Signal, evento.Emitter)
		rdb.Incr(ctx, "ebpf:eventos_totales")
		rdb.RPush(ctx, "ebpf:log", fmt.Sprintf("%d|%s|%d|%d|%d",
			time.Now().Unix(), id, evento.Receiver, evento.Emitter, evento.Signal))
		rdb.LTrim(ctx, "ebpf:log", -100, -1)
	} else {
		log.Printf("[WARN] Docker eliminó %s, pero no se observó su SIGTERM en eBPF.", id)
	}
	return true
}

func esperarEventoEBPF(inicio time.Time, timeout time.Duration) (ebpfEvent, bool) {
	temporizador := time.NewTimer(timeout)
	defer temporizador.Stop()
	for {
		select {
		case evento := <-ebpfEvents:
			if !evento.At.Before(inicio) && evento.Signal == 15 {
				return evento, true
			}
		case <-temporizador.C:
			return ebpfEvent{}, false
		}
	}
}

// Lógica de decisión

func gestionarContenedores(procesos []Proceso, rdb *redis.Client) {
	contenedores := listarContenedoresDocker()

	type candidato struct {
		Contenedor ContenedorDocker
		Score      int64
	}
	var candidatos []candidato
	for _, c := range contenedores {
		for _, p := range procesos {
			if p.PID == c.PID {
				score := int64(p.RssKB) + int64(p.CpuPorc)*1000
				candidatos = append(candidatos, candidato{Contenedor: c, Score: score})
				break
			}
		}
	}

	// Ordenar ascendente: menor consumo primero
	sort.Slice(candidatos, func(i, j int) bool {
		return candidatos[i].Score < candidatos[j].Score
	})

	if len(candidatos) == 0 {
		log.Println("[GESTION] No se encontraron procesos de contenedores en /proc.")
		return
	}

	protegidos := make(map[string]bool)
	bajos, altos := 0, 0
	for _, c := range candidatos {
		switch c.Contenedor.Categoria {
		case "bajo":
			bajos++
		case "alto":
			altos++
		}
	}

	// Se conservan los tres contenedores bajos de menor consumo.
	protegidosBajos := 0
	for _, c := range candidatos {
		if c.Contenedor.Categoria == "bajo" && protegidosBajos < minBajoConsumo {
			protegidos[c.Contenedor.ID] = true
			protegidosBajos++
		}
	}
	// Se conservan los dos contenedores altos de mayor consumo.
	protegidosAltos := 0
	for i := len(candidatos) - 1; i >= 0 && protegidosAltos < minAltoConsumo; i-- {
		if candidatos[i].Contenedor.Categoria == "alto" {
			protegidos[candidatos[i].Contenedor.ID] = true
			protegidosAltos++
		}
	}

	eliminados := 0
	for _, c := range candidatos {
		if protegidos[c.Contenedor.ID] {
			continue
		}
		if c.Contenedor.Categoria != "intruso" && c.Contenedor.Categoria != "bajo" && c.Contenedor.Categoria != "alto" {
			log.Printf("[GESTION] Contenedor %s sin categoría reconocida; se conserva.", c.Contenedor.ID)
			continue
		}
		if matarContenedor(c.Contenedor.ID, c.Contenedor.PID, rdb) {
			eliminados++
		}
	}
	log.Printf("[GESTION] %d eliminados, %d protegidos; categorías detectadas: %d bajos, %d altos.", eliminados, len(protegidos), bajos, altos)
}

// Registro en Valkey de métricas de memoria y procesos

func guardarEnValkey(rdb *redis.Client, memTotal, memFree, memUsed int, procesos []Proceso) {
	ts := time.Now().Unix()

	rdb.Set(ctx, "mem:total", memTotal, 0)
	rdb.Set(ctx, "mem:free", memFree, 0)
	rdb.Set(ctx, "mem:used", memUsed, 0)
	rdb.ZAdd(ctx, "mem:historia", redis.Z{Score: float64(ts), Member: memUsed})

	for _, p := range procesos {
		key := fmt.Sprintf("proc:%d:%s", p.PID, carnet)
		rdb.HSet(ctx, key, map[string]interface{}{
			"nombre": p.Nombre, "rss": p.RssKB, "vsz": p.VszKB,
			"mem_porc": p.MemPorc, "cpu_porc": p.CpuPorc, "timestamp": ts,
		})
		rdb.Expire(ctx, key, 10*time.Minute)
	}
	log.Println("[VALKEY] Datos guardados correctamente.")
}

// MAIN

func main() {
	log.Println("=== Daemon Proyecto 2 SO1 - Carnet " + carnet + " ===")

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("No se pudo conectar a Valkey: %v", err)
	}
	log.Println("[INIT] Conectado a Valkey.")

	iniciarGrafana()
	time.Sleep(3 * time.Second)
	instalarCronjob()
	cargarModuloKernel()
	time.Sleep(2 * time.Second)
	iniciarSondaEBPF(rdb)

	// Manejo de señal de terminación limpia (Ctrl+C)
	defer func() {
		eliminarCronjob()
		descargarModuloKernel()
		log.Println("Daemon finalizado limpiamente.")
	}()

	log.Println("[LOOP] Iniciando loop principal...")
	for {
		memTotal, memFree, memUsed, procesos, err := leerProc()
		if err != nil {
			log.Printf("[ERROR] Leyendo /proc: %v", err)
			time.Sleep(loopInterval)
			continue
		}

		gestionarContenedores(procesos, rdb)
		guardarEnValkey(rdb, memTotal, memFree, memUsed, procesos)

		time.Sleep(loopInterval)
	}
}
