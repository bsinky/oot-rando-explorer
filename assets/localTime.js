"use strict";

/**
 * Converts a UTC time element to local time.
 * @param {Element} timeEl - The time element to convert.
 */
function convertToLocalTimes(timeEl) {
    const datetime = timeEl.getAttribute('datetime');
    if (datetime) {
        const date = new Date(datetime);
        const localTime = date.toLocaleString();
        timeEl.textContent = localTime;
    }
}

document.addEventListener('DOMContentLoaded', function() {
    const timeElements = document.querySelectorAll('time');
    timeElements.forEach(convertToLocalTimes);
});

document.addEventListener("htmx:afterSwap", function(e) {
    let content = e.target;

    content.querySelectorAll("time").forEach(convertToLocalTimes);
});