// Set unshort server
var unshortPattern = "https://unshort.link";
var directRedirect = false;

function loadOptions() {
    function setData(result) {
        unshortPattern = result.serverUrl || "https://unshort.link";
        directRedirect = result.directRedirect || false;
    }
    chrome.storage.sync.get(["serverUrl","directRedirect"],setData);
}

loadOptions();

// Redirect services via unshort.link
function redirect(requestDetails) {
    if (requestDetails.originUrl != undefined && requestDetails.originUrl.startsWith(unshortPattern)){
        console.log("Skip unlock because origin is "+unshortPattern);
        return
    }
    console.log("Unshort: ",requestDetails.url)
    var p = "/d/"
    if (directRedirect) {
        console.log("direct redirect")
        p = "/"
    }
    return {
        redirectUrl: unshortPattern + p + requestDetails.url
    };
}

// Load available services from server
var req = new XMLHttpRequest();
req.open("GET", unshortPattern + "/providers", true);
req.addEventListener("load", function () {
    let servicesUrls = [];

    let services = JSON.parse(req.response);
    services.forEach(function (item, index) {
        if (item.length == 0){
            return
        }
        servicesUrls.push("*://" + item + "/*")
    });


    chrome.webRequest.onBeforeRequest.addListener(
        redirect,
        { urls: servicesUrls },
        ["blocking"]
    );
});
req.send(null);

