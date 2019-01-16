// RECAPTCHA PLUGIN (v0.0.1)

var RECAPTCHASITEKEY = '6Lds9z0UAAAAAFfF0zUxizO5RB4W3GIExWCUcKW2';

// args is a list, it only currently contains the URI of the requested page 
function onGetRender(args) {
    if (args[0] === "/redirect-test") {
        return {
            route: "/",
            code: 302
        }
    }

    if (args[0] === "/recaptcha-test") {
        document.Find("head").AppendHtml("<script src=\"https://www.google.com/recaptcha/api.js\" async defer></script>")
        document.Find("body").AppendHtml("<form action= \"" + args[0] + "\" method=\"post\"><input name=\"sometext\" type=\"text\"><button type=\"submit\">Send</button></form>")
        document.Find("form").AppendHtml("<div class=\"g-recaptcha\" data-sitekey=\"" + RECAPTCHASITEKEY + "\"></div>")
    }

    return null;
}

function onPostRecieve(args) {
    if (args[0] === "/recaptcha-test") {
        InfoLog("Recieved post request");
        formResponseData = args[1];
        console.log(formResponseData["g-recaptcha-response"])
        return {
            route: "/recaptcha-test"
        }
    }
}

function main() {
    InfoLog("Loaded plugin");

    AddToRobots("Disallow: /cheesecake-test")
    DelFromRobots("Disallow: /cheesecake-test")
}
