// Set unshort server
var unshortPattern = "unshort.link";


// Redirect services via unshort.link
function redirect(requestDetails) {
    console.log("Unshort: ",requestDetails.url)
    return {
        redirectUrl: "https://" + unshortPattern + "/d/" + requestDetails.url
    };
}

// Load available services from server
var req = new XMLHttpRequest();
req.open("GET", "https://" + unshortPattern + "/providers", true);
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

