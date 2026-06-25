// Описываем структуру ответа от нашего Go-бэкенда для строгой типизации
interface SystemStatus {
    status: string;
    redisConnected: boolean;
    rabbitmqConnected: boolean;
    postgresConnected: boolean;
    timestamp: string;
}

// Функция для опроса бэкенда и обновления индикаторов на экране
async function updateSystemStatus(): Promise<void> {
    try {
        const response = await fetch('/api/status');
        if (!response.ok) {
            throw new Error(`Ошибка сервера: ${response.status}`);
        }
        
        const data: SystemStatus = await response.json();

        // 1. Обновляем индикатор PostgreSQL
        const postgresIndicator = document.getElementById('postgres-status');
        if (postgresIndicator) {
            if (data.postgresConnected) {
                postgresIndicator.textContent = "ПОДКЛЮЧЕНО";
                postgresIndicator.style.color = "#2ecc71"; // Зелёный
            } else {
                postgresIndicator.textContent = "ОТКЛЮЧЕНО";
                postgresIndicator.style.color = "#e74c3c"; // Красный
            }
        }

        // 2. Обновляем индикатор RabbitMQ (чтобы старая логика не ломалась)
        const rabbitIndicator = document.getElementById('rabbit-status');
        if (rabbitIndicator) {
            if (data.rabbitmqConnected) {
                rabbitIndicator.textContent = "ПОДКЛЮЧЕНО";
                rabbitIndicator.style.color = "#2ecc71";
            } else {
                rabbitIndicator.textContent = "ОТКЛЮЧЕНО";
                rabbitIndicator.style.color = "#e74c3c";
            }
        }

        // 3. Обновляем индикатор Redis
        const redisIndicator = document.getElementById('redis-status');
        if (redisIndicator) {
            if (data.redisConnected) {
                redisIndicator.textContent = "ПОДКЛЮЧЕНО";
                redisIndicator.style.color = "#2ecc71";
            } else {
                redisIndicator.textContent = "БЕЗ КЭША";
                redisIndicator.style.color = "#f1c40f"; // Жёлтый
            }
        }

    } catch (error) {
        console.error("Не удалось получить статус системы:", error);
    }
}

// Запускаем опрос статуса при загрузке страницы и повторяем каждые 5 секунд
document.addEventListener('DOMContentLoaded', () => {
    updateSystemStatus();
    setInterval(updateSystemStatus, 5000);
});
