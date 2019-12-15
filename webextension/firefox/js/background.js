// Set unshort server
var unshortPattern = "https://unshort.link";
var directRedirect = false;

function getHostname(href) {
  var l = document.createElement("a");
  l.href = href;
  return l;
}

function loadOptions() {
  function setServerUrl(result) {
    unshortPattern = result.serverUrl || "https://unshort.link";
  }
  function setRedirect(result) {
    directRedirect = result.directRedirect || false;
  }

  function onError(error) {
    alert(`Unshort.link Error: ${error}`);
  }

  browser.storage.sync.get("serverUrl").then(setServerUrl, onError);
  browser.storage.sync.get("directRedirect").then(setRedirect, onError);
}

loadOptions();

// Redirect services via unshort.link
function redirect(requestDetails) {
  if (requestDetails.originUrl != undefined) {
    var l = document.createElement("a");
    l.href = requestDetails.originUrl;
    if (
      requestDetails.originUrl.startsWith(unshortPattern) ||
      requestDetails.url.includes(l.hostname)
    ) {
      console.log("Skip unshort because origin is " + unshortPattern);
      return;
    }
  }
  console.log("Unshort: ", requestDetails.url);

  var p = "/d/";
  if (directRedirect) {
    console.log("direct redirect");
    p = "/";
  }

  return {
    redirectUrl: unshortPattern + p + requestDetails.url
  };
}

// Load available services from server
var req = new XMLHttpRequest();
req.open("GET", unshortPattern + "/providers", true);
req.addEventListener("load", function() {
  let servicesUrls = [];

  let services = JSON.parse(req.response);
  services.forEach(function(item, index) {
    if (item.length == 0) {
      return;
    }
    servicesUrls.push("*://" + item + "/*");
  });

  console.log(servicesUrls);

  browser.webRequest.onBeforeRequest.addListener(
    redirect,
    { urls: servicesUrls },
    ["blocking"]
  );
});
req.send(null);
