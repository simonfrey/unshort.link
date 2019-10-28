// Set unshort server
var unshortPattern = "unshort.link";

// Load available services from server
var servicesUrls = [];
var req = new XMLHttpRequest();
req.open("GET", "https://"+unshortPattern+"/providers", false);
req.addEventListener("load", function () {
    let services = JSON.parse(req.response);
    services.forEach(function (item, index) {
        servicesUrls.push("*://" + item + "/*")
    });
});
req.send(null);


// Redirect services via unshort.link
function redirect(requestDetails) {
    return {
        redirectUrl: "https://"+unshortPattern+"/d/" + requestDetails.url
    };
}
browser.webRequest.onBeforeRequest.addListener(
    redirect,
    { urls: servicesUrls },
    ["blocking"]
);