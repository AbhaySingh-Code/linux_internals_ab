# Ptrace Syscall Injection — Concepts & Reference

This project uses Linux's `ptrace(2)` API to attach to a running process,
inspect its memory and registers, and momentarily hijack its execution to
run a syscall on its behalf — then restore it to its original state as if
nothing happened.

It's built from two programs:

- **`binary`** (C) — a test target. Prints its own PID, a secret string,
  and the memory address of that string, then loops forever so you have
  time to attach a debugger to it.
- **`main` / `main2`** (Go) — the tool that attaches to `binary`, reads its
  memory, and injects syscalls into it.

---

## 1. What is `ptrace`?

`ptrace` is the Linux syscall that lets one process (the **tracer**) observe
and control another process (the **tracee**). It's the primitive underneath
`gdb`, `strace`, and `Wine`. Key operations used here:

| Call | Purpose |
|---|---|
| `PTRACE_ATTACH` | Tracer asks the kernel to stop the tracee and become its controller. Sends an implicit stop signal. |
| `PTRACE_PEEKDATA` | Read a word of memory from the tracee's address space. |
| `PTRACE_POKEDATA` | Write a word of memory into the tracee's address space. |
| `PTRACE_GETREGS` | Read the tracee's full register set (RIP, RSP, RAX, etc). |
| `PTRACE_SETREGS` | Overwrite the tracee's register set. |
| `PTRACE_SINGLESTEP` | Resume the tracee for exactly one CPU instruction, then stop it again. |
| `PTRACE_DETACH` | Release control and let the tracee resume normally. |

A tracee can only be controlled by one tracer, and — depending on your
distro's Yama LSM settings — usually only by a tracer with a matching UID
or root privileges.

---

## 2. Attaching and stopping the process

```go
syscall.PtraceAttach(pid)
syscall.Wait4(pid, &status, 0, nil)
```

`PtraceAttach` sends the equivalent of `SIGSTOP` to the target. But the stop
isn't instant from the caller's point of view — the kernel needs a moment to
actually park the tracee. `Wait4` blocks until the tracee's status changes
to "stopped," which is the signal that it's now safe to read/write its
memory and registers.

Skipping the `Wait4` and immediately trying to `PeekData` is a common bug —
the process may not be fully stopped yet.

---

## 3. Reading memory: `PtracePeekData`

```go
syscall.PtracePeekData(pid, uintptr(addr), outBuffer)
```

This copies raw bytes out of the tracee's virtual address space into a
buffer in your own process. Address spaces are **per-process** — the
pointer `0x55af6dd74028` only means something inside `binary`'s memory
layout, and reading it from your Go process requires going through the
kernel via `ptrace`, since your process can't normally see another
process's memory directly.

Go's implementation loops internally over word-sized reads to fill
buffers larger than one machine word, so you can request more than 8
bytes at a time.

---

## 4. Registers: `PtraceGetRegs` / `PtraceSetRegs`

Every running thread has a **register state** — the CPU's current
snapshot of execution: which instruction is next (`RIP`), the stack
pointer (`RSP`), and general-purpose registers (`RAX`, `RDI`, `RSI`, etc).

`PtraceGetRegs` reads this whole struct (`syscall.PtraceRegs`) from the
paused tracee. `PtraceSetRegs` writes a (possibly modified) struct back.

This is the key primitive for injection: **if you can read and rewrite
the register state, you can redirect what the CPU does next** — including
making it jump somewhere it never intended to run.

---

## 5. The Linux x86-64 syscall calling convention

To make the kernel perform an action (open a file, write output, allocate
memory, etc), user-space code executes the `syscall` instruction (machine
opcode `0F 05`) with specific registers pre-loaded:

| Register | Meaning |
|---|---|
| `RAX` | Syscall number (which kernel function to call) |
| `RDI` | 1st argument |
| `RSI` | 2nd argument |
| `RDX` | 3rd argument |
| `R10` | 4th argument |
| `R8`  | 5th argument |
| `R9`  | 6th argument |
| `RAX` (after) | Return value (or a negative errno on failure) |

This convention is fixed by the OS/architecture ABI — it's not something
this project invents, it's what every C program's libc wrappers do under
the hood whenever you call `write()`, `read()`, `mmap()`, etc.

---

## 6. The "gadget": finding an existing `syscall` instruction

To trigger a syscall, the CPU must execute the 2-byte opcode `0F 05`
somewhere in **executable** memory. Rather than injecting new code, this
tool takes a shortcut: it scans `/proc/<pid>/maps` for executable memory
regions, reads them via `PtracePeekData`, and searches the bytes for an
existing `0F 05` sequence.

Because virtually every dynamically linked binary maps in `libc`, and
`libc` contains many `syscall` instructions (for its own internal syscall
wrappers), there's almost always one ready to (ab)use. This means we don't
need write+execute permission on any page — we just borrow an instruction
that's already there.

```
/proc/<pid>/maps  →  list of memory regions + permissions (r/w/x)
                  →  filter to executable regions
                  →  read bytes, search for 0x0F 0x05
                  →  return that address as our "gadget"
```

