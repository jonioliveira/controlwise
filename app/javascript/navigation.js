document.addEventListener("DOMContentLoaded", function () {
  document
    .getElementById("userMenuButton")
    .addEventListener("click", function () {
      const menu = document.getElementById("userMenu");
      menu.classList.toggle("hidden");
    });

  // Close the dropdown if clicked outside
  window.addEventListener("click", function (e) {
    const menuButton = document.getElementById("userMenuButton");
    const menu = document.getElementById("userMenu");
    if (!menuButton.contains(e.target) && !menu.contains(e.target)) {
      menu.classList.add("hidden");
    }
  });
});
