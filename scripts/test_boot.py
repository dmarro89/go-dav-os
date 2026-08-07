import argparse
import os
import subprocess
import sys
import time


def check_log_for(target, log_file, timeout=5):
    start = time.time()
    while time.time() - start < timeout:
        if os.path.exists(log_file):
            with open(log_file, "r", errors="ignore") as f:
                if target in f.read():
                    return True
        time.sleep(0.1)
    return False


def create_disk_image(disk_img):
    print(f"Creating empty {disk_img} (20MB)...")
    with open(disk_img, "wb") as f:
        f.write(b"\0" * (20 * 1024 * 1024))


def start_qemu(iso_path, disk_img, log_file, serial_socket=None):
    cmd = [
        "qemu-system-x86_64",
        "-cdrom",
        iso_path,
        "-drive",
        f"file={disk_img},format=raw",
        "-debugcon",
        f"file:{log_file}",
    ]
    if serial_socket:
        if os.path.exists(serial_socket):
            os.remove(serial_socket)
        cmd.extend(
            ["-serial", f"unix:{serial_socket},server=on,wait=off"],
        )
    else:
        cmd.extend(["-serial", "none"])
    cmd.extend(
        [
            "-monitor",
            "stdio",
            "-display",
            "none",
            "-no-reboot",
            "-no-shutdown",
        ],
    )
    return subprocess.Popen(
        cmd,
        stdin=subprocess.PIPE,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.PIPE,
        text=True,
    )


def stop_qemu(process):
    if process.poll() is None:
        process.terminate()
        try:
            process.wait(timeout=2)
        except subprocess.TimeoutExpired:
            process.kill()
    if process.stdin is not None:
        process.stdin.close()
    if process.stderr is not None:
        process.stderr.close()


def fail_with_log(message, process, log_file):
    print(f"ERROR: {message}")
    if process.poll() is not None:
        print(f"QEMU process exited with code {process.returncode}")
        print("--- QEMU stderr ---")
        print(process.stderr.read())
        print("-------------------")

    print(f"--- {log_file} content ---")
    if os.path.exists(log_file):
        with open(log_file, "r", errors="ignore") as f:
            print(f.read())
    else:
        print(f"{log_file} not found.")
    print("--------------------------")
    sys.exit(1)


def wait_for_boot(process, log_file):
    print("Waiting for boot prompt...")
    if not check_log_for("Welcome to DavOS", log_file, timeout=12):
        fail_with_log("Timeout waiting for DavOS prompt.", process, log_file)
    print("Boot successful.")


def send_shell_command(process, cmd_text):
    print(f"Sending '{cmd_text}' command via QEMU monitor...")
    for ch in cmd_text:
        key = "spc" if ch == " " else ch
        try:
            process.stdin.write(f"sendkey {key}\n")
            process.stdin.flush()
        except BrokenPipeError:
            print("Error: QEMU closed stdin (crashed?)")
            break
        time.sleep(0.1)

    try:
        process.stdin.write("sendkey ret\n")
        process.stdin.flush()
    except BrokenPipeError:
        print("Error: QEMU closed stdin while sending Enter.")
    time.sleep(0.1)


