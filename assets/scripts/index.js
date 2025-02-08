const buttonSearch = document.querySelector(".btn.btn-search");
const buttonFriendRequests = document.querySelector(".btn.btn-friend-requests");

const searchPopup = document.querySelector(".search-overlay");
const friendRequestsPopup = document.querySelector(".friend-requests-overlay");

const searchCloseButton = document.querySelector(".search-overlay .close-btn");
const friendRequestsCloseButton = document.querySelector(".friend-requests-overlay .close-btn");

buttonFriendRequests.addEventListener("click", function() {
    friendRequestsPopup.style.display = "flex";
});

buttonSearch.addEventListener("click", function() {
    searchPopup.style.display = "flex";
});

searchCloseButton.addEventListener("click", function() {
    searchPopup.style.display = "none";
});

friendRequestsCloseButton.addEventListener("click", function() {
    friendRequestsPopup.style.display = "none";
});

searchPopup.addEventListener("click", function(event) {
    if (event.target === searchPopup) {
        searchPopup.style.display = "none";
    }
});

friendRequestsPopup.addEventListener("click", function(event) {
    if (event.target === friendRequestsPopup) {
        friendRequestsPopup.style.display = "none";
    }
});