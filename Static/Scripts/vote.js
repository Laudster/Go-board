document.addEventListener("DOMContentLoaded", () => {
    let ups = document.getElementsByClassName("up")
    let downs = document.getElementsByClassName("down")

    let charge = document.getElementById("charge")

    for (let i = 0; i < ups.length; i++) {

        let status = ups[i].parentElement.querySelector("#status")

        if (status) {
            console.log(status.value)
            if (status.value == "po") {
                ups[i].checked = true
            } else {
                downs[i].checked = true
            }
        }

        ups[i].addEventListener("change", function(){
            if (this.checked === true) {
                downs[i].checked = false
                downs[i].parentElement.querySelector("#charge").value = "↑"
            } else {
                downs[i].parentElement.querySelector("#charge").value = ""
            }

            this.parentElement.submit()
        })
    }

    for (let i = 0; i < downs.length; i++) {
        downs[i].addEventListener("change", function() {
            if (this.checked === true) {
                ups[i].checked = false
                ups[i].parentElement.querySelector("#charge").value = "↓"
            } else {
                ups[i].parentElement.querySelector("#charge").value = ""
            }

            this.parentElement.submit()
        })
    }
})