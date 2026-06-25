// 1. Структура ответа от Go-бэкенда
interface SystemStatus {
    status: string;
    redisConnected: boolean;
    rabbitmqConnected: boolean;
    postgresConnected: boolean;
    timestamp: string;
}

// 2. Обновление индикаторов статуса
async function updateSystemStatus(): Promise<void> {
    try {
        const response = await fetch('/api/status');
        if (!response.ok) throw new Error(`Ошибка: ${response.status}`);
        
        const data: SystemStatus = await response.json();

        const indicators = [
            { id: 'postgres-status', val: data.postgresConnected, on: "ПОДКЛЮЧЕНО", off: "ОТКЛЮЧЕНО" },
            { id: 'rabbit-status',   val: data.rabbitmqConnected,   on: "ПОДКЛЮЧЕНО", off: "ОТКЛЮЧЕНО" },
            { id: 'redis-status',    val: data.redisConnected,      on: "ПОДКЛЮЧЕНО", off: "БЕЗ КЭША" }
        ];

        indicators.forEach(item => {
            const el = document.getElementById(item.id);
            if (el) {
                el.textContent = item.val ? item.on : item.off;
                el.style.color = item.val ? "#2ecc71" : (item.id === 'redis-status' ? "#f1c40f" : "#e74c3c");
            }
        });
    } catch (error) {
        console.error("Не удалось получить статус:", error);
    }
}

// 3. Отправка сообщений
async function sendMessage(): Promise<void> {
    const input = document.getElementById('msgInput') as HTMLInputElement;
    const logs = document.getElementById('logs')!;
    
    if (!input?.value) return;

    try {
        const response = await fetch('/api/send', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ text: input.value })
        });
        
        const result = await response.json();
        logs.insertAdjacentHTML('beforeend', `<div>> ${result.message}</div>`);
        input.value = '';
    } catch (e) {
        logs.insertAdjacentHTML('beforeend', `<div style="color:red">> Ошибка отправки</div>`);
    }
}

// 4. Инициализация
document.addEventListener('DOMContentLoaded', () => {
    updateSystemStatus();
    setInterval(updateSystemStatus, 5000);
});
