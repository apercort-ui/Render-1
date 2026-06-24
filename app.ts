interface BackendStatus {
    status: string;
    redisConnected: boolean;
    rabbitmqConnected: boolean;
    timestamp: string;
}

class Dashboard {
    private statusBadge: HTMLElement;
    private goCardState: HTMLElement;
    private redisCardState: HTMLElement;
    private rabbitCardState: HTMLElement;
    private logConsole: HTMLElement;
    private messageInput: HTMLInputElement;
    private sendButton: HTMLButtonElement;

    // Ссылаемся на ваш живой сервер Render
    private backendUrl: string = "https://render-1-bq28.onrender.com";

    constructor() {
        this.statusBadge = document.getElementById('connection-status')!;
        this.goCardState = document.querySelector('#go-core-card .state')!;
        this.redisCardState = document.querySelector('#redis-card .state')!;
        this.rabbitCardState = document.querySelector('#rabbitmq-card .state')!;
        this.logConsole = document.getElementById('log-console')!;
        this.messageInput = document.getElementById('message-input') as HTMLInputElement;
        this.sendButton = document.getElementById('send-btn') as HTMLButtonElement;

        this.init();
    }

    private init(): void {
        this.addLog("Инициализация системы мониторинга...");
        
        // Первый запрос к бэкенду
        this.checkBackendStatus();
        
        // Каждые 10 секунд обновляем статус компонентов
        setInterval(() => this.checkBackendStatus(), 10000);

        // Навешиваем событие клика на кнопку отправки
        this.sendButton.addEventListener('click', () => this.sendMessageToQueue());
        
        // Позволяем отправлять сообщение по нажатию Enter в инпуте
        this.messageInput.addEventListener('keypress', (e: KeyboardEvent) => {
            if (e.key === 'Enter') {
                this.sendMessageToQueue();
            }
        });
    }

    private async checkBackendStatus(): Promise<void> {
        try {
            const response = await fetch(`${this.backendUrl}/api/status`);
            if (!response.ok) throw new Error(`Статус ответа: ${response.status}`);
            
            const data: BackendStatus = await response.json();
            this.updateUI(data);
        } catch (error: any) {
            this.handleError(error.message || error);
        }
    }

    // Метод отправки сообщения в очередь через ваш Go-бэкенд
    private async sendMessageToQueue(): Promise<void> {
        const message = this.messageInput.value.trim();
        if (!message) return;

        try {
            this.sendButton.disabled = true;
            this.addLog(`Отправка сообщения: "${message}"...`);

            // Отправляем POST-запрос на бэкенд (убедитесь, что в Go настроен этот эндпоинт)
            const response = await fetch(`${this.backendUrl}/api/send`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ text: message })
            });

            if (response.ok) {
                this.addLog(`[Успех] Сообщение доставлено в брокер RabbitMQ.`);
                this.messageInput.value = ''; // Очищаем поле ввода
            } else {
                throw new Error(`Код ошибки бэкенда: ${response.status}`);
            }
        } catch (error: any) {
            this.addLog(`[Ошибка отправки] Очередь недоступна: ${error.message}`);
        } finally {
            this.sendButton.disabled = false;
            this.messageInput.focus();
        }
    }

    private updateUI(data: BackendStatus): void {
        this.statusBadge.innerText = "В эфире 🎉";
        this.statusBadge.className = "status-badge online";

        this.goCardState.innerText = "АКТИВЕН";
        this.goCardState.style.color = "#22c55e";

        this.redisCardState.innerText = data.redisConnected ? "ПОДКЛЮЧЕНО" : "БЕЗ КЭША";
        this.redisCardState.style.color = data.redisConnected ? "#22c55e" : "#eab308";

        this.rabbitCardState.innerText = data.rabbitmqConnected ? "РАБОТАЕТ" : "ОШИБКА";
        this.rabbitCardState.style.color = data.rabbitmqConnected ? "#22c55e" : "#ef4444";

        this.messageInput.disabled = false;
        this.sendButton.disabled = false;
    }

    private handleError(errorMessage: string): void {
        this.statusBadge.innerText = "Ошибка соединения";
        this.statusBadge.className = "status-badge connecting";

        this.goCardState.innerText = "ОФФЛАЙН";
        this.goCardState.style.color = "#ef4444";
        this.redisCardState.innerText = "НЕИЗВЕСТНО";
        this.redisCardState.style.color = "#94a3b8";
        this.rabbitCardState.innerText = "НЕИЗВЕСТНО";
        this.rabbitCardState.style.color = "#94a3b8";

        this.messageInput.disabled = true;
        this.sendButton.disabled = true;

        this.addLog(`[Ошибка связи] Сервер Render недоступен: ${errorMessage}`);
    }

    private addLog(text: string): void {
        const line = document.createElement('div');
        line.className = 'log-line';
        line.innerText = `[${new Date().toLocaleTimeString()}] ${text}`;
        this.logConsole.appendChild(line);
        this.logConsole.scrollTop = this.logConsole.scrollHeight;
    }
}

window.addEventListener('DOMContentLoaded', () => {
    new Dashboard();
});
