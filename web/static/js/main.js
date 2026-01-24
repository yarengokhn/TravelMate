// Mobile Navigation Toggle
document.getElementById('navToggle')?.addEventListener('click', function () {
    const navMenu = document.querySelector('.nav-menu');
    navMenu.classList.toggle('active');
});

// Login Form Handler
document.getElementById('loginForm')?.addEventListener('submit', async function (e) {
    e.preventDefault();

    const formData = {
        email: document.getElementById('email').value,
        password: document.getElementById('password').value
    };

    try {
        const response = await fetch('/api/users/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(formData)
        });

        if (response.ok) {
            window.location.href = '/dashboard';
        } else {
            const error = await response.json();
            showAlert(error.message || 'Login failed', 'error');
        }
    } catch (err) {
        showAlert('An error occurred. Please try again.', 'error');
    }
});

// Register Form Handler
document.getElementById('registerForm')?.addEventListener('submit', async function (e) {
    e.preventDefault();

    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirm_password').value;

    if (password !== confirmPassword) {
        showAlert('Passwords do not match', 'error');
        return;
    }

    const formData = {
        email: document.getElementById('email').value,
        password: password,
        first_name: document.getElementById('first_name').value,
        last_name: document.getElementById('last_name').value
    };

    try {
        const response = await fetch('/api/users/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(formData)
        });

        if (response.ok) {
            showAlert('Registration successful! Redirecting...', 'success');
            setTimeout(() => {
                window.location.href = '/login';
            }, 1500);
        } else {
            const error = await response.json();
            showAlert(error.message || 'Registration failed', 'error');
        }
    } catch (err) {
        showAlert('An error occurred. Please try again.', 'error');
    }
});



