setupPopup(".btn.btn-edit", ".edit-overlay");
document.getElementById("profile_picture").addEventListener("change", function() {
    let label = document.querySelector(".custom-file-upload");
    label.textContent = this.files[0] ? this.files[0].name : "Choose a file";
});