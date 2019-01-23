// RECAPTCHA PLUGIN (v0.0.1)

var RECAPTCHASITEKEY = '6Lds9z0UAAAAAFfF0zUxizO5RB4W3GIExWCUcKW2';

// args is a list, it only currently contains the URI of the requested page 
function onGetRender(uri) {

    logging.Info(session.Get("cheese")[0]);

    if (uri === "/redirect-test") {
        return {
            route: "/",
            code: 302
        }
    }

    if (uri === "/recaptcha-test") {
        document.Find("head").AppendHtml("<script src=\"https://www.google.com/recaptcha/api.js\" async defer></script>")
        document.Find("body").AppendHtml("<form action= \"" + uri + "\" method=\"post\"><input name=\"sometext\" type=\"text\"><button type=\"submit\">Send</button></form>")
        document.Find("form").AppendHtml("<div class=\"g-recaptcha\" data-sitekey=\"" + RECAPTCHASITEKEY + "\"></div>")
    }

    return null;
}

function onPostRecieve(uri, data) {
    if (uri === "/recaptcha-test") {
        logging.Info("Recieved post request");
        console.log(data["g-recaptcha-response"])
        return {
            route: uri 
        }
    }
}

function main() {
    logging.Info("Loaded plugin")

    for (var i = 0; i < 20; i++) {
        robots.Add("Disallow: /cheesecake-test")
    }

    for (var j = 0; j < 20; j++) {
        robots.Del("Disallow: /cheesecake-test")
    }

    session.Set("cheese", "cake");
}
