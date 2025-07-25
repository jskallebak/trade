// web/static/js/dashboard.js
class Dashboard {
    constructor() {
        this.refreshInterval = null;
    }

    init() {
        this.loadDashboardData();
        this.setupEventListeners();
        this.startAutoRefresh();
    }

    loadDashboardData() {
        this.loadMetrics();
        this.loadBotStats();
        this.loadPositions();
    }

    async loadMetrics() {
        try {
            const response = await this.apiCall('/api/dashboard/metrics');
            const data = await response.json();
            
            this.updateElement('totalPnl', `$${Math.abs(data.total_pnl).toLocaleString()}`, 
                data.total_pnl >= 0 ? 'metric-value positive' : 'metric-value negative');
            this.updateElement('annualizedRoi', `${data.annualized_roi}%`);
            this.updateElement('maxDrawdown', `${data.max_drawdown}%`);
            this.updateElement('uptime', `${data.uptime}%`);
        } catch (error) {
            console.error('Error loading metrics:', error);
        }
    }

    async loadBotStats() {
        try {
            const response = await this.apiCall('/api/dashboard/bot-stats');
            const data = await response.json();
            
            const tbody = document.getElementById('botStatsTable');
            tbody.innerHTML = '';
            
            data.forEach(bot => {
                const row = this.createBotStatsRow(bot);
                tbody.appendChild(row);
            });
        } catch (error) {
            console.error('Error loading bot stats:', error);
        }
    }

    async loadPositions() {
        try {
            const response = await this.apiCall('/api/dashboard/positions');
            const data = await response.json();
            
            const tbody = document.getElementById('positionsTable');
            tbody.innerHTML = '';
            
            data.forEach(position => {
                const row = this.createPositionRow(position);
                tbody.appendChild(row);
            });
            
            if (data.length > 0) {
                const ellipsisRow = document.createElement('tr');
                ellipsisRow.innerHTML = '<td colspan="7" style="text-align: center; color: #666; padding: 20px;">...</td>';
                tbody.appendChild(ellipsisRow);
            }
        } catch (error) {
            console.error('Error loading positions:', error);
        }
    }

    createBotStatsRow(bot) {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${bot.name}</td>
            <td><span class="status-badge ${bot.status.toLowerCase()}">${bot.status}</span></td>
            <td>${bot.win_rate}%</td>
            <td>${bot.profit_factor}</td>
            <td>${bot.trades}</td>
            <td class="${bot.pnl >= 0 ? 'positive' : 'negative'}">
                ${bot.pnl >= 0 ? '$' : '-$'}${Math.abs(bot.pnl).toLocaleString()}
            </td>
        `;

        row.addEventListener('click', () => {
            this.showBotDetailsModal(bot);
        });

        return row;
    }

    async handleCreateBot() {
        const submitBtn = document.getElementById('submitBotBtn');
        const form = document.getElementById('createBotForm');
        
        // Get form data
        const formData = new FormData(form);
        const botData = {
            name: formData.get('botName').trim(),
            strategy: formData.get('botStrategy').trim(),
            initial_holding: parseFloat(formData.get('initialHolding')) || 0
        };

        // Validate required fields
        if (!botData.name || !botData.strategy) {
            alert('Bot name and strategy are required!');
            return;
        }

        try {
            // Show loading state
            submitBtn.textContent = 'Creating...';
            submitBtn.disabled = true;

            // Make API call
            const response = await this.apiCall('/api/bots', {
                method: 'POST',
                body: JSON.stringify(botData)
            });

            if (response.ok) {
                // Success - close modal and refresh bot list
                this.hideCreateBotModal();
                this.loadBotStats(); // Refresh the bot table
                // alert('Bot created successfully!');
            } else {
                const error = await response.text();
                alert('Error creating bot: ' + error);
            }
        } catch (error) {
            alert('Network error: ' + error.message);
        } finally {
            // Reset button
            submitBtn.textContent = 'Create Bot';
            submitBtn.disabled = false;
        }
    }

    createPositionRow(position) {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${position.trade_id}</td>
            <td>${position.bot}</td>
            <td><span class="position-badge ${position.position.toLowerCase()}">${position.position}</span></td>
            <td>$${parseFloat(position.entry).toLocaleString()}</td>
            <td>$${parseFloat(position.current).toLocaleString()}</td>
            <td class="${position.pnl >= 0 ? 'positive' : 'negative'}">
                ${position.pnl >= 0 ? '$' : '-$'}${Math.abs(position.pnl).toLocaleString()}
            </td>
            <td>${position.time}</td>
        `;
        return row;
    }

    showCreateBotModal() {
        const modal = document.getElementById('createBotModal');
        if (modal) {
            modal.style.display = 'flex';
            // Clear form
            document.getElementById('createBotForm').reset();
        }
    }

    hideCreateBotModal() {
        const modal = document.getElementById('createBotModal');
        if (modal) {
            modal.style.display = 'none';
        }
    }

