from fastapi import APIRouter, UploadFile, File, HTTPException
from fastapi.responses import FileResponse
import shutil
import os
import subprocess
import uuid
from pathlib import Path
import re

router = APIRouter(prefix="/md-to-docx", tags=["Markdown to DOCX"])

UPLOAD_DIR = Path("/tmp/uploads")
OUTPUT_DIR = Path("/tmp/outputs")
UPLOAD_DIR.mkdir(parents=True, exist_ok=True)
OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

# Helper: Preprocess Markdown for Mermaid
def preprocess_markdown(file_path):
    """
    Reads markdown file, converts <div class="mermaid"> to fenced code blocks,
    and returns path to a temporary file.
    """
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Regex to transform <div class="mermaid">...</div> into ```mermaid...```
        pattern = re.compile(r'<div class="mermaid">\s*(.*?)\s*</div>', re.DOTALL)
        
        def replacement(match):
            code = match.group(1).strip()
            return f"\n```mermaid\n{code}\n```\n"
        
        new_content = pattern.sub(replacement, content)
        
        base = os.path.splitext(file_path)[0]
        temp_path = f"{base}_processed.md"
        
        with open(temp_path, 'w', encoding='utf-8') as f:
            f.write(new_content)
        
        return temp_path
    except Exception as e:
        print(f"Error preprocessing markdown: {e}")
        return file_path

@router.post("/convert/")
async def convert_markdown_to_docx(file: UploadFile = File(...)):
    if not file.filename.endswith(".md"):
        raise HTTPException(status_code=400, detail="Only .md files are allowed")

    request_id = str(uuid.uuid4())
    input_path = UPLOAD_DIR / f"{request_id}_{file.filename}"
    output_filename = f"{Path(file.filename).stem}.docx"
    output_path = OUTPUT_DIR / f"{request_id}_{output_filename}"

    # Save uploaded file
    with open(input_path, "wb") as buffer:
        shutil.copyfileobj(file.file, buffer)

    # Preprocess
    processed_path = preprocess_markdown(str(input_path))

    # Puppeteer Config for Mermaid Filter (needs to be in CWD for filter)
    # We will create it temporarily in the working directory
    puppeteer_config = {
        "executablePath": "/usr/bin/google-chrome",
        "args": ["--no-sandbox", "--disable-setuid-sandbox"]
    }
    import json
    config_path = Path("puppeteer-config.json")
    with open(config_path, "w") as f:
        json.dump(puppeteer_config, f)

    # Convert using Pandoc
    # Command: pandoc input.md -o output.docx -F mermaid-filter
    cmd = [
        "pandoc",
        processed_path,
        "-o", str(output_path),
        "-F", "mermaid-filter"
    ]

    # TODO: Add reference doc support if uploaded or default exists

    try:
        # We need to run this where the puppeteer config is visible if mermaid-filter looks for it in CWD
        subprocess.run(cmd, check=True, cwd=os.getcwd())
    except subprocess.CalledProcessError as e:
        raise HTTPException(status_code=500, detail=f"Conversion failed: {str(e)}")
    finally:
        # Cleanup input files
        if os.path.exists(input_path):
            os.remove(input_path)
            if input_path != Path(processed_path) and os.path.exists(processed_path):
                 os.remove(processed_path)

    return FileResponse(
        path=output_path, 
        filename=output_filename, 
        media_type='application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    )

@router.get("/health")
def health_check():
    return {"status": "ok"}
