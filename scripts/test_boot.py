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


def start_qemu(iso_path, disk_img, log_file):
    cmd = [
        "qemu-system-x86_64",
        "-cdrom",
        iso_path,
        "-drive",
        f"file={disk_img},format=raw",
        "-debugcon",
        f"file:{log_file}",
        "-serial",
        "none",
        "-monitor",
        "stdio",
        "-display",
        "none",
        "-no-reboot",
        "-no-shutdown",
    ]
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


def run_functional_suite(iso_path, disk_img, log_file):
    if os.path.exists(log_file):
        os.remove(log_file)

    process = start_qemu(iso_path, disk_img, log_file)
    try:
        wait_for_boot(process, log_file)

        test_cases = [
            ("help", "Commands:"),
            ("version", "DavOS 0.2.0 (64bit)"),
            ("fatformat", "FAT16 Formatted"),
            ("fatinit", "FAT16 Initialized"),
            ("fatcreate test hi", "File created"),
            ("fatls", "TEST"),
            ("fatread test", "hi"),
        ]

        for cmd_text, expected in test_cases:
            send_shell_command(process, cmd_text)
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


def run_fault_probe(iso_path, disk_img, cmd_text, log_file):
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

            print("Waiting for ring3 protection fault marker ('PF')...")
            if not check_log_for("PF", log_file, timeout=6):
                fail_with_log(
                    f"Did not observe PF marker after '{cmd_text}'.",
                    process,
                    log_file,
                )
            print(f"Test Passed: '{cmd_text}' triggered PF as expected.")
            return
        finally:
            stop_qemu(process)
            time.sleep(1)

    fail_with_log(
        f"Timeout waiting for boot prompt while running '{cmd_text}' after retries.",
        last_process,
        log_file,
    )


def main():
    iso_path = "build/dav-go-os.iso"
    disk_img = "disk.img"

    if not os.path.exists(iso_path):
        print(f"ERROR: ISO not found at {iso_path}. Build it first.")
        sys.exit(1)

    create_disk_image(disk_img)

    print(f"Starting QEMU verification for {iso_path}...")
    run_functional_suite(iso_path, disk_img, "qemu.log")

    # Each probe must run in its own VM instance because a #PF is terminal here.
    kread_disk = "disk_kread.img"
    kwrite_disk = "disk_kwrite.img"
    create_disk_image(kread_disk)
    create_disk_image(kwrite_disk)
    run_fault_probe(iso_path, kread_disk, "run kread", "qemu_kread.log")
    run_fault_probe(iso_path, kwrite_disk, "run kwrite", "qemu_kwrite.log")

    print("All QEMU checks passed.")


if __name__ == "__main__":
    main()
