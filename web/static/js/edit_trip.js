// Activity counter (mevcut activity sayısından başla)
let activityCount = document.querySelectorAll('.activity-item-edit').length;

// Expense counter (mevcut expense sayısından başla)
let expenseCount = document.querySelectorAll('.expense-item-edit').length;

// Add Activity
function addActivity() {
    const container = document.getElementById('activitiesContainer');
    const index = activityCount++;

    const activityHTML = `
        <div class="dynamic-item activity-item-edit">
            <div class="dynamic-item-header">
                <h4><i class="fas fa-hiking"></i> Activity ${index + 1}</h4>
                <button type="button" class="btn-remove" onclick="removeActivity(this)">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Activity Name *</label>
                    <input type="text" name="activities[${index}][name]" required>
                </div>
                <div class="form-group">
                    <label>Date *</label>
                    <input type="date" name="activities[${index}][date]" required>
                </div>
            </div>
            <div class="form-group">
                <label>Location</label>
                <input type="text" name="activities[${index}][location]">
            </div>
            <div class="form-group">
                <label>Description</label>
                <textarea name="activities[${index}][description]" rows="2"></textarea>
            </div>
        </div>
    `;

    container.insertAdjacentHTML('beforeend', activityHTML);
}

// Remove Activity
function removeActivity(button) {
    if (confirm('Are you sure you want to remove this activity?')) {
        button.closest('.dynamic-item').remove();
    }
}

// Add Expense
function addExpense() {
    const container = document.getElementById('expensesContainer');
    const index = expenseCount++;

    const expenseHTML = `
        <div class="dynamic-item expense-item-edit">
            <div class="dynamic-item-header">
                <h4><i class="fas fa-receipt"></i> Expense ${index + 1}</h4>
                <button type="button" class="btn-remove" onclick="removeExpense(this)">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Category *</label>
                    <select name="expenses[${index}][category]" required>
                        <option value="">Select category</option>
                        <option value="food">Food</option>
                        <option value="transport">Transport</option>
                        <option value="accommodation">Accommodation</option>
                        <option value="entertainment">Entertainment</option>
                        <option value="other">Other</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Amount (EUR) *</label>
                    <input type="number" name="expenses[${index}][amount]" step="0.01" min="0" required>
                </div>
                <div class="form-group">
                    <label>Date *</label>
                    <input type="date" name="expenses[${index}][expense_date]" required>
                </div>
            </div>
        </div>
    `;

    container.insertAdjacentHTML('beforeend', expenseHTML);
}

// Remove Expense
function removeExpense(button) {
    if (confirm('Are you sure you want to remove this expense?')) {
        button.closest('.dynamic-item').remove();
    }
}

// Form Submit Handler
document.getElementById('editTripForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = {
        title: document.getElementById('title').value,
        destination: document.getElementById('destination').value,
        start_date: document.getElementById('start_date').value,
        end_date: document.getElementById('end_date').value,
        description: document.getElementById('description').value,
        budget: parseFloat(document.getElementById('budget').value) || 0,
        is_public: document.getElementById('is_public').checked
    };

    try {
        const response = await fetch(`/api/trips/${tripId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify(formData)
        });

        if (response.ok) {
            alert('Trip updated successfully!');
            window.location.href = `/trips/${tripId}`;
        } else {
            const error = await response.text();
            alert('Failed to update trip: ' + error);
        }
    } catch (error) {
        console.error('Error:', error);
        alert('An error occurred while updating the trip');
    }
});