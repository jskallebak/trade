// web/static/js/login.js
class LoginForm {
    constructor() {
        this.form = document.getElementById('loginForm');
        this.loginBtn = document.getElementById('loginBtn');
        this.errorMessage = document.getElementById('errorMessage');
        this.successMessage = document.getElementById('successMessage');
        this.emailInput = document.getElementById('email');
        this.passwordInput = document.getElementById('password');
    }

    init() {
        this.setupEventListeners();
        this.checkExistingAuth();
    }

    setupEventListeners() {
        this.form.addEventListener('submit', (e) => this.handleSubmit(e));

        // Enter key handling for better UX
        [this.emailInput, this.passwordInput].forEach(input => {
            input.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.form.dispatchEvent(new Event('submit'));
                }
            });
        });
    }

    async handleSubmit(e) {
        e.preventDefault();

        this.hideMessages();
        this.setLoading(true);

        const credentials = {
            email: this.emailInput.value.trim(),
            password: this.passwordInput.value
        };

        if (!this.validateCredentials(credentials)) {
            this.setLoading(false);
            return;
        }

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(credentials)
            });

            const data = await response.json();

            if (response.ok) {
                await this.handleLoginSuccess(data);
            } else {
                this.handleLoginError(data.message || 'Login failed. Please check your credentials.');
            }
        } catch (error) {
            console.error('Login error:', error);
            this.handleLoginError('Network error. Please try again.');
        } finally {
            this.setLoading(false);
        }
    }

    validateCredentials(credentials) {
        if (!credentials.email) {
            this.showError('Email address is required.');
            this.emailInput.focus();
            return false;
        }

        if (!this.isValidEmail(credentials.email)) {
            this.showError('Please enter a valid email address.');
            this.emailInput.focus();
            return false;
        }

        if (!credentials.password) {
            this.showError('Password is required.');
            this.passwordInput.focus();
            return false;
        }

        return true;
    }

    isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    async handleLoginSuccess(data) {
        // Store JWT token
        localStorage.setItem('authToken', data.token);

        // Show success message
        this.showSuccess('Login successful! Redirecting...');

        // Clear form
        this.form.reset();

        // Redirect to dashboard
        setTimeout(() => {
            window.location.href = '/';
        }, 1500);
    }

    handleLoginError(message) {
        this.showError(message);
        this.passwordInput.focus();
        this.passwordInput.select();
    }

    setLoading(isLoading) {
        this.loginBtn.disabled = isLoading;
        this.loginBtn.textContent = isLoading ? 'Signing in...' : 'Sign In';

        // Disable inputs during loading
        this.emailInput.disabled = isLoading;
        this.passwordInput.disabled = isLoading;
    }

    hideMessages() {
        this.errorMessage.style.display = 'none';
        this.successMessage.style.display = 'none';
    }

    showError(message) {
        this.errorMessage.textContent = message;
        this.errorMessage.style.display = 'block';
        this.successMessage.style.display = 'none';
    }

    showSuccess(message) {
        this.successMessage.textContent = message;
        this.successMessage.style.display = 'block';
        this.errorMessage.style.display = 'none';
    }

    checkExistingAuth() {
        // If user is already logged in, redirect to dashboard
        const token = localStorage.getItem('authToken');
        if (token) {
            // Optionally verify token is still valid
            this.verifyTokenAndRedirect(token);
        }
    }

    async verifyTokenAndRedirect(token) {
        try {
            const response = await fetch('/api/profile', {
                headers: {
                    'Authorization': 'Bearer ' + token
                }
            });

            if (response.ok) {
                // Token is valid, redirect to dashboard
                window.location.href = '/';
            } else {
                // Token is invalid, remove it
                localStorage.removeItem('authToken');
            }
        } catch (error) {
            // Network error or token invalid
            localStorage.removeItem('authToken');
        }
    }
}

// Initialize login form when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    const loginForm = new LoginForm();
    loginForm.init();
});
