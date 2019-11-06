// Set unshort server
var unshortPattern = "https://unshort.link";


// Redirect services via unshort.link
function redirect(requestDetails) {
    if (requestDetails.originUrl != undefined && requestDetails.originUrl.startsWith(unshortPattern)){
        console.log("Skip unlock because origin is "+unshortPattern);
        return
    }
    console.log("Unshort: ",requestDetails.url)
    return {
        redirectUrl: unshortPattern + "/d/" + requestDetails.url
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

    console.log(servicesUrls)

    browser.webRequest.onBeforeRequest.addListener(
        redirect,
        { urls: servicesUrls },
        ["blocking"]
    );
});
req.send(null);

