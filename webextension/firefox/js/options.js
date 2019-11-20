var serverUrl = document.getElementById("serverUrl");
var directRedirect = document.getElementById("directRedirect");


function saveOptions(e) {
    e.preventDefault();
    browser.storage.sync.set({
        serverUrl: serverUrl.value,
        directRedirect: directRedirect.checked,
    });

    browser.runtime.reload();
}

function restoreOptions() {

    function setServerUrl(result) {
        serverUrl.value = result.serverUrl || "https://unshort.link";
    }
    function setRedirect(result) {
        directRedirect.checked = result.directRedirect || false;
    }

    function onError(error) {
        alert(`Error: ${error}`);
    }

    browser.storage.sync.get("serverUrl").then(setServerUrl, onError);
    browser.storage.sync.get("directRedirect").then(setRedirect, onError);
}

document.addEventListener("DOMContentLoaded", restoreOptions);
document.querySelector("form").addEventListener("submit", saveOptions);