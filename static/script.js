document.addEventListener('DOMContentLoaded', () => {
    // --- MD to DOCX Logic ---
    const dropZone = document.getElementById('drop-zone');
    const fileInput = document.getElementById('file-input');
    const fileInfo = document.getElementById('file-info');
    const filenameDisplay = document.getElementById('filename');
    const convertBtn = document.getElementById('convert-btn');
    const statusDiv = document.getElementById('conversion-status');
    let selectedFile = null;

    dropZone.addEventListener('click', () => fileInput.click());

    dropZone.addEventListener('dragover', (e) => {
        e.preventDefault();
        dropZone.classList.add('dragover');
    });

    dropZone.addEventListener('dragleave', () => {
        dropZone.classList.remove('dragover');
    });

    dropZone.addEventListener('drop', (e) => {
        e.preventDefault();
        dropZone.classList.remove('dragover');
        if (e.dataTransfer.files.length) {
            handleFile(e.dataTransfer.files[0]);
        }
    });

    fileInput.addEventListener('change', (e) => {
        if (e.target.files.length) {
            handleFile(e.target.files[0]);
        }
    });

    function handleFile(file) {
        if (!file.name.endsWith('.md')) {
            alert('Please select a Markdown (.md) file.');
            return;
        }
        selectedFile = file;
        filenameDisplay.textContent = file.name;
        fileInfo.classList.remove('hidden');
        dropZone.classList.add('hidden');
        statusDiv.textContent = '';
    }

    convertBtn.addEventListener('click', async () => {
        if (!selectedFile) return;
        statusDiv.textContent = 'Converting...';
        
        const formData = new FormData();
        formData.append('file', selectedFile);

        try {
            const response = await fetch('/md-to-docx/convert/', {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                // Trigger download
                const blob = await response.blob();
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = selectedFile.name.replace('.md', '.docx');
                document.body.appendChild(a);
                a.click();
                a.remove();
                statusDiv.innerHTML = '<span class="success-msg">Conversion successful! File downloaded.</span>';
            } else {
                statusDiv.innerHTML = '<span class="error-msg">Conversion failed.</span>';
            }
        } catch (err) {
            statusDiv.innerHTML = `<span class="error-msg">Error: ${err.message}</span>`;
        }
    });

    // Git logic removed per user request
});