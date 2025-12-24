const API_URL = 'http://localhost:8000';

function getHeaders() {
    const userId = document.getElementById('userId').value;
    if (!userId) return null;
    return {
        'Content-Type': 'application/json',
        'X-User-Id': userId
    };
}

async function apiCreateAccount() {
    const headers = getHeaders();
    const res = await fetch(`${API_URL}/accounts`, { method: 'POST', headers });
    if (!res.ok) throw new Error(await res.text());
    return true;
}

async function apiGetBalance() {
    const headers = getHeaders();
    if (!headers) return null;
    const res = await fetch(`${API_URL}/accounts/balance`, { method: 'GET', headers });
    if (!res.ok) return null; 
    return await res.json();
}

async function apiDeposit(amount) {
    const res = await fetch(`${API_URL}/accounts/deposit`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ amount: parseInt(amount) })
    });
    if (!res.ok) throw new Error("Ошибка пополнения");
    return true;
}

async function apiCreateOrder(amount) {
    const res = await fetch(`${API_URL}/orders`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ amount: parseInt(amount) })
    });
    if (!res.ok) throw new Error("Ошибка создания заказа");
    return await res.json();
}

async function apiGetOrders() {
    const headers = getHeaders();
    if (!headers) return [];
    const res = await fetch(`${API_URL}/orders`, { method: 'GET', headers });
    if (!res.ok) return [];
    return await res.json();
}