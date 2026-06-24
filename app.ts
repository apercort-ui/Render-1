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

    // Укажите URL вашего запущенного бэкенда на Render
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
        this.addLog("Инициализация проверки связи с Render Go API...");
        
        // Сразу делаем первый запрос
        this.checkBackendStatus();
        
        // Настраиваем периодический опрос (polling) каждые 10 секунд
        setInterval(() => this.checkBackendStatus(), 10000);
    }

    // Асинхронный метод для выполнения fetch-запроса
    private async checkBackendStatus(): Promise<void> {
        try {
            const response = await fetch(`${this.backendUrl}/api/status`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`Ошибка сервера: ${response.status}`);
            }

            const data: BackendStatus = await response.json();
            this.updateUI(data);

        } catch (error: any) {
            this.handleError(error.message || error);
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

        this.addLog("Данные успешно синхронизированы с бэкендом.");
        if (!data.redisConnected) {
            this.addLog("[Предупреждение] Бэкенд работает без кэша Redis.");
        }
    }

    private handleError(errorMessage: string): void {
        this.statusBadge.innerText = "Ошибка соединения";
        this.statusBadge.className = "status-badge connecting"; // оранжевый/красный статус

        this.goCardState.innerText = "ОФФЛАЙН";
        this.goCardState.style.color = "#ef4444";
        this.redisCardState.innerText = "НЕИЗВЕСТНО";
        this.redisCardState.style.color = "#94a3b8";
        this.rabbitCardState.innerText = "НЕИЗВЕСТНО";
        this.rabbitCardState.style.color = "#94a3b8";

        this.messageInput.disabled = true;
        this.sendButton.disabled = true;

        this.addLog(`[Ошибка] Не удалось связаться с бэкендом: ${errorMessage}`);
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
