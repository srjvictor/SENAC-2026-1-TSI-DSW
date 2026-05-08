const API_URL = 'http://localhost:8080/tasks';

// Carrega as tarefas ao abrir a página
document.addEventListener('DOMContentLoaded', loadTasks);

async function loadTasks() {
    try {
        const response = await fetch(API_URL);
        const tasks = await response.json();
        const taskList = document.getElementById('taskList');
        taskList.innerHTML = '';

        if (tasks) {
            tasks.forEach(task => renderTask(task));
        }
    } catch (error) {
        console.error('Erro ao carregar tarefas:', error);
    }
}

function renderTask(task) {
    const taskList = document.getElementById('taskList');
    const li = document.createElement('li');
    if (task.completed) li.classList.add('completed');

    li.innerHTML = `
        <div>
            <input type="checkbox" class="checkbox" ${task.completed ? 'checked' : ''} onchange="toggleTask(${task.id}, this.checked)">
            <span>${task.title}</span>
        </div>
        <button class="delete-btn" onclick="deleteTask(${task.id})">X</button>
    `;
    taskList.appendChild(li);
}

async function addTask() {
    const input = document.getElementById('taskInput');
    const title = input.value.trim();

    if (!title) return;

    try {
        const response = await fetch(API_URL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title: title })
        });

        const newTask = await response.json();
        renderTask(newTask);
        input.value = '';
    } catch (error) {
        console.error('Erro ao adicionar tarefa:', error);
    }
}

async function toggleTask(id, completed) {
    try {
        await fetch(`${API_URL}/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ completed: completed })
        });
        loadTasks(); // Recarrega a lista para atualizar a UI
    } catch (error) {
        console.error('Erro ao atualizar tarefa:', error);
    }
}

async function deleteTask(id) {
    try {
        await fetch(`${API_URL}/${id}`, {
            method: 'DELETE'
        });
        loadTasks(); // Recarrega a lista
    } catch (error) {
        console.error('Erro ao deletar tarefa:', error);
    }
}