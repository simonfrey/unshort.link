var unshortBaseURL = "https://unshort.link";
var directRedirect = false;
var doNotCheckBlacklist = false;
var active = true;

function loadOptions() {
    function setData(result) {
        unshortBaseURL = result.serverUrl || "https://unshort.link";
        directRedirect = result.directRedirect || false;
        doNotCheckBlacklist = result.doNotCheckBlacklist || false;
    }
    chrome.storage.sync.get(
        ["serverUrl", "directRedirect", "doNotCheckBlacklist"],
        setData
    );
}

loadOptions();

// Redirect services via unshort.link
function redirect(requestDetails) {
    if (!active) {
        console.log("Skip unshort plugin button was set to inactive");
        return;
    }

    if (requestDetails.originUrl != undefined) {
        var l = document.createElement("a");
        l.href = requestDetails.originUrl;
        if (
            requestDetails.originUrl.startsWith(unshortBaseURL) ||
            requestDetails.url.includes(l.hostname)
        ) {
            console.log("Skip unshort because origin is " + unshortBaseURL);
            return;
        }
    }
    console.log("Unshort: ", requestDetails.url);
    var p = "/d/";
    if (directRedirect) {
        console.log("direct redirect");
        p = "/";
    } else if (doNotCheckBlacklist) {
        console.log("do not check blacklist");
        p = "/nb/";
    }
    return {
        redirectUrl: unshortBaseURL + p + requestDetails.url
    };
}

// Load available services from server
var req = new XMLHttpRequest();
req.open("GET", unshortBaseURL + "/providers", true);
req.addEventListener("load", function() {
    let servicesUrls = [];

    let services = JSON.parse(req.response);
    services.forEach(function(item, index) {
        if (item.length == 0) {
            return;
        }
        servicesUrls.push("*://" + item + "/*");
    });

    chrome.webRequest.onBeforeRequest.addListener(
        redirect, { urls: servicesUrls }, ["blocking"]
    );
});
req.send(null);

chrome.browserAction.onClicked.addListener(function() {
    if (active) {
        active = false;
        chrome.browserAction.setIcon({ path: "icons/128_deactivated.png" });
    } else {
        active = true;
        chrome.browserAction.setIcon({ path: "icons/128.png" });
    }
});


function handleInstalled(details) {
    if (details.reason == "install") {
        chrome.tabs.create({
            url: unshortBaseURL + "/about?extension=true"
        });
    }
}
chrome.runtime.onInstalled.addListener(handleInstalled);