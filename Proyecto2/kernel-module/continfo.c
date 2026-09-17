#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/sched.h>
#include <linux/ktime.h>
#include <linux/sched/signal.h>
#include <linux/sched/mm.h>
#include <linux/mm.h>
#include <linux/jiffies.h>

#define CARNET "202300540"
#define PROC_NAME "continfo_pr2_so1_" CARNET

MODULE_LICENSE("GPL");
MODULE_AUTHOR("Rene Sebastian Gutierrez Contreras");
MODULE_DESCRIPTION("Sonda de kernel: memoria y procesos para telemetria de contenedores");

static int continfo_show(struct seq_file *m, void *v)
{
    struct sysinfo si;
    struct task_struct *task;
    unsigned long total_ram_kb, free_ram_kb, used_ram_kb;
    unsigned long uptime_ms;

    si_meminfo(&si);
    total_ram_kb = si.totalram * (PAGE_SIZE / 1024);
    free_ram_kb  = si.freeram  * (PAGE_SIZE / 1024);
    used_ram_kb  = total_ram_kb - free_ram_kb;

    seq_printf(m, "MEM_TOTAL_KB:%lu\n", total_ram_kb);
    seq_printf(m, "MEM_FREE_KB:%lu\n", free_ram_kb);
    seq_printf(m, "MEM_USED_KB:%lu\n", used_ram_kb);
    seq_printf(m, "---PROCESOS---\n");
    seq_printf(m, "PID|NOMBRE|CMD|VSZ_KB|RSS_KB|MEM_PORC|CPU_PORC\n");

    uptime_ms = jiffies_to_msecs(jiffies);
    if (uptime_ms == 0)
        uptime_ms = 1; /* evitar division por cero */

    rcu_read_lock();
    for_each_process(task) {
        struct mm_struct *mm;
        unsigned long vsz_kb = 0, rss_kb = 0;
        unsigned long mem_porc = 0, cpu_porc = 0;
        

        task_lock(task);
        mm = task->mm;
        if (mm) {
            vsz_kb = mm->total_vm * (PAGE_SIZE / 1024);
            rss_kb = get_mm_rss(mm) * (PAGE_SIZE / 1024);
            if (total_ram_kb > 0)
                mem_porc = (rss_kb * 100) / total_ram_kb;
        }
        task_unlock(task);

        {
        u64 now_ns = ktime_get_boottime_ns();
        u64 elapsed_ns = now_ns - task->start_time;
        u64 total_time_ns = task->utime + task->stime;
        if (elapsed_ns == 0) elapsed_ns = 1;
        cpu_porc = (unsigned long)((total_time_ns * 100ULL) / elapsed_ns);
    }
        seq_printf(m, "%d|%s|%s|%lu|%lu|%lu|%lu\n",
           task->pid, task->comm, task->comm, vsz_kb, rss_kb, mem_porc, cpu_porc);
    }
    rcu_read_unlock();

    return 0;
}

static int continfo_open(struct inode *inode, struct file *file)
{
    return single_open(file, continfo_show, NULL);
}

static const struct proc_ops continfo_fops = {
    .proc_open    = continfo_open,
    .proc_read    = seq_read,
    .proc_lseek   = seq_lseek,
    .proc_release = single_release,
};

static int __init continfo_init(void)
{
    proc_create(PROC_NAME, 0444, NULL, &continfo_fops);
    printk(KERN_INFO "continfo: modulo cargado, /proc/%s creado\n", PROC_NAME);
    return 0;
}

static void __exit continfo_exit(void)
{
    remove_proc_entry(PROC_NAME, NULL);
    printk(KERN_INFO "continfo: modulo descargado, /proc/%s eliminado\n", PROC_NAME);
}

module_init(continfo_init);
module_exit(continfo_exit);