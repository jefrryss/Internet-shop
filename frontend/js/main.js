function findId(obj) {
    return obj.order_id || obj.orderId || obj.id || obj.ID || obj.Id || "---";
}

function showStatusToast(message, type) {
    let container = document.getElementById('toast-container');
    if (!container) {
        container = document.createElement('div');
        container.id = 'toast-container';
        container.style.cssText = "position: fixed; top: 20px; right: 20px; z-index: 9999;";
        document.body.appendChild(container);
    }
    const toast = document.createElement('div');
    toast.className = `alert alert-${type} shadow-lg mb-2`;
    toast.style.minWidth = "250px";
    toast.innerHTML = message;
    container.appendChild(toast);
    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transition = '0.5s';
        setTimeout(() => toast.remove(), 500);
    }, 4000);
}


async function refreshUserData() {
    const balanceEl = document.getElementById('balanceDisplay');
    const list = document.getElementById('ordersList');
    
    try {
        const data = await apiGetBalance();
        if (data) {
         
            balanceEl.innerText = data.balance !== undefined ? data.balance : data.Balance;
            await handleGetOrders(); 
        } else {
           
            balanceEl.innerText = "Счет не создан";
            list.innerHTML = '<li class="list-group-item text-center text-muted">Создайте счет, чтобы увидеть заказы</li>';
        }
    } catch (e) {
        balanceEl.innerText = "---";
    }
}


document.getElementById('userId').addEventListener('input', () => {
    
    clearTimeout(window.searchTimeout);
    window.searchTimeout = setTimeout(refreshUserData, 500);
});


window.addEventListener('load', refreshUserData);


async function handleCreateAccount() {
    try {
        await apiCreateAccount();
        showStatusToast("Счет успешно создан", "success");
        refreshUserData();
    } catch (e) { alert("Ошибка: " + e.message); }
}

async function handleDeposit() {
    const input = document.getElementById('depositAmount');
    if (!input.value) return;
    try {
        await apiDeposit(input.value);
        showStatusToast("Баланс пополнен", "success");
        input.value = '';
        refreshUserData();
    } catch (e) { alert(e.message); }
}

async function handleCreateOrder() {
    const input = document.getElementById('orderAmount');
    if (!input.value) return;
    try {
        const data = await apiCreateOrder(input.value);
        const orderId = findId(data);
        alert(`Заказ создан! ID: ${orderId}`);
        input.value = '';
        
  
        setTimeout(refreshUserData, 2000);
    } catch (e) { alert(e.message); }
}

async function handleGetOrders() {
    const list = document.getElementById('ordersList');
    try {
        const orders = await apiGetOrders();
        list.innerHTML = '';
        if (orders.length === 0) {
            list.innerHTML = '<li class="list-group-item text-center">Заказов пока нет</li>';
            return;
        }
        orders.reverse().forEach(o => {
            const status = (o.status || o.Status || "").toUpperCase();
            const orderId = findId(o);
            let badge = status === 'PAID' ? 'bg-success' : (status === 'FAILED' ? 'bg-danger' : 'bg-warning text-dark');
            list.innerHTML += `
                <li class="list-group-item d-flex justify-content-between align-items-center">
                    <div style="word-break: break-all; font-size: 0.85rem; padding-right:10px;">
                        <strong>ID: ${orderId}</strong><br>
                        Сумма: ${o.amount || o.Amount} ₽
                    </div>
                    <span class="badge ${badge} rounded-pill">${status}</span>
                </li>`;
        });
    } catch (e) { list.innerHTML = 'Ошибка загрузки'; }
}