---

## 7. The injection sequence

This is the heart of the tool — hijacking execution for **exactly one
instruction**, then reverting:

```
1. PtraceGetRegs        → save the tracee's real, original state
2. find syscall gadget  → locate an executable "syscall" instruction
3. build fake regs:
     RIP = gadget address
     RAX = syscall number
     RDI/RSI/RDX = arguments
4. PtraceSetRegs         → install the fake state into the real tracee
5. PtraceSingleStep      → let the CPU execute ONE instruction (the syscall)
6. Wait4                 → wait for it to stop again after that instruction
7. PtraceGetRegs         → read RAX — this is the syscall's return value
8. PtraceSetRegs(orig)   → restore the ORIGINAL saved state from step 1
9. PtraceDetach          → resume the process; it continues exactly where
                           it left off, unaware anything happened
```

Step 8 is the most important safety step in the whole design. If you skip
it, the tracee's `RIP` is left pointing into the middle of `libc` (right
after the borrowed `syscall` instruction) instead of back in its own
`while(1)` loop — continuing execution from there would almost certainly
crash or corrupt the process, since the surrounding instructions weren't
meant to run in that context.

---

## 8. Why `PtraceSingleStep` instead of `PtraceCont`

`PtraceCont` would resume the tracee freely, and it would keep running
past the `syscall` instruction into whatever code happens to follow the
gadget in `libc` — which is not code you want to execute, since it belongs
to some unrelated library function.

`PtraceSingleStep` executes **one instruction only**, then re-stops the
tracee under kernel control. Since our fake `RIP` points exactly at the
`syscall` opcode, a single step is guaranteed to execute precisely that
instruction (and nothing after it) before we regain control.

---

## 9. Example walkthrough: `write()`

Given `binary` prints:

```
[+] Secret message string value : PEEKABOOSECOND
[+] Target memory address to read: 0x55af6dd74028
```

Running the injector with syscall number `1` (`write`) and:

| Arg | Register | Value |
|---|---|---|
| fd | `RDI` | `1` (stdout) |
| buf | `RSI` | `0x55af6dd74028` |
| count | `RDX` | `15` |

...causes the **kernel**, acting on the tracee's behalf, to copy 15 bytes
starting at that address to the tracee's stdout — so `PEEKABOOSECOND`
appears in the *target's* terminal, even though the target's own C code
never called `write()` or `printf()` for that specific output. The
tracer (your Go tool) triggered it entirely from the outside.

---

## 10. Example walkthrough: `getpid()`

`getpid` (syscall `39`) takes no arguments and has no side effects, which
makes it a good sanity check: the returned value in `RAX` should exactly
match the PID `binary` printed at startup. If it doesn't, something in the
attach/read/restore pipeline is broken.

---

## 11. Extending to 6-argument syscalls (e.g. `mmap`)

Some syscalls (like `mmap`) need more than 3 arguments. x86-64 provides
three more argument registers — `R10`, `R8`, `R9` — for arguments 4–6.
An extended injection helper sets all six before single-stepping,
following the same save → fake → step → restore pattern.

`mmap`'s signature:

```
mmap(addr, length, prot, flags, fd, offset)
```

Using `PROT_READ|PROT_WRITE|PROT_EXEC` and `MAP_PRIVATE|MAP_ANONYMOUS`
with `fd = -1` asks the kernel to hand back a fresh, writable, executable
page inside the **tracee's** address space — useful if you want to write
your own machine code into the target rather than reusing borrowed
gadgets.

**Gotcha:** `mmap`'s return value is either a valid pointer or a negative
`errno` reinterpreted as an unsigned 64-bit number. A "success" address
that looks suspiciously close to `2^64` is actually a negative error code
and should be checked for, e.g. treating anything above
`0xffffffffff000000` as a failure.

---

## 12. Key safety/correctness rules

- **Always restore the full register struct**, not just `RIP`. The
  `syscall` instruction itself clobbers `RCX` and `R11` on x86-64 as part
  of the hardware's syscall/sysret mechanism, so don't assume anything
  besides `RAX` changes only.
- **Single-step exactly once.** Anything else risks running unintended
  code from whatever library you borrowed the gadget from.
- **Permissions:** the tracer needs to be root, or the same UID as the
  tracee with the system's Yama `ptrace_scope` allowing it. Attaching to
  an arbitrary unrelated process typically requires root.
- **Addresses are per-process.** A pointer read from the tracee's `%p`
  output is only meaningful when passed back through `ptrace` calls
  against that same PID — it means nothing in the tracer's own address
  space.

---

## 13. Mental model summary

At its core, this whole technique is one repeating pattern:

```
save state → install fake state → run one instruction → read result → restore state
```

Every syscall you inject — `getpid`, `write`, `mmap`, `kill`, whatever —
is just a different set of numbers plugged into that same pattern. Once
this shape feels natural, the "hard part" left is only ever: *which
syscall number, and what do its arguments mean?* — which is just standard
Linux syscall ABI knowledge, documented in `man 2 <syscall_name>`.
