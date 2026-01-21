function toggleEditMode() {
    const editSection = document.getElementById('editSection');
    if (editSection.style.display === 'none') {
        editSection.style.display = 'block';
        editSection.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    } else {
        editSection.style.display = 'none';
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const editForm = document.getElementById('editProfileForm');
    if (editForm) {
        editForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            const formData = {
                first_name: document.getElementById('first_name').value.trim(),
                last_name: document.getElementById('last_name').value.trim()
            };

            if (!formData.first_name || !formData.last_name) {
                alert('Please fill in both first name and last name');
                return;
            }

            try {
                const response = await fetch('/api/users/profile', {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(formData)
                });

                if (response.ok) {
                    alert('Profile updated successfully!');
                    window.location.reload();
                } else {
                    const data = await response.json();
                    alert('Error: ' + (data.message || 'Failed to update profile'));
                }
            } catch (error) {
                console.error('Error:', error);
                alert('Error updating profile. Please try again.');
            }
        });
    }
});