    setupEventListeners() {
        // Logout functionality
        document.getElementById('logoutBtn')?.addEventListener('click', () => {
            localStorage.removeItem('authToken');
            window.location.href = '/logout';
        });

        // Debug API test
        document.getElementById('testApiBtn')?.addEventListener('click', () => {
            this.testApiEndpoint();
        });

        // Profile button
        document.getElementById('profileBtn')?.addEventListener('click', () => {
            this.loadProfile();
        });

        document.getElementById("createBotBtn")?.addEventListener('click', () => {
            this.showCreateBotModal();
        });

        
        // Modal close events
        document.getElementById('closeBotModal')?.addEventListener('click', () => {
            this.hideCreateBotModal();
        });

        document.getElementById('cancelBotBtn')?.addEventListener('click', () => {
            this.hideCreateBotModal();
        });

        // Close modal when clicking outside
        document.getElementById('createBotModal')?.addEventListener('click', (e) => {
            if (e.target.id === 'createBotModal') {
                this.hideCreateBotModal();
            }
        });

        // Bot Details Modal close events
        document.getElementById('closeBotDetailsModal')?.addEventListener('click', () => {
            this.hideBotDetailsModal();
        });

        document.getElementById('cancelEditBtn')?.addEventListener('click', () => {
            this.hideBotDetailsModal();
        });

        // Close modal when clicking outside
        document.getElementById('botDetailsModal')?.addEventListener('click', (e) => {
            if (e.target.id === 'botDetailsModal') {
                this.hideBotDetailsModal();
            }
        });

                // Bot status control buttons
        document.getElementById('startBotBtn')?.addEventListener('click', () => {
            this.updateBotStatus('RUNNING');
        });

        document.getElementById('stopBotBtn')?.addEventListener('click', () => {
            this.updateBotStatus('STOPPED');
        });

        document.getElementById('pauseBotBtn')?.addEventListener('click', () => {
            this.updateBotStatus('PAUSED');
        });

        // Delete bot button
        document.getElementById('deleteBotBtn')?.addEventListener('click', () => {
            this.deleteBotConfirm();
        })

            // Edit bot form submission
        document.getElementById('editBotForm')?.addEventListener('submit', (e) => {
            e.preventDefault();
            this.saveChanges();
        });

        const form = document.getElementById('createBotForm');
        
        form?.addEventListener('submit', (e) => {
            console.log('Form submit event triggered!'); // ← Add this
            e.preventDefault();
            this.handleCreateBot();
        });


    }

    async saveChanges() {
        if (!this.currentBot) {
            alert('No bot selected');
            return;
        }

        const saveBtn = document.getElementById('saveChangesBtn');
        const form = document.getElementById('editBotForm');

        // Get form data
        const formData = new FormData(form);
        const updatedData = {
            name: formData.get('editBotName').trim(),
            strategy: formData.get('editBotStrategy').trim(),
            initial_holding: parseFloat(formData.get('editInitialHolding')) || 0
        };

        // Validate required fields
        if (!updatedData.name || !updatedData.strategy) {
            alert('Bot name and strategy are required!');
            return;
        }

        try {
            // Show loading state
            saveBtn.textContent = 'Saving...';
            saveBtn.disabled = true;

            // Make API call
            const response = await this.apiCall(`/api/bots/${this.currentBot.id}`, {
                method: 'PUT',
                body: JSON.stringify(updatedData)
            });

            if (response.ok) {
                const updatedBot = await response.json();
                
                // Update the current bot object
                this.currentBot = updatedBot;
                
                // Refresh the bot table to show updated data
                this.loadBotStats();
                
                console.log('Bot updated successfully');
                
                // Optional: Brief success feedback
                saveBtn.textContent = 'Saved!';
                setTimeout(() => {
                    saveBtn.textContent = 'Save Changes';
                }, 1000);
                
            } else {
                const error = await response.text();
                alert('Error updating bot: ' + error);
            }
        } catch (error) {
            alert('Network error: ' + error.message);
        } finally {
            // Reset button
            saveBtn.disabled = false;
            setTimeout(() => {
                if (saveBtn.textContent === 'Saving...') {
                    saveBtn.textContent = 'Save Changes';
                }
            }, 1000);
        }
    }

    deleteBotConfirm() {
        if (!this.currentBot) {
            alert('No bot selected');
            return;
        }

        // Show confirmation dialog
        const confirmed = confirm(`Are you sure you want to delete "${this.currentBot.name}"?\n\nThis action cannot be undone.`);
        
        if (confirmed) {
            this.deleteBot();
        }
    }

