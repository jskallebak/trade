// Profile button
const profileBtn = document.getElementById('profileBtn');
const profileResponse = document.getElementById('profileResponse');

profileBtn.addEventListener('click', async function() {
    try {
        profileBtn.textContent = 'Loading...';
        profileBtn.disabled = true;

        const authToken = localStorage.getItem('authToken');
        const headers = {};
        if (authToken) {
            headers['Authorization'] = 'Bearer ' + authToken;
        }
        
        const response = await fetch('/api/profile', {
            headers: headers
        });
        
        if (response.ok) {
            const userData = await response.json();
            profileResponse.innerHTML = `
                <strong>Your Profile:</strong><br>
                <strong>ID:</strong> ${userData.id}<br>
                <strong>Name:</strong> ${userData.name}<br>
                <strong>Email:</strong> ${userData.email}<br>
                <strong>Created:</strong> ${new Date(userData.created_at).toLocaleDateString()}
            `;
            profileResponse.style.display = 'block';
        } else {
            profileResponse.innerHTML = '<strong>Error:</strong><br>Failed to load profile';
            profileResponse.style.display = 'block';
        }
        
    } catch (error) {
        profileResponse.innerHTML = `<strong>Error:</strong><br>${error.message}`;
        profileResponse.style.display = 'block';
    } finally {
        profileBtn.textContent = 'View Profile';
        profileBtn.disabled = false;
    }
});
