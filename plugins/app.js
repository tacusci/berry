
function onGetRender(uri) {
    if (uri === "/main.go") {
        var pageAndResult = session.Get("login_page");
        if (pageAndResult[1]) {
            document.SetHtml(pageAndResult[0]);
        }
    }
}

function onPostRecieve(uri, data) {}

//this list of routes gets mapped on plugin load, before main() is called
var routesToRegister = ["/main.go", "/images/{imgfile}"];

function main() {
    var loginPage = files.Read("./main.go");
    if (loginPage !== undefined) {
        if (typeof loginPage === 'string') {
            session.Set("login_page", loginPage);
        }
    }
}
