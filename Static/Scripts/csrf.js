function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);

    if (parts.length === 2)
        return parts.pop().split(";").shift();
}

document.addEventListener("DOMContentLoaded", function () {
    const csrfToken = getCookie("csrf_token");

    if (csrfToken) {
        let csrfElements = document.getElementsByClassName("csrf_token")

        for (let i = 0; i < csrfElements.length; i++) {
            csrfElements[i].value = csrfToken
        }
    }
});