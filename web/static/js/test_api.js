// web/static/app.js
document.addEventListener('DOMContentLoaded', function() {
    const testApiBtn = document.getElementById('testApiBtn');
    const apiResponse = document.getElementById('apiResponse');
    
    testApiBtn.addEventListener('click', async function() {
        try {
            testApiBtn.textContent = 'Testing...';
            testApiBtn.disabled = true;
            
            const response = await fetch('/api/hello');
            const data = await response.json();
            
            apiResponse.innerHTML = `
                <strong>API Response:</strong><br>
                ${JSON.stringify(data, null, 2)}
            `;
            apiResponse.style.display = 'block';
            
        } catch (error) {
            apiResponse.innerHTML = `
                <strong>Error:</strong><br>
                ${error.message}
            `;
            apiResponse.style.display = 'block';
        } finally {
            testApiBtn.textContent = 'Test API Endpoint';
            testApiBtn.disabled = false;
        }
    });
});
