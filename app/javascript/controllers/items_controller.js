import { Controller } from "@hotwired/stimulus";

// Connects to data-controller="items"
export default class extends Controller {
  static targets = ["row"];

  onChange(event) {
    window.location = `${window.location.origin}/?type=${event.target.value}`;
  }

  goToItem(event) {
    window.location = `${window.location.origin}/${event.currentTarget.dataset.id}`;
  }

  exportAttendees() {
    window.location = `${window.location.origin}${window.location.pathname}/export${window.location.search}`;
  }
}
