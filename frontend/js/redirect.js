let buttonDirectElement = document.getElementById("unshortDirectButton");
let urlDirectElement = document.getElementById("urlDirect");
buttonDirectElement.addEventListener("click", function () {
    window.location.href = "http://localhost:8080/" + urlDirectElement.value;
});


let buttonShowElement = document.getElementById("unshortShowButton");
let urlShowElement = document.getElementById("urlShow");
buttonShowElement.addEventListener("click", function () {
    window.location.href = "http://localhost:8080/" + urlShowElement.value;
});