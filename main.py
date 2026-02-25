from fastapi import FastAPI
from fastapi.responses import HTMLResponse, FileResponse
from fastapi.staticfiles import StaticFiles
import uvicorn
import importlib.util
import sys
from pathlib import Path

# Add modules to path
sys.path.append(str(Path(__file__).parent))

# Import the md-to-docx router
# We use dynamic import below because the folder name has hyphens
# from md_to_docx import router as md_docs_router

app = FastAPI(
    title="Personnel Productivity Dashboard",
    description="A collection of personal productivity tools and services.",
    version="1.0.0"
)

app.mount("/static", StaticFiles(directory="static"), name="static")

# Mount Routers
# Note: md-to-docx/router.py needs to be importable.
# Since 'md-to-docx' has dashes, it's not a standard package name.
# We might need to rename the folder or use importlib.
# Let's try dynamic import for the folder with dashes
md_to_docx_path = Path("./md-to-docx/router.py")

spec = importlib.util.spec_from_file_location("md_to_docx_router", md_to_docx_path)
md_to_docx_module = importlib.util.module_from_spec(spec)
sys.modules["md_to_docx_router"] = md_to_docx_module
spec.loader.exec_module(md_to_docx_module)

app.include_router(md_to_docx_module.router)


@app.get("/", response_class=HTMLResponse)
async def dashboard():
    return FileResponse("templates/index.html")

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8989)
