document.addEventListener("DOMContentLoaded", function () {
  const addItemButton = document.getElementById("addItemButton");
  const closeModalButton = document.getElementById("closeModalButton");
  const confirmButton = document.getElementById("confirmButton");
  const modal = document.getElementById("budgetItemsModal");
  const searchInput = document.getElementById("searchInput");
  const results = document.getElementById("results");
  const itemsDataInput = document.getElementById('itemsData');
  const form = document.querySelector('form.contents');

  if (addItemButton) {
    addItemButton.addEventListener("click", function (event) {
      event.preventDefault(); // Prevent form submission
      modal.classList.remove("hidden");
    });
  }

  if (closeModalButton) {
    closeModalButton.addEventListener("click", function (event) {
      event.preventDefault(); // Prevent any default action
      modal.classList.add("hidden");
    });
  }

  if (confirmButton) {
    confirmButton.addEventListener("click", function (event) {
      event.preventDefault(); // Prevent form submission
      // Add your confirm logic here
      modal.classList.add("hidden");
    });
  }

  searchInput.addEventListener("input", function () {
    console.log("input event fired");
    const query = searchInput.value;

    fetch(`/items/search?q[name_cont]=${query}`, {
      headers: {
        "X-Requested-With": "XMLHttpRequest",
      },
    })
      .then((response) => response.text())
      .then((html) => {
        results.innerHTML = html;
      })
      .catch((error) => {
        console.error("Error:", error);
      });
  });

  // Event delegation to handle click event for dynamically added buttons
  document.addEventListener("click", function (event) {
    if (event.target && event.target.classList.contains("add-item-button")) {
      const button = event.target;
      const itemId = button.getAttribute("data-item-id");
      const itemName = button.getAttribute("data-item-name");
      const itemMargin = button.getAttribute("data-item-margin");
      const itemPrice = button.getAttribute("data-item-price");
      const itemUnit = button.getAttribute("data-item-unit");

      // Create a new row for the main table
      const budgetItemsTable = document
        .getElementById("budgetItemsTable")
        .getElementsByTagName("tbody")[0];
      const existingRow = budgetItemsTable.querySelector(
        `tr[data-item-id="${itemId}"]`
      );

      if (!existingRow) {
        const newRow = document.createElement("tr");
        newRow.setAttribute("data-item-id", itemId);

        const total = itemPrice * (1 + itemMargin / 100);

        newRow.innerHTML = `
          <td>${itemName}</td>
          <td>${itemPrice}</td>
          <td>${itemMargin}</td>
          <td>${itemUnit}</td>
          <td><input type="number" value="1" class="quantity-input border rounded px-2 py-1" min="1" /></td>
          <td class="total">${total.toFixed(2)}</td>
          <td>
            <button class="remove-item-button bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded" data-item-id="${itemId}">
              Remove
            </button>
          </td>
        `;

        budgetItemsTable.appendChild(newRow);
        updateTotalBudget();
      } else {
        alert("This item is already added to the table.");
      }
    }
  });

  // Update total value when quantity changes
  document.addEventListener("input", function (event) {
    if (event.target && event.target.classList.contains("quantity-input")) {
      const input = event.target;

      let quantity = parseInt(input.value, 10);

      // Ensure the quantity is not less than 1
      if (quantity < 1) {
        quantity = 1;
        input.value = 1; // Reset the input value to 1 if it is less than 1
      }

      const row = input.closest("tr");
      const margin = parseFloat(row.children[2].textContent);
      const price = parseFloat(row.children[1].textContent);
      const total = quantity * price * (1 + margin / 100);
      row.querySelector(".total").textContent = total.toFixed(2);
      updateTotalBudget();
    }
  });

  document.addEventListener("click", function (event) {
    if (event.target && event.target.classList.contains("remove-item-button")) {
      const button = event.target;
      const itemId = button.getAttribute("data-item-id");
      const row = button.closest("tr");
      row.remove();
      updateTotalBudget();
    }
  });

  // Function to update the total budget
  function updateTotalBudget() {
    const totalBudgetElement = document.getElementById("totalBudget");
    const totalElements = document.querySelectorAll(".total");
    let totalBudget = 0;

    totalElements.forEach((element) => {
      totalBudget += parseFloat(element.textContent);
    });

    const budgetMargin = parseFloat(document.getElementById('budgetMargin').value) || 0;
    totalBudget = totalBudget * (1 + budgetMargin / 100);

    totalBudgetElement.textContent = totalBudget.toFixed(2);
  }

  // Update total budget when budget margin changes
  document.getElementById('budgetMargin').addEventListener('input', updateTotalBudget);

  // Serialize table data before form submission
  form.addEventListener('submit', function(event) {
    const rows = document.querySelectorAll('#budgetItemsTable tbody tr');
    const items = [];

    rows.forEach(row => {
      const itemId = row.getAttribute('data-item-id');
      const itemName = row.children[0].textContent;
      const itemQuantity = row.querySelector('.quantity-input').value;
      const itemMargin = parseFloat(row.children[2].textContent);
      const itemPrice = parseFloat(row.children[1].textContent);
      const itemTotal = parseFloat(row.querySelector('.total').textContent);

      items.push({
        id: itemId,
        name: itemName,
        quantity: itemQuantity,
        margin: itemMargin,
        price: itemPrice,
        total: itemTotal
      });
    });

    itemsDataInput.value = JSON.stringify(items);
  });

});
