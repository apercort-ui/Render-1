// Определяем протокол: если сайт на https, то ws должен стать wss
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const socket = new WebSocket(`${protocol}//${window.location.host}/ws`);

const logs = document.getElementById('logs')!;

socket.onopen = () => {
    logs.insertAdjacentHTML('beforeend', `<div>[Система]: Соединение установлено</div>`);
};

socket.onmessage = (event) => {
    try {
        const msg = JSON.parse(event.data);
        logs.insertAdjacentHTML('beforeend', `<div>> ${msg.text}</div>`);
    } catch (e) {
        console.error("Ошибка парсинга:", e);
    }
};

socket.onerror = (err) => {
    logs.insertAdjacentHTML('beforeend', `<div style="color:red">[Ошибка]: WebSocket недоступен</div>`);
};

function sendMessage() {
    const input = document.getElementById('msgInput') as HTMLInputElement;
    if (input.value && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ text: input.value }));
        input.value = '';
    } else {
        logs.insertAdjacentHTML('beforeend', `<div style="color:orange">[Система]: Соединение не готово</div>`);
    }
}
