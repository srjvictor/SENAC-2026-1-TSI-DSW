// Ajustado para o novo endpoint da sua API Go
const API_URL = 'http://localhost:8080/api/v1/tasks';

document.addEventListener('DOMContentLoaded', () => {
    loadTasks();
    
    // Interceptar o submit do formulário evita o recarregamento da página
    document.getElementById('taskForm').addEventListener('submit', function(e) {
        e.preventDefault();
        addTask();
    });
});

async function loadTasks() {
    try {
        const response = await fetch(API_URL);
        const tasks = await response.json();
        
        const taskList = document.getElementById('taskList');
        taskList.innerHTML = ''; // Limpar lista atual (seguro aqui, pois não há dados do usuário)

        if (tasks && tasks.length > 0) {
            tasks.forEach(task => renderTask(task));
        } else {
            taskList.innerHTML = '<p style="text-align:center; color:#6B7280;">Nenhuma tarefa pendente. Você está livre!</p>';
        }
    } catch (error) {
        console.error('Erro ao carregar tarefas:', error);
    }
}

// O SEGREDO DA SEGURANÇA (Prevenção XSS)
// Em vez de injetar uma string com innerHTML, criamos os nós do DOM manualmente.
// Isso garante que se um usuário digitar "<script>alert('hack')</script>"
// o navegador tratará isso estritamente como texto, e não como código executável.
function renderTask(task) {
    const taskList = document.getElementById('taskList');
    
    // Remove a mensagem de "lista vazia" se for a primeira tarefa
    if (taskList.querySelector('p')) {
        taskList.innerHTML = '';
    }

    const li = document.createElement('li');
    li.className = `task-item ${task.completed ? 'completed' : ''}`;
    li.id = `task-${task.id}`; // Para manipulação dinâmica sem recarregar a página

    const contentDiv = document.createElement('div');
    contentDiv.className = 'task-content';

    const checkbox = document.createElement('input');
    checkbox.type = 'checkbox';
    checkbox.className = 'task-checkbox';
    checkbox.checked = task.completed;
    checkbox.addEventListener('change', () => toggleTask(task.id, checkbox.checked));

    const span = document.createElement('span');
    span.className = 'task-text';
    // MÁXIMA SEGURANÇA: textContent escapa automaticamente caracteres HTML perigosos
    span.textContent = task.title;

    const deleteBtn = document.createElement('button');
    deleteBtn.className = 'delete-btn';
    deleteBtn.textContent = 'Excluir';
    deleteBtn.addEventListener('click', () => deleteTask(task.id));

    // Montando as peças (como um Lego)
    contentDiv.appendChild(checkbox);
    contentDiv.appendChild(span);
    li.appendChild(contentDiv);
    li.appendChild(deleteBtn);
    
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

        if (response.ok) {
            const newTask = await response.json();
            renderTask(newTask); // Renderiza apenas a nova tarefa, sem recarregar tudo
            input.value = ''; // Limpa o input
            input.focus();
        }
    } catch (error) {
        console.error('Erro ao adicionar tarefa:', error);
    }
}

async function toggleTask(id, completed) {
    try {
        const response = await fetch(`${API_URL}/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ completed: completed })
        });

        if (response.ok) {
            // Atualiza o DOM dinamicamente (UX muito mais rápida que recarregar a página)
            const li = document.getElementById(`task-${id}`);
            if (completed) {
                li.classList.add('completed');
            } else {
                li.classList.remove('completed');
            }
        }
    } catch (error) {
        console.error('Erro ao atualizar tarefa:', error);
    }
}

async function deleteTask(id) {
    try {
        const response = await fetch(`${API_URL}/${id}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            // Remove o elemento do DOM com uma pequena animação de sumiço
            const li = document.getElementById(`task-${id}`);
            li.style.opacity = '0';
            setTimeout(() => {
                li.remove();
                
                // Se apagou a última, mostra a mensagem vazia recarregando a lista
                const taskList = document.getElementById('taskList');
                if (taskList.children.length === 0) {
                    loadTasks();
                }
            }, 200);
        }
    } catch (error) {
        console.error('Erro ao deletar tarefa:', error);
    }
}