// Delete Trip Function
async function deleteTrip(tripId) {
    if (!confirm('Are you sure you want to delete this trip? This action cannot be undone.')) {
        return;
    }

    try {
        const response = await fetch(`/api/trips/${tripId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            showAlert('Trip deleted successfully', 'success');
            setTimeout(() => {
                window.location.reload();
            }, 1000);
        } else {
            showAlert('Failed to delete trip', 'error');
        }
    } catch (err) {
        showAlert('An error occurred. Please try again.', 'error');
    }
}

// Filter Tabs (Dashboard)
document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', function () {
        document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
        this.classList.add('active');

        const filter = this.dataset.filter;
        filterTrips(filter);
    });
});

// Alert Helper Function
function showAlert(message, type) {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type}`;
    alertDiv.innerHTML = `
        <i class="fas fa-${type === 'success' ? 'check-circle' : 'exclamation-circle'}"></i>
        ${message}
    `;

    const main = document.querySelector('main');
    main.insertBefore(alertDiv, main.firstChild);

    setTimeout(() => {
        alertDiv.remove();
    }, 5000);
}

// Auto-dismiss alerts
setTimeout(() => {
    document.querySelectorAll('.alert').forEach(alert => {
        alert.style.opacity = '0';
        setTimeout(() => alert.remove(), 300);
    });
}, 5000);

let activityCounter = 0;
let expenseCounter = 0;

// Add Activity Field
function addActivity() {
    activityCounter++;
    const container = document.getElementById('activitiesContainer');

    const activityDiv = document.createElement('div');
    activityDiv.className = 'dynamic-field';
    activityDiv.id = `activity-${activityCounter}`;
    activityDiv.innerHTML = `
        <div class="field-header">
            <h4><i class="fas fa-hiking"></i> Activity ${activityCounter}</h4>
            <button type="button" class="btn-remove" onclick="removeField('activity-${activityCounter}')">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <div class="form-group">
            <label>Activity Name *</label>
            <input type="text" class="activity-name" placeholder="e.g., Visit Eiffel Tower" required>
        </div>
        <div class="form-row">
            <div class="form-group">
                <label>Location</label>
                <input type="text" class="activity-location" placeholder="e.g., Champ de Mars, Paris">
            </div>
            <div class="form-group">
                <label>Date *</label>
                <input type="date" class="activity-date" required>
            </div>
        </div>
        <div class="form-group">
            <label>Description</label>
            <textarea class="activity-description" rows="2" placeholder="What will you do?"></textarea>
        </div>
    `;

    container.appendChild(activityDiv);
}

// Add Expense Field
function addExpense() {
    expenseCounter++;
    const container = document.getElementById('expensesContainer');

    const expenseDiv = document.createElement('div');
    expenseDiv.className = 'dynamic-field';
    expenseDiv.id = `expense-${expenseCounter}`;
    expenseDiv.innerHTML = `
        <div class="field-header">
            <h4><i class="fas fa-receipt"></i> Expense ${expenseCounter}</h4>
            <button type="button" class="btn-remove" onclick="removeField('expense-${expenseCounter}')">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <div class="form-row">
            <div class="form-group">
                <label>Category *</label>
                <select class="expense-category" required>
                    <option value="">Select category</option>
                    <option value="Food">🍽️ Food & Dining</option>
                    <option value="Transport">🚗 Transportation</option>
                    <option value="Accommodation">🏨 Accommodation</option>
                    <option value="Entertainment">🎭 Entertainment</option>
                    <option value="Shopping">🛍️ Shopping</option>
                    <option value="Other">📌 Other</option>
                </select>
            </div>
            <div class="form-group">
                <label>Amount (EUR) *</label>
                <input type="number" class="expense-amount" step="0.01" min="0" placeholder="0.00" required>
            </div>
            <div class="form-group">
                <label>Date *</label>
                <input type="date" class="expense-date" required>
            </div>
        </div>
    `;

    container.appendChild(expenseDiv);
}

// Remove Field
function removeField(id) {
    const field = document.getElementById(id);
    if (field) {
        field.remove();
    }
}

// Form Submit Handler
document.getElementById('createTripForm').addEventListener('submit', async function (e) {
    e.preventDefault();

    // Collect basic trip data
    const formData = {
        title: document.getElementById('title').value,
        destination: document.getElementById('destination').value,
        start_date: document.getElementById('start_date').value,
        end_date: document.getElementById('end_date').value,
        description: document.getElementById('description').value,
        budget: parseFloat(document.getElementById('budget').value) || 0,
        is_public: document.getElementById('is_public').checked
    };

    // 🆕 Collect Activities
    const activities = [];
    document.querySelectorAll('#activitiesContainer .dynamic-field').forEach(field => {
        const name = field.querySelector('.activity-name').value;
        const location = field.querySelector('.activity-location').value;
        const date = field.querySelector('.activity-date').value;
        const description = field.querySelector('.activity-description').value;

        if (name && date) {
            activities.push({
                name: name,
                description: description,
                location: location,
                date: date
            });
        }
    });

    // 🆕 Collect Expenses
    const expenses = [];
    document.querySelectorAll('#expensesContainer .dynamic-field').forEach(field => {
        const category = field.querySelector('.expense-category').value;
        const amount = parseFloat(field.querySelector('.expense-amount').value);
        const date = field.querySelector('.expense-date').value;

        if (category && amount && date) {
            expenses.push({
                category: category,
                amount: amount,
                expense_date: date
            });
        }
    });

    // Add to formData if not empty
    if (activities.length > 0) {
        formData.activities = activities;
    }

    if (expenses.length > 0) {
        formData.expenses = expenses;
    }

    try {
        const response = await fetch('/api/trips', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(formData)
        });

        if (response.ok) {
            const result = await response.json();
            alert('✅ Trip created successfully!');
            window.location.href = `/trips/${result.trip.id}`;
        } else {
            const error = await response.text();
            alert('❌ Error: ' + error);
        }
    } catch (error) {
        console.error('Error:', error);
        alert('❌ Failed to create trip');
    }
});



