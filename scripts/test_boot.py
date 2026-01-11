import subprocess
import time
import sys
import os

def check_log_for(target, log_file, timeout=5):
    start = time.time()
    while time.time() - start < timeout:
        if os.path.exists(log_file):
            with open(log_file, 'r', errors='ignore') as f:
                content = f.read()
                if target in content:
                    return True
        time.sleep(0.5)
    return False

def main():
    iso_path = "build/dav-go-os.iso"
    log_file = "qemu.log"
    
    if os.path.exists(log_file):
        os.remove(log_file)
        
    print(f"Starting QEMU verification for {iso_path}...")
    print(f"Log file: {log_file}")
    
    # Run QEMU with debugcon logging to file and monitor on stdio
    # CRITIAL: We use 'file:qemu.log' to capture 0xE9 port output. 
    cmd = [
        "qemu-system-i386",
        "-cdrom", iso_path,
        "-debugcon", f"file:{log_file}",
        "-monitor", "stdio",
        "-display", "none",
        "-nographic" 
    ]
    
    # Start QEMU process
    # We must NOT use subprocess.PIPE for stdout/stderr if we want to interact with stdio monitor
    # But to capture errors, we need a way.
    # Approach: Use a pipe for stderr to capture errors, leave stdout/stdin for monitor interaction.
    
    process = subprocess.Popen(
        cmd,
        stdin=subprocess.PIPE,
        stdout=subprocess.DEVNULL, # We don't need stdout as we check the log file
        stderr=subprocess.PIPE,    # Capture errors
        text=True                  # Use text mode for easier reading
    )
    
    try:
        # 1. Wait for Boot Prompt "DavOS"
        print("Waiting for boot prompt...")
        if not check_log_for("DavOS", log_file, timeout=10):
            print("ERROR: Timeout waiting for DavOS prompt.")
            
            # Check if QEMU crashed
            if process.poll() is not None:
                print(f"QEMU process exited early with code {process.returncode}")
                stderr_out = process.stderr.read()
                print("--- QEMU stderr ---")
                print(stderr_out)
                print("-------------------")
            
            print("--- qemu.log content ---")
            if os.path.exists(log_file):
                with open(log_file, 'r', errors='ignore') as f:
                    print(f.read())
            else:
                print("qemu.log file not found.")
            print("------------------------")
            sys.exit(1)
        print("Boot successful.")
        
        # 2. Test Shell Command "help"
        # Since we are using -monitor stdio, writing to stdin interacts with the QEMU monitor.
        
        print("Sending 'help' command via QEMU monitor...")
        keys = ['h', 'e', 'l', 'p', 'ret']
        for k in keys:
            cmd_str = f"sendkey {k}\n"
            try:
                process.stdin.write(cmd_str)
                process.stdin.flush()
            except BrokenPipeError:
                 print("Error: QEMU closed stdin (crashed?)")
                 break
            time.sleep(0.1) 

        # 3. Wait for Command Output "Commands:"
        print("Waiting for help output...")
        if not check_log_for("Commands:", log_file, timeout=5):
            print("ERROR: Timeout waiting for help command text.")
            if process.poll() is not None:
                print(f"QEMU exited with {process.returncode}")
                print("Stderr:", process.stderr.read())
            sys.exit(1)
            
        print("Test Passed: 'help' command executed successfully.")
        
    finally:
        if process.poll() is None:
            process.terminate()

if __name__ == "__main__":
    main()
