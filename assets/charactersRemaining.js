"use strict";

/**
 * 
 * @param {Element} textArea 
 * @param {Element} charactersRemainingText 
 */
function charactersRemainingUpdate(textArea, charactersRemainingText) {
    let maxLength = textArea.getAttribute("maxlength");
    let value = textArea.value || "";
    let remainingLen = maxLength - value.length;

    charactersRemainingText.innerHTML = `${remainingLen} characters remaining`;
    if (remainingLen < 20) {
        charactersRemainingText.classList.add("charactersRemaining-few")
    }
    else {
        charactersRemainingText.classList.remove("charactersRemaining-few")
    }
}

/**
 * 
 * @param {Element} textArea 
 */
function charactersRemainingInit(textArea) {
    if (!textArea) {
        return;
    }

    let maxLength = textArea.getAttribute("maxlength");
    if (!maxLength) {
        return;
    }

    var charactersRemainingText = document.createElement("div", {
        "class": "charactersRemainingText"
    });
    textArea.parentElement.appendChild(charactersRemainingText);

    charactersRemainingUpdate(textArea, charactersRemainingText);

    textArea.addEventListener("keyup", charactersRemainingUpdate.bind(null, textArea, charactersRemainingText));
}

document.addEventListener("htmx:afterSwap", function(e) {
    let content = e.target;

    content.querySelectorAll(".js-charactersRemaining").forEach(function(textArea) {
        charactersRemainingInit(textArea);
    });
});