    async deleteBot() {
        if (!this.currentBot) {
            return;
        }

        const deleteBtn = document.getElementById('deleteBotBtn');
        const botName = this.currentBot.name;

        try {
            // Show loading state
            deleteBtn.textContent = 'Deleting...';
            deleteBtn.disabled = true;

            // Make API call
            const response = await this.apiCall(`/api/bots/${this.currentBot.id}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                this.currentBot = null;
                // Success - close modal and refresh bot list
                this.hideBotDetailsModal();
                this.loadBotStats(); // Refresh the bot table
                console.log(`Bot "${botName}" deleted successfully`);
            } else {
                const error = await response.text();
                alert('Error deleting bot: ' + error);
            }
        } catch (error) {
            alert('Network error: ' + error.message);
        } finally {
            // Reset button
            deleteBtn.textContent = 'Delete Bot';
            deleteBtn.disabled = false;
        }
    }

    async updateBotStatus(newStatus) {
        if (!this.currentBot) {
            alert('No bot selected');
            return;
        }

        try {
            const response = await this.apiCall(`/api/bots/${this.currentBot.id}/status`, {
                method: 'PUT',
                body: JSON.stringify({ status: newStatus })
            });

            if (response.ok) {
                // Update the current bot object
                this.currentBot.status = newStatus;
                
                // Update the status display in the modal
                const statusElement = document.getElementById('botCurrentStatus');
                statusElement.textContent = newStatus;
                statusElement.className = `status-badge ${newStatus.toLowerCase()}`;
                
                // Refresh the bot table to show updated status
                this.loadBotStats();
                
                console.log(`Bot status updated to ${newStatus}`);
            } else {
                const error = await response.text();
                alert('Error updating bot status: ' + error);
            }
        } catch (error) {
            alert('Network error: ' + error.message);
        }
    }

    async testApiEndpoint() {
        const testApiBtn = document.getElementById('testApiBtn');
        const apiResponse = document.getElementById('apiResponse');
        
        try {
            testApiBtn.textContent = 'Testing...';
            testApiBtn.disabled = true;
            
            const response = await this.apiCall('/api/hello');
            const data = await response.json();
            
            apiResponse.innerHTML = `<strong>API Response:</strong><br>${JSON.stringify(data, null, 2)}`;
            apiResponse.style.display = 'block';
        } catch (error) {
            apiResponse.innerHTML = `<strong>Error:</strong><br>${error.message}`;
            apiResponse.style.display = 'block';
        } finally {
            testApiBtn.textContent = 'Test API Endpoint';
            testApiBtn.disabled = false;
        }
    }

    async loadProfile() {
        const profileBtn = document.getElementById('profileBtn');
        const profileResponse = document.getElementById('profileResponse');
        
        try {
            profileBtn.textContent = 'Loading...';
            profileBtn.disabled = true;
            
            const response = await this.apiCall('/api/profile');
            
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
    }

    startAutoRefresh() {
        // Refresh data every 30 seconds
        this.refreshInterval = setInterval(() => {
            this.loadDashboardData();
        }, 30000);
    }

    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    // Utility methods
    apiCall(url, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };

        const authToken = localStorage.getItem('authToken');
        if (authToken) {
            headers['Authorization'] = 'Bearer ' + authToken
        }

        return fetch(url, {
            ...options,
            headers
        });
    }

    updateElement(id, content, className = null) {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = content;
            if (className) {
                element.className = className;
            }
        }
    }

    showBotDetailsModal(bot) {
        console.log('Opening bot details for:', bot); // Debug log
        console.log('Bot object received:', bot);  // ← Add this
        console.log('Bot ID:', bot.id);
        
        const modal = document.getElementById('botDetailsModal');
        if (!modal) {
            console.error('Bot details modal not found!');
            return;
        }

        // Store the current bot data for later use
        this.currentBot = bot;

        // Populate the form with current bot data
        document.getElementById('editBotName').value = bot.name || '';
        document.getElementById('editBotStrategy').value = bot.strategy || '';
        document.getElementById('editInitialHolding').value = bot.initial_holding || 0;

        // Update status display
        const statusElement = document.getElementById('botCurrentStatus');
        if (statusElement) {
            statusElement.textContent = bot.status || 'UNKNOWN';
            statusElement.className = `status-badge ${(bot.status || '').toLowerCase()}`;
        }

        // Update read-only stats
        document.getElementById('displayWinRate').textContent = `${bot.win_rate || 0}%`;
        document.getElementById('displayProfitFactor').textContent = bot.profit_factor || 0;
        document.getElementById('displayTrades').textContent = bot.trades || 0;
        
        // Format P&L
        const pnl = bot.pnl || 0;
        const pnlElement = document.getElementById('displayPnL');
        pnlElement.textContent = pnl >= 0 ? `$${Math.abs(pnl).toLocaleString()}` : `-$${Math.abs(pnl).toLocaleString()}`;
        pnlElement.className = pnl >= 0 ? 'positive' : 'negative';

        // Show the modal
        modal.style.display = 'flex';
    }

    hideBotDetailsModal() {
        const modal = document.getElementById('botDetailsModal');
        if (modal) {
            modal.style.display = 'none';
        }
        this.currentBot = null;
    }
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    const dashboard = new Dashboard();
    dashboard.init();
    
    // Clean up on page unload
    window.addEventListener('beforeunload', () => {
        dashboard.stopAutoRefresh();
    });
});
