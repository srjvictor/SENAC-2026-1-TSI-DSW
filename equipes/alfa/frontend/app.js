const API_URL = 'http://localhost:8080/api/v1';

// Elementos Auth
const authContainer = document.getElementById('auth-container');
const appContainer = document.getElementById('app-container');
const authForm = document.getElementById('auth-form');
const authTitle = document.getElementById('auth-title');
const authBtn = document.getElementById('auth-btn');
const toggleAuthLink = document.getElementById('toggle-auth-link');
const usernameInput = document.getElementById('username');
const passwordInput = document.getElementById('password');
const authError = document.getElementById('auth-error');
const logoutBtn = document.getElementById('logout-btn');

// Elementos Tarefas
const taskForm = document.getElementById('task-form');
const taskInput = document.getElementById('task-input');
const taskList = document.getElementById('task-list');

let isLoginMode = true;

// Inicialização
document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem('token');
    if (token) {
        showApp();
    } else {
        showAuth();
    }
});

// Headers Dinâmicos com Token
function getHeaders() {
    return {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
    };
}

// Alternar entre Login e Cadastro
toggleAuthLink.addEventListener('click', (e) => {
    e.preventDefault();
    isLoginMode = !isLoginMode;
    authTitle.textContent = isLoginMode ? 'Login' : 'Cadastro';
    authBtn.textContent = isLoginMode ? 'Entrar' : 'Registrar';
    toggleAuthLink.textContent = isLoginMode ? 'Não tem uma conta? Cadastre-se' : 'Já tem uma conta? Faça login';
    authError.textContent = '';
});

// Submissão do Formulário de Auth
authForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const endpoint = isLoginMode ? '/login' : '/register';

    try {
        const response = await fetch(`${API_URL}${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                username: usernameInput.value.trim(),
                password: passwordInput.value.trim()
            })
        });

        if (!response.ok) throw new Error('Falha na autenticação. Verifique os dados.');

        if (!isLoginMode) {
            alert('Cadastro realizado com sucesso! Faça login.');
            toggleAuthLink.click(); // Volta pro login
        } else {
            const data = await response.json();
            localStorage.setItem('token', data.token);
            showApp();
        }
    } catch (error) {
        authError.textContent = error.message;
    }
});

// Logout
logoutBtn.addEventListener('click', () => {
    localStorage.removeItem('token');
    showAuth();
});

// Navegação de Telas
function showApp() {
    authContainer.style.display = 'none';
    appContainer.style.display = 'block';
    loadTasks();
}

function showAuth() {
    appContainer.style.display = 'none';
    authContainer.style.display = 'block';
    usernameInput.value = '';
    passwordInput.value = '';
}

// === Lógica de Tarefas (Protegida) ===

async function loadTasks() {
    try {
        const response = await fetch(`${API_URL}/tasks`, { headers: getHeaders() });
        if (response.status === 401) {
            logoutBtn.click();
            return;
        }
        if (!response.ok) throw new Error('Falha ao buscar tarefas');

        const tasks = await response.json();
        taskList.innerHTML = '';
        tasks.forEach(task => renderTaskElement(task));
    } catch (error) {
        console.error('Erro:', error);
    }
}

function renderTaskElement(task) {
    const li = document.createElement('li');
    li.dataset.id = task.id;
    if (task.completed) li.classList.add('completed');

    const titleSpan = document.createElement('span');
    titleSpan.textContent = task.title;

    const actionsDiv = document.createElement('div');
    actionsDiv.className = 'task-actions';

    if (!task.completed) {
        const completeBtn = document.createElement('button');
        completeBtn.textContent = '✓';
        completeBtn.className = 'btn-complete';
        completeBtn.onclick = () => completeTask(task.id, li);
        actionsDiv.appendChild(completeBtn);
    }

    const deleteBtn = document.createElement('button');
    deleteBtn.textContent = '✗';
    deleteBtn.className = 'btn-delete';
    deleteBtn.onclick = () => deleteTask(task.id, li);
    actionsDiv.appendChild(deleteBtn);

    li.appendChild(titleSpan);
    li.appendChild(actionsDiv);
    taskList.appendChild(li);
}

taskForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const title = taskInput.value.trim();
    if (!title) return;

    try {
        const response = await fetch(`${API_URL}/tasks`, {
            method: 'POST',
            headers: getHeaders(),
            body: JSON.stringify({ title: title })
        });

        if (response.ok) {
            const newTask = await response.json();
            renderTaskElement(newTask);
            taskInput.value = '';
        }
    } catch (error) {
        console.error('Erro ao criar tarefa:', error);
    }
});

async function completeTask(id, liElement) {
    try {
        const response = await fetch(`${API_URL}/tasks/${id}`, {
            method: 'PUT',
            headers: getHeaders()
        });

        if (response.ok) {
            liElement.classList.add('completed');
            const completeBtn = liElement.querySelector('.btn-complete');
            if (completeBtn) completeBtn.remove();
        }
    } catch (error) {
        console.error('Erro ao concluir tarefa:', error);
    }
}

async function deleteTask(id, liElement) {
    try {
        const response = await fetch(`${API_URL}/tasks/${id}`, {
            method: 'DELETE',
            headers: getHeaders()
        });

        if (response.ok) {
            liElement.remove();
        }
    } catch (error) {
        console.error('Erro ao excluir tarefa:', error);
    }
}