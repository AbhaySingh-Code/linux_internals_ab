# Linux Internals: Process Enumeration via Raw Linux APIs

This project implements a native, pure system-level process enumerator in Go. It completely bypasses high-level wrappers (like Go's `os` package or calling out to the external `ps` binary) to interact directly with the Linux kernel's window into system state: the **`/proc` pseudo-filesystem**.

---

## Architecture Overview

In Linux, `/proc` does not exist on disk; it is a virtual filesystem generated on the fly by the kernel. Every running process is assigned a directory named after its unique Process ID (PID). 

Instead of treating directories and files as continuous abstract strings, this tool uses raw Linux kernel file descriptors and directory entries (`dirent`) structures to map out processes.

```text
+------------------------------------------------------------+
|                       Kernel Space                         |
+------------------------------------------------------------+
       ^                          ^                       ^
       | sys_openat               | sys_getdents64        | sys_read
       |                          |                       |
+------+--------------------------+-----------------------+--+
|      |                          |                       |  |
|  1. Open /proc           2. Read directory stream       |  |
|  Enforces O_DIRECTORY    Parses variable-length         |  |
|  returns raw FD          linux_dirent64 blocks          |  |
|                                                         |  |
|                                            3. Inspect Targets
|                                            Reads /proc/[PID]/cmdline
|                                            Reads /proc/[PID]/maps
|                                                            |
|                        Go User Space                       |
+------------------------------------------------------------+
