from fastapi import FastAPI
from fastapi.responses import HTMLResponse
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
    title="Personnel Productivity API Catalog",
    description="A collection of personal productivity tools and services.",
    version="1.0.0"
)

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
async def catalog():
    return """
    <!DOCTYPE html>
    <html>
    <head>
        <title>Productivity API Catalog</title>
        <style>
            body { font_family: sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; line-height: 1.6; }
            h1 { color: #2c3e50; }
            .card { border: 1px solid #ddd; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
            .card h2 { margin-top: 0; }
            .btn { display: inline-block; background: #3498db; color: white; padding: 10px 15px; text-decoration: none; border-radius: 4px; }
            .btn:hover { background: #2980b9; }
        </style>
    </head>
    <body>
        <h1>🚀 Personnel Productivity Tools</h1>
        <p>Welcome to your personal API catalog. Below are the available services:</p>

        <div class="card">
            <h2>📄 Markdown to DOCX</h2>
            <p>Convert Markdown files (with Mermaid diagrams) to styled Word documents.</p>
            <p><strong>Endpoint:</strong> <code>POST /md-to-docx/convert/</code></p>
            <a href="/docs#/Markdown%20to%20DOCX/convert_markdown_to_docx_md_to_docx_convert__post" class="btn">Try it in OpenAPI</a>
        </div>

        <div class="card">
            <h2>📚 API Documentation</h2>
            <p>Explore the full interactive API documentation.</p>
            <a href="/docs" class="btn">Open Swagger UI</a>
            <a href="/redoc" class="btn" style="background: #e67e22;">Open ReDoc</a>
        </div>
    </body>
    </html>
    """

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8989)
