class ServiceMonitor {
    constructor() {
        this.ws = null;
        this.uptimeChart = null;
        this.refreshInterval = null;
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.connectWebSocket();
        this.loadData();
        this.setupAutoRefresh();
        this.initChart();
    }

    setupEventListeners() {
        // Модальные окна
        document.getElementById('addServiceBtn').addEventListener('click', () => {
            this.showModal('addServiceModal');
        });

        document.getElementById('closeModal').addEventListener('click', () => {
            this.hideModal('addServiceModal');
        });

        document.getElementById('closeDetailsModal').addEventListener('click', () => {
            this.hideModal('serviceDetailsModal');
        });

        document.getElementById('cancelAdd').addEventListener('click', () => {
            this.hideModal('addServiceModal');
        });

        // Форма добавления сервиса
        document.getElementById('addServiceForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.addService();
        });

        // Закрытие модальных окон по клику на overlay
        document.querySelectorAll('.modal__overlay').forEach(overlay => {
            overlay.addEventListener('click', (e) => {
                if (e.target === overlay) {
                    this.hideAllModals();
                }
            });
        });

        // Закрытие по Escape
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.hideAllModals();
            }
        });
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/api/v1/ws`;
        
        this.ws = new WebSocket(wsUrl);
        
        this.ws.onopen = () => {
            this.updateConnectionStatus('connected');
            console.log('WebSocket подключен');
        };
        
        this.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.handleWebSocketMessage(data);
            } catch (error) {
                console.error('Ошибка парсинга WebSocket сообщения:', error);
            }
        };
        
        this.ws.onclose = () => {
            this.updateConnectionStatus('disconnected');
            console.log('WebSocket отключен');
            // Переподключение через 5 секунд
            setTimeout(() => this.connectWebSocket(), 5000);
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket ошибка:', error);
            this.updateConnectionStatus('disconnected');
        };
    }

    updateConnectionStatus(status) {
        const statusElement = document.getElementById('connectionStatus');
        statusElement.className = `status-indicator ${status}`;
        
        switch (status) {
            case 'connected':
                statusElement.textContent = 'Подключено';
                break;
            case 'disconnected':
                statusElement.textContent = 'Отключено';
                break;
            default:
                statusElement.textContent = 'Подключение...';
        }
    }

    handleWebSocketMessage(data) {
        switch (data.type) {
            case 'welcome':
                console.log('WebSocket приветствие:', data.message);
                break;
            case 'service_update':
                this.loadServices();
                break;
            case 'alert_update':
                this.loadAlerts();
                break;
            case 'stats_update':
                this.loadStats();
                break;
        }
    }

    async loadData() {
        await Promise.all([
            this.loadServices(),
            this.loadAlerts(),
            this.loadStats()
        ]);
    }

    async loadServices() {
        try {
            const response = await fetch('/api/v1/services');
            if (!response.ok) throw new Error('Ошибка загрузки сервисов');
            
            const services = await response.json();
            this.renderServices(services);
        } catch (error) {
            console.error('Ошибка загрузки сервисов:', error);
            this.showError('Ошибка загрузки сервисов');
        }
    }

    async loadAlerts() {
        try {
            const response = await fetch('/api/v1/alerts');
            if (!response.ok) throw new Error('Ошибка загрузки алертов');
            
            const alerts = await response.json();
            this.renderAlerts(alerts);
        } catch (error) {
            console.error('Ошибка загрузки алертов:', error);
        }
    }

    async loadStats() {
        try {
            const response = await fetch('/api/v1/stats');
            if (!response.ok) throw new Error('Ошибка загрузки статистики');
            
            const stats = await response.json();
            this.renderStats(stats);
        } catch (error) {
            console.error('Ошибка загрузки статистики:', error);
        }
    }

    renderServices(services) {
        const grid = document.getElementById('servicesGrid');
        grid.innerHTML = '';

        if (services.length === 0) {
            grid.innerHTML = `
                <div class="empty-state">
                    <p>Сервисы не добавлены</p>
                    <button class="btn btn--primary" onclick="this.showModal('addServiceModal')">
                        Добавить первый сервис
                    </button>
                </div>
            `;
            return;
        }

        services.forEach(service => {
            const card = this.createServiceCard(service);
            grid.appendChild(card);
        });
    }

    createServiceCard(service) {
        const card = document.createElement('div');
        card.className = `service-card service-card--${service.last_status || 'unknown'}`;
        
        const statusClass = service.last_status === 'healthy' ? 'healthy' : 
                           service.last_status === 'unhealthy' ? 'unhealthy' : 'unknown';
        
        const lastCheck = service.last_check ? 
            new Date(service.last_check).toLocaleString('ru-RU') : 'Не проверялся';
        
        const uptime = service.uptime ? `${service.uptime.toFixed(1)}%` : 'N/A';

        card.innerHTML = `
            <div class="service-card__header">
                <div>
                    <h3 class="service-card__title">${this.escapeHtml(service.name)}</h3>
                    <p class="service-card__url">${this.escapeHtml(service.url)}</p>
                </div>
                <span class="service-card__status service-card__status--${statusClass}">
                    ${this.getStatusIcon(service.last_status)} ${this.getStatusText(service.last_status)}
                </span>
            </div>
            
            <div class="service-card__details">
                <div class="service-card__detail">
                    <div class="service-card__detail-label">Uptime (24ч)</div>
                    <div class="service-card__detail-value">${uptime}</div>
                </div>
                <div class="service-card__detail">
                    <div class="service-card__detail-label">Последняя проверка</div>
                    <div class="service-card__detail-value">${lastCheck}</div>
                </div>
            </div>
            
            <div class="service-card__actions">
                <button class="btn btn--small" onclick="serviceMonitor.showServiceDetails(${service.id})">
                    📊 Детали
                </button>
                <button class="btn btn--small btn--danger" onclick="serviceMonitor.deleteService(${service.id})">
                    🗑️ Удалить
                </button>
            </div>
        `;
        
        return card;
    }

    renderAlerts(alerts) {
        const list = document.getElementById('alertsList');
        list.innerHTML = '';

        if (alerts.length === 0) {
            list.innerHTML = '<p class="empty-state">Алертов нет</p>';
            return;
        }

        alerts.forEach(alert => {
            const item = this.createAlertItem(alert);
            list.appendChild(item);
        });
    }

    createAlertItem(alert) {
        const item = document.createElement('div');
        item.className = `alert-item ${alert.is_resolved ? 'alert-item--resolved' : ''}`;
        
        const createdAt = new Date(alert.created_at).toLocaleString('ru-RU');
        const icon = alert.is_resolved ? '✅' : '🚨';
        
        item.innerHTML = `
            <div class="alert-item__icon">${icon}</div>
            <div class="alert-item__content">
                <div class="alert-item__message">${this.escapeHtml(alert.message)}</div>
                <div class="alert-item__service">${this.escapeHtml(alert.service?.name || 'Неизвестный сервис')}</div>
                <div class="alert-item__time">${createdAt}</div>
            </div>
            ${!alert.is_resolved ? `
                <div class="alert-item__actions">
                    <button class="btn btn--small" onclick="serviceMonitor.resolveAlert(${alert.id})">
                        ✅ Разрешить
                    </button>
                </div>
            ` : ''}
        `;
        
        return item;
    }

    renderStats(stats) {
        document.getElementById('totalServices').textContent = stats.total_services;
        document.getElementById('healthyServices').textContent = stats.healthy_services;
        document.getElementById('unhealthyServices').textContent = stats.unhealthy_services;
        document.getElementById('averageUptime').textContent = `${stats.average_uptime.toFixed(1)}%`;
        document.getElementById('activeAlerts').textContent = stats.active_alerts;
    }

    async addService() {
        const form = document.getElementById('addServiceForm');
        const formData = new FormData(form);
        
        const serviceData = {
            name: formData.get('name'),
            url: formData.get('url'),
            check_interval: parseInt(formData.get('check_interval')) || 30,
            timeout: parseInt(formData.get('timeout')) || 10
        };

        try {
            const response = await fetch('/api/v1/services', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(serviceData)
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Ошибка создания сервиса');
            }

            this.hideModal('addServiceModal');
            form.reset();
            this.loadServices();
            this.showSuccess('Сервис успешно добавлен');
        } catch (error) {
            console.error('Ошибка добавления сервиса:', error);
            this.showError(error.message);
        }
    }

    async deleteService(id) {
        if (!confirm('Вы уверены, что хотите удалить этот сервис?')) {
            return;
        }

        try {
            const response = await fetch(`/api/v1/services/${id}`, {
                method: 'DELETE'
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Ошибка удаления сервиса');
            }

            this.loadServices();
            this.showSuccess('Сервис успешно удален');
        } catch (error) {
            console.error('Ошибка удаления сервиса:', error);
            this.showError(error.message);
        }
    }

    async resolveAlert(id) {
        try {
            const response = await fetch(`/api/v1/alerts/${id}/resolve`, {
                method: 'PUT'
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Ошибка разрешения алерта');
            }

            this.loadAlerts();
            this.showSuccess('Алерт разрешен');
        } catch (error) {
            console.error('Ошибка разрешения алерта:', error);
            this.showError(error.message);
        }
    }

    async showServiceDetails(id) {
        try {
            const [serviceResponse, checksResponse] = await Promise.all([
                fetch(`/api/v1/services/${id}`),
                fetch(`/api/v1/services/${id}/checks?limit=50`)
            ]);

            if (!serviceResponse.ok || !checksResponse.ok) {
                throw new Error('Ошибка загрузки деталей сервиса');
            }

            const service = await serviceResponse.json();
            const checks = await checksResponse.json();

            this.renderServiceDetails(service, checks);
            this.showModal('serviceDetailsModal');
        } catch (error) {
            console.error('Ошибка загрузки деталей сервиса:', error);
            this.showError('Ошибка загрузки деталей сервиса');
        }
    }

    renderServiceDetails(service, checks) {
        const content = document.getElementById('serviceDetailsContent');
        
        const checksTable = checks.length > 0 ? `
            <table class="table">
                <thead>
                    <tr>
                        <th>Время</th>
                        <th>Статус</th>
                        <th>Время ответа</th>
                        <th>Ошибка</th>
                    </tr>
                </thead>
                <tbody>
                    ${checks.map(check => `
                        <tr>
                            <td>${new Date(check.checked_at).toLocaleString('ru-RU')}</td>
                            <td>
                                <span class="service-card__status service-card__status--${check.status}">
                                    ${this.getStatusIcon(check.status)} ${this.getStatusText(check.status)}
                                </span>
                            </td>
                            <td>${check.response_time}ms</td>
                            <td>${check.error_message || '-'}</td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        ` : '<p>История проверок пуста</p>';

        content.innerHTML = `
            <div class="service-details">
                <h4>Информация о сервисе</h4>
                <p><strong>Название:</strong> ${this.escapeHtml(service.name)}</p>
                <p><strong>URL:</strong> ${this.escapeHtml(service.url)}</p>
                <p><strong>Интервал проверки:</strong> ${service.check_interval} сек</p>
                <p><strong>Таймаут:</strong> ${service.timeout} сек</p>
                <p><strong>Создан:</strong> ${new Date(service.created_at).toLocaleString('ru-RU')}</p>
                
                <h4 style="margin-top: 2rem;">История проверок</h4>
                ${checksTable}
            </div>
        `;
    }

    initChart() {
        const ctx = document.getElementById('uptimeChart').getContext('2d');
        this.uptimeChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Uptime (%)',
                    data: [],
                    borderColor: '#2563eb',
                    backgroundColor: 'rgba(37, 99, 235, 0.1)',
                    tension: 0.4,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100,
                        ticks: {
                            callback: function(value) {
                                return value + '%';
                            }
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });
    }

    setupAutoRefresh() {
        // Обновляем данные каждые 30 секунд
        this.refreshInterval = setInterval(() => {
            this.loadData();
        }, 30000);
    }

    showModal(modalId) {
        document.getElementById(modalId).classList.add('active');
    }

    hideModal(modalId) {
        document.getElementById(modalId).classList.remove('active');
    }

    hideAllModals() {
        document.querySelectorAll('.modal').forEach(modal => {
            modal.classList.remove('active');
        });
    }

    showSuccess(message) {
        this.showNotification(message, 'success');
    }

    showError(message) {
        this.showNotification(message, 'error');
    }

    showNotification(message, type) {
        const notification = document.createElement('div');
        notification.className = `notification notification--${type}`;
        notification.textContent = message;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.classList.add('notification--show');
        }, 100);
        
        setTimeout(() => {
            notification.classList.remove('notification--show');
            setTimeout(() => {
                document.body.removeChild(notification);
            }, 300);
        }, 3000);
    }

    getStatusIcon(status) {
        switch (status) {
            case 'healthy': return '✅';
            case 'unhealthy': return '❌';
            default: return '❓';
        }
    }

    getStatusText(status) {
        switch (status) {
            case 'healthy': return 'Работает';
            case 'unhealthy': return 'Не работает';
            default: return 'Неизвестно';
        }
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Инициализация приложения
let serviceMonitor;
document.addEventListener('DOMContentLoaded', () => {
    serviceMonitor = new ServiceMonitor();
});

// Добавляем стили для уведомлений
const notificationStyles = `
    .notification {
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 1rem 1.5rem;
        border-radius: 8px;
        color: white;
        font-weight: 500;
        z-index: 10000;
        transform: translateX(100%);
        transition: transform 0.3s ease;
    }
    
    .notification--show {
        transform: translateX(0);
    }
    
    .notification--success {
        background: #10b981;
    }
    
    .notification--error {
        background: #ef4444;
    }
    
    .empty-state {
        text-align: center;
        padding: 2rem;
        color: #64748b;
    }
`;

const styleSheet = document.createElement('style');
styleSheet.textContent = notificationStyles;
document.head.appendChild(styleSheet);
