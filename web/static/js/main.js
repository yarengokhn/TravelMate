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

function filterTrips(filter) {
    const trips = document.querySelectorAll('.trip-item');
    const now = new Date();

    trips.forEach(trip => {
        const startDate = new Date(trip.dataset.startDate);

        switch (filter) {
            case 'upcoming':
                trip.style.display = startDate > now ? 'block' : 'none';
                break;
            case 'past':
                trip.style.display = startDate < now ? 'block' : 'none';
                break;
            default:
                trip.style.display = 'block';
        }
    });
}

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



