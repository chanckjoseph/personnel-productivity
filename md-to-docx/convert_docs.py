import os
import subprocess
import sys
import re
import shutil

# Configuration
DOCKER_IMAGE_NAME = "covpay-docs-builder"
DOCKERFILE_PATH = "Dockerfile.pandoc"

def run_command(command, cwd=None):
    """Running a shell command and printing output."""
    print(f"Running: {command}")
    try:
        result = subprocess.run(
            command,
            shell=True,
            check=True,
            cwd=cwd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        print(result.stdout)
        return True
    except subprocess.CalledProcessError as e:
        print(f"Error executing command: {command}")
        print(e.stderr)
        return False

def check_docker_image_exists(image_name):
    """Check if the docker image exists locally."""
    try:
        result = subprocess.run(
            f"docker images -q {image_name}",
            shell=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        return bool(result.stdout.strip())
    except Exception:
        return False

def build_docker_image():
    """Build the docker image if it doesn't exist."""
    print(f"Building Docker image '{DOCKER_IMAGE_NAME}'...")
    if not os.path.exists(DOCKERFILE_PATH):
        print(f"Error: {DOCKERFILE_PATH} not found.")
        sys.exit(1)
    
    success = run_command(f"docker build -t {DOCKER_IMAGE_NAME} -f {DOCKERFILE_PATH} .")
    if not success:
        print("Failed to build Docker image.")
        sys.exit(1)
    print("Docker image built successfully.")

def preprocess_markdown(file_path):
    """
    Reads markdown file, converts <div class="mermaid"> to fenced code blocks,
    and returns path to a temporary file.
    """
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Regex to transform <div class="mermaid">...</div> into ```mermaid...```
    # The flag re.DOTALL is essential so . matches newlines
    pattern = re.compile(r'<div class="mermaid">\s*(.*?)\s*</div>', re.DOTALL)
    
    def replacement(match):
        code = match.group(1).strip()
        return f"\n```mermaid\n{code}\n```\n"
    
    new_content = pattern.sub(replacement, content)
    
    base, ext = os.path.splitext(file_path)
    temp_path = f"{base}_temp{ext}"
    
    with open(temp_path, 'w', encoding='utf-8') as f:
        f.write(new_content)
    
    return temp_path

def convert_file(input_file):
    """
    Converts a single markdown file to docx using the docker container.
    """
    if not os.path.exists(input_file):
        print(f"File not found: {input_file}")
        return

    print(f"\nProcessing {input_file}...")
    
    # 1. Preprocess (handle custom mermaid divs)
    temp_file = preprocess_markdown(input_file)
    output_file = os.path.splitext(input_file)[0] + ".docx"
    
    # Get absolute paths for Docker volume mounting
    current_dir = os.getcwd()
    temp_filename = os.path.basename(temp_file)
    output_filename = os.path.basename(output_file)
    
    # 2. Run Docker Container
    # We pass the puppeteer config via -p. We must ensure the file exists or construct one.
    # But wait, pandoc filter arguments are passed differently. 
    # mermaid-filter supports a .puppeteer.json file in the current working directory.
    # We will write a temporary .puppeteer.json to the current directory (which is mounted)
    
    puppeteer_config = {
        "executablePath": "/usr/bin/google-chrome",
        "args": ["--no-sandbox", "--disable-setuid-sandbox"]
    }
    
    # Write puppeteer config to cwd so mermaid-filter picks it up
    puppeteer_config_path = os.path.join(current_dir, ".puppeteer.json")
    with open(puppeteer_config_path, 'w') as f:
        import json
        json.dump(puppeteer_config, f)
        
    # Check for reference doc
    reference_doc = "reference.docx"
    reference_arg = ""
    if os.path.exists(os.path.join(current_dir, reference_doc)):
        print(f"Using style reference: {reference_doc}")
        reference_arg = f'--reference-doc="{reference_doc}"'
        
    docker_cmd = (
        f'docker run --rm '
        f'-v "{current_dir}:/data" '
        f'{DOCKER_IMAGE_NAME} '
        f'"{temp_filename}" -o "{output_filename}" '
        f'-F mermaid-filter {reference_arg}'
    )
    
    try:
        success = run_command(docker_cmd)
    finally:
        # Cleanup config
        if os.path.exists(puppeteer_config_path):
            os.remove(puppeteer_config_path)
    
    # 3. Cleanup
    if os.path.exists(temp_file):
        os.remove(temp_file)
        
    if success:
        print(f"✅ Successfully converted: {output_file}")
    else:
        print(f"❌ Failed to convert: {input_file}")

if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description="Convert Markdown to DOCX with Mermaid support.")
    parser.add_argument('files', nargs='*', help="Markdown files to convert.")
    parser.add_argument('--generate-reference', action='store_true', help="Generate a default reference.docx for styling.")
    args = parser.parse_args()

    # 1. Check/Build Docker Image
    if not check_docker_image_exists(DOCKER_IMAGE_NAME):
        build_docker_image()
    else:
        print(f"Docker image '{DOCKER_IMAGE_NAME}' found.")

    # Feature: Generate default reference doc if requested
    if args.generate_reference:
        print("Generating 'reference.docx' template from Pandoc default...")
        current_dir = os.getcwd()
        docker_cmd = (
            f'docker run --rm '
            f'-v "{current_dir}:/data" '
            f'{DOCKER_IMAGE_NAME} '
            f'-o "reference.docx" --print-default-data-file reference.docx'
        )
        run_command(docker_cmd)
        print("Created 'reference.docx'. Open this file in Word, modify the Styles (Normal, Heading 1, etc.), save it, and run the conversion again.")
        sys.exit(0)

    # 2. Determine files to process
    files_to_convert = args.files if args.files else [
        "CoVPay_Strategic_Analysis.md",
        "CoVPay_technical_spec.md"
    ]

    # 3. Process files
    for file in files_to_convert:
        if os.path.exists(file):
            convert_file(file)
        else:
            print(f"Skipping missing file: {file}")
