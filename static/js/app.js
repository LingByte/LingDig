document.addEventListener('DOMContentLoaded', function() {
    // 表单元素
    const digForm = document.getElementById('digForm');
    const curlForm = document.getElementById('curlForm');
    const serverSelect = document.getElementById('server');
    const customServerGroup = document.getElementById('customServerGroup');
    const customServerInput = document.getElementById('customServer');
    
    // 状态元素
    const loadingState = document.getElementById('loading');
    const emptyState = document.getElementById('emptyState');
    const results = document.getElementById('results');
    const resultStats = document.getElementById('resultStats');
    const recordCount = document.getElementById('recordCount');
    
    // 工具切换
    const tabBtns = document.querySelectorAll('.tab-btn');
    const toolForms = document.querySelectorAll('.tool-form');
    
    let currentTool = 'dns';

    // 工具切换事件
    tabBtns.forEach(btn => {
        btn.addEventListener('click', function() {
            const tool = this.dataset.tool;
            switchTool(tool);
        });
    });

    function switchTool(tool) {
        currentTool = tool;
        
        // 更新标签状态
        tabBtns.forEach(btn => {
            btn.classList.toggle('active', btn.dataset.tool === tool);
        });
        
        // 更新表单显示
        toolForms.forEach(form => {
            form.classList.toggle('active', form.dataset.tool === tool);
        });
        
        // 清空结果
        clearResults();
    }

    function clearResults() {
        results.innerHTML = '';
        resultStats.style.display = 'none';
        emptyState.style.display = 'flex';
        loadingState.style.display = 'none';
    }

    // DNS相关功能
    serverSelect.addEventListener('change', function() {
        if (this.value === 'custom') {
            customServerGroup.style.display = 'block';
            customServerInput.required = true;
        } else {
            customServerGroup.style.display = 'none';
            customServerInput.required = false;
        }
    });

    // DNS查询表单提交
    digForm.addEventListener('submit', function(e) {
        e.preventDefault();
        
        const formData = new FormData(digForm);
        
        if (serverSelect.value === 'custom') {
            formData.set('server', customServerInput.value);
        }

        showLoading('dns');

        fetch('/dig', {
            method: 'POST',
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            hideLoading();
            displayDnsResults(data);
        })
        .catch(error => {
            hideLoading();
            displayError('DNS查询失败: ' + error.message);
        });
    });

    // HTTP请求表单提交
    curlForm.addEventListener('submit', function(e) {
        e.preventDefault();
        
        const formData = new FormData(curlForm);
        
        showLoading('curl');

        fetch('/curl', {
            method: 'POST',
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            hideLoading();
            displayCurlResults(data);
        })
        .catch(error => {
            hideLoading();
            displayError('HTTP请求失败: ' + error.message);
        });
    });

    function showLoading(tool) {
        emptyState.style.display = 'none';
        loadingState.style.display = 'flex';
        results.innerHTML = '';
        resultStats.style.display = 'none';
        
        // 更新按钮状态
        const activeForm = document.querySelector('.tool-form.active');
        const btn = activeForm.querySelector('.query-btn');
        const btnText = btn.querySelector('.btn-text');
        const btnIcon = btn.querySelector('.btn-icon');
        
        if (tool === 'dns') {
            btnText.textContent = '查询中...';
            btnIcon.className = 'btn-icon fas fa-spinner fa-spin';
        } else {
            btnText.textContent = '发送中...';
            btnIcon.className = 'btn-icon fas fa-spinner fa-spin';
        }
        btn.disabled = true;
    }

    function hideLoading() {
        loadingState.style.display = 'none';
        
        // 恢复按钮状态
        const activeForm = document.querySelector('.tool-form.active');
        const btn = activeForm.querySelector('.query-btn');
        const btnText = btn.querySelector('.btn-text');
        const btnIcon = btn.querySelector('.btn-icon');
        
        if (currentTool === 'dns') {
            btnText.textContent = '查询';
            btnIcon.className = 'btn-icon fas fa-search';
        } else {
            btnText.textContent = '发送请求';
            btnIcon.className = 'btn-icon fas fa-paper-plane';
        }
        btn.disabled = false;
    }

    function displayDnsResults(data) {
        if (data.error) {
            displayError(data.error);
            return;
        }

        if (!data.results || data.results.length === 0) {
            displayNoResults('未找到DNS记录');
            return;
        }

        recordCount.textContent = data.results.length;
        resultStats.style.display = 'block';

        let html = '';
        data.results.forEach(record => {
            html += `
                <div class="record-item">
                    <div class="record-header">
                        <div class="record-name">${escapeHtml(record.name)}</div>
                        <span class="record-ttl">TTL ${record.ttl}s</span>
                    </div>
                    <div class="record-main">
                        <span class="record-type">${record.type}</span>
                        <div class="record-value">${escapeHtml(record.value)}</div>
                    </div>
                </div>
            `;
        });

        results.innerHTML = html;
        emptyState.style.display = 'none';
    }

    function displayCurlResults(data) {
        if (data.error) {
            displayError(data.error);
            return;
        }

        // 状态码颜色
        let statusClass = 'success';
        if (data.status_code >= 300 && data.status_code < 400) statusClass = 'redirect';
        else if (data.status_code >= 400 && data.status_code < 500) statusClass = 'client-error';
        else if (data.status_code >= 500) statusClass = 'server-error';

        let html = `
            <div class="http-response">
                <div class="response-status">
                    <span class="status-code ${statusClass}">${data.status_code}</span>
                    <span>${data.status_text}</span>
                    <span class="response-time">${data.response_time_ms}ms</span>
                    ${data.body_size ? `<span class="response-size">${formatBytes(data.body_size)}</span>` : ''}
                </div>
                
                <!-- 请求信息 -->
                <details class="request-info">
                    <summary>请求信息</summary>
                    <div class="info-content">
                        <div class="info-item">
                            <span class="info-key">方法:</span>
                            <span class="info-value">${data.method}</span>
                        </div>
                        <div class="info-item">
                            <span class="info-key">URL:</span>
                            <span class="info-value">${escapeHtml(data.request_info.final_url || data.url)}</span>
                        </div>
                        <div class="info-item">
                            <span class="info-key">协议:</span>
                            <span class="info-value">${data.request_info.protocol || 'HTTP/1.1'}</span>
                        </div>
                        ${data.request_info.tls_version ? `
                        <div class="info-item">
                            <span class="info-key">TLS版本:</span>
                            <span class="info-value">${data.request_info.tls_version}</span>
                        </div>
                        ` : ''}
                        ${data.request_info.remote_addr ? `
                        <div class="info-item">
                            <span class="info-key">远程地址:</span>
                            <span class="info-value">${data.request_info.remote_addr}</span>
                        </div>
                        ` : ''}
                    </div>
                </details>

                <!-- 重定向链 -->
                ${data.redirect_chain && data.redirect_chain.length > 0 ? `
                <details class="redirect-chain">
                    <summary>重定向链 (${data.redirect_chain.length})</summary>
                    <div class="redirect-list">
                        ${data.redirect_chain.map((url, index) => `
                            <div class="redirect-item">
                                <span class="redirect-step">${index + 1}.</span>
                                <span class="redirect-url">${escapeHtml(url)}</span>
                            </div>
                        `).join('')}
                    </div>
                </details>
                ` : ''}

                <!-- 请求头 -->
                <details class="request-headers">
                    <summary>请求头 (${Object.keys(data.request_info.request_headers || {}).length})</summary>
                    <div class="headers-list">
        `;

        for (const [key, value] of Object.entries(data.request_info.request_headers || {})) {
            html += `
                <div class="header-item">
                    <span class="header-key">${escapeHtml(key)}:</span>
                    <span class="header-value">${escapeHtml(value)}</span>
                </div>
            `;
        }

        html += `
                    </div>
                </details>
                
                <!-- 响应头 -->
                <details class="response-headers">
                    <summary>响应头 (${Object.keys(data.headers || {}).length})</summary>
                    <div class="headers-list">
        `;

        for (const [key, value] of Object.entries(data.headers || {})) {
            html += `
                <div class="header-item">
                    <span class="header-key">${escapeHtml(key)}:</span>
                    <span class="header-value">${escapeHtml(value)}</span>
                </div>
            `;
        }

        html += `
                    </div>
                </details>
                
                <!-- 响应体 -->
                <div class="response-body-section">
                    <div class="body-header">
                        <span>响应体</span>
                        ${data.is_binary ? '<span class="binary-badge">二进制</span>' : ''}
                        ${data.content_type ? `<span class="content-type">${data.content_type}</span>` : ''}
                    </div>
                    <div class="response-body">${escapeHtml(data.body_preview || data.body || '')}</div>
                </div>
            </div>
        `;

        results.innerHTML = html;
        emptyState.style.display = 'none';
        resultStats.style.display = 'none';
    }

    function formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    function displayError(message) {
        results.innerHTML = `
            <div class="error-message">
                ${escapeHtml(message)}
            </div>
        `;
        emptyState.style.display = 'none';
        resultStats.style.display = 'none';
    }

    function displayNoResults(message) {
        results.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-inbox"></i>
                <p>${message}</p>
            </div>
        `;
        emptyState.style.display = 'none';
        resultStats.style.display = 'none';
    }

    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
});