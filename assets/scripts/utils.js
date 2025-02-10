/**
 * Fetches the current user's ID from the server.
 * @returns {Promise<number>} - A promise that resolves to the user's ID.
 */
async function getCurrentUserID() {
    try {
        let response = await fetch("http://localhost:8080/current-user-id");
        let data = await response.json();
        return data.userID; 
    } catch (error) {
        console.error("Error fetching user ID:", error);
    }
}

/**
 * Fetches the username of a user by their ID.
 * @param {number} id - The ID of the user.
 * @returns {Promise<string>} - A promise that resolves to the username.
 */
async function getUsernameById(id) {
    try {
        let response = await fetch(`http://localhost:8080/user-username/${id}`);
        let data = await response.json();
        return data.success; 
    } catch (error) {
        console.error("Error fetching username:", error);
    }
}