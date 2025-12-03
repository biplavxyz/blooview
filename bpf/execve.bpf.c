// SPDX-License-Identifier: GPL-2.0
#include "vmlinux.h"
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_helpers.h>

char LICENSE[] SEC("license") = "GPL";

struct {
  __uint(type, BPF_MAP_TYPE_RINGBUF);
  __uint(max_entries, 256 * 1024);
} ringbuf SEC(".maps");

struct event {
  u32 pid;
  char filename[512];
};

/*
 * Tracepoint argument struct:
 * struct trace_event_raw_sys_enter {
 *     long id;
 *     long args[6];
 * };
 *
 * For execve, args[0] = const char *filename
 */
SEC("tp/syscalls/sys_enter_execve")
int handle_execve(struct trace_event_raw_sys_enter *ctx) {
  struct event *evt;
  const char *filename;

  u32 pid = bpf_get_current_pid_tgid() >> 32;

  evt = bpf_ringbuf_reserve(&ringbuf, sizeof(*evt), 0);
  if (!evt)
    return 0;

  evt->pid = pid;

  filename = (const char *)ctx->args[0];

  /* CO-RE safe user read */
  bpf_probe_read_user_str(evt->filename, sizeof(evt->filename), filename);

  bpf_ringbuf_submit(evt, 0);
  return 0;
}
