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
        this.loadBinanceAccounts();
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
            console.log('Bot stats data:', data); // Add this debug log

            const tbody = document.getElementById('botStatsTable');
            tbody.innerHTML = '';

            data.forEach(bot => {
                console.log('Bot data:', bot); // Add this too
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

    showCreateAccountModal() {
        const modal = document.getElementById('createAccountModal');
        if (modal) {
            modal.style.display = 'flex';
            // Clear form
            document.getElementById('createAccountForm').reset();
        }
    }

    hideCreateAccountModal() {
        const modal = document.getElementById('createAccountModal');
        if (modal) {
            modal.style.display = 'none';
        }
    }

    async handleCreateAccount() {
        const submitBtn = document.getElementById('submitAccountBtn');
        const form = document.getElementById('createAccountForm');

        // Get form data
        const formData = new FormData(form);
        const accountData = {
            name: formData.get('accountName').trim(),
            api_key: formData.get('apiKey').trim(),
            api_secret: formData.get('apiSecret').trim(),
            base_url: formData.get('baseUrl').trim() || 'https://api.binance.com'
        };

        // Validate required fields
        if (!accountData.name || !accountData.api_key || !accountData.api_secret) {
            alert('Account name, API key, and API secret are required!');
            return;
        }

        try {
            // Show loading state
            submitBtn.textContent = 'Adding...';
            submitBtn.disabled = true;

            // Make API call
            const response = await this.apiCall('/api/binance-accounts', {
                method: 'POST',
                body: JSON.stringify(accountData)
            });

            if (response.ok) {
                // Success - close modal and refresh accounts list
                this.hideCreateAccountModal();
                this.loadBinanceAccounts(); // Refresh the accounts table
                console.log('Account added successfully');
            } else {
                const error = await response.text();
                alert('Error adding account: ' + error);
            }
        } catch (error) {
            alert('Network error: ' + error.message);
        } finally {
            // Reset button
            submitBtn.textContent = 'Add Account';
            submitBtn.disabled = false;
        }
    }

    async deleteAccount(accountId) {
        // Show confirmation dialog
        const confirmed = confirm('Are you sure you want to delete this Binance account?\n\nThis action cannot be undone.');

        if (confirmed) {
            try {
                const response = await this.apiCall(`/api/binance-accounts/${accountId}`, {
                    method: 'DELETE'
                });

                if (response.ok) {
                    this.loadBinanceAccounts(); // Refresh the accounts table
                    console.log('Account deleted successfully');
                } else {
                    const error = await response.text();
                    alert('Error deleting account: ' + error);
                }
            } catch (error) {
                alert('Network error: ' + error.message);
            }
        }
    }

    async loadBinanceAccounts() {
        try {
            const response = await this.apiCall('/api/binance-accounts');
            const data = await response.json();

            const tbody = document.getElementById('binanceAccountsTable');
            if (tbody) {
                // Store existing balance values before clearing the table
                const existingBalances = {};
                const existingBalanceElements = tbody.querySelectorAll('[id^="balance-"]');
                existingBalanceElements.forEach(el => {
                    const accountId = el.id.replace('balance-', '');
                    existingBalances[accountId] = el.textContent;
                });

                tbody.innerHTML = '';

                if (data.length === 0) {
                    const emptyRow = document.createElement('tr');
                    emptyRow.innerHTML = '<td colspan="5" style="text-align: center; color: #666; padding: 20px;">No accounts configured</td>';
                    tbody.appendChild(emptyRow);
                } else {
                    // Create rows and restore existing balances
                    for (const account of data) {
                        const row = this.createAccountRow(account);
                        tbody.appendChild(row);

                        // Restore previous balance if it exists and is valid
                        const previousBalance = existingBalances[account.id];
                        if (previousBalance && previousBalance !== 'Loading...' && !previousBalance.includes('Error') && !previousBalance.includes('Network')) {
                            const balanceElement = document.getElementById(`balance-${account.id}`);
                            if (balanceElement) {
                                balanceElement.textContent = previousBalance;
                            }
                        }
                    }
                }
            }
        } catch (error) {
            console.error('Error loading Binance accounts:', error);
            const tbody = document.getElementById('binanceAccountsTable');
            if (tbody) {
                tbody.innerHTML = '<tr><td colspan="5" style="text-align: center; color: #e53e3e; padding: 20px;">Error loading accounts</td></tr>';
            }
        }
    }

    async loadAccountBalance(accountId, key, secret) {
        try {
            const accountData = {
                key: key,
                secret: secret
            };

            const response = await this.apiCall('/api/get-margin-account-info', {
                method: 'POST',
                body: JSON.stringify(accountData)
            });

            const balanceElement = document.getElementById(`balance-${accountId}`);

            if (!balanceElement) {
                console.warn(`Balance element for account ${accountId} not found`);
                return;
            }

            if (response.ok) {
                const accInfo = await response.json();
                const balance = parseFloat(accInfo.TotalNetAssetOfUSDT || accInfo.totalNetAssetOfUsdt || 0);

                // Update the balance (this will replace "Loading..." or update existing balance)
                balanceElement.textContent = `$${balance.toLocaleString('en-US', {
                    minimumFractionDigits: 2,
                    maximumFractionDigits: 2
                })}`;

                // Reset any error styling that might have been applied
                balanceElement.style.color = '';

            } else {
                const error = await response.text();
                console.error('Error getting account info:', error);

                // Only show error if we don't have a previous valid balance
                if (balanceElement.textContent === 'Loading...' || balanceElement.textContent.includes('Error') || balanceElement.textContent.includes('Network')) {
                    balanceElement.textContent = 'Error loading';
                    balanceElement.style.color = '#e53e3e';
                }
                // If we have a previous valid balance (starts with $), keep it
            }
        } catch (error) {
            console.error('Network error loading account balance:', error);
            const balanceElement = document.getElementById(`balance-${accountId}`);
            if (balanceElement) {
                // Only show network error if we don't have a previous valid balance
                if (balanceElement.textContent === 'Loading...' || balanceElement.textContent.includes('Error') || balanceElement.textContent.includes('Network')) {
                    balanceElement.textContent = 'Network Error';
                    balanceElement.style.color = '#e53e3e';
                }
                // If we have a previous valid balance (starts with $), keep it
            }
        }
    }

    showEditAccountModal(account) {
        const modal = document.getElementById('editAccountModal');
        if (modal) {
            // Store the current account for later use
            this.currentEditAccount = account;

            // Pre-fill the form with current values
            document.getElementById('editAccountName').value = account.name;
            document.getElementById('editBaseUrl').value = account.base_url;

            // Show the modal
            modal.style.display = 'flex';
        }
    }

    hideEditAccountModal() {
        const modal = document.getElementById('editAccountModal');
        if (modal) {
            modal.style.display = 'none';
            this.currentEditAccount = null; // Clear the stored account
        }
    }

    async handleEditAccount() {
        if (!this.currentEditAccount) {
            alert('No account selected for editing');
            return;
        }

        const saveBtn = document.getElementById('saveAccountBtn');
        const form = document.getElementById('editAccountForm');

        // Get form data
        const formData = new FormData(form);
        const updatedData = {
            name: formData.get('editAccountName').trim(),
            base_url: formData.get('editBaseUrl').trim()
        };

        // Validate required fields
        if (!updatedData.name) {
            alert('Account name is required!');
            return;
        }

        try {
            // Show loading state
            saveBtn.textContent = 'Saving...';
            saveBtn.disabled = true;

            // Make API call
            const response = await this.apiCall(`/api/binance-accounts/${this.currentEditAccount.id}`, {
                method: 'PUT',
                body: JSON.stringify(updatedData)
            });

            if (response.ok) {
                // Success - close modal and refresh accounts list
                this.hideEditAccountModal();
                this.loadBinanceAccounts(); // Refresh the accounts table
                console.log('Account updated successfully');
            } else {
                const error = await response.text();
                alert('Error updating account: ' + error);
            }
        } catch (error) {
            alert('Network error: ' + error.message);
        } finally {
            // Reset button
            saveBtn.textContent = 'Save Changes';
            saveBtn.disabled = false;
        }
    }

    createAccountRow(account) {
        const row = document.createElement('tr');
        const createdDate = new Date(account.created_at).toLocaleDateString();

        const existingBalanceElement = document.getElementById(`balance-${account.id}`);
        const existingBalance = existingBalanceElement ? existingBalanceElement.textContent : 'Loading...'

        // Create the row HTML
        row.innerHTML = `
        <td>${account.name}</td>
        <td id="balance-${account.id}">${existingBalance}</td>
        <td><span class="status-badge ${account.account_active ? 'running' : 'stopped'}">
            ${account.account_active ? 'ACTIVE' : 'INACTIVE'}
        </span></td>
        <td>${createdDate}</td>
        <td>
            <button class="edit-account-btn btn-secondary" style="padding: 4px 8px; font-size: 0.8rem; margin-right: 5px;" data-account-id="${account.id}">
                Edit
            </button>
            <button class="delete-account-btn btn-danger" style="padding: 4px 8px; font-size: 0.8rem;" data-account-id="${account.id}">
                Delete
            </button>
        </td>
    `;

        // Add event listeners
        const editBtn = row.querySelector('.edit-account-btn');
        editBtn.addEventListener('click', () => {
            this.showEditAccountModal(account);
        });

        const deleteBtn = row.querySelector('.delete-account-btn');
        deleteBtn.addEventListener('click', () => {
            this.deleteAccount(account.id);
        });

        // Load account balance asynchronously (don't await it)
        this.loadAccountBalance(account.id, account.api_key, account.api_secret);

        return row; // Return the DOM node immediately
    }

    createBotStatsRow(bot) {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${bot.name}</td>
            <td><span class="status-badge ${bot.status.toLowerCase()}">${bot.status}</span></td>
            <td>${bot.account_name || "No Account"}</td>
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

        // Add account ID if selected
        const accountId = formData.get('botAccount');
        if (accountId) {
            botData.binance_account_id = parseInt(accountId);
        }

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

            // Load available accounts for the dropdown
            this.loadAvailableAccounts();
        }
    }

    async loadAvailableAccounts() {
        try {
            const response = await this.apiCall('/api/binance-accounts');
            if (response.ok) {
                const accounts = await response.json();
                console.log('Available accounts:', accounts); // Debug log

                const dropdown = document.getElementById('botAccount');
                dropdown.innerHTML = '<option value="">No Account</option>';

                if (accounts && Array.isArray(accounts)) {
                    accounts.forEach(account => {
                        console.log('Adding account:', account); // Debug each account
                        const option = document.createElement('option');
                        option.value = account.id;
                        option.textContent = `${account.name} (${account.base_url})`;
                        dropdown.appendChild(option);
                    });
                }
            }
        } catch (error) {
            console.error('Error loading accounts:', error);
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

        // binance get prcie
        document.getElementById('binanceBtn')?.addEventListener('click', () => {
            this.testBinance();
        });

        // binance account info
        document.getElementById('accInfoBtn')?.addEventListener('click', () => {
            this.getAccountInfo();
        });

        document.getElementById('marginBtn')?.addEventListener('click', () => {
            this.getMarginAccountInfo();
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

        // Add these to setupEventListeners()
        document.getElementById('createAccountBtn')?.addEventListener('click', () => {
            this.showCreateAccountModal();
        });

        document.getElementById('closeAccountModal')?.addEventListener('click', () => {
            this.hideCreateAccountModal();
        });

        document.getElementById('createAccountForm')?.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleCreateAccount();
        });

        // Add these to setupEventListeners()
        document.getElementById('cancelAccountBtn')?.addEventListener('click', () => {
            this.hideCreateAccountModal();
        });

        // Close modal when clicking outside
        document.getElementById('createAccountModal')?.addEventListener('click', (e) => {
            if (e.target.id === 'createAccountModal') {
                this.hideCreateAccountModal();
            }
        });

        // Edit account modal event listeners
        document.getElementById('closeEditAccountModal')?.addEventListener('click', () => {
            this.hideEditAccountModal();
        });

        document.getElementById('cancelEditAccountBtn')?.addEventListener('click', () => {
            this.hideEditAccountModal();
        });

        document.getElementById('editAccountForm')?.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleEditAccount();
        });

        // Close modal when clicking outside
        document.getElementById('editAccountModal')?.addEventListener('click', (e) => {
            if (e.target.id === 'editAccountModal') {
                this.hideEditAccountModal();
            }
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

        // Include account ID from dropdown
        const accountId = formData.get('editBotAccount');
        console.log('Selected account ID:', accountId); // Debug log

        if (accountId) {
            updatedData.binance_account_id = parseInt(accountId);
        } else {
            updatedData.binance_account_id = null;
        }

        console.log('Sending updated data:', updatedData); // Debug log

        // Validate required fields
        if (!updatedData.name || !updatedData.strategy) {
            alert('Bot name and strategy are required!');
            return;
        }

        try {
            // Show loading state
            saveBtn.textContent = 'Saving...';
            saveBtn.disabled = true;

            console.log('Making API call to:', `/api/bots/${this.currentBot.id}`); // Debug log

            // Make API call
            const response = await this.apiCall(`/api/bots/${this.currentBot.id}`, {
                method: 'PUT',
                body: JSON.stringify(updatedData)
            });

            console.log('Response status:', response.status, response.ok); // Debug log

            if (response.ok) {
                const updatedBot = await response.json();
                console.log('Updated bot response:', updatedBot);

                // Update the current bot object
                this.currentBot = updatedBot;

                // Refresh ALL related tables to show changes
                this.loadBotStats();
                this.loadBinanceAccounts();
                this.loadAccountsForEdit(updatedBot);

                console.log('Bot updated successfully');

                // Brief success feedback
                saveBtn.textContent = 'Saved!';
                setTimeout(() => {
                    saveBtn.textContent = 'Save Changes';
                }, 1000);
            } else {
                const error = await response.text();
                console.error('Server error:', error); // Debug log
                alert('Error updating bot: ' + error);
            }
        } catch (error) {
            console.error('Network error:', error); // Debug log
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
        this.hideBotDetailsModal();
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

    async testBinance() {
        const binanceBtn = document.getElementById("binanceBtn")
        const binanceResponse = document.getElementById("binanceResponse")

        try {
            const response = await this.apiCall('/api/test-binance')

            if (response.ok) {
                const priceData = await response.json();
                binanceResponse.innerHTML = `
                    <strong>Symbol:</strong> ${priceData.symbol}<br>
                    <strong>Price:</strong> ${priceData.price}<br>
                `;
                binanceResponse.style.display = 'block';
            } else {
                binanceResponse.innerHTML = '<strong>Error:</strong><br>Failed to get price';
                binanceResponse.style.display = 'block';
            }
        } catch (error) {
            binanceResponse.innerHTML = '<strong>Error:</strong><br>Failed to get price';
            binanceResponse.style.display = 'block';
        }
    }

    async getAccountInfo() {
        const accInfoBtn = document.getElementById("accInfoBtn");
        const accInfoResponse = document.getElementById("accInfoResponse");

        try {
            const response = await this.apiCall('/api/get-account-info');

            if (response.ok) {
                const accountData = await response.json();
                let balanceHtml = '<strong>Account Balances:</strong><br><br>';

                accountData.Balances
                    .filter(balance => parseFloat(balance.free || 0) > 0 || parseFloat(balance.locked || 0) > 0)
                    .forEach(balance => {
                        // Show asset name AND the amounts
                        balanceHtml += `<strong>${balance.asset}:</strong><br>`;
                        balanceHtml += `&nbsp;&nbsp;Free: ${balance.free}<br>`;
                        balanceHtml += `&nbsp;&nbsp;Locked: ${balance.locked}<br><br>`;
                    });

                accInfoResponse.innerHTML = balanceHtml;
                accInfoResponse.style.display = 'block';
            }
            else {
                const errorText = await response.text();
                accInfoResponse.innerHTML = `<strong>Server Error:</strong><br>${errorText}`;
                accInfoResponse.style.display = 'block';
            }

        } catch (error) {
            console.error("JavaScript error:", error);
            accInfoResponse.innerHTML = '<strong>Error:</strong><br>Failed to get account info';
            accInfoResponse.style.display = 'block';
        }
    }

    async getMarginAccountInfo() {
        const marginBtn = document.getElementById("marginBtn");
        const marginResponse = document.getElementById("marginResponse");

        try {
            marginBtn.textContent = 'Loading...';
            marginBtn.disabled = true;

            const response = await this.apiCall('/api/get-margin-account-info');
            if (response.ok) {
                const marginData = await response.json();
                console.log("Margin account data:", marginData);

                let marginHtml = '<strong>Cross Margin Account Info:</strong><br><br>';

                // Account overview
                marginHtml += `<strong>Account Status:</strong><br>`;
                marginHtml += `&nbsp;&nbsp;Margin Level: ${marginData.marginLevel}<br>`;
                marginHtml += `&nbsp;&nbsp;Borrow Enabled: ${marginData.borrowEnabled ? 'Yes' : 'No'}<br>`;
                marginHtml += `&nbsp;&nbsp;Trade Enabled: ${marginData.tradeEnabled ? 'Yes' : 'No'}<br><br>`;

                // Portfolio summary (in BTC terms)
                marginHtml += `<strong>Portfolio Summary (BTC):</strong><br>`;
                marginHtml += `&nbsp;&nbsp;Total Assets: ${marginData.totalAssetOfBtc}<br>`;
                marginHtml += `&nbsp;&nbsp;Total Liabilities: ${marginData.totalLiabilityOfBtc}<br>`;
                marginHtml += `&nbsp;&nbsp;Net Assets: ${marginData.totalNetAssetOfBtc}<br><br>`;

                // User assets with balances
                marginHtml += `<strong>Asset Details:</strong><br>`;
                marginData.userAssets
                    .filter(asset =>
                        parseFloat(asset.free || 0) > 0 ||
                        parseFloat(asset.locked || 0) > 0 ||
                        parseFloat(asset.borrowed || 0) > 0
                    )
                    .forEach(asset => {
                        marginHtml += `<strong>${asset.asset}:</strong><br>`;
                        marginHtml += `&nbsp;&nbsp;Free: ${asset.free}<br>`;
                        marginHtml += `&nbsp;&nbsp;Locked: ${asset.locked}<br>`;
                        marginHtml += `&nbsp;&nbsp;Borrowed: ${asset.borrowed}<br>`;
                        marginHtml += `&nbsp;&nbsp;Interest: ${asset.interest}<br>`;
                        marginHtml += `&nbsp;&nbsp;Net Asset: ${asset.netAsset}<br><br>`;
                    });

                marginResponse.innerHTML = marginHtml;
                marginResponse.style.display = 'block';

            } else {
                const errorText = await response.text();
                marginResponse.innerHTML = `<strong>Server Error:</strong><br>${errorText}`;
                marginResponse.style.display = 'block';
            }
        } catch (error) {
            console.error("JavaScript error:", error);
            marginResponse.innerHTML = `<strong>Error:</strong><br>${error.message}`;
            marginResponse.style.display = 'block';
        } finally {
            marginBtn.textContent = 'Get Margin Info';
            marginBtn.disabled = false;
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

        this.loadAccountsForEdit(bot);
        modal.style.display = 'flex';
    }

    async loadAccountsForEdit(bot) {
        console.log('Loading accounts for edit, bot:', bot); // Debug log

        try {
            const response = await this.apiCall('/api/binance-accounts');
            console.log('Accounts response:', response.status, response.ok); // Debug log

            if (response.ok) {
                const accounts = await response.json();
                console.log('Accounts data for edit:', accounts); // Debug log

                const dropdown = document.getElementById('editBotAccount');
                console.log('Dropdown element:', dropdown); // Debug log

                // Clear existing options
                dropdown.innerHTML = '<option value="">No Account</option>';

                if (accounts && Array.isArray(accounts)) {
                    console.log('Adding', accounts.length, 'accounts to edit dropdown'); // Debug log

                    accounts.forEach((account, index) => {
                        console.log('Adding account', index, ':', account); // Debug log
                        const option = document.createElement('option');
                        option.value = account.id;
                        option.textContent = `${account.name} (${account.base_url})`;
                        dropdown.appendChild(option);
                    })

                    if (bot.binance_account_id && bot.binance_account_id > 0) {
                        dropdown.value = bot.binance_account_id;
                        console.log('Pre-selected account ID: ', bot.binance_account_id);
                    } else {
                        dropdown.value = "";
                        console.log('No account linked, selected "No Account');
                    }
                } else {
                    console.log('No accounts or not an array'); // Debug log
                }

                console.log('Final dropdown HTML:', dropdown.innerHTML); // Debug log
            }
        } catch (error) {
            console.error('Error loading accounts for edit:', error);
        }
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
