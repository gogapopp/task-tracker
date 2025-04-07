let currentFilter = 'all';

// Load tasks based on current filter
function loadTasks() {
    if (!token) return;
    
    let url = `${API_URL}/tasks`;
    
    if (currentFilter !== 'all') {
        const completed = currentFilter === 'completed';
        url += `?completed=${completed}`;
    }
    
    $.ajax({
        url: url,
        type: 'GET',
        headers: {
            'Authorization': `Bearer ${token}`
        },
        success: function(response) {
            displayTasks(response.tasks);
        },
        error: function(xhr) {
            console.error('Error loading tasks:', xhr.responseText);
        }
    });
}

// Display tasks in the UI
function displayTasks(tasks) {
    const tasksList = $('#tasks-list');
    tasksList.empty();
    
    if (tasks.length === 0) {
        tasksList.html('<div class="no-tasks-message">No tasks to display</div>');
        return;
    }
    
    tasks.forEach(task => {
        const taskStatus = task.completed ? 'completed' : 'pending';
        const statusText = task.completed ? 'Completed' : 'Pending';
        const statusClass = task.completed ? 'status-completed' : 'status-pending';
        
        const completedDate = task.completed_at ? `Completed on: ${formatDate(new Date(task.completed_at))}` : '';
        
        const taskItem = `
            <div class="task-item ${taskStatus}" data-id="${task.id}">
                <div class="task-header">
                    <h3 class="task-title">${escapeHtml(task.title)}</h3>
                    <div class="task-actions-btn">
                        <button class="edit-task">Edit</button>
                        <button class="delete-task">Delete</button>
                    </div>
                </div>
                <div class="task-description">${escapeHtml(task.description)}</div>
                <div class="task-meta">
                    <div>
                        <span class="task-status ${statusClass}">${statusText}</span>
                        <span class="task-date">Created: ${formatDate(new Date(task.created_at))}</span>
                        ${completedDate ? `<span class="task-completed-date"> • ${completedDate}</span>` : ''}
                    </div>
                </div>
            </div>
        `;
        
        tasksList.append(taskItem);
    });
    
    // Add event listeners for edit and delete buttons
    $('.edit-task').click(function() {
        const taskId = $(this).closest('.task-item').data('id');
        openEditTaskModal(taskId);
    });
    
    $('.delete-task').click(function() {
        const taskId = $(this).closest('.task-item').data('id');
        if (confirm('Are you sure you want to delete this task?')) {
            deleteTask(taskId);
        }
    });
}

// Add new task
$('#add-task-btn').click(function() {
    $('#add-task-form').toggleClass('hidden');
});

$('#cancel-add-task').click(function() {
    $('#add-task-form').addClass('hidden');
    $('#create-task')[0].reset();
});

$('#create-task').submit(function(e) {
    e.preventDefault();
    
    const title = $('#task-title').val().trim();
    const description = $('#task-description').val().trim();
    
    if (!title) {
        alert('Task title is required!');
        return;
    }
    
    $.ajax({
        url: `${API_URL}/tasks`,
        type: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        data: JSON.stringify({
            title: title,
            description: description
        }),
        success: function() {
            // Clear form and hide it
            $('#create-task')[0].reset();
            $('#add-task-form').addClass('hidden');
            // Reload tasks
            loadTasks();
        },
        error: function(xhr) {
            console.error('Error creating task:', xhr.responseText);
            alert('Failed to create task. Please try again.');
        }
    });
});

// Edit task
function openEditTaskModal(taskId) {
    // Get task details
    $.ajax({
        url: `${API_URL}/tasks/${taskId}`,
        type: 'GET',
        headers: {
            'Authorization': `Bearer ${token}`
        },
        success: function(task) {
            // Fill form with task data
            $('#edit-task-id').val(task.id);
            $('#edit-task-title').val(task.title);
            $('#edit-task-description').val(task.description);
            $('#edit-task-completed').prop('checked', task.completed);
            
            // Show modal
            $('#task-edit-modal').removeClass('hidden');
        },
        error: function(xhr) {
            console.error('Error loading task details:', xhr.responseText);
            alert('Failed to load task details. Please try again.');
        }
    });
}

$('#edit-task-form').submit(function(e) {
    e.preventDefault();
    
    const taskId = $('#edit-task-id').val();
    const title = $('#edit-task-title').val().trim();
    const description = $('#edit-task-description').val().trim();
    const completed = $('#edit-task-completed').is(':checked');
    
    if (!title) {
        alert('Task title is required!');
        return;
    }
    
    $.ajax({
        url: `${API_URL}/tasks/${taskId}`,
        type: 'PUT',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        data: JSON.stringify({
            title: title,
            description: description,
            completed: completed
        }),
        success: function() {
            // Close modal and reload tasks
            $('#task-edit-modal').addClass('hidden');
            loadTasks();
        },
        error: function(xhr) {
            console.error('Error updating task:', xhr.responseText);
            alert('Failed to update task. Please try again.');
        }
    });
});

// Delete task
function deleteTask(taskId) {
    $.ajax({
        url: `${API_URL}/tasks/${taskId}`,
        type: 'DELETE',
        headers: {
            'Authorization': `Bearer ${token}`
        },
        success: function() {
            loadTasks();
        },
        error: function(xhr) {
            console.error('Error deleting task:', xhr.responseText);
            alert('Failed to delete task. Please try again.');
        }
    });
}

// Close edit modal
$('.close-modal, .cancel-btn').click(function() {
    $('#task-edit-modal').addClass('hidden');
});

// Filter tasks
$('.filter-btn').click(function() {
    $('.filter-btn').removeClass('active');
    $(this).addClass('active');
    
    currentFilter = $(this).data('filter');
    loadTasks();
});

// Helper functions
function formatDate(date) {
    return date.toLocaleString();
}

function escapeHtml(str) {
    if (!str) return '';
    return str
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}