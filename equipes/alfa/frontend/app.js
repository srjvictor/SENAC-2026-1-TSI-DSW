const API_URL = 'http://localhost:8080/api/v1/tasks';

const taskForm = document.getElementById('task-form');
const taskInput = document.getElementById('task-input');
const taskList = document.getElementById('task-list');

document.addEventListener('DOMContentLoaded', loadTasks);

async function loadTasks() {
    try {
        const response = await fetch(API_URL);
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
        completeBtn.title = 'Marcar como concluída';
        completeBtn.onclick = () => completeTask(task.id, li);
        actionsDiv.appendChild(completeBtn);
    }

    const deleteBtn = document.createElement('button');
    deleteBtn.textContent = '✗';
    deleteBtn.className = 'btn-delete';
    deleteBtn.title = 'Excluir tarefa';
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
        const response = await fetch(API_URL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
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