def run_agent_transport_probe(process, log_file, serial_socket):
    probe = subprocess.Popen(
        [
            sys.executable,
            "scripts/agent_transport_probe.py",
            "--socket",
            serial_socket,
        ],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    try:
        send_shell_command(process, "agent transport ping")
        if not check_log_for("Agent transport: connected", log_file, timeout=6):
            fail_with_log("Agent transport probe did not complete.", process, log_file)
        try:
            probe.wait(timeout=2)
        except subprocess.TimeoutExpired:
            fail_with_log("Host transport probe did not exit.", process, log_file)
        if probe.returncode != 0:
            print("--- Agent transport probe stderr ---")
            print(probe.stderr.read())
            print("------------------------------------")
            fail_with_log("Host transport probe failed.", process, log_file)

        send_shell_command(process, "agent transport ping")
        if not check_log_for("agent: transport timeout", log_file, timeout=6):
            fail_with_log("Transport timeout was not reported.", process, log_file)

        send_shell_command(process, "version")
        if not check_log_for("DavOS 0.5.0 (64bit)", log_file, timeout=6):
            fail_with_log("Shell did not recover from transport timeout.", process, log_file)
    finally:
        if probe.poll() is None:
            probe.terminate()
            probe.wait(timeout=2)
        if probe.stdout is not None:
            probe.stdout.close()
        if probe.stderr is not None:
            probe.stderr.close()


def run_functional_suite(iso_path, disk_img, log_file):
    if os.path.exists(log_file):
        os.remove(log_file)

    serial_socket = "qemu-agent.sock"
    process = start_qemu(iso_path, disk_img, log_file, serial_socket)
    try:
        wait_for_boot(process, log_file)
        run_agent_transport_probe(process, log_file, serial_socket)

        test_cases = [
            ("help", ["Commands:", "agent"]),
            ("version", ["DavOS 0.5.0 (64bit)"]),
            ("write notes hi", ["ok"]),
            ("agent show files", ["notes  size=2", "agent: files listed"]),
            (
                "agent context",
                [
                    "Agent context:",
                    "last input: agent show files",
                    "last intent: list_files",
                    "last action: list_files",
                    "request count: 1",
                    "planner mode: deterministic",
                ],
            ),
            ("agent help", ["Agent commands:", "agent delete <name>"]),
            ("agent show history", ["agent: history listed"]),
            ("agent show version", ["DavOS 0.5.0 (64bit)", "agent: version shown"]),
            ("agent show ticks", ["agent: ticks shown"]),
            ("agent show memorymap", ["base=0x", "agent: memory map shown"]),
            ("agent read notes", ["hi", "agent: file read"]),
            ("agent stat notes", ["page=0x", "size=2", "agent: file stat"]),
            ("agent delete notes", ["Agent understood: delete file \"notes\"", "This action may remove data.", "Confirm? type: yes"]),
            ("yes", ["ok"]),
            ("agent show files", ["agent: no files"]),
            ("agent mode", ["Current planner: deterministic"]),
            ("agent mode deterministic", ["Planner switched to: deterministic"]),
            ("agent mode llm", ["agent: llm bridge unavailable"]),
            ("agent mode", ["Current planner: deterministic"]),
            ("fatformat", ["FAT16 Formatted"]),
            ("fatinit", ["FAT16 Initialized"]),
            ("fatcreate test hi", ["File created"]),
            ("fatls", ["TEST"]),
            ("fatread test", ["hi"]),
            ("layout", ["current layout:"]),
            ("layout us", ["layout: switched to us"]),
            ("layout it", ["layout: switched to it"]),
            ("run hello", ["hello from userland", "Process exited with status 0"]),
        ]

        for cmd_text, expected_outputs in test_cases:
            send_shell_command(process, cmd_text)
            for expected in expected_outputs:
                print(f"Waiting for '{expected}' output...")
                if not check_log_for(expected, log_file, timeout=6):
                    fail_with_log(
                        f"Timeout waiting for '{expected}' from command '{cmd_text}'.",
                        process,
                        log_file,
                    )
            print(f"Test Passed: '{cmd_text}' command executed successfully.")
    finally:
        stop_qemu(process)
        if os.path.exists(serial_socket):
            os.remove(serial_socket)


def run_fault_probe(iso_path, disk_img, cmd_text, log_file, fault_marker="PF"):
    last_process = None
    for attempt in range(1, 4):
        if os.path.exists(log_file):
            os.remove(log_file)

        process = start_qemu(iso_path, disk_img, log_file)
        last_process = process
        try:
            print(f"Waiting for boot prompt for '{cmd_text}' (attempt {attempt}/3)...")
            if not check_log_for("Welcome to DavOS", log_file, timeout=12):
                print("Boot prompt not seen, retrying...")
                continue

            print("Boot successful.")
            send_shell_command(process, cmd_text)

            print(f"Waiting for ring3 protection fault marker ('{fault_marker}')...")
            if not check_log_for(fault_marker, log_file, timeout=6):
                fail_with_log(
                    f"Did not observe {fault_marker} marker after '{cmd_text}'.",
                    process,
                    log_file,
                )
            print(f"Test Passed: '{cmd_text}' triggered {fault_marker} as expected.")
            return
        finally:
            stop_qemu(process)
            time.sleep(1)

    fail_with_log(
        f"Timeout waiting for boot prompt while running '{cmd_text}' after retries.",
        last_process,
        log_file,
    )


def parse_args():
    parser = argparse.ArgumentParser(
        description="Run QEMU boot verification for DavOS.",
    )
    suite = parser.add_mutually_exclusive_group()
    suite.add_argument(
        "--functional",
        action="store_true",
        help="Run only the functional shell suite (skip fault probes).",
    )
    suite.add_argument(
        "--fault-probes",
        action="store_true",
        help="Run only the kread / kwrite / kpriv fault probes (skip the functional suite).",
    )
    return parser.parse_args()


def main():
    args = parse_args()
    run_functional = not args.fault_probes
    run_faults = not args.functional

    iso_path = "build/dav-go-os.iso"

    if not os.path.exists(iso_path):
        print(f"ERROR: ISO not found at {iso_path}. Build it first.")
        sys.exit(1)

    print(f"Starting QEMU verification for {iso_path}...")

    if run_functional:
        disk_img = "disk.img"
        create_disk_image(disk_img)
        run_functional_suite(iso_path, disk_img, "qemu.log")

    if run_faults:
        # Each probe must run in its own VM instance because a #PF is terminal here.
        kread_disk = "disk_kread.img"
        kwrite_disk = "disk_kwrite.img"
        kpriv_disk = "disk_kpriv.img"
        create_disk_image(kread_disk)
        create_disk_image(kwrite_disk)
        create_disk_image(kpriv_disk)
        run_fault_probe(iso_path, kread_disk, "run kread", "qemu_kread.log", "PF")
        run_fault_probe(iso_path, kwrite_disk, "run kwrite", "qemu_kwrite.log", "PF")
        run_fault_probe(iso_path, kpriv_disk, "run kpriv", "qemu_kpriv.log", "GP")

    print("All QEMU checks passed.")


if __name__ == "__main__":
    main()
