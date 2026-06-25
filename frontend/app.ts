const socket = new WebSocket(`ws://${window.location.host}/ws`);
const logs = document.getElementById('logs')!;

socket.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    logs.insertAdjacentHTML('beforeend', `<div>> ${msg.text}</div>`);
};

function sendMessage() {
    const input = document.getElementById('msgInput') as HTMLInputElement;
    if (input.value) {
        socket.send(JSON.stringify({ text: input.value }));
        input.value = '';
    }
}
