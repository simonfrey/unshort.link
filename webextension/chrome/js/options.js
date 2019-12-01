var serverUrl = document.getElementById("serverUrl");
var directRedirect = document.getElementById("directRedirect");


function saveOptions(e) {
    e.preventDefault();
    chrome.storage.sync.set({
        serverUrl: serverUrl.value,
        directRedirect: directRedirect.checked,
    });
    chrome.runtime.reload();
}

function restoreOptions() {
    function setData(result) {
        serverUrl.value = result.serverUrl || "https://unshort.link";
        directRedirect.checked = result.directRedirect || false;
    }

    chrome.storage.sync.get(["serverUrl","directRedirect"],setData);
}

document.addEventListener("DOMContentLoaded", restoreOptions);
document.querySelector("form").addEventListener("submit", saveOptions);