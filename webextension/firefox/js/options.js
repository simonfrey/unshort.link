var serverUrl = document.getElementById("serverUrl");
var directRedirect = document.getElementById("directRedirect");
var doNotCheckBlacklist = document.getElementById("doNotCheckBlacklist");


function saveOptions(e) {
    e.preventDefault();
    browser.storage.sync.set({
        serverUrl: serverUrl.value,
        directRedirect: directRedirect.checked,
        doNotCheckBlacklist: doNotCheckBlacklist.checked,
    });

    browser.runtime.reload();
}


function restoreOptions() {
    function setData(result) {
        serverUrl.value = result.serverUrl || "https://unshort.link";
        directRedirect.checked = result.directRedirect || false;
        doNotCheckBlacklist.checked = result.doNotCheckBlacklist || false;
    }

    function onError(error) {
        alert(`Error: ${error}`);
    }

    browser.storage.sync.get(["serverUrl", "directRedirect", "doNotCheckBlacklist"]).then(setData, onError);
}

document.addEventListener("DOMContentLoaded", restoreOptions);
document.querySelector("form").addEventListener("submit", saveOptions);