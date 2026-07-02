# Linux Internals: Mini Debugger via PTRACE API

This project demonstrates the core mechanics of Linux systems programming and process debugging using native Go and C. By leveraging the Linux `ptrace` (Process Trace) system call interface, the Go binary intercepts a running C process, halts its execution, reads an 8-byte word from its virtual memory space, and detaches cleanly.

---

## Architecture Overview

When a tracer attaches to a target process via `ptrace`, the kernel enforces a parent-child relationship for debugging tracing states. 

```text
  +------------------+                     +------------------+
  |    Go Tracer     |                     |    C Target      |
  |   (Debugger)     |                     |    (Tracee)      |
  +--------+---------+                     +--------+---------+
           |                                        |
           | 1. PTRACE_ATTACH (Sends SIGSTOP)       |
           |--------------------------------------->| [HALTED]
           |                                        |
           | 2. Wait4() blocks until halted         |
           |<---------------------------------------|
           |                                        |
           | 3. PTRACE_PEEKDATA (Read 8-bytes)      |
           |--------------------------------------->| Reads .data
           |<---------------------------------------| string memory
           |                                        |
           | 4. PTRACE_DETACH (Resumes target)      |
           |--------------------------------------->| [RUNNING]
           v                                        v
