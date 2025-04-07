const API_URL = '/api';
let token = localStorage.getItem('token');

// Check if user is already logged in
function checkAuthStatus() {
    if (token) {
        $.ajax({
            url: `${API_URL}/user`,
            type: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            success: function(response) {
                showAuthenticatedUI(response);
                loadTasks();
            },
            error: function() {
                // Token invalid or expired
                localStorage.removeItem('token');
                token = null;
                showLoginUI();
            }
        });
    } else {
        showLoginUI();
    }
}

function showLoginUI() {
    $('#auth-container').removeClass('hidden');
    $('#task-container').addClass('hidden');
    $('#user-info').addClass('hidden');
}

function showAuthenticatedUI(user) {
    $('#auth-container').addClass('hidden');
    $('#task-container').removeClass('hidden');
    $('#user-info').removeClass('hidden');
    $('#user-email').text(user.email);
}

// Handle login form submission
$('#login').submit(function(e) {
    e.preventDefault();
    $('#login-error').text('');
    
    const email = $('#login-email').val();
    const password = $('#login-password').val();
    
    $.ajax({
        url: `${API_URL}/auth/login`,
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
            email: email,
            password: password
        }),
        success: function(response) {
            token = response.token;
            localStorage.setItem('token', token);
            
            // Get current user info
            $.ajax({
                url: `${API_URL}/user`,
                type: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                success: function(userResponse) {
                    showAuthenticatedUI(userResponse);
                    loadTasks();
                },
                error: function(xhr) {
                    $('#login-error').text('Error loading user information');
                }
            });
        },
        error: function(xhr) {
            let errorMessage = 'Login failed';
            
            if (xhr.status === 401) {
                errorMessage = 'Invalid email or password';
            } else if (xhr.responseJSON && xhr.responseJSON.message) {
                errorMessage = xhr.responseJSON.message;
            }
            
            $('#login-error').text(errorMessage);
        }
    });
});

// Handle register form submission
$('#register').submit(function(e) {
    e.preventDefault();
    $('#register-error').text('');
    
    const email = $('#register-email').val();
    const password = $('#register-password').val();
    
    if (password.length < 8) {
        $('#register-error').text('Password must be at least 8 characters');
        return;
    }
    
    $.ajax({
        url: `${API_URL}/user`,
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
            email: email,
            password: password
        }),
        success: function(response) {
            token = response.token;
            localStorage.setItem('token', token);
            
            // Get current user info
            $.ajax({
                url: `${API_URL}/user`,
                type: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                success: function(userResponse) {
                    showAuthenticatedUI(userResponse);
                    loadTasks();
                },
                error: function(xhr) {
                    $('#register-error').text('Error loading user information');
                }
            });
        },
        error: function(xhr) {
            let errorMessage = 'Registration failed';
            
            if (xhr.status === 409) {
                errorMessage = 'Email is already taken';
            } else if (xhr.responseJSON && xhr.responseJSON.message) {
                errorMessage = xhr.responseJSON.message;
            }
            
            $('#register-error').text(errorMessage);
        }
    });
});

// Handle logout
$('#logout-btn').click(function() {
    localStorage.removeItem('token');
    token = null;
    showLoginUI();
});

// Tab switching
$('.tab-btn').click(function() {
    const tabId = $(this).data('tab');
    
    $('.tab-btn').removeClass('active');
    $(this).addClass('active');
    
    $('.form-container').removeClass('active');
    $(`#${tabId}-form`).addClass('active');
});