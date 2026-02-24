import subprocess
import sys
import argparse
import time
import webbrowser

IMAGE_NAME = "personnel-productivity"
CONTAINER_NAME = "personnel-app"
PORT = 8989

def run_command(command, check=True):
    print(f"Running: {command}")
    try:
        subprocess.run(command, shell=True, check=check)
        return True
    except subprocess.CalledProcessError as e:
        print(f"Error: {e}")
        return False

def build_image():
    print("--- Building Docker Image ---")
    run_command(f"docker build -t {IMAGE_NAME} .")

def stop_container():
    print(f"--- Stopping Container '{CONTAINER_NAME}' ---")
    # check if running
    run_command(f"docker stop {CONTAINER_NAME}", check=False)
    run_command(f"docker rm {CONTAINER_NAME}", check=False)

def start_container(background=True):
    stop_container() # Ensure clean state
    print(f"--- Starting Container '{CONTAINER_NAME}' ---")
    
    # We mount the current directory to /app so we can develop live if needed (with reload)
    # But for a stable "spin up", we might rely on COPY in Dockerfile.
    # Let's assume production-like run for "spin up".
    
    detach_flag = "-d" if background else ""
    
    # Check if we should mount for development
    # volume_mount = f"-v \"{os.getcwd()}:/app\"" 
    # For now, let's rely on the built image content to ensure consistency.
    
    cmd = (
        f"docker run {detach_flag} "
        f"--name {CONTAINER_NAME} "
        f"-p {PORT}:{PORT} "
        f"{IMAGE_NAME} "
        f"uvicorn main:app --host 0.0.0.0 --port {PORT}"
    )
    
    if run_command(cmd):
        print(f"✅ Container started on port {PORT}")
        print(f"   Catalog: http://localhost:{PORT}")
        print(f"   Docs:    http://localhost:{PORT}/docs")
        
        if background:
            time.sleep(2) # Wait a bit for startup
            webbrowser.open(f"http://localhost:{PORT}")

def logs():
    run_command(f"docker logs -f {CONTAINER_NAME}")

def main():
    parser = argparse.ArgumentParser(description="Docker Lifecycle Manager")
    subparsers = parser.add_subparsers(dest="action", help="Action to perform")
    
    subparsers.add_parser("build", help="Build the Docker image")
    subparsers.add_parser("up", help="Start the container (background)")
    subparsers.add_parser("down", help="Stop and remove the container")
    subparsers.add_parser("logs", help="View container logs")
    subparsers.add_parser("restart", help="Stop, Build, and Start")

    args = parser.parse_args()

    if args.action == "build":
        build_image()
    elif args.action == "up":
        start_container()
    elif args.action == "down":
        stop_container()
    elif args.action == "logs":
        logs()
    elif args.action == "restart":
        stop_container()
        build_image()
        start_container()
    else:
        parser.print_help()

if __name__ == "__main__":
    main()
